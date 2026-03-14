package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jchanning/gocase/internal/llm"
	"github.com/jchanning/gocase/internal/models"
)

type adminFlowLLM struct {
	available bool
}

func (m *adminFlowLLM) GenerateQuestions(ctx context.Context, notesText string, cfg llm.GenerateConfig) ([]models.QuestionUpload, error) {
	return []models.QuestionUpload{{QuestionText: "Generated", Points: 1, Options: []string{"A", "B", "C", "D"}, CorrectIndex: 0}}, nil
}

func (m *adminFlowLLM) IsAvailable() bool {
	return m.available
}

type adminNotesTestStore struct {
	test                 *models.Test
	updatedNotesTestID   int
	updatedNotesFilename *string
}

func (m *adminNotesTestStore) GetOrCreateSubject(ctx context.Context, name, description string) (int, error) {
	return 0, nil
}
func (m *adminNotesTestStore) GetOrCreateTopic(ctx context.Context, subjectID int, name, description string) (int, error) {
	return 0, nil
}
func (m *adminNotesTestStore) Create(ctx context.Context, test *models.Test) error { return nil }
func (m *adminNotesTestStore) UpdateTestNotes(ctx context.Context, testID int, notesFilename *string) error {
	m.updatedNotesTestID = testID
	m.updatedNotesFilename = notesFilename
	return nil
}
func (m *adminNotesTestStore) CreateQuestion(ctx context.Context, question *models.Question) error {
	return nil
}
func (m *adminNotesTestStore) CreateAnswerOption(ctx context.Context, option *models.AnswerOption) error {
	return nil
}
func (m *adminNotesTestStore) GetAll(ctx context.Context) ([]models.Test, error) { return nil, nil }
func (m *adminNotesTestStore) GetSubjects(ctx context.Context) ([]models.Subject, error) {
	return nil, nil
}
func (m *adminNotesTestStore) GetByID(ctx context.Context, id int) (*models.Test, error) {
	return m.test, nil
}
func (m *adminNotesTestStore) DeleteTest(ctx context.Context, testID int) error    { return nil }
func (m *adminNotesTestStore) Update(ctx context.Context, test *models.Test) error { return nil }
func (m *adminNotesTestStore) UpdateQuestion(ctx context.Context, question *models.Question) error {
	return nil
}
func (m *adminNotesTestStore) UpdateAnswerOption(ctx context.Context, option *models.AnswerOption) error {
	return nil
}

func TestGenerateFromNotes_ReturnsServiceUnavailableWhenLLMDisabled(t *testing.T) {
	handler := NewAdminHandler(&adminNotesTestStore{}, &mockAdminUserStore{}, &adminFlowLLM{available: false})
	req := httptest.NewRequest(http.MethodPost, "/admin/generate", nil)
	rr := httptest.NewRecorder()

	handler.GenerateFromNotes(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when LLM unavailable, got %d", rr.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if success, ok := payload["success"].(bool); !ok || success {
		t.Fatalf("expected unsuccessful response, got %#v", payload)
	}
}

func TestGenerateFromNotes_RequiresDocumentAndValidFileType(t *testing.T) {
	handler := NewAdminHandler(&adminNotesTestStore{}, &mockAdminUserStore{}, &adminFlowLLM{available: true})

	missingReq := httptest.NewRequest(http.MethodPost, "/admin/generate", nil)
	missingRR := httptest.NewRecorder()
	handler.GenerateFromNotes(missingRR, missingReq)
	if missingRR.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when document missing, got %d", missingRR.Code)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("subject", "Math")
	part, err := writer.CreateFormFile("document", "notes.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	_, _ = part.Write([]byte("plain text"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/admin/generate", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	handler.GenerateFromNotes(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid notes file type, got %d", rr.Code)
	}
}

func TestRemoveAndServeNotes_Flows(t *testing.T) {
	handler := NewAdminHandler(&adminNotesTestStore{}, &mockAdminUserStore{}, nil)

	invalidReq := httptest.NewRequest(http.MethodPost, "/admin/test/bad/remove-notes", nil)
	invalidReq.SetPathValue("id", "bad")
	invalidRR := httptest.NewRecorder()
	handler.RemoveTestNotes(invalidRR, invalidReq)
	if invalidRR.Code != http.StatusOK {
		t.Fatalf("expected JSON response for invalid id, got %d", invalidRR.Code)
	}

	store := &adminNotesTestStore{test: &models.Test{ID: 5}}
	handler = NewAdminHandler(store, &mockAdminUserStore{}, nil)
	req := httptest.NewRequest(http.MethodPost, "/admin/test/5/remove-notes", nil)
	req.SetPathValue("id", "5")
	rr := httptest.NewRecorder()
	handler.RemoveTestNotes(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected successful notes removal response, got %d", rr.Code)
	}
	if store.updatedNotesTestID != 5 || store.updatedNotesFilename != nil {
		t.Fatalf("expected notes filename cleared for test 5, got test=%d filename=%#v", store.updatedNotesTestID, store.updatedNotesFilename)
	}

	serveStore := &adminNotesTestStore{test: &models.Test{ID: 6, NotesFilename: strPtr("missing.pdf")}}
	handler = NewAdminHandler(serveStore, &mockAdminUserStore{}, nil)
	serveReq := httptest.NewRequest(http.MethodGet, "/admin/test/6/notes", nil)
	serveReq.SetPathValue("id", "6")
	serveRR := httptest.NewRecorder()
	handler.ServeTestNotes(serveRR, serveReq)

	if serveRR.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing notes file, got %d", serveRR.Code)
	}
}

func strPtr(v string) *string { return &v }
