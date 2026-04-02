package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"room-booking-service-go/internal/api/httpx"
	"room-booking-service-go/internal/app"
	"room-booking-service-go/internal/config"
	platformauth "room-booking-service-go/internal/platform/auth"
	"room-booking-service-go/internal/platform/postgres"
	"room-booking-service-go/internal/platform/timeutil"
)

const buildVersion = "dev"

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		slog.Error("application exited with error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return err
	}

	pool, err := postgres.NewPool(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	if len(args) > 0 {
		return runMigrations(ctx, pool, cfg, args)
	}

	if cfg.AutoMigrate {
		if err := postgres.Bootstrap(ctx, pool); err != nil {
			return err
		}
	}

	store := postgres.NewStore(pool)
	authSigner := platformauth.Signer{
		Secret:   cfg.JWTSecret,
		Issuer:   cfg.JWTIssuer,
		Lifetime: 24 * time.Hour,
		Now:      timeutil.NowUTC,
	}

	roomService := app.RoomService{Rooms: store.Rooms}
	materializer := app.SlotMaterializer{Slots: store.Slots}
	scheduleService := app.ScheduleService{
		Rooms:     store.Rooms,
		Schedules: store.Schedules,
	}
	slotService := app.SlotService{
		Rooms:        store.Rooms,
		Schedules:    store.Schedules,
		Slots:        store.Slots,
		Materializer: materializer,
	}
	bookingService := app.BookingService{
		Users:    store.Users,
		Slots:    store.Slots,
		Bookings: store.Bookings,
		Now:      timeutil.NowUTC,
	}

	handler := httpx.NewRouter(httpx.RouterDependencies{
		BuildVersion: buildVersion,
		Now:          timeutil.NowUTC,
		DBPing: func(ctx context.Context) error {
			return pool.Ping(ctx)
		},
		AuthSigner:      authSigner,
		Store:           store,
		RoomService:     roomService,
		ScheduleService: scheduleService,
		SlotService:     slotService,
		BookingService:  bookingService,
		Materializer:    materializer,
	})

	server := &http.Server{
		Addr:              cfg.HTTPListenAddress(),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("starting http server", "addr", server.Addr, "env", cfg.AppEnv)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool, cfg config.Config, args []string) error {
	if len(args) < 2 || args[0] != "migrate" {
		return fmt.Errorf("unsupported command")
	}

	action := args[1]
	if action != "up" {
		return fmt.Errorf("unsupported migrate action: %s", action)
	}

	slog.Info("running migrations", "action", action, "database_url_set", cfg.DatabaseURL != "")
	return postgres.Bootstrap(ctx, pool)
}
