package repo

import (
	"context"
	"database/sql"
	"encoding/json"

	"gocamping/internal/model"
	"gocamping/internal/timeutil"
)

type RiskRepo struct{ DB *sql.DB }

func (r *RiskRepo) Save(ctx context.Context, rep *model.RiskReport) error {
	detail, _ := json.Marshal(rep)
	return r.DB.QueryRowContext(ctx,
		`INSERT INTO risk_reports (trip_id,level,score,detail,dispersion,computed_at) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		rep.TripID, rep.Level, rep.Score, detail, rep.Dispersion, timeutil.NowNaive(),
	).Scan(&rep.ID)
}

func (r *RiskRepo) Latest(ctx context.Context, tripID int64) (*model.RiskReport, error) {
	var raw []byte
	var id int64
	err := r.DB.QueryRowContext(ctx,
		`SELECT id, detail FROM risk_reports WHERE trip_id=$1 ORDER BY id DESC LIMIT 1`, tripID,
	).Scan(&id, &raw)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var rep model.RiskReport
	if err := json.Unmarshal(raw, &rep); err != nil {
		return nil, err
	}
	rep.ID = id
	if rep.Hits == nil {
		rep.Hits = []model.RiskHit{}
	}
	return &rep, nil
}

func (r *RiskRepo) LogNotify(ctx context.Context, channel string, payload []byte) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO notify_log (channel,payload,created_at) VALUES ($1,$2,$3)`,
		channel, payload, timeutil.NowNaive())
	return err
}
