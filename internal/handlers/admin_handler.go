package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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

// EditTest displays the edit page for a test (admin-only)
func (h *AdminHandler) EditTest(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)
	testIDStr := r.PathValue("id")
	testID, _ := strconv.Atoi(testIDStr)

	// Get existing test
	test, err := h.testRepo.GetByID(r.Context(), testID)
	if err != nil {
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}

	subjects, err := h.testRepo.GetSubjects(r.Context())
	if err != nil {
		subjects = []models.Subject{}
	}

	data := map[string]interface{}{
		"Session":  session,
		"Test":     test,
		"Subjects": subjects,
	}

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).ParseFiles("views/layout.html", "views/edit_test.html")
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

	// Redirect back to admin dashboard
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// UpdateTest handles full test updates including metadata and notes
func (h *AdminHandler) UpdateTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	testIDStr := r.PathValue("id")
	testID, err := strconv.Atoi(testIDStr)
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	// Try to parse multipart form first (for file uploads), then fall back to regular form
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil { // 32 MB max
			http.Error(w, fmt.Sprintf("Failed to parse multipart form data: %v", err), http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse form data: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Get the test first
	test, err := h.testRepo.GetByID(r.Context(), testID)
	if err != nil {
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}

	// Check if this is a full update or just notes update
	titleParam := r.FormValue("title")
	if titleParam != "" {
		// Full test update - only update test fields
		test.Title = titleParam
		test.Description = r.FormValue("description")
		test.ExamStandard = r.FormValue("exam_standard")
		test.Difficulty = r.FormValue("difficulty")
		test.PassingScore = parseIntOrDefault(r.FormValue("passing_score"), 60)
		test.TimeLimitMinutes = parseIntOrDefault(r.FormValue("time_limit_minutes"), 10)

		log.Printf("Updating test %d: title=%s, description=%s", testID, test.Title, test.Description)

		// Update test in database
		if err := h.testRepo.Update(r.Context(), test); err != nil {
			log.Printf("Error updating test: %v", err)
			http.Error(w, fmt.Sprintf("Failed to update test: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("Test %d updated successfully", testID)

		// Update questions
		for idx, q := range test.Questions {
			questionText := r.FormValue(fmt.Sprintf("question_%d_text", idx))
			pointsStr := r.FormValue(fmt.Sprintf("question_%d_points", idx))

			if questionText != "" {
				q.QuestionText = questionText
				q.Points = parseIntOrDefault(pointsStr, 1)

				log.Printf("Updating question %d: text=%s, points=%d", q.ID, q.QuestionText, q.Points)

				if err := h.testRepo.UpdateQuestion(r.Context(), &q); err != nil {
					log.Printf("Error updating question: %v", err)
					http.Error(w, fmt.Sprintf("Failed to update question: %v", err), http.StatusInternalServerError)
					return
				}
			}

			// Update answer options and set correct answer
			correctOptionID := parseIntOrDefault(r.FormValue(fmt.Sprintf("question_%d_correct_option", idx)), 0)
			for optIdx, opt := range q.Options {
				optionText := r.FormValue(fmt.Sprintf("question_%d_option_%d_text", idx, optIdx))
				if optionText != "" {
					opt.OptionText = optionText
					opt.IsCorrect = (opt.ID == correctOptionID)

					log.Printf("Updating option %d: text=%s, isCorrect=%v", opt.ID, opt.OptionText, opt.IsCorrect)

					if err := h.testRepo.UpdateAnswerOption(r.Context(), &opt); err != nil {
						log.Printf("Error updating answer option: %v", err)
						http.Error(w, fmt.Sprintf("Failed to update answer option: %v", err), http.StatusInternalServerError)
						return
					}
				}
			}
		}

		log.Printf("Redirecting to /admin/test/%d/edit", testID)
		http.Redirect(w, r, fmt.Sprintf("/admin/test/%d/edit", testID), http.StatusSeeOther)
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
			http.Error(w, fmt.Sprintf("Failed to save notes: %v", err), http.StatusInternalServerError)
			return
		}

		test.NotesFilename = &filename
	}

	// Update test in database
	if err := h.testRepo.UpdateTestNotes(r.Context(), testID, test.NotesFilename); err != nil {
		log.Printf("Error updating test notes: %v", err)
		http.Error(w, "Failed to update test", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/test/%d/edit", testID), http.StatusSeeOther)
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

// parseIntOrDefault parses a string as int or returns default
func parseIntOrDefault(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return defaultVal
}
