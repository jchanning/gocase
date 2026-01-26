package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"my-app/internal/auth"
	"my-app/internal/models"
	"my-app/internal/repository"
	"my-app/internal/validation"
)

// TeacherHandler handles teacher requests
type TeacherHandler struct {
	testRepo    *repository.TestRepository
	userRepo    *repository.UserRepository
	attemptRepo *repository.AttemptRepository
}

// NewTeacherHandler creates a new teacher handler
func NewTeacherHandler(testRepo *repository.TestRepository, userRepo *repository.UserRepository, attemptRepo *repository.AttemptRepository) *TeacherHandler {
	return &TeacherHandler{
		testRepo:    testRepo,
		userRepo:    userRepo,
		attemptRepo: attemptRepo,
	}
}

// ShowDashboard displays the teacher dashboard
func (h *TeacherHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)

	// Get tests created by this teacher
	tests, err := h.testRepo.GetByCreator(r.Context(), session.UserID)
	if err != nil {
		log.Printf("Error fetching tests: %v", err)
		tests = []models.Test{}
	}

	// Get statistics for each test
	testStats := make(map[int]map[string]interface{})
	for _, test := range tests {
		attempts, _ := h.attemptRepo.GetByTestID(r.Context(), test.ID)

		totalAttempts := len(attempts)
		var totalScore float64
		completedAttempts := 0

		for _, attempt := range attempts {
			if attempt.CompletedAt != nil && attempt.Score != nil {
				completedAttempts++
				totalScore += float64(*attempt.Score)
			}
		}

		avgScore := 0.0
		if completedAttempts > 0 {
			avgScore = totalScore / float64(completedAttempts)
		}

		testStats[test.ID] = map[string]interface{}{
			"total_attempts":     totalAttempts,
			"completed_attempts": completedAttempts,
			"average_score":      fmt.Sprintf("%.1f", avgScore),
		}
	}

	data := map[string]interface{}{
		"Session":   session,
		"Tests":     tests,
		"TestStats": testStats,
	}

	tmpl, err := template.ParseFiles("views/layout.html", "views/teacher_dashboard.html")
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

// ShowUpload displays the test upload page
func (h *TeacherHandler) ShowUpload(w http.ResponseWriter, r *http.Request) {
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

	tmpl, err := template.ParseFiles("views/layout.html", "views/teacher_upload.html")
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

// UploadTest handles JSON test upload for teachers
func (h *TeacherHandler) UploadTest(w http.ResponseWriter, r *http.Request) {
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

// ShowCreateTest displays the create/edit test page
func (h *TeacherHandler) ShowCreateTest(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)

	subjects, err := h.testRepo.GetSubjects(r.Context())
	if err != nil {
		log.Printf("Error fetching subjects: %v", err)
		subjects = []models.Subject{}
	}

	// Check if this is an edit request
	testID := r.URL.Query().Get("id")
	var test *models.Test
	if testID != "" {
		id, _ := strconv.Atoi(testID)
		test, _ = h.testRepo.GetByID(r.Context(), id)
	}

	data := map[string]interface{}{
		"Session":  session,
		"Subjects": subjects,
		"Test":     test,
		"IsEdit":   test != nil,
	}

	tmpl, err := template.ParseFiles("views/layout.html", "views/create_test.html")
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

// CreateTest handles test creation via form submission
func (h *TeacherHandler) CreateTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := auth.GetSessionData(r)
	r.ParseMultipartForm(32 << 20) // 32MB max

	// Parse form data
	test := &models.Test{
		Title:            r.FormValue("title"),
		Description:      r.FormValue("description"),
		ExamStandard:     r.FormValue("exam_standard"),
		Difficulty:       r.FormValue("difficulty"),
		PassingScore:     parseIntOrDefault(r.FormValue("passing_score"), 60),
		TimeLimitMinutes: parseIntOrDefault(r.FormValue("time_limit_minutes"), 10),
		CreatedBy:        &session.UserID,
	}

	// Parse subject and topic
	if subjectIDStr := r.FormValue("subject_id"); subjectIDStr != "" {
		subjectID, _ := strconv.Atoi(subjectIDStr)
		test.SubjectID = &subjectID
	}

	// Validate test
	validator := validation.NewTestValidator()
	if !validator.ValidateTest(test) {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "Validation failed", http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"errors":  validator.GetErrorMessages(),
		})
		return
	}

	// Create test
	if err := h.testRepo.Create(r.Context(), test); err != nil {
		log.Printf("Error creating test: %v", err)
		http.Error(w, "Failed to create test", http.StatusInternalServerError)
		return
	}

	// Handle questions
	numQuestions := parseIntOrDefault(r.FormValue("num_questions"), 0)
	for i := 1; i <= numQuestions; i++ {
		questionText := r.FormValue(fmt.Sprintf("question_%d_text", i))
		if questionText == "" {
			continue
		}

		question := &models.Question{
			TestID:        test.ID,
			QuestionText:  questionText,
			QuestionOrder: i,
			Points:        parseIntOrDefault(r.FormValue(fmt.Sprintf("question_%d_points", i)), 1),
		}

		// Handle image upload
		if file, handler, err := r.FormFile(fmt.Sprintf("question_%d_image", i)); err == nil {
			defer file.Close()
			imagePath, err := h.saveUploadedImage(test.ID, i, file, handler.Filename)
			if err == nil && imagePath != "" {
				question.ImageURL = &imagePath
			}
		}

		if err := h.testRepo.CreateQuestion(r.Context(), question); err != nil {
			log.Printf("Error creating question: %v", err)
			continue
		}

		// Create answer options
		for j := 1; j <= 4; j++ {
			optionText := r.FormValue(fmt.Sprintf("question_%d_option_%d", i, j))
			if optionText == "" {
				continue
			}

			isCorrect := r.FormValue(fmt.Sprintf("question_%d_correct", i)) == strconv.Itoa(j)

			option := &models.AnswerOption{
				QuestionID:  question.ID,
				OptionText:  optionText,
				IsCorrect:   isCorrect,
				OptionOrder: j,
			}

			if err := h.testRepo.CreateAnswerOption(r.Context(), option); err != nil {
				log.Printf("Error creating option: %v", err)
			}
		}
	}

	// Redirect to test edit page
	http.Redirect(w, r, fmt.Sprintf("/teacher/test/%d/edit", test.ID), http.StatusSeeOther)
}

// EditTest handles test updates
func (h *TeacherHandler) EditTest(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)
	testIDStr := r.PathValue("id")
	testID, _ := strconv.Atoi(testIDStr)

	// Get existing test
	test, err := h.testRepo.GetByID(r.Context(), testID)
	if err != nil {
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if test.CreatedBy == nil || *test.CreatedBy != session.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
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

// UpdateTest handles test updates
func (h *TeacherHandler) UpdateTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := auth.GetSessionData(r)
	testIDStr := r.PathValue("id")
	testID, _ := strconv.Atoi(testIDStr)

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

	// Get existing test
	test, err := h.testRepo.GetByID(r.Context(), testID)
	if err != nil {
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if test.CreatedBy == nil || *test.CreatedBy != session.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Update test fields
	test.Title = r.FormValue("title")
	test.Description = r.FormValue("description")
	test.ExamStandard = r.FormValue("exam_standard")
	test.Difficulty = r.FormValue("difficulty")
	test.PassingScore = parseIntOrDefault(r.FormValue("passing_score"), 60)
	test.TimeLimitMinutes = parseIntOrDefault(r.FormValue("time_limit_minutes"), 10)

	log.Printf("Updating test %d: title=%s, description=%s", testID, test.Title, test.Description)

	// Validate
	validator := validation.NewTestValidator()
	if !validator.ValidateTest(test) {
		http.Error(w, "Validation failed", http.StatusBadRequest)
		return
	}

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

	log.Printf("Redirecting to /teacher/test/%d/edit", testID)
	http.Redirect(w, r, fmt.Sprintf("/teacher/test/%d/edit", test.ID), http.StatusSeeOther)
}

// PreviewTest displays a preview of the test for the teacher
func (h *TeacherHandler) PreviewTest(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)
	testIDStr := r.PathValue("id")
	testID, _ := strconv.Atoi(testIDStr)

	test, err := h.testRepo.GetByID(r.Context(), testID)
	if err != nil {
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}

	// Verify ownership or admin
	if test.CreatedBy != nil && *test.CreatedBy != session.UserID && session.Role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	data := map[string]interface{}{
		"Session": session,
		"Test":    test,
	}

	tmpl, err := template.ParseFiles("views/layout.html", "views/test_preview.html")
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

// PublishTest publishes a test
func (h *TeacherHandler) PublishTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := auth.GetSessionData(r)
	testIDStr := r.PathValue("id")
	testID, _ := strconv.Atoi(testIDStr)

	// Get test to verify ownership
	test, err := h.testRepo.GetByID(r.Context(), testID)
	if err != nil {
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}

	if test.CreatedBy == nil || *test.CreatedBy != session.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.testRepo.PublishTest(r.Context(), testID); err != nil {
		log.Printf("Error publishing test: %v", err)
		http.Error(w, "Failed to publish test", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Test published successfully",
	})
}

// UnpublishTest unpublishes a test
func (h *TeacherHandler) UnpublishTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := auth.GetSessionData(r)
	testIDStr := r.PathValue("id")
	testID, _ := strconv.Atoi(testIDStr)

	// Get test to verify ownership
	test, err := h.testRepo.GetByID(r.Context(), testID)
	if err != nil {
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}

	if test.CreatedBy == nil || *test.CreatedBy != session.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.testRepo.UnpublishTest(r.Context(), testID); err != nil {
		log.Printf("Error unpublishing test: %v", err)
		http.Error(w, "Failed to unpublish test", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Test unpublished successfully",
	})
}

// DeleteTest deletes a test
func (h *TeacherHandler) DeleteTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := auth.GetSessionData(r)
	testIDStr := r.PathValue("id")
	testID, _ := strconv.Atoi(testIDStr)

	// Get test to verify ownership
	test, err := h.testRepo.GetByID(r.Context(), testID)
	if err != nil {
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}

	if test.CreatedBy == nil || *test.CreatedBy != session.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.testRepo.DeleteTest(r.Context(), testID); err != nil {
		log.Printf("Error deleting test: %v", err)
		http.Error(w, "Failed to delete test", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Test deleted successfully",
		})
	} else {
		http.Redirect(w, r, "/teacher/dashboard", http.StatusSeeOther)
	}
}

// saveUploadedImage saves an uploaded image file
func (h *TeacherHandler) saveUploadedImage(testID int, questionNumber int, file io.ReadCloser, filename string) (string, error) {
	// Create uploads directory if it doesn't exist
	uploadDir := filepath.Join("assets", "uploads", strconv.Itoa(testID))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Generate filename
	ext := filepath.Ext(filename)
	if !isAllowedImageType(ext) {
		return "", fmt.Errorf("unsupported image type")
	}

	fname := fmt.Sprintf("question_%d%s", questionNumber, ext)
	fpath := filepath.Join(uploadDir, fname)
	webPath := fmt.Sprintf("/assets/uploads/%d/%s", testID, fname)

	// Save file
	out, err := os.Create(fpath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}

	return webPath, nil
}

// Helper functions

func isAllowedImageType(ext string) bool {
	ext = strings.ToLower(ext)
	allowed := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	return allowed[ext]
}
