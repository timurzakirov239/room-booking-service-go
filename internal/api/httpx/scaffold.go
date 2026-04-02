package httpx

import (
	"encoding/json"
	"net/http"
)

type EndpointScaffold struct {
	Resource string
	Action   string
}

func (h EndpointScaffold) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "NOT_IMPLEMENTED",
				"message": h.Resource + " " + h.Action + " is not implemented yet",
			},
		})
	})
}
