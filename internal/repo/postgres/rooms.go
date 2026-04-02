package postgres

import (
	"context"

	"room-booking-service-go/internal/repo"
)

type roomsRepo struct {
	queries queryRunner
}

func (r *roomsRepo) Create(ctx context.Context, params repo.CreateRoomParams) (repo.Room, error) {
	row := r.queries.queryRow(ctx, `
		INSERT INTO rooms (name, description, capacity)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, capacity, created_at
	`, params.Name, params.Description, params.Capacity)

	return scanRoom(row)
}

func (r *roomsRepo) GetByID(ctx context.Context, id string) (repo.Room, error) {
	row := r.queries.queryRow(ctx, `
		SELECT id, name, description, capacity, created_at
		FROM rooms
		WHERE id = $1
	`, id)

	return scanRoom(row)
}

func (r *roomsRepo) List(ctx context.Context) ([]repo.Room, error) {
	rows, err := r.queries.query(ctx, `
		SELECT id, name, description, capacity, created_at
		FROM rooms
		ORDER BY created_at ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]repo.Room, 0)
	for rows.Next() {
		item, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
