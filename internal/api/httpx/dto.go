package httpx

import (
	"time"

	"room-booking-service-go/internal/repo"
)

type roomDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Capacity    *int32  `json:"capacity,omitempty"`
	CreatedAt   *string `json:"createdAt,omitempty"`
}

type scheduleDTO struct {
	ID         string  `json:"id,omitempty"`
	RoomID     string  `json:"roomId"`
	DaysOfWeek []int16 `json:"daysOfWeek"`
	StartTime  string  `json:"startTime"`
	EndTime    string  `json:"endTime"`
}

type slotDTO struct {
	ID     string `json:"id"`
	RoomID string `json:"roomId"`
	Start  string `json:"start"`
	End    string `json:"end"`
}

type bookingDTO struct {
	ID             string  `json:"id"`
	SlotID         string  `json:"slotId"`
	UserID         string  `json:"userId"`
	Status         string  `json:"status"`
	ConferenceLink *string `json:"conferenceLink,omitempty"`
	CreatedAt      *string `json:"createdAt,omitempty"`
}

type paginationDTO struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

func mapRoom(item repo.Room) roomDTO {
	return roomDTO{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Capacity:    item.Capacity,
		CreatedAt:   optionalRFC3339(item.CreatedAt),
	}
}

func mapRooms(items []repo.Room) []roomDTO {
	result := make([]roomDTO, 0, len(items))
	for _, item := range items {
		result = append(result, mapRoom(item))
	}
	return result
}

func mapSchedule(item repo.Schedule) scheduleDTO {
	return scheduleDTO{
		ID:         item.ID,
		RoomID:     item.RoomID,
		DaysOfWeek: item.DaysOfWeek,
		StartTime:  item.StartTime.UTC().Format("15:04"),
		EndTime:    item.EndTime.UTC().Format("15:04"),
	}
}

func mapSlot(item repo.Slot) slotDTO {
	return slotDTO{
		ID:     item.ID,
		RoomID: item.RoomID,
		Start:  item.StartAt.UTC().Format(time.RFC3339),
		End:    item.EndAt.UTC().Format(time.RFC3339),
	}
}

func mapSlots(items []repo.Slot) []slotDTO {
	result := make([]slotDTO, 0, len(items))
	for _, item := range items {
		result = append(result, mapSlot(item))
	}
	return result
}

func mapBooking(item repo.Booking) bookingDTO {
	return bookingDTO{
		ID:             item.ID,
		SlotID:         item.SlotID,
		UserID:         item.UserID,
		Status:         item.Status,
		ConferenceLink: item.ConferenceLink,
		CreatedAt:      optionalRFC3339(item.CreatedAt),
	}
}

func mapBookings(items []repo.Booking) []bookingDTO {
	result := make([]bookingDTO, 0, len(items))
	for _, item := range items {
		result = append(result, mapBooking(item))
	}
	return result
}

func optionalRFC3339(value time.Time) *string {
	if value.IsZero() {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}
