package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// SessionStore manages user sessions (in-memory for simplicity, use Redis in production)
type SessionStore struct {
	sessions map[string]*SessionData
}

// SessionData holds session information
type SessionData struct {
	UserID    int
	Username  string
	Role      string
	CreatedAt time.Time
}

// NewSessionStore creates a new session store
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*SessionData),
	}
}

// Create creates a new session for a user
func (s *SessionStore) Create(userID int, username, role string) (string, error) {
	// Generate a random session token
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}

	s.sessions[token] = &SessionData{
		UserID:    userID,
		Username:  username,
		Role:      role,
		CreatedAt: time.Now(),
	}

	return token, nil
}

// Get retrieves session data by token
func (s *SessionStore) Get(token string) (*SessionData, bool) {
	session, exists := s.sessions[token]
	if !exists {
		return nil, false
	}

	// Check if session has expired (24 hours)
	if time.Since(session.CreatedAt) > 24*time.Hour {
		delete(s.sessions, token)
		return nil, false
	}

	return session, true
}

// Delete removes a session
func (s *SessionStore) Delete(token string) {
	delete(s.sessions, token)
}

// GetSessionToken extracts the session token from the request cookie
func GetSessionToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// SetSessionCookie sets the session cookie in the response
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie clears the session cookie
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a password with a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateSessionToken generates a random session token
func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
