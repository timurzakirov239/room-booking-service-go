package domain

import "strings"

const (
	BookingStatusActive    = "active"
	BookingStatusCancelled = "cancelled"
)

func NormalizeBookingStatus(status string) string {
	return strings.TrimSpace(strings.ToLower(status))
}

func ValidateBookingStatus(status string) error {
	normalized := NormalizeBookingStatus(status)
	if normalized == BookingStatusActive || normalized == BookingStatusCancelled {
		return nil
	}

	return ErrInvalidBookingStatus
}

func IsBookingCancelled(status string) bool {
	return NormalizeBookingStatus(status) == BookingStatusCancelled
}
