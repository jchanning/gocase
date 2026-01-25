package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"my-app/internal/auth"
	"my-app/internal/models"
	"my-app/internal/repository"
	"my-app/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

// AdminHandler handles admin/teacher requests
type AdminHandler struct {
	testRepo *repository.TestRepository
	userRepo *repository.UserRepository
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(testRepo *repository.TestRepository, userRepo *repository.UserRepository) *AdminHandler {
	return &AdminHandler{
		testRepo: testRepo,
		userRepo: userRepo,
	}
}

// ShowAdmin displays the admin dashboard
func (h *AdminHandler) ShowAdmin(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)

	tests, err := h.testRepo.GetAll(r.Context())
	if err != nil {
		log.Printf("Error fetching tests: %v", err)
		tests = []models.Test{}
	}

	subjects, err := h.testRepo.GetSubjects(r.Context())
	if err != nil {
		log.Printf("Error fetching subjects: %v", err)
		subjects = []models.Subject{}
	}

	data := map[string]interface{}{
		"Session":  session,
		"Tests":    tests,
		"Subjects": subjects,
	}

	tmpl, err := template.ParseFiles("views/layout.html", "views/admin.html")
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ShowWizard renders the manual test creation wizard for admins.
func (h *AdminHandler) ShowWizard(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)

	subjects, err := h.testRepo.GetSubjects(r.Context())
	if err != nil {
		log.Printf("Error fetching subjects: %v", err)
		subjects = []models.Subject{}
	}

	data := map[string]interface{}{
		"Session":  session,
		"Subjects": subjects,
	}

	tmpl, err := template.ParseFiles("views/layout.html", "views/admin_wizard.html")
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// UploadTest handles JSON test upload
func (h *AdminHandler) UploadTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := auth.GetSessionData(r)

	var testUpload models.TestUpload
	if err := json.NewDecoder(r.Body).Decode(&testUpload); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Invalid JSON format: %v", err),
		})
		return
	}

	if validationErrors := validateTestUpload(testUpload); len(validationErrors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"errors":  validationErrors,
		})
		return
	}

	test, err := persistTestUpload(r.Context(), h.testRepo, testUpload, session.UserID)
	if err != nil {
		log.Printf("Error creating test: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to create test: %v", err),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"test_id": test.ID,
		"message": "Test uploaded successfully",
	})
}

// CreateWizardTest handles JSON submissions from the wizard UI.
func (h *AdminHandler) CreateWizardTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := auth.GetSessionData(r)

	var testUpload models.TestUpload
	if err := json.NewDecoder(r.Body).Decode(&testUpload); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Invalid JSON format: %v", err),
		})
		return
	}

	if validationErrors := validateTestUpload(testUpload); len(validationErrors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"errors":  validationErrors,
		})
		return
	}

	test, err := persistTestUpload(r.Context(), h.testRepo, testUpload, session.UserID)
	if err != nil {
		log.Printf("Error creating test: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to create test: %v", err),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"test_id": test.ID,
		"message": "Test created successfully",
	})
}

// ShowManagement displays the admin management page
func (h *AdminHandler) ShowManagement(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)

	subjects, err := h.testRepo.GetSubjects(r.Context())
	if err != nil {
		log.Printf("Error fetching subjects: %v", err)
		subjects = []models.Subject{}
	}

	data := map[string]interface{}{
		"Session":       session,
		"Subjects":      subjects,
		"Difficulties":  models.ValidDifficulties,
		"ExamStandards": models.ValidExamStandards,
	}

	tmpl, err := template.ParseFiles("views/layout.html", "views/admin_manage.html")
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// CreateSubject handles subject creation
func (h *AdminHandler) CreateSubject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	subjectID, err := h.testRepo.GetOrCreateSubject(r.Context(), name, description)
	if err != nil {
		log.Printf("Error creating subject: %v", err)
		http.Error(w, "Failed to create subject", http.StatusInternalServerError)
		return
	}

	// Return HTML fragment for HTMX
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<div class="flex items-center justify-between p-3 bg-gray-50 rounded border">
		<div>
			<span class="font-semibold">%s</span>
			<span class="text-gray-600 text-sm ml-2">%s</span>
		</div>
		<button hx-delete="/admin/manage/subjects/%d" hx-confirm="Are you sure?" hx-target="closest div" hx-swap="outerHTML"
				class="text-red-600 hover:text-red-800 px-3 py-1">Delete</button>
	</div>`, name, description, subjectID)
}

// DeleteSubject handles subject deletion
func (h *AdminHandler) DeleteSubject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract subject ID from URL path
	// This will be handled by the router
	w.WriteHeader(http.StatusOK)
}

// ShowUserManagement displays the user management page
func (h *AdminHandler) ShowUserManagement(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)

	// Get all users
	users, err := h.userRepo.GetAllUsers(r.Context())
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		users = []models.User{}
	}

	data := map[string]interface{}{
		"Session": session,
		"Users":   users,
	}

	tmpl, err := template.ParseFiles("views/layout.html", "views/admin_users.html")
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// CreateUser creates a new user as admin
func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")
	role := r.FormValue("role")

	// Validate inputs
	if email == "" || username == "" || password == "" || role == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Validate role
	validRoles := map[string]bool{"student": true, "teacher": true, "admin": true}
	if !validRoles[role] {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Email:        email,
		Username:     username,
		PasswordHash: string(hashedPassword),
		Role:         role,
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Initialize user stats
	h.userRepo.InitializeUserStats(r.Context(), user.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user_id": user.ID,
		"message": "User created successfully",
	})
}

// UpdateUserRole updates a user's role
func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.PathValue("id")
	userID, _ := strconv.Atoi(userIDStr)
	newRole := r.FormValue("role")

	// Validate role
	validRoles := map[string]bool{"student": true, "teacher": true, "admin": true}
	if !validRoles[newRole] {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	if err := h.userRepo.UpdateUserRole(r.Context(), userID, newRole); err != nil {
		log.Printf("Error updating user role: %v", err)
		http.Error(w, "Failed to update user role", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User role updated successfully",
	})
}

// ResetUserPassword resets a user's password
func (h *AdminHandler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.PathValue("id")
	userID, _ := strconv.Atoi(userIDStr)
	newPassword := r.FormValue("password")

	if newPassword == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Failed to reset password", http.StatusInternalServerError)
		return
	}

	if err := h.userRepo.UpdatePasswordHash(r.Context(), userID, string(hashedPassword)); err != nil {
		log.Printf("Error resetting password: %v", err)
		http.Error(w, "Failed to reset password", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Password reset successfully",
	})
}

// DeleteUser deletes a user
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.PathValue("id")
	userID, _ := strconv.Atoi(userIDStr)

	if err := h.userRepo.DeleteUser(r.Context(), userID); err != nil {
		log.Printf("Error deleting user: %v", err)
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User deleted successfully",
	})
}

// DeleteTest removes a test and its questions (admin-only).
func (h *AdminHandler) DeleteTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	testIDStr := r.PathValue("id")
	testID, err := strconv.Atoi(testIDStr)
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	if err := h.testRepo.DeleteTest(r.Context(), testID); err != nil {
		log.Printf("Error deleting test: %v", err)
		http.Error(w, "Failed to delete test", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Test deleted",
	})
}

// UpdateTest handles test notes upload/update
func (h *AdminHandler) UpdateTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	testIDStr := r.PathValue("id")
	testID, err := strconv.Atoi(testIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid test ID",
		})
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32 MB max
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to parse form data",
		})
		return
	}

	// Get the test first
	test, err := h.testRepo.GetByID(r.Context(), testID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Test not found",
		})
		return
	}

	// Handle notes file upload
	file, header, err := r.FormFile("notes_file")
	if err == nil {
		defer file.Close()

		// Delete old notes file if exists
		if test.NotesFilename != nil && *test.NotesFilename != "" {
			storage.DeleteNotesFile(*test.NotesFilename)
		}

		// Save new notes file
		filename, err := storage.SaveNotesFile(file, header.Filename)
		if err != nil {
			log.Printf("Error saving notes file: %v", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("Failed to save notes: %v", err),
			})
			return
		}

		test.NotesFilename = &filename
	}

	// Update test in database
	if err := h.testRepo.UpdateTestNotes(r.Context(), testID, test.NotesFilename); err != nil {
		log.Printf("Error updating test notes: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to update test",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Test updated successfully",
	})
}

// RemoveTestNotes removes notes from a test
func (h *AdminHandler) RemoveTestNotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	testIDStr := r.PathValue("id")
	testID, err := strconv.Atoi(testIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid test ID",
		})
		return
	}

	// Get the test first
	test, err := h.testRepo.GetByID(r.Context(), testID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Test not found",
		})
		return
	}

	// Delete notes file if exists
	if test.NotesFilename != nil && *test.NotesFilename != "" {
		if err := storage.DeleteNotesFile(*test.NotesFilename); err != nil {
			log.Printf("Error deleting notes file: %v", err)
		}
	}

	// Update test in database (set notes to null)
	if err := h.testRepo.UpdateTestNotes(r.Context(), testID, nil); err != nil {
		log.Printf("Error updating test notes: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to remove notes",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Notes removed successfully",
	})
}

// ServeTestNotes serves the notes file for a test
func (h *AdminHandler) ServeTestNotes(w http.ResponseWriter, r *http.Request) {
	testIDStr := r.PathValue("id")
	testID, err := strconv.Atoi(testIDStr)
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	// Get the test
	test, err := h.testRepo.GetByID(r.Context(), testID)
	if err != nil || test.NotesFilename == nil || *test.NotesFilename == "" {
		http.Error(w, "Notes not found", http.StatusNotFound)
		return
	}

	// Get file path
	filePath := storage.GetNotesFilePath(*test.NotesFilename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Notes file not found", http.StatusNotFound)
		return
	}

	// Serve file inline (not as download)
	w.Header().Set("Content-Disposition", "inline")
	http.ServeFile(w, r, filePath)
}
