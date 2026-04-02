package app

import (
	"context"
	"errors"
	"time"

	"room-booking-service-go/internal/domain"
	"room-booking-service-go/internal/repo"
)

type BookingService struct {
	Users    repo.UsersRepository
	Slots    repo.SlotsRepository
	Bookings repo.BookingsRepository
	Now      func() time.Time
}

type Actor struct {
	UserID string
	Role   string
}

type CreateBookingInput struct {
	Actor          Actor
	SlotID         string
	ConferenceLink *string
}

func (s BookingService) ListMy(ctx context.Context, actor Actor) ([]repo.Booking, error) {
	if actor.Role != domain.RoleUser {
		return nil, domain.ErrUserRoleNotAllowed
	}

	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}

	return s.Bookings.ListByUser(ctx, repo.ListBookingsByUserParams{
		UserID: actor.UserID,
		From:   &now,
	})
}

type CancelBookingInput struct {
	Actor     Actor
	BookingID string
}

func (s BookingService) Create(ctx context.Context, input CreateBookingInput) (repo.Booking, error) {
	if !domain.CanCreateBooking(input.Actor.Role) {
		return repo.Booking{}, domain.ErrUserRoleNotAllowed
	}

	user, err := s.Users.GetByID(ctx, input.Actor.UserID)
	if err != nil {
		return repo.Booking{}, err
	}
	if !domain.CanCreateBooking(user.Role) {
		return repo.Booking{}, domain.ErrUserRoleNotAllowed
	}

	slot, err := s.Slots.GetByID(ctx, input.SlotID)
	if err != nil {
		return repo.Booking{}, err
	}

	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	if !slot.StartAt.UTC().After(now) {
		return repo.Booking{}, domain.ErrSlotInPast
	}

	booking, err := s.Bookings.Create(ctx, repo.CreateBookingParams{
		SlotID:         slot.ID,
		UserID:         user.ID,
		Status:         domain.BookingStatusActive,
		ConferenceLink: input.ConferenceLink,
	})
	if err != nil {
		return repo.Booking{}, err
	}

	return booking, nil
}

func (s BookingService) Cancel(ctx context.Context, input CancelBookingInput) (repo.Booking, error) {
	booking, err := s.Bookings.GetByID(ctx, input.BookingID)
	if err != nil {
		return repo.Booking{}, err
	}

	if domain.IsBookingCancelled(booking.Status) {
		return booking, nil
	}

	if !domain.CanCancelBooking(input.Actor.Role, input.Actor.UserID, booking.UserID) {
		return repo.Booking{}, domain.ErrForbiddenBooking
	}

	cancelledBooking, err := s.Bookings.Cancel(ctx, booking.ID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return booking, nil
		}
		return repo.Booking{}, err
	}

	return cancelledBooking, nil
}
