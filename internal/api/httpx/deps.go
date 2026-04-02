package httpx

import (
	"context"
	"time"

	"room-booking-service-go/internal/app"
	"room-booking-service-go/internal/repo/postgres"
	platformauth "room-booking-service-go/internal/platform/auth"
)

type RouterDependencies struct {
	BuildVersion   string
	Now            func() time.Time
	DBPing         func(context.Context) error
	AuthSigner     platformauth.Signer
	Store          postgres.Store
	RoomService    app.RoomService
	BookingService app.BookingService
	Materializer   app.SlotMaterializer
}
