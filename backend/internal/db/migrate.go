package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"gocamping/internal/logger"
)

//go:embed sql/001_init.sql
var migFS embed.FS

func SchemaSQL() (string, error) {
	b, err := migFS.ReadFile("sql/001_init.sql")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Migrate runs SQL migrations under a session-scoped advisory lock.
// Must use sql.DB.Conn so lock/unlock stay on the same session.
func Migrate(ctx context.Context, db *sql.DB, sqlText string) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("migrate conn: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock(88421001)"); err != nil {
		return fmt.Errorf("advisory lock: %w", err)
	}
	defer func() {
		if _, uerr := conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock(88421001)"); uerr != nil {
			logger.Warn("advisory unlock", "err", uerr)
		}
	}()

	if _, err := conn.ExecContext(ctx, sqlText); err != nil {
		return fmt.Errorf("migrate exec: %w", err)
	}
	logger.Info("migrations applied")
	return nil
}
