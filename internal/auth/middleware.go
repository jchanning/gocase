package auth

import (
	"context"
	"net/http"
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

// RequireAuth is middleware that requires authentication
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := GetSessionToken(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session, exists := m.store.Get(token)
		if !exists {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add session data to request context
		ctx := context.WithValue(r.Context(), sessionDataKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole is middleware that requires a specific role
func (m *Middleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := GetSessionData(r)
			if session == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
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
