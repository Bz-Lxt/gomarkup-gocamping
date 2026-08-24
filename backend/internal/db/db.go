package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"gocamping/internal/logger"
)

func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(8)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := pingWait(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func pingWait(ctx context.Context, db *sql.DB) error {
	var last error
	for i := 0; i < 30; i++ {
		if err := db.PingContext(ctx); err == nil {
			return nil
		} else {
			last = err
			logger.Warn("postgres ping retry", "err", err, "n", i)
			time.Sleep(time.Second)
		}
	}
	return fmt.Errorf("postgres not ready: %w", last)
}
