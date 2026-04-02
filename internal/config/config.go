package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultHTTPPort = "8080"
)

type Config struct {
	AppEnv          string
	HTTPPort        string
	DatabaseURL     string
	DatabaseMaxConns int32
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		HTTPPort:    getEnv("HTTP_PORT", defaultHTTPPort),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	maxConns, err := getEnvInt32("DATABASE_MAX_CONNS", 4)
	if err != nil {
		return Config{}, fmt.Errorf("parse DATABASE_MAX_CONNS: %w", err)
	}

	cfg.DatabaseMaxConns = maxConns

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func (c Config) HTTPListenAddress() string {
	return ":" + c.HTTPPort
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getEnvInt32(key string, fallback int32) (int32, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(parsed), nil
}
