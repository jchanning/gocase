package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jchanning/gocase/internal/auth"
	"github.com/jchanning/gocase/internal/models"
	"github.com/jchanning/gocase/internal/repository"
)

type mockTeacherTestStore struct {
	test          *models.Test
	publishedID   int
	unpublishedID int
	deletedID     int
	publishErr    error
}

func (m *mockTeacherTestStore) GetOrCreateSubject(ctx context.Context, name, description string) (int, error) {
	return 0, nil
}

func (m *mockTeacherTestStore) GetOrCreateTopic(ctx context.Context, subjectID int, name, description string) (int, error) {
	return 0, nil
}

func (m *mockTeacherTestStore) Create(ctx context.Context, test *models.Test) error { return nil }
func (m *mockTeacherTestStore) UpdateTestNotes(ctx context.Context, testID int, notesFilename *string) error {
	return nil
}
func (m *mockTeacherTestStore) CreateQuestion(ctx context.Context, question *models.Question) error {
	return nil
}
func (m *mockTeacherTestStore) CreateAnswerOption(ctx context.Context, option *models.AnswerOption) error {
	return nil
}
func (m *mockTeacherTestStore) GetByCreator(ctx context.Context, userID int) ([]models.Test, error) {
	return nil, nil
}
func (m *mockTeacherTestStore) GetSubjects(ctx context.Context) ([]models.Subject, error) {
	return nil, nil
}
func (m *mockTeacherTestStore) GetByID(ctx context.Context, id int) (*models.Test, error) {
	return m.test, nil
}
func (m *mockTeacherTestStore) Update(ctx context.Context, test *models.Test) error { return nil }
func (m *mockTeacherTestStore) UpdateQuestion(ctx context.Context, question *models.Question) error {
	return nil
}
func (m *mockTeacherTestStore) UpdateAnswerOption(ctx context.Context, option *models.AnswerOption) error {
	return nil
}
func (m *mockTeacherTestStore) PublishTest(ctx context.Context, testID int) error {
	m.publishedID = testID
	return m.publishErr
}
func (m *mockTeacherTestStore) UnpublishTest(ctx context.Context, testID int) error {
	m.unpublishedID = testID
	return nil
}
func (m *mockTeacherTestStore) DeleteTest(ctx context.Context, testID int) error {
	m.deletedID = testID
	return nil
}

type mockTeacherUserStore struct {
	students []models.User
}

func (m *mockTeacherUserStore) GetUsersByRole(ctx context.Context, role string) ([]models.User, error) {
	return m.students, nil
}

type mockTeacherAttemptStore struct{}

func (m *mockTeacherAttemptStore) GetByTestID(ctx context.Context, testID int) ([]models.TestAttempt, error) {
	return nil, nil
}

type mockTeacherAssignmentStore struct {
	created []models.TestAssignment
}

func (m *mockTeacherAssignmentStore) GetByTeacher(ctx context.Context, teacherID int) ([]models.TestAssignment, error) {
	return nil, nil
}

func (m *mockTeacherAssignmentStore) Create(ctx context.Context, a *models.TestAssignment) error {
	m.created = append(m.created, *a)
	return nil
}

type mockAdminTestStore struct {
	createdSubjectID int
}

func (m *mockAdminTestStore) GetOrCreateSubject(ctx context.Context, name, description string) (int, error) {
	return m.createdSubjectID, nil
}
func (m *mockAdminTestStore) GetOrCreateTopic(ctx context.Context, subjectID int, name, description string) (int, error) {
	return 0, nil
}
func (m *mockAdminTestStore) Create(ctx context.Context, test *models.Test) error { return nil }
func (m *mockAdminTestStore) UpdateTestNotes(ctx context.Context, testID int, notesFilename *string) error {
	return nil
}
func (m *mockAdminTestStore) CreateQuestion(ctx context.Context, question *models.Question) error {
	return nil
}
func (m *mockAdminTestStore) CreateAnswerOption(ctx context.Context, option *models.AnswerOption) error {
	return nil
}
func (m *mockAdminTestStore) GetAll(ctx context.Context) ([]models.Test, error) { return nil, nil }
func (m *mockAdminTestStore) GetSubjects(ctx context.Context) ([]models.Subject, error) {
	return nil, nil
}
func (m *mockAdminTestStore) GetByID(ctx context.Context, id int) (*models.Test, error) {
	return nil, nil
}
func (m *mockAdminTestStore) DeleteTest(ctx context.Context, testID int) error    { return nil }
func (m *mockAdminTestStore) Update(ctx context.Context, test *models.Test) error { return nil }
func (m *mockAdminTestStore) UpdateQuestion(ctx context.Context, question *models.Question) error {
	return nil
}
func (m *mockAdminTestStore) UpdateAnswerOption(ctx context.Context, option *models.AnswerOption) error {
	return nil
}

type mockAdminUserStore struct {
	createdUser       *models.User
	initializedUserID int
	updatedRoleUserID int
	updatedRole       string
}

func (m *mockAdminUserStore) GetAllUsers(ctx context.Context) ([]models.User, error) { return nil, nil }
func (m *mockAdminUserStore) Create(ctx context.Context, user *models.User) error {
	m.createdUser = user
	user.ID = 44
	return nil
}
func (m *mockAdminUserStore) InitializeUserStats(ctx context.Context, userID int) error {
	m.initializedUserID = userID
	return nil
}
func (m *mockAdminUserStore) UpdateUserRole(ctx context.Context, userID int, role string) error {
	m.updatedRoleUserID = userID
	m.updatedRole = role
	return nil
}
func (m *mockAdminUserStore) UpdatePasswordHash(ctx context.Context, userID int, newHash string) error {
	return nil
}
func (m *mockAdminUserStore) DeleteUser(ctx context.Context, userID int) error { return nil }

func TestTeacherPublishTest_RequiresOwnership(t *testing.T) {
	ownerID := 99
	store := auth.NewSessionStore()
	token, _ := store.Create(7, "teacher", "teacher")
	mw := auth.NewMiddleware(store)

	handler := NewTeacherHandler(
		&mockTeacherTestStore{test: &models.Test{ID: 5, CreatedBy: &ownerID}},
		&mockTeacherUserStore{},
		&mockTeacherAttemptStore{},
		&mockTeacherAssignmentStore{},
	)

	req := httptest.NewRequest(http.MethodPost, "/teacher/test/5/publish", nil)
	req.SetPathValue("id", "5")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rr := httptest.NewRecorder()

	mw.RequireAuth(http.HandlerFunc(handler.PublishTest)).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden for non-owner publish attempt, got %d", rr.Code)
	}
}

func TestTeacherPublishAndUnpublish_SucceedsForOwner(t *testing.T) {
	ownerID := 7
	store := auth.NewSessionStore()
	token, _ := store.Create(ownerID, "teacher", "teacher")
	mw := auth.NewMiddleware(store)
	testStore := &mockTeacherTestStore{test: &models.Test{ID: 5, CreatedBy: &ownerID}}
	handler := NewTeacherHandler(testStore, &mockTeacherUserStore{}, &mockTeacherAttemptStore{}, &mockTeacherAssignmentStore{})

	for _, tc := range []struct {
		name     string
		path     string
		handler  http.HandlerFunc
		expected string
	}{
		{name: "publish", path: "/teacher/test/5/publish", handler: handler.PublishTest, expected: "Test published successfully"},
		{name: "unpublish", path: "/teacher/test/5/unpublish", handler: handler.UnpublishTest, expected: "Test unpublished successfully"},
	} {
		req := httptest.NewRequest(http.MethodPost, tc.path, nil)
		req.SetPathValue("id", "5")
		req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
		rr := httptest.NewRecorder()

		mw.RequireAuth(http.HandlerFunc(tc.handler)).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", tc.name, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), tc.expected) {
			t.Fatalf("%s: expected response to contain %q, got %q", tc.name, tc.expected, rr.Body.String())
		}
	}

	if testStore.publishedID != 5 || testStore.unpublishedID != 5 {
		t.Fatalf("expected publish and unpublish to target test 5, got publish=%d unpublish=%d", testStore.publishedID, testStore.unpublishedID)
	}
}

func TestTeacherPublishTest_RequiresApprovedReview(t *testing.T) {
	ownerID := 7
	store := auth.NewSessionStore()
	token, _ := store.Create(ownerID, "teacher", "teacher")
	mw := auth.NewMiddleware(store)
	testStore := &mockTeacherTestStore{
		test:       &models.Test{ID: 5, CreatedBy: &ownerID, ReviewStatus: "draft"},
		publishErr: repository.ErrReviewApprovalRequired,
	}
	handler := NewTeacherHandler(testStore, &mockTeacherUserStore{}, &mockTeacherAttemptStore{}, &mockTeacherAssignmentStore{})

	req := httptest.NewRequest(http.MethodPost, "/teacher/test/5/publish", nil)
	req.SetPathValue("id", "5")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rr := httptest.NewRecorder()

	mw.RequireAuth(http.HandlerFunc(handler.PublishTest)).ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected conflict when review approval missing, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "approved before it can be published") {
		t.Fatalf("expected review gate message, got %q", rr.Body.String())
	}
}

func TestTeacherAssignTest_ValidatesDueDateAndCreatesAssignments(t *testing.T) {
	store := auth.NewSessionStore()
	token, _ := store.Create(7, "teacher", "teacher")
	mw := auth.NewMiddleware(store)
	assignmentStore := &mockTeacherAssignmentStore{}
	handler := NewTeacherHandler(&mockTeacherTestStore{}, &mockTeacherUserStore{}, &mockTeacherAttemptStore{}, assignmentStore)

	invalidReq := httptest.NewRequest(http.MethodPost, "/teacher/test/8/assign", strings.NewReader(url.Values{"student_ids": {"1"}}.Encode()))
	invalidReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	invalidReq.SetPathValue("id", "8")
	invalidReq.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	invalidRR := httptest.NewRecorder()
	mw.RequireAuth(http.HandlerFunc(handler.AssignTest)).ServeHTTP(invalidRR, invalidReq)
	if invalidRR.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request when due date missing, got %d", invalidRR.Code)
	}

	form := url.Values{}
	form.Set("due_date", "2026-03-20")
	form.Add("student_ids", "2")
	form.Add("student_ids", "bad-id")
	form.Add("student_ids", "3")
	req := httptest.NewRequest(http.MethodPost, "/teacher/test/8/assign", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "8")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rr := httptest.NewRecorder()

	mw.RequireAuth(http.HandlerFunc(handler.AssignTest)).ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect after assignment, got %d", rr.Code)
	}
	if location := rr.Header().Get("Location"); location != "/teacher/dashboard" {
		t.Fatalf("expected redirect to teacher dashboard, got %s", location)
	}
	if len(assignmentStore.created) != 2 {
		t.Fatalf("expected 2 valid assignments to be created, got %d", len(assignmentStore.created))
	}
	if assignmentStore.created[0].AssignedTo != 2 || assignmentStore.created[1].AssignedTo != 3 {
		t.Fatalf("unexpected assignment targets: %#v", assignmentStore.created)
	}
	if assignmentStore.created[0].AssignedBy == nil || *assignmentStore.created[0].AssignedBy != 7 {
		t.Fatalf("expected assigner to be current teacher, got %#v", assignmentStore.created[0].AssignedBy)
	}
}

func TestAdminCreateSubject_ReturnsHTMXFragment(t *testing.T) {
	handler := NewAdminHandler(&mockAdminTestStore{createdSubjectID: 12}, &mockAdminUserStore{}, nil)
	form := url.Values{}
	form.Set("name", "Physics")
	form.Set("description", "Science subject")
	req := httptest.NewRequest(http.MethodPost, "/admin/manage/subjects", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.CreateSubject(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "/admin/manage/subjects/12") {
		t.Fatalf("expected HTMX fragment with subject ID, got %q", rr.Body.String())
	}
}

func TestAdminCreateUser_ValidatesRoleAndInitializesStats(t *testing.T) {
	userStore := &mockAdminUserStore{}
	handler := NewAdminHandler(&mockAdminTestStore{}, userStore, nil)

	invalidForm := url.Values{}
	invalidForm.Set("email", "staff@example.com")
	invalidForm.Set("username", "staff")
	invalidForm.Set("password", "secret123")
	invalidForm.Set("role", "superadmin")
	invalidReq := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(invalidForm.Encode()))
	invalidReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	invalidRR := httptest.NewRecorder()
	handler.CreateUser(invalidRR, invalidReq)
	if invalidRR.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid role to fail with 400, got %d", invalidRR.Code)
	}

	form := url.Values{}
	form.Set("email", "staff@example.com")
	form.Set("username", "staff")
	form.Set("password", "secret123")
	form.Set("role", "teacher")
	req := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.CreateUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected successful user create, got %d", rr.Code)
	}
	if userStore.createdUser == nil || userStore.createdUser.Role != "teacher" {
		t.Fatalf("expected created teacher user, got %#v", userStore.createdUser)
	}
	if userStore.initializedUserID != 44 {
		t.Fatalf("expected stats initialization for created user 44, got %d", userStore.initializedUserID)
	}
}

func TestAdminUpdateUserRole_ValidatesAndUpdates(t *testing.T) {
	userStore := &mockAdminUserStore{}
	handler := NewAdminHandler(&mockAdminTestStore{}, userStore, nil)

	badForm := url.Values{}
	badForm.Set("role", "owner")
	badReq := httptest.NewRequest(http.MethodPost, "/admin/users/9/role", strings.NewReader(badForm.Encode()))
	badReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	badReq.SetPathValue("id", "9")
	badRR := httptest.NewRecorder()
	handler.UpdateUserRole(badRR, badReq)
	if badRR.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid role to fail with 400, got %d", badRR.Code)
	}

	form := url.Values{}
	form.Set("role", "admin")
	req := httptest.NewRequest(http.MethodPost, "/admin/users/9/role", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "9")
	rr := httptest.NewRecorder()
	handler.UpdateUserRole(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected successful role update, got %d", rr.Code)
	}
	if userStore.updatedRoleUserID != 9 || userStore.updatedRole != "admin" {
		t.Fatalf("expected role update for user 9 to admin, got user=%d role=%s", userStore.updatedRoleUserID, userStore.updatedRole)
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if success, ok := payload["success"].(bool); !ok || !success {
		t.Fatalf("expected success response, got %#v", payload)
	}
}

var _ = time.Now
