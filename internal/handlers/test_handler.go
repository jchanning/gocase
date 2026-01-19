package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"my-app/internal/auth"
	"my-app/internal/models"
	"my-app/internal/repository"
)

// TestHandler handles test-related requests
type TestHandler struct {
	testRepo    *repository.TestRepository
	attemptRepo *repository.AttemptRepository
	userRepo    *repository.UserRepository
}

// NewTestHandler creates a new test handler
func NewTestHandler(testRepo *repository.TestRepository, attemptRepo *repository.AttemptRepository, userRepo *repository.UserRepository) *TestHandler {
	return &TestHandler{
		testRepo:    testRepo,
		attemptRepo: attemptRepo,
		userRepo:    userRepo,
	}
}

// ListTests displays all available tests
func (h *TestHandler) ListTests(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)

	tests, err := h.testRepo.GetAll(r.Context())
	if err != nil {
		log.Printf("Error fetching tests: %v", err)
		http.Error(w, "Failed to load tests", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Session": session,
		"Tests":   tests,
	}

	tmpl, err := template.ParseFiles("views/layout.html", "views/tests_list.html")
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

// StartTest creates a new test attempt and displays the first question
func (h *TestHandler) StartTest(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)
	testIDStr := r.URL.Query().Get("id")
	testID, err := strconv.Atoi(testIDStr)
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	// Get test details (just to validate it exists)
	_, err = h.testRepo.GetByID(r.Context(), testID)
	if err != nil {
		log.Printf("Error fetching test: %v", err)
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}

	// Create new attempt
	attempt := &models.TestAttempt{
		UserID:    session.UserID,
		TestID:    testID,
		StartedAt: time.Now(),
		Status:    "in_progress",
	}

	if err := h.attemptRepo.Create(r.Context(), attempt); err != nil {
		log.Printf("Error creating attempt: %v", err)
		http.Error(w, "Failed to start test", http.StatusInternalServerError)
		return
	}

	// Redirect to take test page
	http.Redirect(w, r, "/test/take?attempt_id="+strconv.Itoa(attempt.ID), http.StatusSeeOther)
}

// TakeTest displays the test questions
func (h *TestHandler) TakeTest(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)
	attemptIDStr := r.URL.Query().Get("attempt_id")
	attemptID, err := strconv.Atoi(attemptIDStr)
	if err != nil {
		http.Error(w, "Invalid attempt ID", http.StatusBadRequest)
		return
	}

	// Get attempt
	attempt, err := h.attemptRepo.GetByID(r.Context(), attemptID)
	if err != nil {
		log.Printf("Error fetching attempt: %v", err)
		http.Error(w, "Attempt not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if attempt.UserID != session.UserID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Check if already completed
	if attempt.Status == "completed" {
		http.Redirect(w, r, "/test/results?attempt_id="+attemptIDStr, http.StatusSeeOther)
		return
	}

	// Get test with questions
	test, err := h.testRepo.GetByID(r.Context(), attempt.TestID)
	if err != nil {
		log.Printf("Error fetching test: %v", err)
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}

	// Get existing answers
	answers, err := h.attemptRepo.GetAnswersByAttemptID(r.Context(), attemptID)
	if err != nil {
		log.Printf("Error fetching answers: %v", err)
	}

	// Create map of answered questions
	answeredMap := make(map[int]int) // questionID -> selectedOptionID
	for _, answer := range answers {
		if answer.SelectedOptionID != nil {
			answeredMap[answer.QuestionID] = *answer.SelectedOptionID
		}
	}

	// Hide correct answers from students (they shouldn't see this during the test)
	for i := range test.Questions {
		for j := range test.Questions[i].Options {
			test.Questions[i].Options[j].IsCorrect = false
		}
	}

	data := map[string]interface{}{
		"Session":   session,
		"Test":      test,
		"Attempt":   attempt,
		"Answered":  answeredMap,
		"TimeLimit": test.TimeLimitMinutes * 60, // Convert to seconds for JS timer
	}

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"mod": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a % b
		},
	}).ParseFiles("views/layout.html", "views/take_test.html")
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

// SubmitAnswer handles AJAX submission of individual answers
func (h *TestHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := auth.GetSessionData(r)

	var req struct {
		AttemptID  int `json:"attempt_id"`
		QuestionID int `json:"question_id"`
		OptionID   int `json:"option_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Verify attempt belongs to user
	attempt, err := h.attemptRepo.GetByID(r.Context(), req.AttemptID)
	if err != nil || attempt.UserID != session.UserID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Get the question to find correct answer
	test, err := h.testRepo.GetByID(r.Context(), attempt.TestID)
	if err != nil {
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}

	// Find the question and check if answer is correct
	var isCorrect bool
	for _, q := range test.Questions {
		if q.ID == req.QuestionID {
			for _, opt := range q.Options {
				if opt.ID == req.OptionID {
					isCorrect = opt.IsCorrect
					break
				}
			}
			break
		}
	}

	// Save the answer
	answer := &models.StudentAnswer{
		AttemptID:        req.AttemptID,
		QuestionID:       req.QuestionID,
		SelectedOptionID: &req.OptionID,
		IsCorrect:        &isCorrect,
	}

	if err := h.attemptRepo.SaveAnswer(r.Context(), answer); err != nil {
		log.Printf("Error saving answer: %v", err)
		http.Error(w, "Failed to save answer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Answer saved",
	})
}

// SubmitTest completes the test and calculates the score
func (h *TestHandler) SubmitTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := auth.GetSessionData(r)
	attemptIDStr := r.FormValue("attempt_id")
	attemptID, err := strconv.Atoi(attemptIDStr)
	if err != nil {
		http.Error(w, "Invalid attempt ID", http.StatusBadRequest)
		return
	}

	// Verify attempt belongs to user
	attempt, err := h.attemptRepo.GetByID(r.Context(), attemptID)
	if err != nil || attempt.UserID != session.UserID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Get all answers
	answers, err := h.attemptRepo.GetAnswersByAttemptID(r.Context(), attemptID)
	if err != nil {
		log.Printf("Error fetching answers: %v", err)
		http.Error(w, "Failed to calculate score", http.StatusInternalServerError)
		return
	}

	// Calculate score
	score := 0
	totalPoints := 0
	test, _ := h.testRepo.GetByID(r.Context(), attempt.TestID)

	for _, q := range test.Questions {
		totalPoints += q.Points
		for _, answer := range answers {
			if answer.QuestionID == q.ID && answer.IsCorrect != nil && *answer.IsCorrect {
				score += q.Points
			}
		}
	}

	// Complete the attempt
	if err := h.attemptRepo.Complete(r.Context(), attemptID, score, totalPoints); err != nil {
		log.Printf("Error completing attempt: %v", err)
		http.Error(w, "Failed to submit test", http.StatusInternalServerError)
		return
	}

	// Update user stats
	stats, err := h.userRepo.GetUserStats(r.Context(), session.UserID)
	if err != nil {
		// Initialize if doesn't exist
		h.userRepo.InitializeUserStats(r.Context(), session.UserID)
		stats, _ = h.userRepo.GetUserStats(r.Context(), session.UserID)
	}

	stats.TestsCompleted++
	stats.TotalPoints += score
	percentageScore := (float64(score) / float64(totalPoints)) * 100
	if percentageScore >= float64(test.PassingScore) {
		stats.TestsPassed++
	}

	h.userRepo.UpdateUserStats(r.Context(), stats)

	// Redirect to results
	http.Redirect(w, r, "/test/results?attempt_id="+attemptIDStr, http.StatusSeeOther)
}

// ViewResults displays test results
func (h *TestHandler) ViewResults(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)
	attemptIDStr := r.URL.Query().Get("attempt_id")
	attemptID, err := strconv.Atoi(attemptIDStr)
	if err != nil {
		http.Error(w, "Invalid attempt ID", http.StatusBadRequest)
		return
	}

	// Get attempt
	attempt, err := h.attemptRepo.GetByID(r.Context(), attemptID)
	if err != nil {
		log.Printf("Error fetching attempt: %v", err)
		http.Error(w, "Attempt not found", http.StatusNotFound)
		return
	}

	// Verify ownership or admin/teacher role
	if attempt.UserID != session.UserID && session.Role != "admin" && session.Role != "teacher" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Get test
	test, err := h.testRepo.GetByID(r.Context(), attempt.TestID)
	if err != nil {
		log.Printf("Error fetching test: %v", err)
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}

	// Get answers
	answers, err := h.attemptRepo.GetAnswersByAttemptID(r.Context(), attemptID)
	if err != nil {
		log.Printf("Error fetching answers: %v", err)
	}

	// Create map of answers
	answerMap := make(map[int]*models.StudentAnswer)
	for i := range answers {
		answerMap[answers[i].QuestionID] = &answers[i]
	}

	// Calculate percentage
	var percentage float64
	if attempt.TotalPoints != nil && *attempt.TotalPoints > 0 && attempt.Score != nil {
		percentage = (float64(*attempt.Score) / float64(*attempt.TotalPoints)) * 100
	}

	passed := percentage >= float64(test.PassingScore)

	data := map[string]interface{}{
		"Session":    session,
		"Test":       test,
		"Attempt":    attempt,
		"Answers":    answerMap,
		"Percentage": percentage,
		"Passed":     passed,
	}

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"mod": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a % b
		},
	}).ParseFiles("views/layout.html", "views/test_results.html")
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
