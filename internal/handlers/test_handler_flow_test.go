package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jchanning/gocase/internal/auth"
	"github.com/jchanning/gocase/internal/models"
	"github.com/jchanning/gocase/internal/repository"
)

type flowTestCatalogStore struct {
	tests map[int]*models.Test
}

func (m *flowTestCatalogStore) GetAll(ctx context.Context) ([]models.Test, error) {
	return nil, nil
}

func (m *flowTestCatalogStore) GetSubjects(ctx context.Context) ([]models.Subject, error) {
	return nil, nil
}

func (m *flowTestCatalogStore) GetByID(ctx context.Context, id int) (*models.Test, error) {
	return m.tests[id], nil
}

func (m *flowTestCatalogStore) GetRecommendation(ctx context.Context, subjectID int, difficulty string, excludeTestID, userID int) (*models.Test, error) {
	return nil, nil
}

type flowTestAttemptStore struct {
	nextID     int
	attempts   map[int]*models.TestAttempt
	answers    map[int][]models.StudentAnswer
	lastAnswer *models.StudentAnswer
	streaks    repository.StreakStats
}

func (m *flowTestAttemptStore) SearchAttempts(ctx context.Context, filter repository.AttemptSearchFilter) ([]models.TestAttempt, error) {
	return nil, nil
}

func (m *flowTestAttemptStore) Create(ctx context.Context, attempt *models.TestAttempt) error {
	m.nextID++
	attempt.ID = m.nextID
	copyAttempt := *attempt
	m.attempts[attempt.ID] = &copyAttempt
	return nil
}

func (m *flowTestAttemptStore) GetByID(ctx context.Context, id int) (*models.TestAttempt, error) {
	return m.attempts[id], nil
}

func (m *flowTestAttemptStore) GetAnswersByAttemptID(ctx context.Context, attemptID int) ([]models.StudentAnswer, error) {
	return m.answers[attemptID], nil
}

func (m *flowTestAttemptStore) SaveAnswer(ctx context.Context, answer *models.StudentAnswer) error {
	copyAnswer := *answer
	m.lastAnswer = &copyAnswer
	m.answers[answer.AttemptID] = append(m.answers[answer.AttemptID], copyAnswer)
	return nil
}

func (m *flowTestAttemptStore) Complete(ctx context.Context, attemptID, score, totalPoints int) error {
	completedAt := time.Now()
	attempt := m.attempts[attemptID]
	attempt.CompletedAt = &completedAt
	attempt.Score = &score
	attempt.TotalPoints = &totalPoints
	attempt.Status = "completed"
	return nil
}

func (m *flowTestAttemptStore) GetUserStreakStats(ctx context.Context, userID int) (repository.StreakStats, error) {
	return m.streaks, nil
}

type flowTestUserStatsStore struct {
	stats        *models.UserStats
	updatedStats *models.UserStats
}

func (m *flowTestUserStatsStore) GetUserStats(ctx context.Context, userID int) (*models.UserStats, error) {
	return m.stats, nil
}

func (m *flowTestUserStatsStore) InitializeUserStats(ctx context.Context, userID int) error {
	m.stats = &models.UserStats{UserID: userID}
	return nil
}

func (m *flowTestUserStatsStore) UpdateUserStats(ctx context.Context, stats *models.UserStats) error {
	m.updatedStats = stats
	return nil
}

type flowTestAssignmentStore struct {
	studentID int
	testID    int
}

func (m *flowTestAssignmentStore) MarkCompleted(ctx context.Context, studentID, testID int) error {
	m.studentID = studentID
	m.testID = testID
	return nil
}

func TestStudentTestLifecycle_StartAnswerSubmit(t *testing.T) {
	subjectID := 3
	catalog := &flowTestCatalogStore{tests: map[int]*models.Test{
		8: {
			ID:           8,
			SubjectID:    &subjectID,
			Difficulty:   "Easy",
			PassingScore: 60,
			Questions: []models.Question{{
				ID:      101,
				Points:  5,
				Options: []models.AnswerOption{{ID: 201, IsCorrect: true}, {ID: 202, IsCorrect: false}},
			}},
		},
	}}
	attemptStore := &flowTestAttemptStore{attempts: map[int]*models.TestAttempt{}, answers: map[int][]models.StudentAnswer{}, streaks: repository.StreakStats{Current: 2, Best: 4}}
	userStatsStore := &flowTestUserStatsStore{stats: &models.UserStats{UserID: 21}}
	assignmentStore := &flowTestAssignmentStore{}
	handler := NewTestHandler(catalog, attemptStore, userStatsStore, assignmentStore)

	store := auth.NewSessionStore()
	token, _ := store.Create(21, "student", "student")
	mw := auth.NewMiddleware(store)

	startReq := httptest.NewRequest(http.MethodGet, "/test/start?id=8", nil)
	startReq.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	startRR := httptest.NewRecorder()
	mw.RequireAuth(http.HandlerFunc(handler.StartTest)).ServeHTTP(startRR, startReq)

	if startRR.Code != http.StatusSeeOther {
		t.Fatalf("expected start redirect, got %d", startRR.Code)
	}
	if location := startRR.Header().Get("Location"); location != "/test/take?attempt_id=1" {
		t.Fatalf("expected redirect to attempt 1, got %s", location)
	}
	if attemptStore.attempts[1] == nil || attemptStore.attempts[1].UserID != 21 || attemptStore.attempts[1].TestID != 8 {
		t.Fatalf("expected created attempt for student 21 and test 8, got %#v", attemptStore.attempts[1])
	}

	answerReq := httptest.NewRequest(http.MethodPost, "/test/answer", strings.NewReader(`{"attempt_id":1,"question_id":101,"option_id":201}`))
	answerReq.Header.Set("Content-Type", "application/json")
	answerReq.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	answerRR := httptest.NewRecorder()
	mw.RequireAuth(http.HandlerFunc(handler.SubmitAnswer)).ServeHTTP(answerRR, answerReq)

	if answerRR.Code != http.StatusOK {
		t.Fatalf("expected successful answer submission, got %d", answerRR.Code)
	}
	if attemptStore.lastAnswer == nil || attemptStore.lastAnswer.IsCorrect == nil || !*attemptStore.lastAnswer.IsCorrect {
		t.Fatalf("expected saved correct answer, got %#v", attemptStore.lastAnswer)
	}

	submitReq := httptest.NewRequest(http.MethodPost, "/test/submit", strings.NewReader("attempt_id=1"))
	submitReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	submitReq.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	submitRR := httptest.NewRecorder()
	mw.RequireAuth(http.HandlerFunc(handler.SubmitTest)).ServeHTTP(submitRR, submitReq)

	if submitRR.Code != http.StatusSeeOther {
		t.Fatalf("expected submit redirect, got %d", submitRR.Code)
	}
	if location := submitRR.Header().Get("Location"); location != "/test/results?attempt_id=1" {
		t.Fatalf("expected redirect to results, got %s", location)
	}
	if attemptStore.attempts[1].Status != "completed" {
		t.Fatalf("expected attempt to be completed, got %#v", attemptStore.attempts[1])
	}
	if userStatsStore.updatedStats == nil || userStatsStore.updatedStats.TestsCompleted != 1 || userStatsStore.updatedStats.TestsPassed != 1 {
		t.Fatalf("expected updated passing stats, got %#v", userStatsStore.updatedStats)
	}
	if assignmentStore.studentID != 21 || assignmentStore.testID != 8 {
		t.Fatalf("expected assignment completion for student 21 test 8, got student=%d test=%d", assignmentStore.studentID, assignmentStore.testID)
	}
}

func TestSubmitAnswer_RejectsOtherUsersAttempt(t *testing.T) {
	handler := NewTestHandler(
		&flowTestCatalogStore{tests: map[int]*models.Test{8: {ID: 8}}},
		&flowTestAttemptStore{attempts: map[int]*models.TestAttempt{1: {ID: 1, UserID: 99, TestID: 8}}, answers: map[int][]models.StudentAnswer{}},
		&flowTestUserStatsStore{stats: &models.UserStats{UserID: 21}},
		&flowTestAssignmentStore{},
	)

	store := auth.NewSessionStore()
	token, _ := store.Create(21, "student", "student")
	mw := auth.NewMiddleware(store)

	req := httptest.NewRequest(http.MethodPost, "/test/answer", strings.NewReader(`{"attempt_id":1,"question_id":101,"option_id":201}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rr := httptest.NewRecorder()

	mw.RequireAuth(http.HandlerFunc(handler.SubmitAnswer)).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden for another user's attempt, got %d", rr.Code)
	}
	var _ map[string]any
	_ = json.Valid(rr.Body.Bytes())
}
