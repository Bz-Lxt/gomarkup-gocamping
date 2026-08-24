package repo

import (
	"context"
	"database/sql"
	"errors"

	"gocamping/internal/model"
	"gocamping/internal/timeutil"
)

type UserRepo struct{ DB *sql.DB }

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	return r.DB.QueryRowContext(ctx,
		`INSERT INTO users (username, password_hash, nickname, role, created_at)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		u.Username, u.PasswordHash, u.Nickname, u.Role, timeutil.NowNaive(),
	).Scan(&u.ID)
}

func (r *UserRepo) ByUsername(ctx context.Context, name string) (*model.User, error) {
	return r.scan(r.DB.QueryRowContext(ctx,
		`SELECT id, username, password_hash, nickname, role, created_at FROM users WHERE username=$1`, name))
}

func (r *UserRepo) ByID(ctx context.Context, id int64) (*model.User, error) {
	return r.scan(r.DB.QueryRowContext(ctx,
		`SELECT id, username, password_hash, nickname, role, created_at FROM users WHERE id=$1`, id))
}

func (r *UserRepo) List(ctx context.Context, limit, offset int) ([]model.User, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, username, password_hash, nickname, role, created_at FROM users ORDER BY id LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Nickname, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	if out == nil {
		out = []model.User{}
	}
	return out, rows.Err()
}

func (r *UserRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

func (r *UserRepo) scan(row *sql.Row) (*model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Nickname, &u.Role, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
