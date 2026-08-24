package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"gocamping/internal/api"
	"gocamping/internal/config"
	"gocamping/internal/db"
	"gocamping/internal/dem"
	"gocamping/internal/geo"
	"gocamping/internal/logger"
	"gocamping/internal/notify"
	"gocamping/internal/repo"
	"gocamping/internal/risk"
	"gocamping/internal/service"
	"gocamping/internal/ws"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.LogLevel, cfg.Env)
	ctx := context.Background()

	sqlDB, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db open", "err", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	raw, err := db.SchemaSQL()
	if err != nil {
		logger.Error("read migration", "err", err)
		os.Exit(1)
	}
	if err := db.Migrate(ctx, sqlDB, raw); err != nil {
		logger.Error("migrate", "err", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("redis ping", "err", err)
		os.Exit(1)
	}

	users := &repo.UserRepo{DB: sqlDB}
	routes := &repo.RouteRepo{DB: sqlDB}
	teams := &repo.TeamRepo{DB: sqlDB}
	trips := &repo.TripRepo{DB: sqlDB}
	if err := service.Seed(ctx, users, routes, teams, trips); err != nil {
		logger.Error("seed", "err", err)
		os.Exit(1)
	}

	app := &service.App{
		Users:  users,
		Routes: routes,
		Teams:  teams,
		Trips:  trips,
		Tracks: &repo.TrackRepo{DB: sqlDB},
		SOS:    &repo.SOSRepo{DB: sqlDB},
		Risks:  &repo.RiskRepo{DB: sqlDB},
		DEM:    dem.New(cfg.DEMProvider, cfg.DEMHTTPURL),
		Notify: notify.New(cfg.NotifyProvider, cfg.NotifyHTTPURL),
		Hub:    ws.NewHub(),
		Grid:   geo.NewGrid(),
		Redis:  rdb,
		PosHz:  ws.NewThrottle(),
		RiskHz: risk.NewThrottle(time.Second),
	}

	ready := func() error {
		if err := sqlDB.Ping(); err != nil {
			return err
		}
		return rdb.Ping(context.Background()).Err()
	}
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           api.NewRouter(cfg, app, ready),
		ReadHeaderTimeout: 8 * time.Second,
	}

	go func() {
		logger.Info("listen", "addr", cfg.HTTPAddr, "tz", "Asia/Shanghai")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	shctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	_ = srv.Shutdown(shctx)
}
