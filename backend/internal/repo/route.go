package repo

import (
	"context"
	"database/sql"
	"encoding/json"

	"gocamping/internal/model"
	"gocamping/internal/timeutil"
)

type RouteRepo struct{ DB *sql.DB }

func (r *RouteRepo) Create(ctx context.Context, b *model.RouteBook) error {
	now := timeutil.NowNaive()
	return r.DB.QueryRowContext(ctx,
		`INSERT INTO route_books (owner_id,title,description,visibility,version,distance_m,ascent_m,geometry_hash,status,created_at,updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id`,
		b.OwnerID, b.Title, b.Description, b.Visibility, b.Version, b.DistanceM, b.AscentM, b.GeometryHash, b.Status, now, now,
	).Scan(&b.ID)
}

func (r *RouteRepo) Update(ctx context.Context, b *model.RouteBook) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE route_books SET title=$1,description=$2,visibility=$3,version=version+1,distance_m=$4,ascent_m=$5,geometry_hash=$6,status=$7,updated_at=$8 WHERE id=$9`,
		b.Title, b.Description, b.Visibility, b.DistanceM, b.AscentM, b.GeometryHash, b.Status, timeutil.NowNaive(), b.ID)
	return err
}

func (r *RouteRepo) Get(ctx context.Context, id int64) (*model.RouteBook, error) {
	var b model.RouteBook
	err := r.DB.QueryRowContext(ctx,
		`SELECT id,owner_id,title,description,visibility,version,distance_m,ascent_m,geometry_hash,status,created_at,updated_at FROM route_books WHERE id=$1`,
		id).Scan(&b.ID, &b.OwnerID, &b.Title, &b.Description, &b.Visibility, &b.Version, &b.DistanceM, &b.AscentM, &b.GeometryHash, &b.Status, &b.CreatedAt, &b.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	wps, err := r.Waypoints(ctx, id)
	if err != nil {
		return nil, err
	}
	b.Waypoints = wps
	return &b, nil
}

func (r *RouteRepo) List(ctx context.Context, ownerID int64, vis string) ([]model.RouteBook, error) {
	q := `SELECT id,owner_id,title,description,visibility,version,distance_m,ascent_m,geometry_hash,status,created_at,updated_at FROM route_books WHERE 1=1`
	args := []any{}
	n := 1
	if ownerID > 0 {
		q += ` AND owner_id=$` + itoa(n)
		args = append(args, ownerID)
		n++
	}
	if vis != "" {
		q += ` AND visibility=$` + itoa(n)
		args = append(args, vis)
		n++
	}
	q += ` ORDER BY id DESC LIMIT 100`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.RouteBook
	for rows.Next() {
		var b model.RouteBook
		if err := rows.Scan(&b.ID, &b.OwnerID, &b.Title, &b.Description, &b.Visibility, &b.Version, &b.DistanceM, &b.AscentM, &b.GeometryHash, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	if out == nil {
		out = []model.RouteBook{}
	}
	return out, rows.Err()
}

func (r *RouteRepo) ReplaceWaypoints(ctx context.Context, routeID int64, wps []model.Waypoint) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM waypoints WHERE route_id=$1`, routeID); err != nil {
		return err
	}
	for i, w := range wps {
		poly, _ := json.Marshal(w.Polygon)
		if w.Polygon == nil {
			poly = []byte("[]")
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO waypoints (route_id,seq,type,lat,lon,elevation,radius_m,polygon,risk_weight,note)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			routeID, i, w.Type, w.Lat, w.Lon, w.Elevation, w.RadiusM, poly, w.RiskWeight, w.Note); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *RouteRepo) Waypoints(ctx context.Context, routeID int64) ([]model.Waypoint, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id,route_id,seq,type,lat,lon,elevation,radius_m,polygon,risk_weight,note FROM waypoints WHERE route_id=$1 ORDER BY seq`,
		routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Waypoint
	for rows.Next() {
		var w model.Waypoint
		var poly []byte
		if err := rows.Scan(&w.ID, &w.RouteID, &w.Seq, &w.Type, &w.Lat, &w.Lon, &w.Elevation, &w.RadiusM, &poly, &w.RiskWeight, &w.Note); err != nil {
			return nil, err
		}
		if len(poly) > 0 {
			_ = json.Unmarshal(poly, &w.Polygon)
		}
		if w.Polygon == nil {
			w.Polygon = [][2]float64{}
		}
		out = append(out, w)
	}
	if out == nil {
		out = []model.Waypoint{}
	}
	return out, rows.Err()
}

func (r *RouteRepo) PutElevation(ctx context.Context, hash, provider string, profile []byte) error {
	_, err := r.DB.ExecContext(ctx,
		`INSERT INTO elevation_cache (geometry_hash, profile, provider, created_at) VALUES ($1,$2,$3,$4)
		 ON CONFLICT (geometry_hash) DO UPDATE SET profile=EXCLUDED.profile, provider=EXCLUDED.provider`,
		hash, profile, provider, timeutil.NowNaive())
	return err
}

func (r *RouteRepo) GetElevation(ctx context.Context, hash string) ([]byte, string, error) {
	var raw []byte
	var provider string
	err := r.DB.QueryRowContext(ctx, `SELECT profile, provider FROM elevation_cache WHERE geometry_hash=$1`, hash).Scan(&raw, &provider)
	if err == sql.ErrNoRows {
		return nil, "", nil
	}
	return raw, provider, err
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
