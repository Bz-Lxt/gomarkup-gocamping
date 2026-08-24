package service

import (
	"context"

	"gocamping/internal/logger"
	"gocamping/internal/model"
	"gocamping/internal/repo"
)

func Seed(ctx context.Context, users *repo.UserRepo, routes *repo.RouteRepo, teams *repo.TeamRepo, trips *repo.TripRepo) error {
	if u, _ := users.ByUsername(ctx, "leader"); u != nil {
		return nil
	}
	accounts := []struct{ name, nick, role, pw string }{
		{"leader", "队长林野", "member", "leader123"},
		{"member", "队员阿山", "member", "member123"},
		{"admin", "营地管理员", "admin", "admin123"},
	}
	var ids []int64
	for _, a := range accounts {
		hash, err := HashPassword(a.pw)
		if err != nil {
			return err
		}
		u := &model.User{Username: a.name, PasswordHash: hash, Nickname: a.nick, Role: a.role}
		if err := users.Create(ctx, u); err != nil {
			return err
		}
		ids = append(ids, u.ID)
	}
	r := 80.0
	book := &model.RouteBook{
		OwnerID: ids[0], Title: "清凉峰西坡小众线", Description: "避开主峰索道，沿西坡水源线穿越到废弃林场营地。",
		Visibility: "public", Status: "active", Version: 1,
		Waypoints: []model.Waypoint{
			{Type: "waypoint", Lat: 30.1240, Lon: 118.8520, Note: "西坡入口", RiskWeight: 1},
			{Type: "water", Lat: 30.1315, Lon: 118.8610, Note: "山涧取水点", RiskWeight: 1},
			{Type: "danger", Lat: 30.1360, Lon: 118.8680, Note: "碎石滑坡带", RiskWeight: 4, RadiusM: &r},
			{Type: "waypoint", Lat: 30.1420, Lon: 118.8740, Note: "脊线鞍部", RiskWeight: 1},
			{Type: "camp", Lat: 30.1488, Lon: 118.8812, Note: "废弃林场营地", RiskWeight: 1},
		},
	}
	app := &App{Routes: routes}
	if _, err := app.SaveRoute(ctx, ids[0], book); err != nil {
		return err
	}
	rid := book.ID
	t, err := (&App{Teams: teams}).CreateTeam(ctx, ids[0], "清凉峰夜宿小队", &rid)
	if err != nil {
		return err
	}
	_ = teams.AddMember(ctx, t.ID, ids[1], "member")
	tr := &model.Trip{TeamID: t.ID, RouteID: &rid, Status: model.TripDraft}
	if err := trips.Create(ctx, tr); err != nil {
		return err
	}
	logger.Info("seed ready", "leader", ids[0], "member", ids[1], "admin", ids[2], "route", rid, "team", t.ID, "trip", tr.ID)
	return nil
}
