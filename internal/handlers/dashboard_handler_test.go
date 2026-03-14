package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/jchanning/gocase/internal/auth"
	"github.com/jchanning/gocase/internal/models"
)

type mockDashboardUserStore struct {
	stats            *models.UserStats
	statsErr         error
	achievements     []models.UserAchievement
	achievementsErr  error
	getStatsCalls    int
	getAchievesCalls int
}

func (m *mockDashboardUserStore) GetUserStats(ctx context.Context, userID int) (*models.UserStats, error) {
	m.getStatsCalls++
	if m.statsErr != nil {
		return nil, m.statsErr
	}
	if m.stats == nil {
		return &models.UserStats{}, nil
	}
	return m.stats, nil
}

func (m *mockDashboardUserStore) GetUserAchievements(ctx context.Context, userID int) ([]models.UserAchievement, error) {
	m.getAchievesCalls++
	if m.achievementsErr != nil {
		return nil, m.achievementsErr
	}
	return m.achievements, nil
}

type mockDashboardAttemptStore struct {
	attempts          []models.TestAttempt
	attemptsErr       error
	testStats         map[string]interface{}
	testStatsErr      error
	getAttemptsCalls  int
	getTestStatsCalls int
}

func (m *mockDashboardAttemptStore) GetUserAttempts(ctx context.Context, userID int, limit int) ([]models.TestAttempt, error) {
	m.getAttemptsCalls++
	if m.attemptsErr != nil {
		return nil, m.attemptsErr
	}
	return m.attempts, nil
}

func (m *mockDashboardAttemptStore) GetUserTestStats(ctx context.Context, userID int) (map[string]interface{}, error) {
	m.getTestStatsCalls++
	if m.testStatsErr != nil {
		return nil, m.testStatsErr
	}
	return m.testStats, nil
}

type mockDashboardAssignmentStore struct {
	assignments       []models.TestAssignment
	assignmentsErr    error
	getByStudentCalls int
}

func (m *mockDashboardAssignmentStore) GetByStudent(ctx context.Context, studentID int) ([]models.TestAssignment, error) {
	m.getByStudentCalls++
	if m.assignmentsErr != nil {
		return nil, m.assignmentsErr
	}
	return m.assignments, nil
}

func TestDashboardLoadData_IncludesAssignmentsForStudents(t *testing.T) {
	userStore := &mockDashboardUserStore{
		stats:        &models.UserStats{UserID: 7, TestsCompleted: 3},
		achievements: []models.UserAchievement{{ID: 1}},
	}
	attemptStore := &mockDashboardAttemptStore{
		attempts:  []models.TestAttempt{{ID: 10}},
		testStats: map[string]interface{}{"average": 85},
	}
	assignmentStore := &mockDashboardAssignmentStore{
		assignments: []models.TestAssignment{{ID: 20, Status: "pending"}},
	}

	handler := NewDashboardHandler(userStore, attemptStore, assignmentStore)
	data := handler.loadDashboardData(context.Background(), &auth.SessionData{UserID: 7, Role: "student"})

	assignments, ok := data["Assignments"].([]models.TestAssignment)
	if !ok {
		t.Fatal("expected Assignments to be a []models.TestAssignment")
	}
	if len(assignments) != 1 || assignments[0].ID != 20 {
		t.Fatalf("expected student assignments to be loaded, got %#v", assignments)
	}
	if assignmentStore.getByStudentCalls != 1 {
		t.Fatalf("expected assignment lookup once, got %d", assignmentStore.getByStudentCalls)
	}
}

func TestDashboardLoadData_SkipsAssignmentsForTeachers(t *testing.T) {
	handler := NewDashboardHandler(
		&mockDashboardUserStore{},
		&mockDashboardAttemptStore{testStats: map[string]interface{}{}},
		&mockDashboardAssignmentStore{},
	)

	data := handler.loadDashboardData(context.Background(), &auth.SessionData{UserID: 8, Role: "teacher"})

	if data["Assignments"] != nil {
		t.Fatalf("expected no assignments for teacher dashboard, got %#v", data["Assignments"])
	}
}

func TestDashboardLoadData_FallsBackOnRepositoryErrors(t *testing.T) {
	assignmentStore := &mockDashboardAssignmentStore{assignmentsErr: errors.New("db down")}
	handler := NewDashboardHandler(
		&mockDashboardUserStore{statsErr: errors.New("stats down"), achievementsErr: errors.New("achievements down")},
		&mockDashboardAttemptStore{attemptsErr: errors.New("attempts down"), testStatsErr: errors.New("test stats down")},
		assignmentStore,
	)

	data := handler.loadDashboardData(context.Background(), &auth.SessionData{UserID: 9, Role: "student"})

	stats, ok := data["Stats"].(*models.UserStats)
	if !ok || stats == nil {
		t.Fatal("expected fallback stats value")
	}
	attempts, ok := data["Attempts"].([]models.TestAttempt)
	if !ok || len(attempts) != 0 {
		t.Fatalf("expected empty attempts fallback, got %#v", data["Attempts"])
	}
	achievements, ok := data["Achievements"].([]models.UserAchievement)
	if !ok || len(achievements) != 0 {
		t.Fatalf("expected empty achievements fallback, got %#v", data["Achievements"])
	}
	testStats, ok := data["TestStats"].(map[string]interface{})
	if !ok || len(testStats) != 0 {
		t.Fatalf("expected empty test stats fallback, got %#v", data["TestStats"])
	}
	if data["Assignments"] != nil {
		t.Fatalf("expected nil assignments on fetch error, got %#v", data["Assignments"])
	}
	if assignmentStore.getByStudentCalls != 1 {
		t.Fatalf("expected assignment lookup once, got %d", assignmentStore.getByStudentCalls)
	}
}
