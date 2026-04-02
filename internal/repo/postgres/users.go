package postgres

import (
	"context"

	"room-booking-service-go/internal/repo"
)

type usersRepo struct {
	queries queryRunner
}

func (r *usersRepo) Create(ctx context.Context, params repo.CreateUserParams) (repo.User, error) {
	row := r.queries.queryRow(ctx, `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, email, password_hash, role, created_at
	`, params.Email, params.PasswordHash, params.Role)

	return scanUser(row)
}

func (r *usersRepo) GetByID(ctx context.Context, id string) (repo.User, error) {
	row := r.queries.queryRow(ctx, `
		SELECT id, email, password_hash, role, created_at
		FROM users
		WHERE id = $1
	`, id)

	return scanUser(row)
}

func (r *usersRepo) GetByEmail(ctx context.Context, email string) (repo.User, error) {
	row := r.queries.queryRow(ctx, `
		SELECT id, email, password_hash, role, created_at
		FROM users
		WHERE email = $1
	`, email)

	return scanUser(row)
}
