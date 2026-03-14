package handlers

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/jchanning/gocase/internal/auth"
	"github.com/jchanning/gocase/internal/models"
	"github.com/jchanning/gocase/internal/repository"
)

type mockTestCatalogStore struct {
	recommendation               *models.Test
	recommendationErr            error
	lastRecommendationID         int
	lastRecommendationDifficulty string
}

func (m *mockTestCatalogStore) GetAll(ctx context.Context) ([]models.Test, error) {
	return nil, nil
}

func (m *mockTestCatalogStore) GetSubjects(ctx context.Context) ([]models.Subject, error) {
	return nil, nil
}

func (m *mockTestCatalogStore) GetByID(ctx context.Context, id int) (*models.Test, error) {
	return nil, nil
}

func (m *mockTestCatalogStore) GetRecommendation(ctx context.Context, subjectID int, difficulty string, excludeTestID, userID int) (*models.Test, error) {
	m.lastRecommendationID = subjectID
	m.lastRecommendationDifficulty = difficulty
	if m.recommendationErr != nil {
		return nil, m.recommendationErr
	}
	return m.recommendation, nil
}

type mockTestAttemptStore struct {
	streaks          repository.StreakStats
	streaksErr       error
	searchCalls      int
	lastSearchFilter repository.AttemptSearchFilter
}

func (m *mockTestAttemptStore) SearchAttempts(ctx context.Context, filter repository.AttemptSearchFilter) ([]models.TestAttempt, error) {
	m.searchCalls++
	m.lastSearchFilter = filter
	return nil, nil
}

func (m *mockTestAttemptStore) Create(ctx context.Context, attempt *models.TestAttempt) error {
	return nil
}

func (m *mockTestAttemptStore) GetByID(ctx context.Context, id int) (*models.TestAttempt, error) {
	return nil, nil
}

func (m *mockTestAttemptStore) GetAnswersByAttemptID(ctx context.Context, attemptID int) ([]models.StudentAnswer, error) {
	return nil, nil
}

func (m *mockTestAttemptStore) SaveAnswer(ctx context.Context, answer *models.StudentAnswer) error {
	return nil
}

func (m *mockTestAttemptStore) Complete(ctx context.Context, attemptID, score, totalPoints int) error {
	return nil
}

func (m *mockTestAttemptStore) GetUserStreakStats(ctx context.Context, userID int) (repository.StreakStats, error) {
	if m.streaksErr != nil {
		return repository.StreakStats{}, m.streaksErr
	}
	return m.streaks, nil
}

type mockTestUserStatsStore struct {
	stats             *models.UserStats
	getErr            error
	updatedStats      *models.UserStats
	initializedUserID int
}

func (m *mockTestUserStatsStore) GetUserStats(ctx context.Context, userID int) (*models.UserStats, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.stats, nil
}

func (m *mockTestUserStatsStore) InitializeUserStats(ctx context.Context, userID int) error {
	m.initializedUserID = userID
	if m.stats == nil {
		m.stats = &models.UserStats{UserID: userID}
	}
	return nil
}

func (m *mockTestUserStatsStore) UpdateUserStats(ctx context.Context, stats *models.UserStats) error {
	m.updatedStats = stats
	return nil
}

type mockTestAssignmentStore struct {
	completedStudentID int
	completedTestID    int
}

func (m *mockTestAssignmentStore) MarkCompleted(ctx context.Context, studentID, testID int) error {
	m.completedStudentID = studentID
	m.completedTestID = testID
	return nil
}

func TestBuildHistoryFilters_LimitsStudentsToOwnAttempts(t *testing.T) {
	req := httptest.NewRequest("GET", "/history?student=someone&test=algebra", nil)
	filters := buildHistoryFilters(req, &auth.SessionData{UserID: 42, Role: "student"})

	if filters.UserID == nil || *filters.UserID != 42 {
		t.Fatalf("expected student history to be restricted to current user, got %#v", filters.UserID)
	}
	if filters.AttemptSearchFilter.UserID == nil || *filters.AttemptSearchFilter.UserID != 42 {
		t.Fatalf("expected search filter to carry current user ID, got %#v", filters.AttemptSearchFilter.UserID)
	}
}

func TestBuildHistoryFilters_DoesNotRestrictTeachers(t *testing.T) {
	req := httptest.NewRequest("GET", "/history?student=someone&test=algebra", nil)
	filters := buildHistoryFilters(req, &auth.SessionData{UserID: 77, Role: "teacher"})

	if filters.UserID != nil {
		t.Fatalf("expected teacher history to remain unrestricted, got %#v", filters.UserID)
	}
	if filters.AttemptSearchFilter.UserID != nil {
		t.Fatalf("expected teacher search filter to remain unrestricted, got %#v", filters.AttemptSearchFilter.UserID)
	}
}

func TestCalculateAttemptScore(t *testing.T) {
	correct := true
	incorrect := false
	test := &models.Test{Questions: []models.Question{{ID: 1, Points: 2}, {ID: 2, Points: 3}}}
	answers := []models.StudentAnswer{{QuestionID: 1, IsCorrect: &correct}, {QuestionID: 2, IsCorrect: &incorrect}}

	score, totalPoints := calculateAttemptScore(test, answers)
	if score != 2 || totalPoints != 5 {
		t.Fatalf("expected score=2 totalPoints=5, got score=%d totalPoints=%d", score, totalPoints)
	}
}

func TestUpdateUserStatsAfterSubmission_InitializesMissingStats(t *testing.T) {
	userStore := &mockTestUserStatsStore{getErr: errors.New("not found")}
	attemptStore := &mockTestAttemptStore{streaks: repository.StreakStats{Current: 3, Best: 5}}
	handler := NewTestHandler(&mockTestCatalogStore{}, attemptStore, userStore, &mockTestAssignmentStore{})

	handler.updateUserStatsAfterSubmission(context.Background(), &auth.SessionData{UserID: 12}, &models.Test{PassingScore: 60}, 8, 10)

	if userStore.initializedUserID != 12 {
		t.Fatalf("expected stats initialization for user 12, got %d", userStore.initializedUserID)
	}
	if userStore.updatedStats == nil {
		t.Fatal("expected updated stats to be persisted")
	}
	if userStore.updatedStats.TestsCompleted != 1 || userStore.updatedStats.TestsPassed != 1 {
		t.Fatalf("unexpected stats update: %#v", userStore.updatedStats)
	}
	if userStore.updatedStats.TotalPoints != 8 {
		t.Fatalf("expected total points 8, got %d", userStore.updatedStats.TotalPoints)
	}
	if userStore.updatedStats.CurrentStreak != 3 || userStore.updatedStats.BestStreak != 5 {
		t.Fatalf("expected streaks 3/5, got %d/%d", userStore.updatedStats.CurrentStreak, userStore.updatedStats.BestStreak)
	}
}

func TestBuildRecommendation_OnlyForPassedTestsWithNextDifficulty(t *testing.T) {
	subjectID := 3
	recommended := &models.Test{ID: 91, Title: "Next Test"}
	catalog := &mockTestCatalogStore{recommendation: recommended}
	handler := NewTestHandler(catalog, &mockTestAttemptStore{}, &mockTestUserStatsStore{}, &mockTestAssignmentStore{})

	rec := handler.buildRecommendation(context.Background(), &auth.SessionData{UserID: 5}, &models.Test{ID: 10, SubjectID: &subjectID, Difficulty: "Easy"}, true)
	if rec == nil || rec.ID != 91 {
		t.Fatalf("expected recommendation to be returned, got %#v", rec)
	}
	if catalog.lastRecommendationDifficulty != "Medium" {
		t.Fatalf("expected next difficulty Medium, got %s", catalog.lastRecommendationDifficulty)
	}

	if handler.buildRecommendation(context.Background(), &auth.SessionData{UserID: 5}, &models.Test{ID: 10, SubjectID: &subjectID, Difficulty: "Hard"}, true) != nil {
		t.Fatal("expected no recommendation for highest difficulty")
	}
	if handler.buildRecommendation(context.Background(), &auth.SessionData{UserID: 5}, &models.Test{ID: 10, SubjectID: &subjectID, Difficulty: "Easy"}, false) != nil {
		t.Fatal("expected no recommendation when test was not passed")
	}
}
