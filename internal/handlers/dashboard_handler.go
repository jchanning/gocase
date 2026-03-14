package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/jchanning/gocase/internal/auth"
	"github.com/jchanning/gocase/internal/models"
	"github.com/jchanning/gocase/internal/repository"
)

// DashboardHandler handles dashboard requests
type DashboardHandler struct {
	userRepo       *repository.UserRepository
	testRepo       *repository.TestRepository
	attemptRepo    *repository.AttemptRepository
	assignmentRepo *repository.AssignmentRepository
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(userRepo *repository.UserRepository, testRepo *repository.TestRepository, attemptRepo *repository.AttemptRepository, assignmentRepo *repository.AssignmentRepository) *DashboardHandler {
	return &DashboardHandler{
		userRepo:       userRepo,
		testRepo:       testRepo,
		attemptRepo:    attemptRepo,
		assignmentRepo: assignmentRepo,
	}
}

// ShowDashboard displays the user dashboard
func (h *DashboardHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)

	// Get user stats
	stats, err := h.userRepo.GetUserStats(r.Context(), session.UserID)
	if err != nil {
		log.Printf("Error fetching user stats: %v", err)
		stats = &models.UserStats{} // Use empty stats
	}

	// Get recent attempts
	attempts, err := h.attemptRepo.GetUserAttempts(r.Context(), session.UserID, 10)
	if err != nil {
		log.Printf("Error fetching attempts: %v", err)
		attempts = []models.TestAttempt{}
	}

	// Get user achievements
	achievements, err := h.userRepo.GetUserAchievements(r.Context(), session.UserID)
	if err != nil {
		log.Printf("Error fetching achievements: %v", err)
		achievements = []models.UserAchievement{}
	}

	// Get test statistics
	testStats, err := h.attemptRepo.GetUserTestStats(r.Context(), session.UserID)
	if err != nil {
		log.Printf("Error fetching test stats: %v", err)
		testStats = make(map[string]interface{})
	}

	// Get assigned tests for students
	var assignments []models.TestAssignment
	if session.Role == "student" {
		assignments, err = h.assignmentRepo.GetByStudent(r.Context(), session.UserID)
		if err != nil {
			log.Printf("Error fetching assignments: %v", err)
			assignments = nil
		}
	}

	data := map[string]interface{}{
		"Session":      session,
		"Stats":        stats,
		"Attempts":     attempts,
		"Achievements": achievements,
		"TestStats":    testStats,
		"Assignments":  assignments,
	}

	tmpl, err := template.ParseFiles("views/layout.html", "views/dashboard.html")
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
