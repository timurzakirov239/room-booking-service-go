package repo

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("repo: not found")
	ErrConflict = errors.New("repo: conflict")
)

type User struct {
	ID           string
	Email        string
	PasswordHash *string
	Role         string
	CreatedAt    time.Time
}

type Room struct {
	ID          string
	Name        string
	Description *string
	Capacity    *int32
	CreatedAt   time.Time
}

type Schedule struct {
	ID         string
	RoomID     string
	DaysOfWeek []int16
	StartTime  time.Time
	EndTime    time.Time
	CreatedAt  time.Time
}

type Slot struct {
	ID         string
	RoomID     string
	ScheduleID *string
	StartAt    time.Time
	EndAt      time.Time
	CreatedAt  time.Time
}

type Booking struct {
	ID             string
	SlotID         string
	UserID         string
	Status         string
	ConferenceLink *string
	CreatedAt      time.Time
}

type CreateUserParams struct {
	Email        string
	PasswordHash *string
	Role         string
}

type CreateRoomParams struct {
	Name        string
	Description *string
	Capacity    *int32
}

type CreateScheduleParams struct {
	RoomID     string
	DaysOfWeek []int16
	StartTime  time.Time
	EndTime    time.Time
}

type CreateSlotParams struct {
	RoomID     string
	ScheduleID *string
	StartAt    time.Time
	EndAt      time.Time
}

type CreateBookingParams struct {
	SlotID         string
	UserID         string
	Status         string
	ConferenceLink *string
}

type ListSlotsParams struct {
	RoomID string
	From   time.Time
	To     time.Time
}

type ListBookingsByUserParams struct {
	UserID string
	From   *time.Time
}

type ListBookingsParams struct {
	Page     int
	PageSize int
}

type UsersRepository interface {
	Create(context.Context, CreateUserParams) (User, error)
	GetByID(context.Context, string) (User, error)
	GetByEmail(context.Context, string) (User, error)
}

type RoomsRepository interface {
	Create(context.Context, CreateRoomParams) (Room, error)
	GetByID(context.Context, string) (Room, error)
	List(context.Context) ([]Room, error)
}

type SchedulesRepository interface {
	Create(context.Context, CreateScheduleParams) (Schedule, error)
	GetByRoomID(context.Context, string) (Schedule, error)
}

type SlotsRepository interface {
	Create(context.Context, CreateSlotParams) (Slot, error)
	GetByID(context.Context, string) (Slot, error)
	ListByRoomAndRange(context.Context, ListSlotsParams) ([]Slot, error)
}

type BookingsRepository interface {
	Create(context.Context, CreateBookingParams) (Booking, error)
	GetByID(context.Context, string) (Booking, error)
	List(context.Context, ListBookingsParams) ([]Booking, int, error)
	ListByUser(context.Context, ListBookingsByUserParams) ([]Booking, error)
	Cancel(context.Context, string) (Booking, error)
}
