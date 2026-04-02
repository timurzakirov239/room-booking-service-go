package postgres

import (
	"time"

	"room-booking-service-go/internal/repo"
)

func scanUser(row rowScanner) (repo.User, error) {
	var item repo.User
	if err := row.Scan(&item.ID, &item.Email, &item.PasswordHash, &item.Role, &item.CreatedAt); err != nil {
		return repo.User{}, normalizeError(err)
	}
	return item, nil
}

func scanRoom(row rowScanner) (repo.Room, error) {
	var item repo.Room
	if err := row.Scan(&item.ID, &item.Name, &item.Description, &item.Capacity, &item.CreatedAt); err != nil {
		return repo.Room{}, normalizeError(err)
	}
	return item, nil
}

func scanSchedule(row rowScanner) (repo.Schedule, error) {
	var item repo.Schedule
	var startTime time.Time
	var endTime time.Time
	if err := row.Scan(&item.ID, &item.RoomID, &item.DaysOfWeek, &startTime, &endTime, &item.CreatedAt); err != nil {
		return repo.Schedule{}, normalizeError(err)
	}
	item.StartTime = startTime.UTC()
	item.EndTime = endTime.UTC()
	return item, nil
}

func scanSlot(row rowScanner) (repo.Slot, error) {
	var item repo.Slot
	if err := row.Scan(&item.ID, &item.RoomID, &item.ScheduleID, &item.StartAt, &item.EndAt, &item.CreatedAt); err != nil {
		return repo.Slot{}, normalizeError(err)
	}
	item.StartAt = item.StartAt.UTC()
	item.EndAt = item.EndAt.UTC()
	item.CreatedAt = item.CreatedAt.UTC()
	return item, nil
}

func scanBooking(row rowScanner) (repo.Booking, error) {
	var item repo.Booking
	if err := row.Scan(&item.ID, &item.SlotID, &item.UserID, &item.Status, &item.ConferenceLink, &item.CreatedAt); err != nil {
		return repo.Booking{}, normalizeError(err)
	}
	item.CreatedAt = item.CreatedAt.UTC()
	return item, nil
}
