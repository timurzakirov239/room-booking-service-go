package app

import (
	"context"
	"time"

	"room-booking-service-go/internal/domain"
	"room-booking-service-go/internal/repo"
)

type SlotMaterializer struct {
	Slots repo.SlotsRepository
}

type MaterializeSlotsParams struct {
	Schedule repo.Schedule
	From     time.Time
	To       time.Time
}

func (s SlotMaterializer) PlanMissingSlots(ctx context.Context, params MaterializeSlotsParams) ([]repo.CreateSlotParams, error) {
	windows, err := domain.BuildRollingSlotWindows(
		params.Schedule.DaysOfWeek,
		params.Schedule.StartTime,
		params.Schedule.EndTime,
		params.From,
		params.To,
	)
	if err != nil {
		return nil, err
	}

	existingSlots, err := s.Slots.ListByRoomAndRange(ctx, repo.ListSlotsParams{
		RoomID: params.Schedule.RoomID,
		From:   params.From.UTC(),
		To:     params.To.UTC(),
	})
	if err != nil {
		return nil, err
	}

	existing := make(map[string]struct{}, len(existingSlots))
	for _, slot := range existingSlots {
		existing[slotIdentityKey(slot.RoomID, slot.StartAt, slot.EndAt)] = struct{}{}
	}

	planned := make([]repo.CreateSlotParams, 0)
	for _, window := range windows {
		key := slotIdentityKey(params.Schedule.RoomID, window.StartAt, window.EndAt)
		if _, ok := existing[key]; ok {
			continue
		}

		scheduleID := params.Schedule.ID
		planned = append(planned, repo.CreateSlotParams{
			RoomID:     params.Schedule.RoomID,
			ScheduleID: &scheduleID,
			StartAt:    window.StartAt.UTC(),
			EndAt:      window.EndAt.UTC(),
		})
	}

	return planned, nil
}

func slotIdentityKey(roomID string, startAt time.Time, endAt time.Time) string {
	return roomID + "|" + startAt.UTC().Format(time.RFC3339) + "|" + endAt.UTC().Format(time.RFC3339)
}
