package postgres

import (
	"context"

	"room-booking-service-go/internal/repo"
)

type slotsRepo struct {
	queries queryRunner
}

func (r *slotsRepo) Create(ctx context.Context, params repo.CreateSlotParams) (repo.Slot, error) {
	row := r.queries.queryRow(ctx, `
		INSERT INTO slots (room_id, schedule_id, start_at, end_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, room_id, schedule_id, start_at, end_at, created_at
	`, params.RoomID, params.ScheduleID, params.StartAt.UTC(), params.EndAt.UTC())

	return scanSlot(row)
}

func (r *slotsRepo) GetByID(ctx context.Context, id string) (repo.Slot, error) {
	row := r.queries.queryRow(ctx, `
		SELECT id, room_id, schedule_id, start_at, end_at, created_at
		FROM slots
		WHERE id = $1
	`, id)

	return scanSlot(row)
}

func (r *slotsRepo) ListByRoomAndRange(ctx context.Context, params repo.ListSlotsParams) ([]repo.Slot, error) {
	rows, err := r.queries.query(ctx, `
		SELECT id, room_id, schedule_id, start_at, end_at, created_at
		FROM slots
		WHERE room_id = $1
		  AND start_at >= $2
		  AND start_at < $3
		ORDER BY start_at ASC, id ASC
	`, params.RoomID, params.From.UTC(), params.To.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]repo.Slot, 0)
	for rows.Next() {
		item, err := scanSlot(rows)
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
