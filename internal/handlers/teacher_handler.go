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
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate JSON structure
	if testUpload.Title == "" {
		http.Error(w, "Test title is required", http.StatusBadRequest)
		return
	}
	if len(testUpload.Questions) == 0 {
		http.Error(w, "At least one question is required", http.StatusBadRequest)
		return
	}
	for i, q := range testUpload.Questions {
		if q.QuestionText == "" {
			http.Error(w, fmt.Sprintf("Question %d text is required", i+1), http.StatusBadRequest)
			return
		}
		if len(q.Options) != 4 {
			http.Error(w, fmt.Sprintf("Question %d must have exactly 4 options", i+1), http.StatusBadRequest)
			return
		}
		if q.CorrectIndex < 0 || q.CorrectIndex > 3 {
			http.Error(w, fmt.Sprintf("Question %d has invalid correct index", i+1), http.StatusBadRequest)
			return
		}
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

		// Create answer options
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
