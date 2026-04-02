package postgres

import (
	"context"

	"room-booking-service-go/internal/repo"
)

type schedulesRepo struct {
	queries queryRunner
}

func (r *schedulesRepo) Create(ctx context.Context, params repo.CreateScheduleParams) (repo.Schedule, error) {
	row := r.queries.queryRow(ctx, `
		INSERT INTO schedules (room_id, days_of_week, start_time, end_time)
		VALUES ($1, $2, $3, $4)
		RETURNING id, room_id, days_of_week, start_time, end_time, created_at
	`, params.RoomID, params.DaysOfWeek, params.StartTime.UTC(), params.EndTime.UTC())

	return scanSchedule(row)
}

func (r *schedulesRepo) GetByRoomID(ctx context.Context, roomID string) (repo.Schedule, error) {
	row := r.queries.queryRow(ctx, `
		SELECT id, room_id, days_of_week, start_time, end_time, created_at
		FROM schedules
		WHERE room_id = $1
	`, roomID)

	return scanSchedule(row)
}
