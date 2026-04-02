package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"room-booking-service-go/internal/repo"
)

type rowScanner interface {
	Scan(dest ...any) error
}

type rowsScanner interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close()
}

type queryRunner struct {
	pool *pgxpool.Pool
}

func (q queryRunner) queryRow(ctx context.Context, sql string, args ...any) rowScanner {
	return q.pool.QueryRow(ctx, sql, args...)
}

func (q queryRunner) query(ctx context.Context, sql string, args ...any) (rowsScanner, error) {
	return q.pool.Query(ctx, sql, args...)
}

func normalizeError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return repo.ErrConflict
		}
	}

	return err
}
