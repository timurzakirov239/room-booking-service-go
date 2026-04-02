package httpx

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type RouterDependencies struct {
	BuildVersion string
	Now          func() time.Time
	DBPing       func(context.Context) error
}

type infoResponse struct {
	Status      string `json:"status"`
	Service     string `json:"service"`
	Version     string `json:"version"`
	NowUTC      string `json:"nowUtc"`
	DatabaseOK  bool   `json:"databaseOk"`
}

func NewRouter(deps RouterDependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /_info", func(w http.ResponseWriter, r *http.Request) {
		writeInfo(w, r, deps)
	})

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
