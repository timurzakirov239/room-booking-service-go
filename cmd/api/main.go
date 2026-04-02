package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"room-booking-service-go/internal/api/httpx"
	"room-booking-service-go/internal/config"
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
		return runMigrations(ctx, cfg, args)
	}

	handler := httpx.NewRouter(httpx.RouterDependencies{
		BuildVersion: buildVersion,
		Now:          timeutil.NowUTC,
		DBPing: func(ctx context.Context) error {
			return pool.Ping(ctx)
		},
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

func runMigrations(ctx context.Context, cfg config.Config, args []string) error {
	if len(args) < 2 || args[0] != "migrate" {
		return fmt.Errorf("unsupported command")
	}

	action := args[1]
	if action != "up" && action != "down" {
		return fmt.Errorf("unsupported migrate action: %s", action)
	}

	slog.Info("migration scaffold invoked", "action", action, "database_url_set", cfg.DatabaseURL != "")
	_ = ctx
	return nil
}
