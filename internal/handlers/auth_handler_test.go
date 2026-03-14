package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jchanning/gocase/internal/auth"
	"github.com/jchanning/gocase/internal/models"
)

type mockAuthUserStore struct {
	user              *models.User
	getByEmailErr     error
	createErr         error
	createdUser       *models.User
	initializedUserID int
}

func (m *mockAuthUserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.user, m.getByEmailErr
}

func (m *mockAuthUserStore) Create(ctx context.Context, user *models.User) error {
	m.createdUser = user
	if m.createErr != nil {
		return m.createErr
	}
	if user.ID == 0 {
		user.ID = 99
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = user.CreatedAt
	}
	return nil
}

func (m *mockAuthUserStore) InitializeUserStats(ctx context.Context, userID int) error {
	m.initializedUserID = userID
	return nil
}

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

func TestLogin_RedirectsStudentToDashboard(t *testing.T) {
	store := auth.NewSessionStore()
	hash, err := auth.HashPassword("secret123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	handler := NewAuthHandler(&mockAuthUserStore{user: &models.User{
		ID:           42,
		Email:        "student@example.com",
		PasswordHash: hash,
		Username:     "student-user",
		Role:         "student",
	}}, store)

	form := url.Values{}
	form.Set("email", "student@example.com")
	form.Set("password", "secret123")
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect status, got %d", rr.Code)
	}
	if location := rr.Header().Get("Location"); location != "/dashboard" {
		t.Fatalf("expected redirect to /dashboard, got %s", location)
	}
	if len(rr.Result().Cookies()) == 0 {
		t.Fatal("expected session cookie to be set")
	}
}

func TestRegister_DefaultsStudentRoleAndInitializesStats(t *testing.T) {
	store := auth.NewSessionStore()
	mockRepo := &mockAuthUserStore{}
	handler := NewAuthHandler(mockRepo, store)

	form := url.Values{}
	form.Set("email", "new@user.com")
	form.Set("password", "secret123")
	form.Set("username", "newuser")

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect status, got %d", rr.Code)
	}
	if location := rr.Header().Get("Location"); location != "/dashboard" {
		t.Fatalf("expected redirect to /dashboard, got %s", location)
	}
	if mockRepo.createdUser == nil {
		t.Fatal("expected user to be created")
	}
	if mockRepo.createdUser.Role != "student" {
		t.Fatalf("expected default role student, got %s", mockRepo.createdUser.Role)
	}
	if mockRepo.initializedUserID != 99 {
		t.Fatalf("expected InitializeUserStats to run for created student, got user ID %d", mockRepo.initializedUserID)
	}
	if len(rr.Result().Cookies()) == 0 {
		t.Fatal("expected session cookie to be set")
	}
}

func TestRegister_InvalidRoleDoesNotPanic(t *testing.T) {
	withRepoRootWorkingDir(t)

	store := auth.NewSessionStore()
	handler := NewAuthHandler(&mockAuthUserStore{}, store)

	form := url.Values{}
	form.Set("email", "new@user.com")
	form.Set("password", "secret123")
	form.Set("username", "newuser")
	form.Set("role", "superadmin")

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected register form re-render with status 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Invalid role specified") {
		t.Fatalf("expected error message in response body, got %q", body)
	}
}

func TestRegister_CreateErrorRerendersForm(t *testing.T) {
	withRepoRootWorkingDir(t)

	store := auth.NewSessionStore()
	handler := NewAuthHandler(&mockAuthUserStore{createErr: errors.New("duplicate")}, store)

	form := url.Values{}
	form.Set("email", "new@user.com")
	form.Set("password", "secret123")
	form.Set("username", "newuser")

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected register form re-render with status 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Email already exists or invalid data") {
		t.Fatalf("expected create error message in response body, got %q", body)
	}
}

func withRepoRootWorkingDir(t *testing.T) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	repoRoot := filepath.Clean(filepath.Join(wd, "..", ".."))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("chdir to repo root: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
}
