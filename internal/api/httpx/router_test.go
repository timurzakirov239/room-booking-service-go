package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestInfoRouteReturns200AndPayloadWhenDatabasePingSucceeds(t *testing.T) {
	fixedNow := time.Date(2026, 4, 2, 9, 0, 0, 0, time.FixedZone("LOCAL", 3*60*60))
	handler := NewRouter(RouterDependencies{
		BuildVersion: "test-build",
		Now: func() time.Time {
			return fixedNow
		},
		DBPing: func(_ context.Context) error {
			return nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/_info", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
	}
	if got := res.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var payload infoResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if payload.Status != "ok" {
		t.Fatalf("status payload = %q, want ok", payload.Status)
	}
	if payload.Service != "room-booking-service-go" {
		t.Fatalf("service = %q, want room-booking-service-go", payload.Service)
	}
	if payload.Version != "test-build" {
		t.Fatalf("version = %q, want test-build", payload.Version)
	}
	if payload.NowUTC != fixedNow.UTC().Format(time.RFC3339) {
		t.Fatalf("nowUtc = %q, want %q", payload.NowUTC, fixedNow.UTC().Format(time.RFC3339))
	}
	if !payload.DatabaseOK {
		t.Fatal("databaseOk = false, want true")
	}
}

func TestInfoRouteStillReturns200WhenDatabasePingFails(t *testing.T) {
	handler := NewRouter(RouterDependencies{
		DBPing: func(_ context.Context) error {
			return errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/_info", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
	}

	var payload infoResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if payload.DatabaseOK {
		t.Fatal("databaseOk = true, want false")
	}
}
