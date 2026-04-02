package domain

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

func CanCreateBooking(role string) bool {
	return role == RoleUser
}

func CanCancelBooking(actorRole string, actorUserID string, bookingUserID string) bool {
	if actorRole == RoleAdmin {
		return true
	}

	return actorRole == RoleUser && actorUserID == bookingUserID
}
