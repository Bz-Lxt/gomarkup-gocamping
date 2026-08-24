package repo

import (
	"context"
	"database/sql"

	"gocamping/internal/model"
	"gocamping/internal/timeutil"
)

type TripRepo struct{ DB *sql.DB }

func (r *TripRepo) Create(ctx context.Context, t *model.Trip) error {
	return r.DB.QueryRowContext(ctx,
		`INSERT INTO trips (team_id,route_id,status,created_at) VALUES ($1,$2,$3,$4) RETURNING id`,
		t.TeamID, t.RouteID, t.Status, timeutil.NowNaive(),
	).Scan(&t.ID)
}

func (r *TripRepo) Get(ctx context.Context, id int64) (*model.Trip, error) {
	var t model.Trip
	err := r.DB.QueryRowContext(ctx,
		`SELECT id,team_id,route_id,status,started_at,finished_at,created_at FROM trips WHERE id=$1`, id,
	).Scan(&t.ID, &t.TeamID, &t.RouteID, &t.Status, &t.StartedAt, &t.FinishedAt, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *TripRepo) ListByTeam(ctx context.Context, teamID int64) ([]model.Trip, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id,team_id,route_id,status,started_at,finished_at,created_at FROM trips WHERE team_id=$1 ORDER BY id DESC`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Trip
	for rows.Next() {
		var t model.Trip
		if err := rows.Scan(&t.ID, &t.TeamID, &t.RouteID, &t.Status, &t.StartedAt, &t.FinishedAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if out == nil {
		out = []model.Trip{}
	}
	return out, rows.Err()
}

func (r *TripRepo) SetStatus(ctx context.Context, id int64, status string) error {
	now := timeutil.NowNaive()
	switch status {
	case model.TripActive:
		_, err := r.DB.ExecContext(ctx, `UPDATE trips SET status=$1, started_at=COALESCE(started_at,$2) WHERE id=$3`, status, now, id)
		return err
	case model.TripFinished:
		_, err := r.DB.ExecContext(ctx, `UPDATE trips SET status=$1, finished_at=$2 WHERE id=$3`, status, now, id)
		return err
	default:
		_, err := r.DB.ExecContext(ctx, `UPDATE trips SET status=$1 WHERE id=$2`, status, id)
		return err
	}
}
