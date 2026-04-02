package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"room-booking-service-go/internal/app"
	"room-booking-service-go/internal/domain"
	"room-booking-service-go/internal/repo"
	platformauth "room-booking-service-go/internal/platform/auth"
)

type createRoomRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Capacity    *int32  `json:"capacity"`
}

type createBookingRequest struct {
	SlotID               string `json:"slotId"`
	CreateConferenceLink bool   `json:"createConferenceLink"`
}

func handleRoomsList(deps RouterDependencies) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rooms, err := deps.RoomService.List(r.Context())
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list rooms")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"rooms": rooms})
	})
}

func handleRoomsCreate(deps RouterDependencies) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input createRoomRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
			return
		}
		if strings.TrimSpace(input.Name) == "" {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "name is required")
			return
		}

		room, err := deps.RoomService.Create(r.Context(), app.CreateRoomInput{
			Name:        strings.TrimSpace(input.Name),
			Description: input.Description,
			Capacity:    input.Capacity,
		})
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create room")
			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{"room": room})
	})
}

func handleBookingsCreate(deps RouterDependencies) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := platformauth.ClaimsFromContext(r.Context())
		if !ok {
			writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing auth context")
			return
		}

		var input createBookingRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
			return
		}
		if strings.TrimSpace(input.SlotID) == "" {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "slotId is required")
			return
		}

		var conferenceLink *string
		if input.CreateConferenceLink {
			conferenceLink = buildConferenceLink(strings.TrimSpace(input.SlotID))
		}

		booking, err := deps.BookingService.Create(r.Context(), app.CreateBookingInput{
			Actor:          app.Actor{UserID: claims.UserID, Role: claims.Role},
			SlotID:         strings.TrimSpace(input.SlotID),
			ConferenceLink: conferenceLink,
		})
		if err != nil {
			switch {
			case errors.Is(err, repo.ErrNotFound):
				writeAPIError(w, http.StatusNotFound, "SLOT_NOT_FOUND", "slot not found")
			case errors.Is(err, repo.ErrConflict):
				writeAPIError(w, http.StatusConflict, "SLOT_ALREADY_BOOKED", "slot is already booked")
			case errors.Is(err, domain.ErrSlotInPast):
				writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "slot is in the past")
			case errors.Is(err, domain.ErrUserRoleNotAllowed):
				writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "booking is allowed only for user role")
			default:
				writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create booking")
			}
			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{"booking": booking})
	})
}

func handleBookingsMy(deps RouterDependencies) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := platformauth.ClaimsFromContext(r.Context())
		if !ok {
			writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing auth context")
			return
		}

		bookings, err := deps.BookingService.ListMy(r.Context(), app.Actor{UserID: claims.UserID, Role: claims.Role})
		if err != nil {
			if errors.Is(err, domain.ErrUserRoleNotAllowed) {
				writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
				return
			}
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list bookings")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"bookings": bookings})
	})
}

func handleBookingsCancel(deps RouterDependencies) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := platformauth.ClaimsFromContext(r.Context())
		if !ok {
			writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing auth context")
			return
		}

		bookingID := strings.TrimSpace(r.PathValue("bookingId"))
		if bookingID == "" {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "bookingId is required")
			return
		}

		booking, err := deps.BookingService.Cancel(r.Context(), app.CancelBookingInput{
			Actor: app.Actor{UserID: claims.UserID, Role: claims.Role},
			BookingID: bookingID,
		})
		if err != nil {
			switch {
			case errors.Is(err, repo.ErrNotFound):
				writeAPIError(w, http.StatusNotFound, "BOOKING_NOT_FOUND", "booking not found")
			case errors.Is(err, domain.ErrForbiddenBooking), errors.Is(err, domain.ErrUserRoleNotAllowed):
				writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "cannot cancel another user's booking")
			default:
				writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to cancel booking")
			}
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"booking": booking})
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeAPIError(w http.ResponseWriter, statusCode int, code string, message string) {
	writeJSON(w, statusCode, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func buildConferenceLink(slotID string) *string {
	value := fmt.Sprintf("https://meet.example.local/bookings/%s", slotID)
	return &value
}

func nowUTC(now func() time.Time) time.Time {
	if now != nil {
		return now().UTC()
	}
	return time.Now().UTC()
}
