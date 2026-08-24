package service_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"gocamping/internal/geo"
	"gocamping/internal/repo"
	"gocamping/internal/risk"
	"gocamping/internal/service"
	"gocamping/internal/ws"
)

func TestLivePositionPersistsRiskAfterCallerCancellation(t *testing.T) {
	state := newRiskContextState()
	db := sql.OpenDB(&riskContextConnector{state: state})
	t.Cleanup(func() { _ = db.Close() })

	riskRepo := &repo.RiskRepo{DB: db}
	app := &service.App{
		Teams:  &repo.TeamRepo{DB: db},
		Trips:  &repo.TripRepo{DB: db},
		Tracks: &repo.TrackRepo{DB: db},
		Risks:  riskRepo,
		Grid:   geo.NewGrid(),
		Hub:    ws.NewHub(),
		PosHz:  ws.NewThrottle(),
		RiskHz: risk.NewThrottle(time.Nanosecond),
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
		state.unblockSave()
	})

	elevation := 120.0
	if err := app.LivePosition(ctx, 7, 42, 30.26, 119.72, &elevation); err != nil {
		t.Fatalf("LivePosition returned an error: %v", err)
	}

	waitForRiskSignal(t, state.saveStarted, "risk persistence to start")

	cancel()
	state.unblockSave()
	waitForRiskSignal(t, state.saveFinished, "risk persistence to finish")

	report, err := riskRepo.Latest(context.Background(), 42)
	if err != nil {
		t.Fatalf("read latest risk report: %v", err)
	}
	if report == nil {
		t.Fatal("LivePosition returned successfully, but its risk report was lost when the caller context was canceled")
	}
	if report.TripID != 42 {
		t.Fatalf("latest risk report belongs to trip %d, want 42", report.TripID)
	}
}

func waitForRiskSignal(t *testing.T, ch <-chan struct{}, operation string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", operation)
	}
}

type riskContextState struct {
	saveStarted  chan struct{}
	releaseSave  chan struct{}
	saveFinished chan struct{}

	startOnce   sync.Once
	releaseOnce sync.Once
	finishOnce  sync.Once

	mu     sync.Mutex
	report []byte
}

func newRiskContextState() *riskContextState {
	return &riskContextState{
		saveStarted:  make(chan struct{}),
		releaseSave:  make(chan struct{}),
		saveFinished: make(chan struct{}),
	}
}

func (s *riskContextState) signalStarted() {
	s.startOnce.Do(func() { close(s.saveStarted) })
}

func (s *riskContextState) unblockSave() {
	s.releaseOnce.Do(func() { close(s.releaseSave) })
}

func (s *riskContextState) signalFinished() {
	s.finishOnce.Do(func() { close(s.saveFinished) })
}

type riskContextConnector struct {
	state *riskContextState
}

func (c *riskContextConnector) Connect(ctx context.Context) (driver.Conn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &riskContextConn{state: c.state}, nil
}

func (c *riskContextConnector) Driver() driver.Driver {
	return &riskContextDriver{state: c.state}
}

type riskContextDriver struct {
	state *riskContextState
}

func (d *riskContextDriver) Open(string) (driver.Conn, error) {
	return &riskContextConn{state: d.state}, nil
}

type riskContextConn struct {
	state *riskContextState
}

func (c *riskContextConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepared statements are not supported")
}

func (c *riskContextConn) Close() error {
	return nil
}

func (c *riskContextConn) Begin() (driver.Tx, error) {
	return &riskContextTx{}, nil
}

func (c *riskContextConn) BeginTx(
	ctx context.Context,
	_ driver.TxOptions,
) (driver.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &riskContextTx{}, nil
}

func (c *riskContextConn) ExecContext(
	ctx context.Context,
	query string,
	_ []driver.NamedValue,
) (driver.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	switch {
	case strings.Contains(query, "INSERT INTO track_points"):
		return riskContextResult(1), nil
	case strings.Contains(query, "UPDATE team_members SET state"):
		return riskContextResult(1), nil
	default:
		return nil, errors.New("unexpected ExecContext query")
	}
}

func (c *riskContextConn) QueryContext(
	ctx context.Context,
	query string,
	args []driver.NamedValue,
) (driver.Rows, error) {
	now := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)

	switch {
	case strings.Contains(query, "FROM trips WHERE id="):
		return oneRiskContextRow(
			[]string{
				"id", "team_id", "route_id", "status",
				"started_at", "finished_at", "created_at",
			},
			int64(42), int64(11), nil, "active", now, nil, now,
		), nil

	case strings.Contains(query, "SELECT role FROM team_members"):
		return oneRiskContextRow([]string{"role"}, "leader"), nil

	case strings.Contains(query, "FROM teams WHERE id="):
		return oneRiskContextRow(
			[]string{
				"id", "leader_id", "route_id", "name",
				"invite_code", "status", "created_at",
			},
			int64(11), int64(7), nil, "ridge team",
			"ABC123", "open", now,
		), nil

	case strings.Contains(query, "FROM team_members m JOIN users"):
		return oneRiskContextRow(
			[]string{
				"id", "team_id", "user_id", "nickname",
				"role", "joined_at", "state",
			},
			int64(101), int64(11), int64(7), "leader",
			"leader", now, "online",
		), nil

	case strings.Contains(query, "INSERT INTO risk_reports"):
		c.state.signalStarted()
		<-c.state.releaseSave
		defer c.state.signalFinished()

		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(args) != 6 {
			return nil, errors.New("unexpected risk insert arguments")
		}
		raw, ok := args[3].Value.([]byte)
		if !ok {
			return nil, errors.New("risk detail is not JSON bytes")
		}

		c.state.mu.Lock()
		c.state.report = append([]byte(nil), raw...)
		c.state.mu.Unlock()

		return oneRiskContextRow([]string{"id"}, int64(1)), nil

	case strings.Contains(query, "FROM risk_reports WHERE trip_id="):
		c.state.mu.Lock()
		raw := append([]byte(nil), c.state.report...)
		c.state.mu.Unlock()

		if raw == nil {
			return &riskContextRows{
				columns: []string{"id", "detail"},
			}, nil
		}
		return oneRiskContextRow(
			[]string{"id", "detail"},
			int64(1), raw,
		), nil

	default:
		return nil, errors.New("unexpected QueryContext query")
	}
}

type riskContextTx struct{}

func (*riskContextTx) Commit() error   { return nil }
func (*riskContextTx) Rollback() error { return nil }

type riskContextResult int64

func (r riskContextResult) LastInsertId() (int64, error) {
	return int64(r), nil
}

func (r riskContextResult) RowsAffected() (int64, error) {
	return int64(r), nil
}

type riskContextRows struct {
	columns []string
	values  [][]driver.Value
	next    int
}

func oneRiskContextRow(
	columns []string,
	values ...driver.Value,
) driver.Rows {
	return &riskContextRows{
		columns: columns,
		values:  [][]driver.Value{values},
	}
}

func (r *riskContextRows) Columns() []string {
	return r.columns
}

func (r *riskContextRows) Close() error {
	return nil
}

func (r *riskContextRows) Next(dest []driver.Value) error {
	if r.next >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.next])
	r.next++
	return nil
}
