package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// Principal holds the authenticated caller's identity, extracted from a JWT.
type Principal struct {
	InspectorID string
	CompanyID   string
	Role        string
}

// TokenVerifier verifies a raw bearer token and returns the caller's Principal.
// Satisfied by internal/identity/infrastructure/auth.JWTService via an adapter.
type TokenVerifier interface {
	VerifyToken(token string) (Principal, error)
}

type contextKey int

const principalKey contextKey = iota

// PrincipalFromContext retrieves the authenticated Principal from the request context.
// Returns false if there is no authenticated principal (request was not routed through
// the Authenticate middleware).
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey).(Principal)
	return p, ok
}

// Authenticate returns middleware that requires a valid JWT bearer token.
// Requests without a valid token receive a 401 Unauthorized response.
func Authenticate(v TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hdr := r.Header.Get("Authorization")
			if !strings.HasPrefix(hdr, "Bearer ") {
				writeUnauthorized(w)
				return
			}

			token := strings.TrimPrefix(hdr, "Bearer ")
			principal, err := v.VerifyToken(token)
			if err != nil {
				writeUnauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), principalKey, principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
}
