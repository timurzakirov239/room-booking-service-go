package httpx

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"room-booking-service-go/internal/api/authx"
	"room-booking-service-go/internal/domain"
)

type infoResponse struct {
	Status     string `json:"status"`
	Service    string `json:"service"`
	Version    string `json:"version"`
	NowUTC     string `json:"nowUtc"`
	DatabaseOK bool   `json:"databaseOk"`
}

func NewRouter(deps RouterDependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /_info", func(w http.ResponseWriter, r *http.Request) {
		writeInfo(w, r, deps)
	})

	mux.Handle("POST /dummyLogin", authx.DummyLoginHandler{Signer: deps.AuthSigner}.Handler())

	authMiddleware := authx.Middleware{Signer: deps.AuthSigner}

	mux.Handle("GET /rooms/list", authMiddleware.RequireAuth(handleRoomsList(deps)))
	mux.Handle("POST /rooms/create", authMiddleware.RequireRole(domain.RoleAdmin, handleRoomsCreate(deps)))
	mux.Handle("POST /rooms/{roomId}/schedule/create", authMiddleware.RequireRole(domain.RoleAdmin, handleScheduleCreate(deps)))
	mux.Handle("GET /rooms/{roomId}/slots/list", authMiddleware.RequireAuth(handleSlotsList(deps)))
	mux.Handle("POST /bookings/create", authMiddleware.RequireRole(domain.RoleUser, handleBookingsCreate(deps)))
	mux.Handle("GET /bookings/list", authMiddleware.RequireRole(domain.RoleAdmin, handleBookingsList(deps)))
	mux.Handle("GET /bookings/my", authMiddleware.RequireRole(domain.RoleUser, handleBookingsMy(deps)))
	mux.Handle("POST /bookings/{bookingId}/cancel", authMiddleware.RequireRole(domain.RoleUser, handleBookingsCancel(deps)))

	return mux
}

func writeInfo(w http.ResponseWriter, r *http.Request, deps RouterDependencies) {
	w.Header().Set("Content-Type", "application/json")

	databaseOK := true
	if deps.DBPing != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := deps.DBPing(ctx); err != nil {
			databaseOK = false
		}
	}

	now := time.Now().UTC()
	if deps.Now != nil {
		now = deps.Now().UTC()
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(infoResponse{
		Status:     "ok",
		Service:    "room-booking-service-go",
		Version:    deps.BuildVersion,
		NowUTC:     now.Format(time.RFC3339),
		DatabaseOK: databaseOK,
	})
}
