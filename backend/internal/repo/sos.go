package repo

import (
	"context"
	"database/sql"

	"gocamping/internal/model"
	"gocamping/internal/timeutil"
)

type SOSRepo struct{ DB *sql.DB }

func (r *SOSRepo) Create(ctx context.Context, e *model.SOSEvent) error {
	return r.DB.QueryRowContext(ctx,
		`INSERT INTO sos_events (trip_id,member_id,type,lat,lon,reason,status,created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		e.TripID, e.MemberID, e.Type, e.Lat, e.Lon, e.Reason, e.Status, timeutil.NowNaive(),
	).Scan(&e.ID)
}

func (r *SOSRepo) Get(ctx context.Context, id int64) (*model.SOSEvent, error) {
	var e model.SOSEvent
	err := r.DB.QueryRowContext(ctx,
		`SELECT id,trip_id,member_id,type,lat,lon,reason,status,created_at,resolved_at FROM sos_events WHERE id=$1`, id,
	).Scan(&e.ID, &e.TripID, &e.MemberID, &e.Type, &e.Lat, &e.Lon, &e.Reason, &e.Status, &e.CreatedAt, &e.ResolvedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &e, err
}

func (r *SOSRepo) List(ctx context.Context, tripID int64) ([]model.SOSEvent, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT s.id,s.trip_id,s.member_id,u.nickname,s.type,s.lat,s.lon,s.reason,s.status,s.created_at,s.resolved_at
		 FROM sos_events s JOIN users u ON u.id=s.member_id WHERE s.trip_id=$1 ORDER BY s.id DESC`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.SOSEvent
	for rows.Next() {
		var e model.SOSEvent
		if err := rows.Scan(&e.ID, &e.TripID, &e.MemberID, &e.Nickname, &e.Type, &e.Lat, &e.Lon, &e.Reason, &e.Status, &e.CreatedAt, &e.ResolvedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if out == nil {
		out = []model.SOSEvent{}
	}
	return out, rows.Err()
}

func (r *SOSRepo) Resolve(ctx context.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE sos_events SET status='resolved', resolved_at=$1 WHERE id=$2`, timeutil.NowNaive(), id)
	return err
}

func (r *SOSRepo) HasOpenAuto(ctx context.Context, tripID, memberID int64) (bool, error) {
	var n int
	err := r.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM sos_events WHERE trip_id=$1 AND member_id=$2 AND type='auto' AND status='open'`,
		tripID, memberID).Scan(&n)
	return n > 0, err
}
