package app

import (
	"context"
	"strings"
	"time"

	"room-booking-service-go/internal/domain"
	"room-booking-service-go/internal/repo"
)

type ScheduleService struct {
	Rooms     repo.RoomsRepository
	Schedules repo.SchedulesRepository
}

type CreateScheduleInput struct {
	RoomID     string
	DaysOfWeek []int16
	StartTime  string
	EndTime    string
}

func (s ScheduleService) Create(ctx context.Context, input CreateScheduleInput) (repo.Schedule, error) {
	if _, err := s.Rooms.GetByID(ctx, input.RoomID); err != nil {
		return repo.Schedule{}, err
	}

	startTime, err := parseScheduleClock(input.StartTime)
	if err != nil {
		return repo.Schedule{}, domain.ErrInvalidTimeRange
	}
	endTime, err := parseScheduleClock(input.EndTime)
	if err != nil {
		return repo.Schedule{}, domain.ErrInvalidTimeRange
	}

	if len(input.DaysOfWeek) == 0 {
		return repo.Schedule{}, domain.ErrInvalidTimeRange
	}
	for _, day := range input.DaysOfWeek {
		if day < 1 || day > 7 {
			return repo.Schedule{}, domain.ErrInvalidTimeRange
		}
	}

	return s.Schedules.Create(ctx, repo.CreateScheduleParams{
		RoomID:     input.RoomID,
		DaysOfWeek: input.DaysOfWeek,
		StartTime:  startTime,
		EndTime:    endTime,
	})
}

func parseScheduleClock(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	return time.Parse("15:04", trimmed)
}
