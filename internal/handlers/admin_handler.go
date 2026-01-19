package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"my-app/internal/auth"
	"my-app/internal/models"
	"my-app/internal/repository"
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
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Get or create subject
	subjectID, err := h.testRepo.GetOrCreateSubject(r.Context(), testUpload.Subject, "")
	if err != nil {
		log.Printf("Error creating subject: %v", err)
		http.Error(w, "Failed to create subject", http.StatusInternalServerError)
		return
	}

	// Get or create topic
	var topicID *int
	if testUpload.Topic != "" {
		id, err := h.testRepo.GetOrCreateTopic(r.Context(), subjectID, testUpload.Topic, "")
		if err != nil {
			log.Printf("Error creating topic: %v", err)
			http.Error(w, "Failed to create topic", http.StatusInternalServerError)
			return
		}
		topicID = &id
	}

	// Create test
	test := &models.Test{
		Title:            testUpload.Title,
		Description:      testUpload.Description,
		SubjectID:        &subjectID,
		TopicID:          topicID,
		ExamStandard:     testUpload.ExamStandard,
		Difficulty:       testUpload.Difficulty,
		TimeLimitMinutes: testUpload.TimeLimitMinutes,
		PassingScore:     testUpload.PassingScore,
		CreatedBy:        &session.UserID,
	}

	if err := h.testRepo.Create(r.Context(), test); err != nil {
		log.Printf("Error creating test: %v", err)
		http.Error(w, "Failed to create test", http.StatusInternalServerError)
		return
	}

	// Create questions and options
	for i, q := range testUpload.Questions {
		question := &models.Question{
			TestID:        test.ID,
			QuestionText:  q.QuestionText,
			QuestionOrder: i + 1,
			Points:        q.Points,
		}

		if q.ImageURL != "" {
			question.ImageURL = &q.ImageURL
		}

		if err := h.testRepo.CreateQuestion(r.Context(), question); err != nil {
			log.Printf("Error creating question: %v", err)
			continue
		}

		// Create answer options (expecting 4)
		for j, optText := range q.Options {
			option := &models.AnswerOption{
				QuestionID:  question.ID,
				OptionText:  optText,
				IsCorrect:   j == q.CorrectIndex,
				OptionOrder: j + 1,
			}

			if err := h.testRepo.CreateAnswerOption(r.Context(), option); err != nil {
				log.Printf("Error creating option: %v", err)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"test_id": test.ID,
		"message": "Test uploaded successfully",
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
		"Session":  session,
		"Subjects": subjects,
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
