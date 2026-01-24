package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"my-app/internal/auth"
)

// Ensures non-admin users cannot register teacher/admin roles.
func TestRegister_ForbidsElevatedRoleForNonAdmin(t *testing.T) {
	store := auth.NewSessionStore()
	token, _ := store.Create(10, "StudentUser", "student")
	mw := auth.NewMiddleware(store)

	handler := NewAuthHandler(nil, store)

	form := url.Values{}
	form.Set("email", "new@user.com")
	form.Set("password", "secret123")
	form.Set("username", "newuser")
	form.Set("role", "teacher")

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})

	rr := httptest.NewRecorder()

	// Chain through RequireAuth to populate session context, then call handler.
	wrapped := mw.RequireAuth(http.HandlerFunc(handler.Register))
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when student attempts to register teacher, got %d", rr.Code)
	}
}
