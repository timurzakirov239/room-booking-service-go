package app

import (
	"context"
	"time"

	"room-booking-service-go/internal/repo"
)

type SlotService struct {
	Rooms        repo.RoomsRepository
	Schedules    repo.SchedulesRepository
	Slots        repo.SlotsRepository
	Materializer SlotMaterializer
}

func (s SlotService) ListAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]repo.Slot, error) {
	if _, err := s.Rooms.GetByID(ctx, roomID); err != nil {
		return nil, err
	}

	schedule, err := s.Schedules.GetByRoomID(ctx, roomID)
	if err != nil {
		if err == repo.ErrNotFound {
			return []repo.Slot{}, nil
		}
		return nil, err
	}

	dayStart := time.Date(date.UTC().Year(), date.UTC().Month(), date.UTC().Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	planned, err := s.Materializer.PlanMissingSlots(ctx, MaterializeSlotsParams{
		Schedule: schedule,
		From:     dayStart,
		To:       dayEnd,
	})
	if err != nil {
		return nil, err
	}

	for _, createParams := range planned {
		if _, err := s.Slots.Create(ctx, createParams); err != nil && err != repo.ErrConflict {
			return nil, err
		}
	}

	return s.Slots.ListByRoomAndRange(ctx, repo.ListSlotsParams{
		RoomID: roomID,
		From:   dayStart,
		To:     dayEnd,
	})
}
