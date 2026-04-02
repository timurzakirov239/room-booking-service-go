package app

import (
	"context"

	"room-booking-service-go/internal/repo"
)

type RoomService struct {
	Rooms repo.RoomsRepository
}

type CreateRoomInput struct {
	Name        string
	Description *string
	Capacity    *int32
}

func (s RoomService) List(ctx context.Context) ([]repo.Room, error) {
	return s.Rooms.List(ctx)
}

func (s RoomService) Create(ctx context.Context, input CreateRoomInput) (repo.Room, error) {
	return s.Rooms.Create(ctx, repo.CreateRoomParams{
		Name:        input.Name,
		Description: input.Description,
		Capacity:    input.Capacity,
	})
}
