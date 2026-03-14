package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jchanning/gocase/internal/version"
)

type contextKey string

const sessionDataKey contextKey = "sessionData"

// Middleware creates an authentication middleware
type Middleware struct {
	store *SessionStore
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(store *SessionStore) *Middleware {
	return &Middleware{store: store}
}

// isJSONRequest returns true when the request carries a JSON body (i.e. an AJAX fetch).
func isJSONRequest(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Content-Type"), "application/json") ||
		strings.Contains(r.Header.Get("Accept"), "application/json")
}

// respondUnauthorized sends a JSON 401 for AJAX requests or redirects for normal browser requests.
func respondUnauthorized(w http.ResponseWriter, r *http.Request) {
	if isJSONRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck
			"error": "Session expired. Please log in again.",
		})
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// RequireAuth is middleware that requires authentication
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := GetSessionToken(r)
		if err != nil {
			respondUnauthorized(w, r)
			return
		}

		session, exists := m.store.Get(token)
		if !exists {
			respondUnauthorized(w, r)
			return
		}

		// Copy session to avoid mutating the shared store entry, then inject app version.
		sessionCopy := *session
		sessionCopy.AppVersion = version.Version
		ctx := context.WithValue(r.Context(), sessionDataKey, &sessionCopy)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole is middleware that requires a specific role
func (m *Middleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := GetSessionData(r)
			if session == nil {
				respondUnauthorized(w, r)
				return
			}

			// Check if user has one of the required roles
			hasRole := false
			for _, role := range roles {
				if session.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetSessionData retrieves session data from request context
func GetSessionData(r *http.Request) *SessionData {
	session, ok := r.Context().Value(sessionDataKey).(*SessionData)
	if !ok {
		return nil
	}
	return session
}
