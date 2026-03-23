package handlers

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jchanning/gocase/internal/auth"
	"github.com/jchanning/gocase/internal/models"
	"github.com/jchanning/gocase/internal/repository"
)

// revisionSyllabusStore provides the published syllabus data needed by RevisionHandler.
type revisionSyllabusStore interface {
	GetPublished(ctx context.Context) ([]models.Syllabus, error)
	GetByID(ctx context.Context, id int) (*models.Syllabus, error)
	GetAllTopicsOrdered(ctx context.Context, syllabusID int) ([]models.SyllabusTopic, error)
}

// revisionPlanStore provides plan and session persistence.
type revisionPlanStore interface {
	GetPlansByUser(ctx context.Context, userID int) ([]models.RevisionPlan, error)
	GetPlanByUserAndSyllabus(ctx context.Context, userID, syllabusID int) (*models.RevisionPlan, error)
	GetPlanByID(ctx context.Context, id int) (*models.RevisionPlan, error)
	CreatePlan(ctx context.Context, plan *models.RevisionPlan, sessions []models.RevisionSession) error
	DeletePlan(ctx context.Context, id int) error
	UpdateSessionStatus(ctx context.Context, sessionID int, status, notes string) error
}

// RevisionHandler handles student revision planner requests.
type RevisionHandler struct {
	syllabusRepo revisionSyllabusStore
	revisionRepo revisionPlanStore
}

// NewRevisionHandler creates a new RevisionHandler.
func NewRevisionHandler(syllabusRepo revisionSyllabusStore, revisionRepo revisionPlanStore) *RevisionHandler {
	return &RevisionHandler{syllabusRepo: syllabusRepo, revisionRepo: revisionRepo}
}

func (h *RevisionHandler) render(w http.ResponseWriter, r *http.Request, data map[string]interface{}, files ...string) {
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		log.Printf("RevisionHandler template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("RevisionHandler render error: %v", err)
	}
}

// ShowRevision is the student revision home showing active plans and available syllabi.
func (h *RevisionHandler) ShowRevision(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)

	plans, err := h.revisionRepo.GetPlansByUser(r.Context(), session.UserID)
	if err != nil {
		log.Printf("RevisionHandler.ShowRevision plans error: %v", err)
		plans = []models.RevisionPlan{}
	}

	published, err := h.syllabusRepo.GetPublished(r.Context())
	if err != nil {
		log.Printf("RevisionHandler.ShowRevision syllabi error: %v", err)
		published = []models.Syllabus{}
	}

	// Filter out syllabi the student already has a plan for.
	plannedIDs := make(map[int]bool)
	for _, p := range plans {
		plannedIDs[p.SyllabusID] = true
	}
	available := make([]models.Syllabus, 0, len(published))
	for _, s := range published {
		if !plannedIDs[s.ID] {
			available = append(available, s)
		}
	}

	// Compute DaysUntilExam for each plan.
	for i := range plans {
		days := int(time.Until(plans[i].ExamDate.Truncate(24*time.Hour)).Hours() / 24)
		if days < 0 {
			days = 0
		}
		plans[i].DaysUntilExam = days
	}

	h.render(w, r, map[string]interface{}{
		"Session":          session,
		"Plans":            plans,
		"AvailableSyllabi": available,
	}, "views/layout.html", "views/revision_planner.html")
}

// StudyDayOption is a weekday option for the create-plan form.
type StudyDayOption struct {
	Index   int
	Name    string
	Default bool
}

// studyDayOptions returns the Mon–Sun day options for the revision create form.
func studyDayOptions() []StudyDayOption {
	days := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	opts := make([]StudyDayOption, len(days))
	for i, name := range days {
		opts[i] = StudyDayOption{Index: i, Name: name, Default: i >= 1 && i <= 5}
	}
	return opts
}

// ShowCreatePlan renders the plan creation form for a specific syllabus.
func (h *RevisionHandler) ShowCreatePlan(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)

	syllabusIDStr := r.URL.Query().Get("syllabus_id")
	syllabusID, err := strconv.Atoi(syllabusIDStr)
	if err != nil || syllabusID <= 0 {
		http.Redirect(w, r, "/revision", http.StatusSeeOther)
		return
	}

	sy, err := h.syllabusRepo.GetByID(r.Context(), syllabusID)
	if err != nil || sy == nil || !sy.IsPublished {
		http.Redirect(w, r, "/revision", http.StatusSeeOther)
		return
	}

	// Check if a plan already exists for this syllabus.
	existing, _ := h.revisionRepo.GetPlanByUserAndSyllabus(r.Context(), session.UserID, syllabusID)
	if existing != nil {
		http.Redirect(w, r, "/revision/plan/"+strconv.Itoa(existing.ID), http.StatusSeeOther)
		return
	}

	// Compute total estimated hours for the syllabus.
	topics, _ := h.syllabusRepo.GetAllTopicsOrdered(r.Context(), syllabusID)
	totalHours := 0.0
	for _, t := range topics {
		totalHours += t.EstimatedHours
	}

	h.render(w, r, map[string]interface{}{
		"Session":         session,
		"Syllabus":        sy,
		"TotalHours":      totalHours,
		"TopicCount":      len(topics),
		"MinExamDate":     time.Now().AddDate(0, 0, 2).Format("2006-01-02"),
		"StudyDayOptions": studyDayOptions(),
	}, "views/layout.html", "views/revision_create.html")
}

// CreatePlan handles POST to generate and persist a new revision plan.
func (h *RevisionHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	syllabusID, err := strconv.Atoi(r.FormValue("syllabus_id"))
	if err != nil || syllabusID <= 0 {
		http.Redirect(w, r, "/revision", http.StatusSeeOther)
		return
	}

	examDate, err := time.Parse("2006-01-02", r.FormValue("exam_date"))
	if err != nil {
		http.Error(w, "Invalid exam date", http.StatusBadRequest)
		return
	}
	if time.Until(examDate.Truncate(24*time.Hour)) < 48*time.Hour {
		http.Error(w, "Exam date must be at least 2 days from today", http.StatusBadRequest)
		return
	}

	hoursPerDay, err := strconv.ParseFloat(r.FormValue("hours_per_day"), 64)
	if err != nil || hoursPerDay < 0.5 || hoursPerDay > 12 {
		http.Error(w, "Hours per day must be between 0.5 and 12", http.StatusBadRequest)
		return
	}

	// Parse study_days checkboxes (0=Sun … 6=Sat)
	studyDays := []int{}
	for _, v := range r.Form["study_days"] {
		if d, err := strconv.Atoi(v); err == nil && d >= 0 && d <= 6 {
			studyDays = append(studyDays, d)
		}
	}
	if len(studyDays) == 0 {
		http.Error(w, "At least one study day is required", http.StatusBadRequest)
		return
	}

	sy, err := h.syllabusRepo.GetByID(r.Context(), syllabusID)
	if err != nil || sy == nil {
		http.NotFound(w, r)
		return
	}

	topics, err := h.syllabusRepo.GetAllTopicsOrdered(r.Context(), syllabusID)
	if err != nil || len(topics) == 0 {
		http.Error(w, "This syllabus has no topics to schedule", http.StatusBadRequest)
		return
	}

	plan := &models.RevisionPlan{
		UserID:      session.UserID,
		SyllabusID:  syllabusID,
		ExamDate:    examDate,
		HoursPerDay: hoursPerDay,
		StudyDays:   studyDays,
	}

	sessions := repository.GenerateSessions(plan, topics)
	if len(sessions) == 0 {
		http.Error(w, "No study sessions could be scheduled with the given inputs", http.StatusBadRequest)
		return
	}

	if err := h.revisionRepo.CreatePlan(r.Context(), plan, sessions); err != nil {
		log.Printf("RevisionHandler.CreatePlan error: %v", err)
		http.Error(w, "Failed to create revision plan", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/revision/plan/"+strconv.Itoa(plan.ID), http.StatusSeeOther)
}

// CalendarDay represents a single day cell in the revision calendar.
type CalendarDay struct {
	Date     time.Time
	IsToday  bool
	IsPast   bool
	Sessions []models.RevisionSession
}

// CalendarWeek is a row of 7 days (Sunday–Saturday), nil for padding cells.
type CalendarWeek [7]*CalendarDay

// CalendarMonth holds a full month grid for template rendering.
type CalendarMonth struct {
	Year      int
	Month     time.Month
	MonthName string
	Weeks     []CalendarWeek
	Prev      string // query string for previous month link
	Next      string // query string for next month link
}

// ShowPlan renders the monthly calendar view for a revision plan.
func (h *RevisionHandler) ShowPlan(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	plan, err := h.revisionRepo.GetPlanByID(r.Context(), id)
	if err != nil || plan == nil {
		http.NotFound(w, r)
		return
	}
	if plan.UserID != session.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Determine which month to show.
	now := time.Now()
	year := now.Year()
	month := now.Month()
	if y, err := strconv.Atoi(r.URL.Query().Get("year")); err == nil {
		year = y
	}
	if m, err := strconv.Atoi(r.URL.Query().Get("month")); err == nil && m >= 1 && m <= 12 {
		month = time.Month(m)
	}

	cal := buildCalendar(year, month, plan.Sessions)

	// Compute stats.
	total := len(plan.Sessions)
	completed := 0
	for _, s := range plan.Sessions {
		if s.Status == "completed" {
			completed++
		}
	}
	progress := 0
	if total > 0 {
		progress = (completed * 100) / total
	}
	daysUntilExam := int(time.Until(plan.ExamDate.Truncate(24*time.Hour)).Hours() / 24)
	if daysUntilExam < 0 {
		daysUntilExam = 0
	}

	h.render(w, r, map[string]interface{}{
		"Session":       session,
		"Plan":          plan,
		"Calendar":      cal,
		"Total":         total,
		"Completed":     completed,
		"Progress":      progress,
		"DaysUntilExam": daysUntilExam,
		"DayNames":      []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
	}, "views/layout.html", "views/revision_plan.html")
}

// buildCalendar organises sessions into a weekly grid for the given month.
func buildCalendar(year int, month time.Month, sessions []models.RevisionSession) CalendarMonth {
	today := time.Now().Truncate(24 * time.Hour)

	// Index sessions by date string.
	byDate := make(map[string][]models.RevisionSession)
	for _, s := range sessions {
		key := s.SessionDate.Format("2006-01-02")
		byDate[key] = append(byDate[key], s)
	}

	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()

	var weeks []CalendarWeek
	var week CalendarWeek
	startWeekday := int(firstDay.Weekday()) // 0=Sun

	// Pad beginning of first week.
	col := startWeekday
	for d := 1; d <= daysInMonth; d++ {
		date := time.Date(year, month, d, 0, 0, 0, 0, time.UTC)
		cell := &CalendarDay{
			Date:     date,
			IsToday:  date.Equal(today),
			IsPast:   date.Before(today),
			Sessions: byDate[date.Format("2006-01-02")],
		}
		week[col] = cell
		col++
		if col == 7 {
			weeks = append(weeks, week)
			week = CalendarWeek{}
			col = 0
		}
	}
	if col > 0 {
		weeks = append(weeks, week)
	}

	prevYear, prevMonth := year, month-1
	if prevMonth < 1 {
		prevMonth = 12
		prevYear--
	}
	nextYear, nextMonth := year, month+1
	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}

	return CalendarMonth{
		Year:      year,
		Month:     month,
		MonthName: firstDay.Format("January 2006"),
		Weeks:     weeks,
		Prev:      "?year=" + strconv.Itoa(prevYear) + "&month=" + strconv.Itoa(int(prevMonth)),
		Next:      "?year=" + strconv.Itoa(nextYear) + "&month=" + strconv.Itoa(int(nextMonth)),
	}
}

// CompleteSession marks a session as completed.
func (h *RevisionHandler) CompleteSession(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	notes := strings.TrimSpace(r.FormValue("notes"))
	if err := h.revisionRepo.UpdateSessionStatus(r.Context(), id, "completed", notes); err != nil {
		log.Printf("RevisionHandler.CompleteSession error: %v", err)
	}
	// Redirect back to the plan that owns this session.
	planIDStr := r.FormValue("plan_id")
	if planID, err := strconv.Atoi(planIDStr); err == nil {
		http.Redirect(w, r, "/revision/plan/"+strconv.Itoa(planID)+buildMonthQuery(r), http.StatusSeeOther)
		return
	}
	_ = session
	http.Redirect(w, r, "/revision", http.StatusSeeOther)
}

// SkipSession marks a session as skipped.
func (h *RevisionHandler) SkipSession(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err := h.revisionRepo.UpdateSessionStatus(r.Context(), id, "skipped", ""); err != nil {
		log.Printf("RevisionHandler.SkipSession error: %v", err)
	}
	planIDStr := r.FormValue("plan_id")
	if planID, err := strconv.Atoi(planIDStr); err == nil {
		http.Redirect(w, r, "/revision/plan/"+strconv.Itoa(planID)+buildMonthQuery(r), http.StatusSeeOther)
		return
	}
	_ = session
	http.Redirect(w, r, "/revision", http.StatusSeeOther)
}

// DeletePlan deletes a revision plan (and all its sessions via CASCADE).
func (h *RevisionHandler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSessionData(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// Verify ownership before deleting.
	plan, err := h.revisionRepo.GetPlanByID(r.Context(), id)
	if err != nil || plan == nil || plan.UserID != session.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.revisionRepo.DeletePlan(r.Context(), id); err != nil {
		log.Printf("RevisionHandler.DeletePlan error: %v", err)
	}
	http.Redirect(w, r, "/revision", http.StatusSeeOther)
}

// buildMonthQuery preserves the current year/month query parameters when redirecting.
func buildMonthQuery(r *http.Request) string {
	year := r.FormValue("year")
	month := r.FormValue("month")
	if year != "" && month != "" {
		return "?year=" + year + "&month=" + month
	}
	return ""
}
