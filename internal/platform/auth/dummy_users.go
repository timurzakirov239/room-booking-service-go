package auth

import (
	"fmt"

	"room-booking-service-go/internal/domain"
)

const (
	DummyAdminUserID = "11111111-1111-1111-1111-111111111111"
	DummyUserUserID  = "22222222-2222-2222-2222-222222222222"
)

func DummyUserIDForRole(role string) (string, error) {
	switch role {
	case domain.RoleAdmin:
		return DummyAdminUserID, nil
	case domain.RoleUser:
		return DummyUserUserID, nil
	default:
		return "", fmt.Errorf("unsupported role: %s", role)
	}
}
