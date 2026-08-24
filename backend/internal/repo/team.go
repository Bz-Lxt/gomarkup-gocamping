package repo

import (
	"context"
	"database/sql"

	"gocamping/internal/model"
	"gocamping/internal/timeutil"
)

type TeamRepo struct{ DB *sql.DB }

func (r *TeamRepo) Create(ctx context.Context, t *model.Team) error {
	return r.DB.QueryRowContext(ctx,
		`INSERT INTO teams (leader_id,route_id,name,invite_code,status,created_at) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		t.LeaderID, t.RouteID, t.Name, t.InviteCode, t.Status, timeutil.NowNaive(),
	).Scan(&t.ID)
}

func (r *TeamRepo) Get(ctx context.Context, id int64) (*model.Team, error) {
	var t model.Team
	err := r.DB.QueryRowContext(ctx,
		`SELECT id,leader_id,route_id,name,invite_code,status,created_at FROM teams WHERE id=$1`, id,
	).Scan(&t.ID, &t.LeaderID, &t.RouteID, &t.Name, &t.InviteCode, &t.Status, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ms, err := r.Members(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Members = ms
	return &t, nil
}

func (r *TeamRepo) ByInvite(ctx context.Context, code string) (*model.Team, error) {
	var id int64
	err := r.DB.QueryRowContext(ctx, `SELECT id FROM teams WHERE invite_code=$1`, code).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func (r *TeamRepo) ListByUser(ctx context.Context, userID int64) ([]model.Team, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT t.id FROM teams t JOIN team_members m ON m.team_id=t.id WHERE m.user_id=$1 ORDER BY t.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	out := make([]model.Team, 0, len(ids))
	for _, id := range ids {
		t, err := r.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		if t != nil {
			out = append(out, *t)
		}
	}
	return out, nil
}

func (r *TeamRepo) AddMember(ctx context.Context, teamID, userID int64, role string) error {
	_, err := r.DB.ExecContext(ctx,
		`INSERT INTO team_members (team_id,user_id,role,joined_at,state) VALUES ($1,$2,$3,$4,'offline')
		 ON CONFLICT (team_id,user_id) DO NOTHING`,
		teamID, userID, role, timeutil.NowNaive())
	return err
}

func (r *TeamRepo) RemoveMember(ctx context.Context, teamID, userID int64) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, teamID, userID)
	return err
}

func (r *TeamRepo) SetState(ctx context.Context, teamID, userID int64, state string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE team_members SET state=$1 WHERE team_id=$2 AND user_id=$3`, state, teamID, userID)
	return err
}

func (r *TeamRepo) Members(ctx context.Context, teamID int64) ([]model.TeamMember, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT m.id,m.team_id,m.user_id,u.nickname,m.role,m.joined_at,m.state
		 FROM team_members m JOIN users u ON u.id=m.user_id WHERE m.team_id=$1 ORDER BY m.id`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.TeamMember
	for rows.Next() {
		var m model.TeamMember
		if err := rows.Scan(&m.ID, &m.TeamID, &m.UserID, &m.Nickname, &m.Role, &m.JoinedAt, &m.State); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if out == nil {
		out = []model.TeamMember{}
	}
	return out, rows.Err()
}

func (r *TeamRepo) IsMember(ctx context.Context, teamID, userID int64) (bool, string, error) {
	var role string
	err := r.DB.QueryRowContext(ctx, `SELECT role FROM team_members WHERE team_id=$1 AND user_id=$2`, teamID, userID).Scan(&role)
	if err == sql.ErrNoRows {
		return false, "", nil
	}
	return err == nil, role, err
}
