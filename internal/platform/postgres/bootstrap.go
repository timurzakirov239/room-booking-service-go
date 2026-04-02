package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	platformauth "room-booking-service-go/internal/platform/auth"
)

func Bootstrap(ctx context.Context, pool *pgxpool.Pool) error {
	if err := applySQLMigrations(ctx, pool); err != nil {
		return fmt.Errorf("apply sql migrations: %w", err)
	}
	if err := seedDummyUsers(ctx, pool); err != nil {
		return fmt.Errorf("seed dummy users: %w", err)
	}
	return nil
}

func applySQLMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	alreadyApplied, err := schemaLooksInitialized(ctx, pool)
	if err != nil {
		return fmt.Errorf("check existing schema: %w", err)
	}
	if alreadyApplied {
		return nil
	}

	files := []string{
		"db/migrations/000001_init_placeholder.sql",
	}

	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		sql := extractUpSQL(string(content))
		if strings.TrimSpace(sql) == "" {
			continue
		}

		if _, err := pool.Exec(ctx, sql); err != nil {
			return fmt.Errorf("exec %s: %w", path, err)
		}
	}

	return nil
}

func schemaLooksInitialized(ctx context.Context, pool *pgxpool.Pool) (bool, error) {
	const sql = `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public'
			  AND table_name = 'users'
		)
	`

	var exists bool
	if err := pool.QueryRow(ctx, sql).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func extractUpSQL(content string) string {
	startMarker := "-- +goose Up"
	endMarker := "-- +goose Down"

	start := strings.Index(content, startMarker)
	if start == -1 {
		return strings.TrimSpace(content)
	}
	content = content[start+len(startMarker):]

	if end := strings.Index(content, endMarker); end != -1 {
		content = content[:end]
	}

	return strings.TrimSpace(content)
}

func seedDummyUsers(ctx context.Context, pool *pgxpool.Pool) error {
	const sql = `
		INSERT INTO users (id, email, role)
		VALUES
			($1, $2, 'admin'),
			($3, $4, 'user')
		ON CONFLICT (id) DO UPDATE
		SET email = EXCLUDED.email,
		    role = EXCLUDED.role
	`

	_, err := pool.Exec(ctx, sql,
		platformauth.DummyAdminUserID, "admin@example.local",
		platformauth.DummyUserUserID, "user@example.local",
	)
	return err
}
