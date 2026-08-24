package repo

import (
	"context"
	"database/sql"

	"gocamping/internal/model"
	"gocamping/internal/timeutil"
)

type TrackRepo struct{ DB *sql.DB }

func (r *TrackRepo) InsertPoints(ctx context.Context, pts []model.TrackPoint) (int, error) {
	if len(pts) == 0 {
		return 0, nil
	}
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	n := 0
	for _, p := range pts {
		res, err := tx.ExecContext(ctx,
			`INSERT INTO track_points (trip_id,member_id,lat,lon,elevation,accuracy,speed,recorded_at,source,is_noise,fingerprint)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT (fingerprint) DO NOTHING`,
			p.TripID, p.MemberID, p.Lat, p.Lon, p.Elevation, p.Accuracy, p.Speed, p.RecordedAt, p.Source, p.IsNoise, p.Fingerprint)
		if err != nil {
			return 0, err
		}
		af, _ := res.RowsAffected()
		n += int(af)
	}
	return n, tx.Commit()
}

func (r *TrackRepo) List(ctx context.Context, tripID, memberID int64) ([]model.TrackPoint, error) {
	q := `SELECT id,trip_id,member_id,lat,lon,elevation,accuracy,speed,recorded_at,source,is_noise,fingerprint
	      FROM track_points WHERE trip_id=$1 AND is_noise=false`
	args := []any{tripID}
	if memberID > 0 {
		q += ` AND member_id=$2`
		args = append(args, memberID)
	}
	q += ` ORDER BY recorded_at`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.TrackPoint
	for rows.Next() {
		var p model.TrackPoint
		if err := rows.Scan(&p.ID, &p.TripID, &p.MemberID, &p.Lat, &p.Lon, &p.Elevation, &p.Accuracy, &p.Speed, &p.RecordedAt, &p.Source, &p.IsNoise, &p.Fingerprint); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if out == nil {
		out = []model.TrackPoint{}
	}
	return out, rows.Err()
}

func (r *TrackRepo) Fingerprints(ctx context.Context, tripID, memberID int64) (map[string]struct{}, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT fingerprint FROM track_points WHERE trip_id=$1 AND member_id=$2`, tripID, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[string]struct{}{}
	for rows.Next() {
		var fp string
		if err := rows.Scan(&fp); err != nil {
			return nil, err
		}
		m[fp] = struct{}{}
	}
	return m, rows.Err()
}

func (r *TrackRepo) Count(ctx context.Context, tripID int64) (int, error) {
	var n int
	err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM track_points WHERE trip_id=$1`, tripID).Scan(&n)
	return n, err
}

func (r *TrackRepo) GetBatch(ctx context.Context, tripID, memberID int64, hash string) (*model.TrackBatch, error) {
	var b model.TrackBatch
	err := r.DB.QueryRowContext(ctx,
		`SELECT id,trip_id,member_id,batch_hash,point_count,accepted,rejected,processed_at FROM track_batches
		 WHERE trip_id=$1 AND member_id=$2 AND batch_hash=$3`, tripID, memberID, hash,
	).Scan(&b.ID, &b.TripID, &b.MemberID, &b.BatchHash, &b.PointCount, &b.Accepted, &b.Rejected, &b.ProcessedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &b, err
}

func (r *TrackRepo) SaveBatch(ctx context.Context, b *model.TrackBatch) error {
	return r.DB.QueryRowContext(ctx,
		`INSERT INTO track_batches (trip_id,member_id,batch_hash,point_count,accepted,rejected,processed_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		b.TripID, b.MemberID, b.BatchHash, b.PointCount, b.Accepted, b.Rejected, timeutil.NowNaive(),
	).Scan(&b.ID)
}

func (r *TrackRepo) ReplaceSegments(ctx context.Context, tripID, memberID int64, segs []model.TrackSegment) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM track_segments WHERE trip_id=$1 AND member_id=$2`, tripID, memberID); err != nil {
		return err
	}
	for _, s := range segs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO track_segments (trip_id,member_id,start_at,end_at,distance_m,is_gap) VALUES ($1,$2,$3,$4,$5,$6)`,
			tripID, memberID, s.StartAt, s.EndAt, s.DistanceM, s.IsGap); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *TrackRepo) Segments(ctx context.Context, tripID, memberID int64) ([]model.TrackSegment, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id,trip_id,member_id,start_at,end_at,distance_m,is_gap FROM track_segments WHERE trip_id=$1 AND member_id=$2 ORDER BY start_at`,
		tripID, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.TrackSegment
	for rows.Next() {
		var s model.TrackSegment
		if err := rows.Scan(&s.ID, &s.TripID, &s.MemberID, &s.StartAt, &s.EndAt, &s.DistanceM, &s.IsGap); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if out == nil {
		out = []model.TrackSegment{}
	}
	return out, rows.Err()
}
