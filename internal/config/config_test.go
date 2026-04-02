package config

import (
	"testing"
)

func TestLoadFromEnvUsesDefaultsAndRequiredDatabaseURL(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("DATABASE_MAX_CONNS", "")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Fatalf("AppEnv = %q, want development", cfg.AppEnv)
	}
	if cfg.HTTPPort != "8080" {
		t.Fatalf("HTTPPort = %q, want 8080", cfg.HTTPPort)
	}
	if cfg.DatabaseURL != "postgres://example" {
		t.Fatalf("DatabaseURL = %q, want postgres://example", cfg.DatabaseURL)
	}
	if cfg.DatabaseMaxConns != 4 {
		t.Fatalf("DatabaseMaxConns = %d, want 4", cfg.DatabaseMaxConns)
	}
	if cfg.HTTPListenAddress() != ":8080" {
		t.Fatalf("HTTPListenAddress() = %q, want :8080", cfg.HTTPListenAddress())
	}
}

func TestLoadFromEnvRespectsExplicitValues(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://custom")
	t.Setenv("DATABASE_MAX_CONNS", "7")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	if cfg.AppEnv != "test" {
		t.Fatalf("AppEnv = %q, want test", cfg.AppEnv)
	}
	if cfg.HTTPPort != "9090" {
		t.Fatalf("HTTPPort = %q, want 9090", cfg.HTTPPort)
	}
	if cfg.DatabaseMaxConns != 7 {
		t.Fatalf("DatabaseMaxConns = %d, want 7", cfg.DatabaseMaxConns)
	}
}

func TestLoadFromEnvRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_MAX_CONNS", "")

	_, err := LoadFromEnv()
	if err == nil {
		t.Fatal("LoadFromEnv() error = nil, want error")
	}
}

func TestLoadFromEnvRejectsInvalidDatabaseMaxConns(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("DATABASE_MAX_CONNS", "invalid")

	_, err := LoadFromEnv()
	if err == nil {
		t.Fatal("LoadFromEnv() error = nil, want error")
	}
}
