package handlers

import (
	"context"
	"html/template"
	"log"
	"net/http"

	"github.com/jchanning/gocase/internal/auth"
	"github.com/jchanning/gocase/internal/models"
)

type dashboardUserStore interface {
	GetUserStats(ctx context.Context, userID int) (*models.UserStats, error)
	GetUserAchievements(ctx context.Context, userID int) ([]models.UserAchievement, error)
}

type dashboardAttemptStore interface {
	GetUserAttempts(ctx context.Context, userID int, limit int) ([]models.TestAttempt, error)
	GetUserTestStats(ctx context.Context, userID int) (map[string]interface{}, error)
}

type dashboardAssignmentStore interface {
	GetByStudent(ctx context.Context, studentID int) ([]models.TestAssignment, error)
}

// DashboardHandler handles dashboard requests
type DashboardHandler struct {
	userRepo       dashboardUserStore
	attemptRepo    dashboardAttemptStore
	assignmentRepo dashboardAssignmentStore
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(userRepo dashboardUserStore, attemptRepo dashboardAttemptStore, assignmentRepo dashboardAssignmentStore) *DashboardHandler {
	return &DashboardHandler{
		userRepo:       userRepo,
		attemptRepo:    attemptRepo,
		assignmentRepo: assignmentRepo,
	}
}

func (h *DashboardHandler) loadDashboardData(ctx context.Context, session *auth.SessionData) map[string]interface{} {
	stats, err := h.userRepo.GetUserStats(ctx, session.UserID)
	if err != nil {
		log.Printf("Error fetching user stats: %v", err)
		stats = &models.UserStats{}
	}

	attempts, err := h.attemptRepo.GetUserAttempts(ctx, session.UserID, 10)
	if err != nil {
		log.Printf("Error fetching attempts: %v", err)
		attempts = []models.TestAttempt{}
	}

	achievements, err := h.userRepo.GetUserAchievements(ctx, session.UserID)
	if err != nil {
		log.Printf("Error fetching achievements: %v", err)
		achievements = []models.UserAchievement{}
	}

	testStats, err := h.attemptRepo.GetUserTestStats(ctx, session.UserID)
	if err != nil {
		log.Printf("Error fetching test stats: %v", err)
		testStats = make(map[string]interface{})
	}

	var assignments []models.TestAssignment
	var assignmentData interface{}
	if session.Role == "student" {
		assignments, err = h.assignmentRepo.GetByStudent(ctx, session.UserID)
		if err != nil {
			log.Printf("Error fetching assignments: %v", err)
			assignments = nil
		} else {
			assignmentData = assignments
		}
	}

	return map[string]interface{}{
		"Session":      session,
		"Stats":        stats,
		"Attempts":     attempts,
		"Achievements": achievements,
		"TestStats":    testStats,
		"Assignments":  assignmentData,
	}
}

// ShowDashboard displays the user dashboard
func (h *DashboardHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)
	data := h.loadDashboardData(r.Context(), session)

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
