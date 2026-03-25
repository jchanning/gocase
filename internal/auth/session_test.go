package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ---- SessionStore.Create ----

func TestSessionStore_Create_ReturnsNonEmptyToken(t *testing.T) {
	store := NewSessionStore()
	token, err := store.Create(1, "alice", "student")
	if err != nil {
		t.Fatalf("unexpected error from Create: %v", err)
	}
	if token == "" {
		t.Fatal("expected a non-empty session token")
	}
}

func TestSessionStore_Create_TokensAreUnique(t *testing.T) {
	store := NewSessionStore()
	token1, _ := store.Create(1, "alice", "student")
	token2, _ := store.Create(2, "bob", "teacher")
	if token1 == token2 {
		t.Fatal("expected distinct tokens for different sessions")
	}
}

// ---- SessionStore.Get ----

func TestSessionStore_Get_ReturnsCorrectData(t *testing.T) {
	store := NewSessionStore()
	token, _ := store.Create(42, "carol", "admin")

	data, ok := store.Get(token)
	if !ok {
		t.Fatal("expected session to exist after Create")
	}
	if data.UserID != 42 {
		t.Errorf("expected UserID 42, got %d", data.UserID)
	}
	if data.Username != "carol" {
		t.Errorf("expected Username %q, got %q", "carol", data.Username)
	}
	if data.Role != "admin" {
		t.Errorf("expected Role %q, got %q", "admin", data.Role)
	}
}

func TestSessionStore_Get_ReturnsFalseForUnknownToken(t *testing.T) {
	store := NewSessionStore()
	_, ok := store.Get("nonexistent-token")
	if ok {
		t.Fatal("expected Get to return false for an unknown token")
	}
}

func TestSessionStore_Get_ReturnsFalseForExpiredSession(t *testing.T) {
	store := NewSessionStore()
	token, _ := store.Create(1, "alice", "student")

	// Manually back-date the session to simulate expiry.
	store.sessions[token].CreatedAt = time.Now().Add(-25 * time.Hour)

	_, ok := store.Get(token)
	if ok {
		t.Fatal("expected expired session to be rejected")
	}
}

func TestSessionStore_Get_RemovesExpiredSession(t *testing.T) {
	store := NewSessionStore()
	token, _ := store.Create(1, "alice", "student")
	store.sessions[token].CreatedAt = time.Now().Add(-25 * time.Hour)

	store.Get(token) // triggers eviction

	if _, exists := store.sessions[token]; exists {
		t.Fatal("expected expired session to be evicted from the store")
	}
}

// ---- SessionStore.Delete ----

func TestSessionStore_Delete_RemovesSession(t *testing.T) {
	store := NewSessionStore()
	token, _ := store.Create(1, "alice", "student")

	store.Delete(token)

	_, ok := store.Get(token)
	if ok {
		t.Fatal("expected session to be gone after Delete")
	}
}

func TestSessionStore_Delete_IsIdempotent(t *testing.T) {
	store := NewSessionStore()
	token, _ := store.Create(1, "alice", "student")
	store.Delete(token)
	// Second delete must not panic.
	store.Delete(token)
}

// ---- GetSessionToken ----

func TestGetSessionToken_ExtractsCookieValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "abc123"})

	got, err := GetSessionToken(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abc123" {
		t.Errorf("expected token %q, got %q", "abc123", got)
	}
}

func TestGetSessionToken_ErrorWhenCookieMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := GetSessionToken(req)
	if err == nil {
		t.Fatal("expected error when session_token cookie is absent")
	}
}

// ---- HashPassword / CheckPasswordHash ----

func TestHashPassword_ProducesValidHash(t *testing.T) {
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash == "secret123" {
		t.Fatal("hash must not equal the plaintext password")
	}
}

func TestCheckPasswordHash_CorrectPassword(t *testing.T) {
	hash, _ := HashPassword("correct")
	if !CheckPasswordHash("correct", hash) {
		t.Fatal("expected correct password to match hash")
	}
}

func TestCheckPasswordHash_WrongPassword(t *testing.T) {
	hash, _ := HashPassword("correct")
	if CheckPasswordHash("wrong", hash) {
		t.Fatal("expected wrong password to not match hash")
	}
}
