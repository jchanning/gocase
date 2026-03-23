package repository

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jchanning/gocase/internal/models"
)

// RevisionRepository handles database operations for revision plans and sessions.
type RevisionRepository struct {
	pool dbQuerier
}

// NewRevisionRepository creates a new RevisionRepository.
func NewRevisionRepository(pool dbQuerier) *RevisionRepository {
	return &RevisionRepository{pool: pool}
}

// GetPlansByUser returns all revision plans for a user, with syllabus join and session counts.
func (r *RevisionRepository) GetPlansByUser(ctx context.Context, userID int) ([]models.RevisionPlan, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT rp.id, rp.user_id, rp.syllabus_id, rp.exam_date,
		       rp.hours_per_day, rp.study_days, rp.created_at, rp.updated_at,
		       sy.title, sy.exam_standard, sy.is_published,
		       s.id, s.name,
		       (SELECT COUNT(*) FROM revision_sessions rs WHERE rs.plan_id = rp.id) AS total,
		       (SELECT COUNT(*) FROM revision_sessions rs WHERE rs.plan_id = rp.id AND rs.status = 'completed') AS completed
		FROM revision_plans rp
		JOIN syllabi sy ON sy.id = rp.syllabus_id
		LEFT JOIN subjects s ON s.id = sy.subject_id
		WHERE rp.user_id = $1
		ORDER BY rp.exam_date`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []models.RevisionPlan
	now := time.Now()
	for rows.Next() {
		var p models.RevisionPlan
		var daysRaw string
		var syTitle, syStandard string
		var syPublished bool
		var subID *int
		var subName *string
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.SyllabusID, &p.ExamDate,
			&p.HoursPerDay, &daysRaw, &p.CreatedAt, &p.UpdatedAt,
			&syTitle, &syStandard, &syPublished,
			&subID, &subName,
			&p.TotalSessions, &p.CompletedCount,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(daysRaw), &p.StudyDays); err != nil {
			p.StudyDays = []int{1, 2, 3, 4, 5}
		}
		sy := &models.Syllabus{
			ID:           p.SyllabusID,
			Title:        syTitle,
			ExamStandard: syStandard,
			IsPublished:  syPublished,
		}
		if subID != nil {
			sy.Subject = &models.Subject{ID: *subID, Name: *subName}
		}
		p.Syllabus = sy
		p.DaysUntilExam = int(math.Ceil(p.ExamDate.Sub(now).Hours() / 24))
		if p.TotalSessions > 0 {
			p.Progress = (p.CompletedCount * 100) / p.TotalSessions
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

// GetPlanByUserAndSyllabus returns the single plan a user has for a given syllabus, or nil.
func (r *RevisionRepository) GetPlanByUserAndSyllabus(ctx context.Context, userID, syllabusID int) (*models.RevisionPlan, error) {
	var p models.RevisionPlan
	var daysRaw string
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, syllabus_id, exam_date, hours_per_day, study_days, created_at, updated_at
		FROM revision_plans WHERE user_id=$1 AND syllabus_id=$2`, userID, syllabusID).Scan(
		&p.ID, &p.UserID, &p.SyllabusID, &p.ExamDate,
		&p.HoursPerDay, &daysRaw, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(daysRaw), &p.StudyDays); err != nil {
		p.StudyDays = []int{1, 2, 3, 4, 5}
	}
	return &p, nil
}

// GetPlanByID returns a full revision plan with all sessions and their topic details.
func (r *RevisionRepository) GetPlanByID(ctx context.Context, id int) (*models.RevisionPlan, error) {
	var p models.RevisionPlan
	var daysRaw string
	var syTitle, syStandard string
	var syPublished bool
	var subID *int
	var subName *string
	err := r.pool.QueryRow(ctx, `
		SELECT rp.id, rp.user_id, rp.syllabus_id, rp.exam_date,
		       rp.hours_per_day, rp.study_days, rp.created_at, rp.updated_at,
		       sy.title, sy.exam_standard, sy.is_published,
		       s.id, s.name
		FROM revision_plans rp
		JOIN syllabi sy ON sy.id = rp.syllabus_id
		LEFT JOIN subjects s ON s.id = sy.subject_id
		WHERE rp.id = $1`, id).Scan(
		&p.ID, &p.UserID, &p.SyllabusID, &p.ExamDate,
		&p.HoursPerDay, &daysRaw, &p.CreatedAt, &p.UpdatedAt,
		&syTitle, &syStandard, &syPublished,
		&subID, &subName,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(daysRaw), &p.StudyDays); err != nil {
		p.StudyDays = []int{1, 2, 3, 4, 5}
	}
	sy := &models.Syllabus{
		ID:           p.SyllabusID,
		Title:        syTitle,
		ExamStandard: syStandard,
		IsPublished:  syPublished,
	}
	if subID != nil {
		sy.Subject = &models.Subject{ID: *subID, Name: *subName}
	}
	p.Syllabus = sy
	p.DaysUntilExam = int(math.Ceil(p.ExamDate.Sub(time.Now()).Hours() / 24))

	sessions, err := r.getSessionsForPlan(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Sessions = sessions
	p.TotalSessions = len(sessions)
	for _, s := range sessions {
		if s.Status == "completed" {
			p.CompletedCount++
		}
	}
	return &p, nil
}

func (r *RevisionRepository) getSessionsForPlan(ctx context.Context, planID int) ([]models.RevisionSession, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT rs.id, rs.plan_id, rs.session_date, rs.syllabus_topic_id,
		       rs.hours_allocated, rs.status, COALESCE(rs.notes,''), rs.completed_at, rs.created_at,
		       st.id, st.title, COALESCE(st.notes_content,''), st.estimated_hours
		FROM revision_sessions rs
		JOIN syllabus_topics st ON st.id = rs.syllabus_topic_id
		WHERE rs.plan_id = $1
		ORDER BY rs.session_date, rs.id`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.RevisionSession
	for rows.Next() {
		var s models.RevisionSession
		var topicID int
		var topicTitle, topicNotes string
		var topicHours float64
		if err := rows.Scan(
			&s.ID, &s.PlanID, &s.SessionDate, &s.SyllabusTopicID,
			&s.HoursAllocated, &s.Status, &s.Notes, &s.CompletedAt, &s.CreatedAt,
			&topicID, &topicTitle, &topicNotes, &topicHours,
		); err != nil {
			return nil, err
		}
		s.Topic = &models.SyllabusTopic{
			ID:             topicID,
			Title:          topicTitle,
			NotesContent:   topicNotes,
			EstimatedHours: topicHours,
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// transactor allows executing a transaction against the pool.
type transactor interface {
	dbQuerier
	Begin(ctx context.Context) (pgx.Tx, error)
}

// CreatePlan inserts the plan and all its sessions atomically inside a transaction.
func (r *RevisionRepository) CreatePlan(ctx context.Context, plan *models.RevisionPlan, sessions []models.RevisionSession) error {
	tx, err := r.pool.(transactor).Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	daysJSON, err := json.Marshal(plan.StudyDays)
	if err != nil {
		return err
	}

	if err := tx.QueryRow(ctx, `
		INSERT INTO revision_plans (user_id, syllabus_id, exam_date, hours_per_day, study_days)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, created_at, updated_at`,
		plan.UserID, plan.SyllabusID, plan.ExamDate.Format("2006-01-02"),
		plan.HoursPerDay, string(daysJSON),
	).Scan(&plan.ID, &plan.CreatedAt, &plan.UpdatedAt); err != nil {
		return err
	}

	for i := range sessions {
		sessions[i].PlanID = plan.ID
		if err := tx.QueryRow(ctx, `
			INSERT INTO revision_sessions
			    (plan_id, session_date, syllabus_topic_id, hours_allocated, status)
			VALUES ($1,$2,$3,$4,'scheduled')
			RETURNING id, created_at`,
			sessions[i].PlanID,
			sessions[i].SessionDate.Format("2006-01-02"),
			sessions[i].SyllabusTopicID,
			sessions[i].HoursAllocated,
		).Scan(&sessions[i].ID, &sessions[i].CreatedAt); err != nil {
			return err
		}
	}

	plan.Sessions = sessions
	return tx.Commit(ctx)
}

// DeletePlan removes a plan and all its sessions.
func (r *RevisionRepository) DeletePlan(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM revision_plans WHERE id=$1`, id)
	return err
}

// UpdateSessionStatus changes a session's status and sets completed_at if necessary.
func (r *RevisionRepository) UpdateSessionStatus(ctx context.Context, sessionID int, status, notes string) error {
	if status == "completed" {
		_, err := r.pool.Exec(ctx, `
			UPDATE revision_sessions
			SET status=$1, notes=$2, completed_at=CURRENT_TIMESTAMP
			WHERE id=$3`, status, notes, sessionID)
		return err
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE revision_sessions
		SET status=$1, notes=$2, completed_at=NULL
		WHERE id=$3`, status, notes, sessionID)
	return err
}

// GenerateSessions builds RevisionSession records from the plan inputs and ordered topics.
// It does not persist them — call CreatePlan to save.
func GenerateSessions(plan *models.RevisionPlan, topics []models.SyllabusTopic) []models.RevisionSession {
	if len(topics) == 0 {
		return nil
	}

	studyDaysSet := make(map[int]bool, len(plan.StudyDays))
	for _, d := range plan.StudyDays {
		studyDaysSet[d] = true
	}

	today := time.Now().Truncate(24 * time.Hour)
	examDay := plan.ExamDate.Truncate(24 * time.Hour)

	var availableDays []time.Time
	for d := today.AddDate(0, 0, 1); !d.After(examDay); d = d.AddDate(0, 0, 1) {
		if studyDaysSet[int(d.Weekday())] {
			availableDays = append(availableDays, d)
		}
	}

	if len(availableDays) == 0 {
		return nil
	}

	totalAvailable := float64(len(availableDays)) * plan.HoursPerDay

	totalEstimated := 0.0
	for _, t := range topics {
		totalEstimated += t.EstimatedHours
	}

	scale := 1.0
	if totalEstimated > totalAvailable {
		scale = totalAvailable / totalEstimated
	}

	var sessions []models.RevisionSession
	dayIdx := 0
	hoursUsedToday := 0.0

	for _, topic := range topics {
		remaining := roundHours(topic.EstimatedHours * scale)
		for remaining > 0.001 && dayIdx < len(availableDays) {
			dayRemaining := roundHours(plan.HoursPerDay - hoursUsedToday)
			chunk := roundHours(math.Min(remaining, dayRemaining))

			sessions = append(sessions, models.RevisionSession{
				SessionDate:     availableDays[dayIdx],
				SyllabusTopicID: topic.ID,
				HoursAllocated:  chunk,
				Status:          "scheduled",
			})

			remaining = roundHours(remaining - chunk)
			hoursUsedToday = roundHours(hoursUsedToday + chunk)

			if hoursUsedToday >= plan.HoursPerDay-0.001 {
				dayIdx++
				hoursUsedToday = 0
			}
		}
	}

	return sessions
}

func roundHours(h float64) float64 {
	return math.Round(h*100) / 100
}
