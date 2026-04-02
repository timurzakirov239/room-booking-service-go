package domain

import "errors"

var (
	ErrInvalidTimeRange     = errors.New("domain: invalid time range")
	ErrInvalidSlotRange     = errors.New("domain: invalid slot range")
	ErrInvalidBookingStatus = errors.New("domain: invalid booking status")
	ErrForbiddenBooking     = errors.New("domain: booking action forbidden")
	ErrSlotInPast           = errors.New("domain: slot is in the past")
	ErrSlotAlreadyCancelled = errors.New("domain: booking already cancelled")
	ErrUserRoleNotAllowed   = errors.New("domain: user role not allowed")
)
