package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	platformauth "room-booking-service-go/internal/platform/auth"
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

func TestDummyLoginReturnsTokenForValidRole(t *testing.T) {
	fixedNow := time.Date(2026, 4, 2, 13, 0, 0, 0, time.UTC)
	handler := NewRouter(RouterDependencies{
		AuthSigner: platformauth.Signer{
			Secret:   "test-secret",
			Issuer:   "test-issuer",
			Lifetime: time.Hour,
			Now: func() time.Time {
				return fixedNow
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBufferString(`{"role":"user"}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
	}

	var payload struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if payload.Token == "" {
		t.Fatal("token is empty")
	}

	claims, err := (platformauth.Signer{Secret: "test-secret", Now: func() time.Time { return fixedNow }}).Verify(payload.Token)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if claims.Role != "user" {
		t.Fatalf("role = %q, want user", claims.Role)
	}
	if claims.UserID != platformauth.DummyUserUserID {
		t.Fatalf("user_id = %q, want %q", claims.UserID, platformauth.DummyUserUserID)
	}
}

func TestDummyLoginRejectsInvalidRole(t *testing.T) {
	handler := NewRouter(RouterDependencies{
		AuthSigner: platformauth.Signer{Secret: "test-secret"},
	})

	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBufferString(`{"role":"guest"}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.Code)
	}
}

func TestProtectedRouteRequiresAuth(t *testing.T) {
	handler := NewRouter(RouterDependencies{
		AuthSigner: platformauth.Signer{Secret: "test-secret"},
	})

	req := httptest.NewRequest(http.MethodGet, "/rooms/list", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", res.Code)
	}
}

func TestProtectedRouteRejectsWrongRole(t *testing.T) {
	signer := platformauth.Signer{Secret: "test-secret", Lifetime: time.Hour}
	token, err := signer.Sign(platformauth.Claims{UserID: platformauth.DummyUserUserID, Role: "user"})
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	handler := NewRouter(RouterDependencies{AuthSigner: signer})
	req := httptest.NewRequest(http.MethodGet, "/bookings/list", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", res.Code)
	}
}

func TestProtectedRouteReturnsPlaceholderWhenAuthorized(t *testing.T) {
	signer := platformauth.Signer{Secret: "test-secret", Lifetime: time.Hour}
	token, err := signer.Sign(platformauth.Claims{UserID: platformauth.DummyUserUserID, Role: "user"})
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	handler := NewRouter(RouterDependencies{AuthSigner: signer})
	req := httptest.NewRequest(http.MethodGet, "/bookings/my", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want 501", res.Code)
	}

	var payload map[string]map[string]string
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if payload["error"]["code"] != "NOT_IMPLEMENTED" {
		t.Fatalf("error.code = %q, want NOT_IMPLEMENTED", payload["error"]["code"])
	}
}
