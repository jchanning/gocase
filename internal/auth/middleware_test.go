package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Tests RequireAuth + RequireRole chaining to ensure role gating works as expected.
func TestRequireRole_AllowsAdmin(t *testing.T) {
	store := NewSessionStore()
	token, _ := store.Create(1, "AdminUser", "admin")
	mw := NewMiddleware(store)

	handler := mw.RequireAuth(mw.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for admin role, got %d", rr.Code)
	}
}

func TestRequireRole_DeniesWrongRole(t *testing.T) {
	store := NewSessionStore()
	token, _ := store.Create(2, "StudentUser", "student")
	mw := NewMiddleware(store)

	handler := mw.RequireAuth(mw.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for student role, got %d", rr.Code)
	}
}

func TestRequireAuth_RedirectsWhenMissingSession(t *testing.T) {
	store := NewSessionStore()
	mw := NewMiddleware(store)

	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect when no session, got %d", rr.Code)
	}
}
