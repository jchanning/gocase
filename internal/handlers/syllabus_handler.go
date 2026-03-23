package handlers

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jchanning/gocase/internal/auth"
	"github.com/jchanning/gocase/internal/models"
)

// syllabusStore defines the repository methods used by SyllabusHandler.
type syllabusStore interface {
	GetAll(ctx context.Context) ([]models.Syllabus, error)
	GetByID(ctx context.Context, id int) (*models.Syllabus, error)
	Create(ctx context.Context, sy *models.Syllabus) error
	Update(ctx context.Context, sy *models.Syllabus) error
	Delete(ctx context.Context, id int) error
	Publish(ctx context.Context, id int) error
	Unpublish(ctx context.Context, id int) error
	AddSection(ctx context.Context, sec *models.SyllabusSection) error
	UpdateSection(ctx context.Context, sec *models.SyllabusSection) error
	DeleteSection(ctx context.Context, id int) error
	AddTopic(ctx context.Context, t *models.SyllabusTopic) error
	UpdateTopic(ctx context.Context, t *models.SyllabusTopic) error
	DeleteTopic(ctx context.Context, id int) error
	LinkTest(ctx context.Context, topicID, testID int) error
	UnlinkTest(ctx context.Context, topicID, testID int) error
	SearchTests(ctx context.Context, subjectID *int, titleFilter string) ([]models.Test, error)
}

// syllabusSubjectStore provides subject listing for the create/edit forms.
type syllabusSubjectStore interface {
	GetSubjects(ctx context.Context) ([]models.Subject, error)
}

// SyllabusHandler handles admin/teacher syllabus management requests.
type SyllabusHandler struct {
	repo        syllabusStore
	subjectRepo syllabusSubjectStore
}

// NewSyllabusHandler creates a new SyllabusHandler.
func NewSyllabusHandler(repo syllabusStore, subjectRepo syllabusSubjectStore) *SyllabusHandler {
	return &SyllabusHandler{repo: repo, subjectRepo: subjectRepo}
}

func (h *SyllabusHandler) render(w http.ResponseWriter, r *http.Request, data map[string]interface{}, files ...string) {
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		log.Printf("SyllabusHandler template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("SyllabusHandler render error: %v", err)
	}
}

// List shows all syllabi to admin/teacher users.
func (h *SyllabusHandler) List(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)

	syllabi, err := h.repo.GetAll(r.Context())
	if err != nil {
		log.Printf("SyllabusHandler.List error: %v", err)
		syllabi = []models.Syllabus{}
	}

	subjects, err := h.subjectRepo.GetSubjects(r.Context())
	if err != nil {
		log.Printf("SyllabusHandler.List subjects error: %v", err)
		subjects = []models.Subject{}
	}

	h.render(w, r, map[string]interface{}{
		"Session":  session,
		"Syllabi":  syllabi,
		"Subjects": subjects,
	}, "views/layout.html", "views/syllabus_manage.html")
}

// ShowCreate renders the create-syllabus form.
func (h *SyllabusHandler) ShowCreate(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)

	subjects, err := h.subjectRepo.GetSubjects(r.Context())
	if err != nil {
		subjects = []models.Subject{}
	}

	h.render(w, r, map[string]interface{}{
		"Session":       session,
		"Subjects":      subjects,
		"ExamStandards": models.ValidExamStandards,
		"Syllabus":      &models.Syllabus{},
		"SubjectIDInt":  0,
		"IsNew":         true,
	}, "views/layout.html", "views/syllabus_edit.html")
}

// Create handles the POST to create a new syllabus.
func (h *SyllabusHandler) Create(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	sy := &models.Syllabus{
		Title:        strings.TrimSpace(r.FormValue("title")),
		Description:  strings.TrimSpace(r.FormValue("description")),
		ExamStandard: r.FormValue("exam_standard"),
		CreatedBy:    &session.UserID,
	}

	if subjectIDStr := r.FormValue("subject_id"); subjectIDStr != "" {
		if id, err := strconv.Atoi(subjectIDStr); err == nil {
			sy.SubjectID = &id
		}
	}

	if sy.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if err := h.repo.Create(r.Context(), sy); err != nil {
		log.Printf("SyllabusHandler.Create error: %v", err)
		http.Error(w, "Failed to create syllabus", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/syllabus/"+strconv.Itoa(sy.ID), http.StatusSeeOther)
}

// ShowEdit renders the edit page for a syllabus.
func (h *SyllabusHandler) ShowEdit(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	sy, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		log.Printf("SyllabusHandler.ShowEdit error: %v", err)
		http.NotFound(w, r)
		return
	}

	subjects, err := h.subjectRepo.GetSubjects(r.Context())
	if err != nil {
		subjects = []models.Subject{}
	}

	// Collect tests matching the syllabus subject for the link panel
	var linkedSubjectID *int
	if sy.SubjectID != nil {
		linkedSubjectID = sy.SubjectID
	}
	availableTests, err := h.repo.SearchTests(r.Context(), linkedSubjectID, "")
	if err != nil {
		availableTests = []models.Test{}
	}

	subjectIDInt := 0
	if sy.SubjectID != nil {
		subjectIDInt = *sy.SubjectID
	}

	h.render(w, r, map[string]interface{}{
		"Session":        session,
		"Syllabus":       sy,
		"Subjects":       subjects,
		"ExamStandards":  models.ValidExamStandards,
		"AvailableTests": availableTests,
		"SubjectIDInt":   subjectIDInt,
		"IsNew":          false,
	}, "views/layout.html", "views/syllabus_edit.html")
}

// Update handles the POST to update syllabus metadata.
func (h *SyllabusHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	sy := &models.Syllabus{
		ID:           id,
		Title:        strings.TrimSpace(r.FormValue("title")),
		Description:  strings.TrimSpace(r.FormValue("description")),
		ExamStandard: r.FormValue("exam_standard"),
	}
	if subjectIDStr := r.FormValue("subject_id"); subjectIDStr != "" {
		if sid, err := strconv.Atoi(subjectIDStr); err == nil {
			sy.SubjectID = &sid
		}
	}

	if err := h.repo.Update(r.Context(), sy); err != nil {
		log.Printf("SyllabusHandler.Update error: %v", err)
		http.Error(w, "Failed to update syllabus", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/syllabus/"+strconv.Itoa(id), http.StatusSeeOther)
}

// Publish marks a syllabus as visible to students.
func (h *SyllabusHandler) Publish(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.repo.Publish(r.Context(), id); err != nil {
		log.Printf("SyllabusHandler.Publish error: %v", err)
	}
	http.Redirect(w, r, "/admin/syllabus/"+strconv.Itoa(id), http.StatusSeeOther)
}

// Unpublish hides a syllabus from students.
func (h *SyllabusHandler) Unpublish(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.repo.Unpublish(r.Context(), id); err != nil {
		log.Printf("SyllabusHandler.Unpublish error: %v", err)
	}
	http.Redirect(w, r, "/admin/syllabus/"+strconv.Itoa(id), http.StatusSeeOther)
}

// AddSection handles POST to add a new section to a syllabus.
func (h *SyllabusHandler) AddSection(w http.ResponseWriter, r *http.Request) {
	syllabusID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	order := 0
	if o, err := strconv.Atoi(r.FormValue("section_order")); err == nil {
		order = o
	}

	sec := &models.SyllabusSection{
		SyllabusID:   syllabusID,
		Title:        strings.TrimSpace(r.FormValue("title")),
		SectionOrder: order,
	}
	if sec.Title == "" {
		http.Redirect(w, r, "/admin/syllabus/"+strconv.Itoa(syllabusID), http.StatusSeeOther)
		return
	}
	if err := h.repo.AddSection(r.Context(), sec); err != nil {
		log.Printf("SyllabusHandler.AddSection error: %v", err)
	}
	http.Redirect(w, r, "/admin/syllabus/"+strconv.Itoa(syllabusID), http.StatusSeeOther)
}

// UpdateSection handles POST to rename or reorder a section.
func (h *SyllabusHandler) UpdateSection(w http.ResponseWriter, r *http.Request) {
	syllabusID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	sid, err := strconv.Atoi(chi.URLParam(r, "sid"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	order := 0
	if o, err := strconv.Atoi(r.FormValue("section_order")); err == nil {
		order = o
	}
	sec := &models.SyllabusSection{
		ID:           sid,
		Title:        strings.TrimSpace(r.FormValue("title")),
		SectionOrder: order,
	}
	if err := h.repo.UpdateSection(r.Context(), sec); err != nil {
		log.Printf("SyllabusHandler.UpdateSection error: %v", err)
	}
	http.Redirect(w, r, "/admin/syllabus/"+strconv.Itoa(syllabusID), http.StatusSeeOther)
}

// DeleteSection handles POST to delete a section.
func (h *SyllabusHandler) DeleteSection(w http.ResponseWriter, r *http.Request) {
	syllabusID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	sid, err := strconv.Atoi(chi.URLParam(r, "sid"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.repo.DeleteSection(r.Context(), sid); err != nil {
		log.Printf("SyllabusHandler.DeleteSection error: %v", err)
	}
	http.Redirect(w, r, "/admin/syllabus/"+strconv.Itoa(syllabusID), http.StatusSeeOther)
}

// AddTopic handles POST to add a new topic to a syllabus section.
func (h *SyllabusHandler) AddTopic(w http.ResponseWriter, r *http.Request) {
	syllabusID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	t := &models.SyllabusTopic{
		SyllabusID:     syllabusID,
		Title:          strings.TrimSpace(r.FormValue("title")),
		Description:    strings.TrimSpace(r.FormValue("description")),
		NotesContent:   strings.TrimSpace(r.FormValue("notes_content")),
		EstimatedHours: 1.0,
	}
	if h, err := strconv.ParseFloat(r.FormValue("estimated_hours"), 64); err == nil && h > 0 {
		t.EstimatedHours = h
	}
	if o, err := strconv.Atoi(r.FormValue("topic_order")); err == nil {
		t.TopicOrder = o
	}
	if sid, err := strconv.Atoi(r.FormValue("section_id")); err == nil {
		t.SectionID = &sid
	}

	if t.Title == "" {
		http.Redirect(w, r, "/admin/syllabus/"+strconv.Itoa(syllabusID), http.StatusSeeOther)
		return
	}
	if err := h.repo.AddTopic(r.Context(), t); err != nil {
		log.Printf("SyllabusHandler.AddTopic error: %v", err)
	}
	http.Redirect(w, r, "/admin/syllabus/"+strconv.Itoa(syllabusID), http.StatusSeeOther)
}

// UpdateTopic handles POST to update an existing topic.
func (h *SyllabusHandler) UpdateTopic(w http.ResponseWriter, r *http.Request) {
	syllabusID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	tid, err := strconv.Atoi(chi.URLParam(r, "tid"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	t := &models.SyllabusTopic{
		ID:           tid,
		Title:        strings.TrimSpace(r.FormValue("title")),
		Description:  strings.TrimSpace(r.FormValue("description")),
		NotesContent: strings.TrimSpace(r.FormValue("notes_content")),
	}
	if hr, err := strconv.ParseFloat(r.FormValue("estimated_hours"), 64); err == nil && hr > 0 {
		t.EstimatedHours = hr
	} else {
		t.EstimatedHours = 1.0
	}
	if o, err := strconv.Atoi(r.FormValue("topic_order")); err == nil {
		t.TopicOrder = o
	}
	if sid, err := strconv.Atoi(r.FormValue("section_id")); err == nil {
		t.SectionID = &sid
	}

	if err := h.repo.UpdateTopic(r.Context(), t); err != nil {
		log.Printf("SyllabusHandler.UpdateTopic error: %v", err)
	}
	http.Redirect(w, r, "/admin/syllabus/"+strconv.Itoa(syllabusID), http.StatusSeeOther)
}

// DeleteTopic handles POST to delete a topic.
func (h *SyllabusHandler) DeleteTopic(w http.ResponseWriter, r *http.Request) {
	syllabusID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	tid, err := strconv.Atoi(chi.URLParam(r, "tid"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.repo.DeleteTopic(r.Context(), tid); err != nil {
		log.Printf("SyllabusHandler.DeleteTopic error: %v", err)
	}
	http.Redirect(w, r, "/admin/syllabus/"+strconv.Itoa(syllabusID), http.StatusSeeOther)
}

// LinkTest handles POST to link a test to a syllabus topic.
func (h *SyllabusHandler) LinkTest(w http.ResponseWriter, r *http.Request) {
	syllabusID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	tid, err := strconv.Atoi(chi.URLParam(r, "tid"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	testID, err := strconv.Atoi(r.FormValue("test_id"))
	if err != nil {
		http.Redirect(w, r, "/admin/syllabus/"+strconv.Itoa(syllabusID), http.StatusSeeOther)
		return
	}
	if err := h.repo.LinkTest(r.Context(), tid, testID); err != nil {
		log.Printf("SyllabusHandler.LinkTest error: %v", err)
	}
	http.Redirect(w, r, "/admin/syllabus/"+strconv.Itoa(syllabusID), http.StatusSeeOther)
}

// UnlinkTest handles POST to remove a test link from a syllabus topic.
func (h *SyllabusHandler) UnlinkTest(w http.ResponseWriter, r *http.Request) {
	syllabusID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	tid, err := strconv.Atoi(chi.URLParam(r, "tid"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	testID, err := strconv.Atoi(chi.URLParam(r, "testid"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.repo.UnlinkTest(r.Context(), tid, testID); err != nil {
		log.Printf("SyllabusHandler.UnlinkTest error: %v", err)
	}
	http.Redirect(w, r, "/admin/syllabus/"+strconv.Itoa(syllabusID), http.StatusSeeOther)
}
