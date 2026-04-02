APP_NAME := room-booking-service-go

.PHONY: build run test migrate-up migrate-down

build:
	go build ./...

run:
	go run ./cmd/api

test:
	go test ./...

migrate-up:
	go run ./cmd/api migrate up

migrate-down:
	go run ./cmd/api migrate down
