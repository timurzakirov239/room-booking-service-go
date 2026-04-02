package authx

import (
	"encoding/json"
	"net/http"

	platformauth "room-booking-service-go/internal/platform/auth"
)

type DummyLoginHandler struct {
	Signer platformauth.Signer
}

type dummyLoginRequest struct {
	Role string `json:"role"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

func (h DummyLoginHandler) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input dummyLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
			return
		}

		userID, err := platformauth.DummyUserIDForRole(input.Role)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid role")
			return
		}

		token, err := h.Signer.Sign(platformauth.Claims{
			UserID: userID,
			Role:   input.Role,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to issue token")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(tokenResponse{Token: token})
	})
}
