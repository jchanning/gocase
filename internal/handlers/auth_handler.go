package handlers

import (
	"html/template"
	"log"
	"net/http"

	"my-app/internal/auth"
	"my-app/internal/models"
	"my-app/internal/repository"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	userRepo     *repository.UserRepository
	sessionStore *auth.SessionStore
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo *repository.UserRepository, sessionStore *auth.SessionStore) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		sessionStore: sessionStore,
	}
}

// ShowLogin displays the login page
func (h *AuthHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("views/layout.html", "views/login.html")
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Session": nil,
	}

	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Login handles login form submission
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	// Get user by email
	user, err := h.userRepo.GetByEmail(r.Context(), email)
	if err != nil {
		log.Printf("Login failed for %s: user not found", email)
		h.showLoginError(w, "Invalid email or password")
		return
	}

	// Check password
	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		log.Printf("Login failed for %s: invalid password", email)
		h.showLoginError(w, "Invalid email or password")
		return
	}

	// Create session
	token, err := h.sessionStore.Create(user.ID, user.Username, user.Role)
	if err != nil {
		log.Printf("Error creating session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	auth.SetSessionCookie(w, token)

	// Redirect based on role
	if user.Role == "admin" || user.Role == "teacher" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

// ShowRegister displays the registration page
func (h *AuthHandler) ShowRegister(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("views/layout.html", "views/register.html")
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Check if current user is admin
	sessionData := auth.GetSessionData(r)
	isAdmin := sessionData != nil && sessionData.Role == "admin"

	data := map[string]interface{}{
		"Session": sessionData,
		"IsAdmin": isAdmin,
	}

	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Register handles registration form submission
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	username := r.FormValue("username")
	role := r.FormValue("role")

	// Default to student if role not provided
	if role == "" {
		role = "student"
	}

	// Validate role
	if role != "student" && role != "teacher" && role != "admin" {
		h.showRegisterError(w, "Invalid role specified")
		return
	}

	// Check if trying to create teacher/admin account
	if role == "teacher" || role == "admin" {
		// Only admins can create teacher/admin accounts
		sessionData := auth.GetSessionData(r)
		if sessionData == nil || sessionData.Role != "admin" {
			http.Error(w, "Forbidden: Only administrators can create teacher or admin accounts", http.StatusForbidden)
			return
		}
	}

	// Hash password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create user
	user := &models.User{
		Email:        email,
		PasswordHash: passwordHash,
		Username:     username,
		Role:         role,
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		log.Printf("Error creating user: %v", err)
		h.showRegisterError(w, "Email already exists or invalid data")
		return
	}

	// Initialize user stats for students
	if role == "student" {
		h.userRepo.InitializeUserStats(r.Context(), user.ID)
	}

	// Create session
	token, err := h.sessionStore.Create(user.ID, user.Username, user.Role)
	if err != nil {
		log.Printf("Error creating session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	auth.SetSessionCookie(w, token)

	// Redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetSessionToken(r)
	if err == nil {
		h.sessionStore.Delete(token)
	}

	auth.ClearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// showLoginError displays login page with error
func (h *AuthHandler) showLoginError(w http.ResponseWriter, errorMsg string) {
	tmpl, _ := template.ParseFiles("views/layout.html", "views/login.html")
	data := map[string]interface{}{
		"Session": nil,
		"Error":   errorMsg,
	}
	tmpl.ExecuteTemplate(w, "layout.html", data)
}

// showRegisterError displays register page with error
func (h *AuthHandler) showRegisterError(w http.ResponseWriter, errorMsg string) {
	tmpl, _ := template.ParseFiles("views/layout.html", "views/register.html")
	sessionData := auth.GetSessionData(nil)
	isAdmin := sessionData != nil && sessionData.Role == "admin"
	data := map[string]interface{}{
		"Session": sessionData,
		"IsAdmin": isAdmin,
		"Error":   errorMsg,
	}
	tmpl.ExecuteTemplate(w, "layout.html", data)
}
