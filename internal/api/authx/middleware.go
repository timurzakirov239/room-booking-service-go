package authx

import (
	"encoding/json"
	"net/http"
	"strings"

	platformauth "room-booking-service-go/internal/platform/auth"
)

type Middleware struct {
	Signer platformauth.Signer
}

func (m Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerTokenFromHeader(r.Header.Get("Authorization"))
		if !ok {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid bearer token")
			return
		}

		claims, err := m.Signer.Verify(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
			return
		}

		next.ServeHTTP(w, r.WithContext(platformauth.WithClaims(r.Context(), claims)))
	})
}

func (m Middleware) RequireRole(requiredRole string, next http.Handler) http.Handler {
	return m.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := platformauth.ClaimsFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing auth context")
			return
		}
		if claims.Role != requiredRole {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
			return
		}

		next.ServeHTTP(w, r)
	}))
}

func bearerTokenFromHeader(header string) (string, bool) {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}

func writeError(w http.ResponseWriter, statusCode int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
