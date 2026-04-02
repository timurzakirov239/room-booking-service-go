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

	"room-booking-service-go/internal/app"
	"room-booking-service-go/internal/domain"
	"room-booking-service-go/internal/repo"
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

type fakeRoomsRepo struct {
	items []repo.Room
	createResult repo.Room
	createErr error
}

func (f *fakeRoomsRepo) Create(_ context.Context, _ repo.CreateRoomParams) (repo.Room, error) {
	return f.createResult, f.createErr
}

func (f *fakeRoomsRepo) GetByID(_ context.Context, id string) (repo.Room, error) {
	for _, item := range f.items {
		if item.ID == id {
			return item, nil
		}
	}
	if f.createResult.ID == id {
		return f.createResult, nil
	}
	return repo.Room{}, repo.ErrNotFound
}

func (f *fakeRoomsRepo) List(_ context.Context) ([]repo.Room, error) {
	return f.items, nil
}

type fakeUsersRepo struct {
	user repo.User
	err error
}

func (f *fakeUsersRepo) Create(_ context.Context, _ repo.CreateUserParams) (repo.User, error) {
	return repo.User{}, nil
}

func (f *fakeUsersRepo) GetByID(_ context.Context, _ string) (repo.User, error) {
	return f.user, f.err
}

func (f *fakeUsersRepo) GetByEmail(_ context.Context, _ string) (repo.User, error) {
	return repo.User{}, repo.ErrNotFound
}

type fakeSlotsRepo struct {
	slot repo.Slot
	err error
	items []repo.Slot
	createErr error
}

func (f *fakeSlotsRepo) Create(_ context.Context, params repo.CreateSlotParams) (repo.Slot, error) {
	if f.createErr != nil {
		return repo.Slot{}, f.createErr
	}
	created := repo.Slot{
		ID:         "slot-created",
		RoomID:     params.RoomID,
		ScheduleID: params.ScheduleID,
		StartAt:    params.StartAt,
		EndAt:      params.EndAt,
	}
	f.items = append(f.items, created)
	return created, nil
}

func (f *fakeSlotsRepo) GetByID(_ context.Context, _ string) (repo.Slot, error) {
	return f.slot, f.err
}

func (f *fakeSlotsRepo) ListByRoomAndRange(_ context.Context, _ repo.ListSlotsParams) ([]repo.Slot, error) {
	return f.items, nil
}

type fakeBookingsRepo struct {
	created repo.Booking
	createErr error
	items []repo.Booking
	listTotal int
	getByID repo.Booking
	getByIDErr error
	cancelled repo.Booking
	cancelErr error
}

type fakeSchedulesRepo struct {
	created repo.Schedule
	createErr error
	getByRoom repo.Schedule
	getByRoomErr error
}

func (f *fakeSchedulesRepo) Create(_ context.Context, _ repo.CreateScheduleParams) (repo.Schedule, error) {
	return f.created, f.createErr
}

func (f *fakeSchedulesRepo) GetByRoomID(_ context.Context, _ string) (repo.Schedule, error) {
	return f.getByRoom, f.getByRoomErr
}

func (f *fakeBookingsRepo) Create(_ context.Context, _ repo.CreateBookingParams) (repo.Booking, error) {
	return f.created, f.createErr
}

func (f *fakeBookingsRepo) GetByID(_ context.Context, _ string) (repo.Booking, error) {
	return f.getByID, f.getByIDErr
}

func (f *fakeBookingsRepo) List(_ context.Context, _ repo.ListBookingsParams) ([]repo.Booking, int, error) {
	total := f.listTotal
	if total == 0 {
		total = len(f.items)
	}
	return f.items, total, nil
}

func (f *fakeBookingsRepo) ListByUser(_ context.Context, _ repo.ListBookingsByUserParams) ([]repo.Booking, error) {
	return f.items, nil
}

func (f *fakeBookingsRepo) Cancel(_ context.Context, _ string) (repo.Booking, error) {
	return f.cancelled, f.cancelErr
}

func TestRoomsCreateAndListScenario(t *testing.T) {
	adminSigner := platformauth.Signer{Secret: "test-secret", Lifetime: time.Hour}
	adminToken, err := adminSigner.Sign(platformauth.Claims{UserID: platformauth.DummyAdminUserID, Role: domain.RoleAdmin})
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	createdRoom := repo.Room{ID: "room-1", Name: "Atlas"}
	roomsRepo := &fakeRoomsRepo{
		items: []repo.Room{createdRoom},
		createResult: createdRoom,
	}

	handler := NewRouter(RouterDependencies{
		AuthSigner: adminSigner,
		RoomService: app.RoomService{Rooms: roomsRepo},
	})

	createReq := httptest.NewRequest(http.MethodPost, "/rooms/create", bytes.NewBufferString(`{"name":"Atlas"}`))
	createReq.Header.Set("Authorization", "Bearer "+adminToken)
	createRes := httptest.NewRecorder()
	handler.ServeHTTP(createRes, createReq)

	if createRes.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", createRes.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/rooms/list", nil)
	listReq.Header.Set("Authorization", "Bearer "+adminToken)
	listRes := httptest.NewRecorder()
	handler.ServeHTTP(listRes, listReq)

	if listRes.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200", listRes.Code)
	}
}

func TestScheduleCreateScenario(t *testing.T) {
	adminSigner := platformauth.Signer{Secret: "test-secret", Lifetime: time.Hour}
	adminToken, err := adminSigner.Sign(platformauth.Claims{UserID: platformauth.DummyAdminUserID, Role: domain.RoleAdmin})
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	createdRoom := repo.Room{ID: "room-1", Name: "Atlas"}
	createdSchedule := repo.Schedule{
		ID:         "schedule-1",
		RoomID:     "room-1",
		DaysOfWeek: []int16{1},
		StartTime:  time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:    time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC),
	}
	roomsRepo := &fakeRoomsRepo{items: []repo.Room{createdRoom}}
	schedulesRepo := &fakeSchedulesRepo{created: createdSchedule}

	handler := NewRouter(RouterDependencies{
		AuthSigner: adminSigner,
		ScheduleService: app.ScheduleService{
			Rooms:     roomsRepo,
			Schedules: schedulesRepo,
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/rooms/room-1/schedule/create", bytes.NewBufferString(`{"daysOfWeek":[1],"startTime":"09:00","endTime":"10:00"}`))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", res.Code)
	}
}

func TestSlotsListScenario(t *testing.T) {
	userSigner := platformauth.Signer{Secret: "test-secret", Lifetime: time.Hour}
	userToken, err := userSigner.Sign(platformauth.Claims{UserID: platformauth.DummyUserUserID, Role: domain.RoleUser})
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	room := repo.Room{ID: "room-1", Name: "Atlas"}
	schedule := repo.Schedule{
		ID:         "schedule-1",
		RoomID:     "room-1",
		DaysOfWeek: []int16{5},
		StartTime:  time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:    time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC),
	}
	slotsRepo := &fakeSlotsRepo{}

	handler := NewRouter(RouterDependencies{
		AuthSigner: userSigner,
		SlotService: app.SlotService{
			Rooms:        &fakeRoomsRepo{items: []repo.Room{room}},
			Schedules:    &fakeSchedulesRepo{getByRoom: schedule},
			Slots:        slotsRepo,
			Materializer: app.SlotMaterializer{Slots: slotsRepo},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/rooms/room-1/slots/list?date=2026-04-03", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
	}
}

func TestBookingsListScenario(t *testing.T) {
	adminSigner := platformauth.Signer{Secret: "test-secret", Lifetime: time.Hour}
	adminToken, err := adminSigner.Sign(platformauth.Claims{UserID: platformauth.DummyAdminUserID, Role: domain.RoleAdmin})
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	booking := repo.Booking{ID: "booking-1", SlotID: "slot-1", UserID: platformauth.DummyUserUserID, Status: domain.BookingStatusActive}
	bookingRepo := &fakeBookingsRepo{items: []repo.Booking{booking}, listTotal: 1}

	handler := NewRouter(RouterDependencies{
		AuthSigner: adminSigner,
		BookingService: app.BookingService{
			Bookings: bookingRepo,
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/bookings/list?page=1&pageSize=20", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
	}
}

func TestBookingCreateScenario(t *testing.T) {
	userSigner := platformauth.Signer{Secret: "test-secret", Lifetime: time.Hour}
	userToken, err := userSigner.Sign(platformauth.Claims{UserID: platformauth.DummyUserUserID, Role: domain.RoleUser})
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	futureSlot := repo.Slot{ID: "slot-1", StartAt: time.Now().UTC().Add(time.Hour)}
	createdBooking := repo.Booking{ID: "booking-1", SlotID: "slot-1", UserID: platformauth.DummyUserUserID, Status: domain.BookingStatusActive}
	bookingService := app.BookingService{
		Users:    &fakeUsersRepo{user: repo.User{ID: platformauth.DummyUserUserID, Role: domain.RoleUser}},
		Slots:    &fakeSlotsRepo{slot: futureSlot},
		Bookings: &fakeBookingsRepo{created: createdBooking},
		Now:      time.Now,
	}

	handler := NewRouter(RouterDependencies{AuthSigner: userSigner, BookingService: bookingService})
	req := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewBufferString(`{"slotId":"slot-1","createConferenceLink":true}`))
	req.Header.Set("Authorization", "Bearer "+userToken)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", res.Code)
	}
}

func TestBookingsMyAndCancelScenario(t *testing.T) {
	userSigner := platformauth.Signer{Secret: "test-secret", Lifetime: time.Hour}
	userToken, err := userSigner.Sign(platformauth.Claims{UserID: platformauth.DummyUserUserID, Role: domain.RoleUser})
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	booking := repo.Booking{ID: "booking-1", SlotID: "slot-1", UserID: platformauth.DummyUserUserID, Status: domain.BookingStatusActive}
	bookingRepo := &fakeBookingsRepo{
		items: []repo.Booking{booking},
		getByID: booking,
		cancelled: repo.Booking{ID: "booking-1", SlotID: "slot-1", UserID: platformauth.DummyUserUserID, Status: domain.BookingStatusCancelled},
	}
	bookingService := app.BookingService{
		Users:    &fakeUsersRepo{user: repo.User{ID: platformauth.DummyUserUserID, Role: domain.RoleUser}},
		Slots:    &fakeSlotsRepo{},
		Bookings: bookingRepo,
		Now:      time.Now,
	}

	handler := NewRouter(RouterDependencies{AuthSigner: userSigner, BookingService: bookingService})

	myReq := httptest.NewRequest(http.MethodGet, "/bookings/my", nil)
	myReq.Header.Set("Authorization", "Bearer "+userToken)
	myRes := httptest.NewRecorder()
	handler.ServeHTTP(myRes, myReq)

	if myRes.Code != http.StatusOK {
		t.Fatalf("my status = %d, want 200", myRes.Code)
	}

	cancelReq := httptest.NewRequest(http.MethodPost, "/bookings/booking-1/cancel", nil)
	cancelReq.Header.Set("Authorization", "Bearer "+userToken)
	cancelRes := httptest.NewRecorder()
	handler.ServeHTTP(cancelRes, cancelReq)

	if cancelRes.Code != http.StatusOK {
		t.Fatalf("cancel status = %d, want 200", cancelRes.Code)
	}
}
