package postgres

import (
	"errors"

	"github.com/jackc/pgconn"
)

func isAlreadyAppliedError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	switch pgErr.Code {
	case "42P07", "42710":
		return true
	default:
		return false
	}
}
