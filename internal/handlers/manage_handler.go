package handlers

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/jchanning/gocase/internal/auth"
	"github.com/jchanning/gocase/internal/models"
)

type manageTestStore interface {
	GetAll(ctx context.Context) ([]models.Test, error)
	GetByCreator(ctx context.Context, userID int) ([]models.Test, error)
	GetByID(ctx context.Context, id int) (*models.Test, error)
	GetSubjects(ctx context.Context) ([]models.Subject, error)
	SubmitForReview(ctx context.Context, testID, actorID int) error
	ApproveReview(ctx context.Context, testID, reviewerID int, notes string) error
	RequestChanges(ctx context.Context, testID, reviewerID int, notes string) error
}

type manageFeedbackStore interface {
	ListIssues(ctx context.Context, status string) ([]models.TestFeedbackIssue, error)
	UpdateIssueReview(ctx context.Context, issueID int, status, response string, reviewerID int) error
}

// ManageHandler renders the combined staff create/manage experience and review workflows.
type ManageHandler struct {
	testRepo     manageTestStore
	feedbackRepo manageFeedbackStore
}

func NewManageHandler(testRepo manageTestStore, feedbackRepo manageFeedbackStore) *ManageHandler {
	return &ManageHandler{testRepo: testRepo, feedbackRepo: feedbackRepo}
}

func (h *ManageHandler) canManageTest(session *auth.SessionData, test *models.Test) bool {
	if session == nil || test == nil {
		return false
	}
	if session.Role == "admin" {
		return true
	}
	return test.CreatedBy != nil && *test.CreatedBy == session.UserID
}

func (h *ManageHandler) ShowManage(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var (
		tests []models.Test
		err   error
	)
	if session.Role == "admin" {
		tests, err = h.testRepo.GetAll(r.Context())
	} else {
		tests, err = h.testRepo.GetByCreator(r.Context(), session.UserID)
	}
	if err != nil {
		log.Printf("Error fetching manage tests: %v", err)
		tests = []models.Test{}
	}

	subjects, err := h.testRepo.GetSubjects(r.Context())
	if err != nil {
		log.Printf("Error fetching subjects: %v", err)
		subjects = []models.Subject{}
	}

	feedbackStatus := r.URL.Query().Get("feedback_status")
	issues := []models.TestFeedbackIssue{}
	if h.feedbackRepo != nil {
		issues, err = h.feedbackRepo.ListIssues(r.Context(), feedbackStatus)
		if err != nil {
			log.Printf("Error fetching feedback issues: %v", err)
			issues = []models.TestFeedbackIssue{}
		}
	}

	pendingReview := []models.Test{}
	for _, test := range tests {
		if test.ReviewStatus == "pending_review" {
			pendingReview = append(pendingReview, test)
		}
	}

	data := map[string]interface{}{
		"Session":         session,
		"Tests":           tests,
		"PendingReview":   pendingReview,
		"FeedbackIssues":  issues,
		"FeedbackStatus":  feedbackStatus,
		"Subjects":        subjects,
		"Difficulties":    models.ValidDifficulties,
		"ExamStandards":   models.ValidExamStandards,
		"FeedbackUpdated": r.URL.Query().Get("message") == "feedback-updated",
		"ReviewUpdated":   r.URL.Query().Get("message"),
	}

	tmpl, err := template.ParseFiles("views/layout.html", "views/manage.html")
	if err != nil {
		log.Printf("Error parsing manage templates: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("Error executing manage template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *ManageHandler) SubmitForReview(w http.ResponseWriter, r *http.Request) {
	h.handleTestReviewAction(w, r, func(testID int, session *auth.SessionData, notes string) error {
		return h.testRepo.SubmitForReview(r.Context(), testID, session.UserID)
	}, "review-submitted")
}

func (h *ManageHandler) ApproveTest(w http.ResponseWriter, r *http.Request) {
	h.handleTestReviewAction(w, r, func(testID int, session *auth.SessionData, notes string) error {
		return h.testRepo.ApproveReview(r.Context(), testID, session.UserID, notes)
	}, "review-approved")
}

func (h *ManageHandler) RequestChanges(w http.ResponseWriter, r *http.Request) {
	h.handleTestReviewAction(w, r, func(testID int, session *auth.SessionData, notes string) error {
		return h.testRepo.RequestChanges(r.Context(), testID, session.UserID, notes)
	}, "review-changes-requested")
}

func (h *ManageHandler) handleTestReviewAction(w http.ResponseWriter, r *http.Request, action func(testID int, session *auth.SessionData, notes string) error, message string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	session := auth.GetSessionData(r)
	testID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}
	test, err := h.testRepo.GetByID(r.Context(), testID)
	if err != nil {
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}
	if !h.canManageTest(session, test) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	if err := action(testID, session, r.FormValue("review_notes")); err != nil {
		log.Printf("Error applying review action: %v", err)
		http.Error(w, "Failed to update review", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/manage?message="+message, http.StatusSeeOther)
}

func (h *ManageHandler) UpdateFeedbackIssue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	session := auth.GetSessionData(r)
	if session == nil || (session.Role != "teacher" && session.Role != "admin") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	issueID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid issue ID", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	status := r.FormValue("status")
	validStatuses := map[string]bool{"open": true, "in_review": true, "resolved": true, "dismissed": true}
	if !validStatuses[status] {
		http.Error(w, "Invalid feedback status", http.StatusBadRequest)
		return
	}
	if err := h.feedbackRepo.UpdateIssueReview(r.Context(), issueID, status, r.FormValue("review_response"), session.UserID); err != nil {
		log.Printf("Error updating feedback issue: %v", err)
		http.Error(w, "Failed to update issue", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/manage?message=feedback-updated", http.StatusSeeOther)
}
