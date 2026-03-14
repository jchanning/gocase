package repository

import (
	"context"

	"github.com/jchanning/gocase/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AssignmentRepository handles test assignment database operations.
type AssignmentRepository struct {
	pool *pgxpool.Pool
}

// NewAssignmentRepository creates a new AssignmentRepository.
func NewAssignmentRepository(pool *pgxpool.Pool) *AssignmentRepository {
	return &AssignmentRepository{pool: pool}
}

// Create inserts a new test assignment.
func (r *AssignmentRepository) Create(ctx context.Context, a *models.TestAssignment) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO test_assignments (test_id, assigned_by, assigned_to, due_date, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id, assigned_at`,
		a.TestID, a.AssignedBy, a.AssignedTo, a.DueDate,
	).Scan(&a.ID, &a.AssignedAt)
}

// GetByStudent returns all assignments for a student, sorted by due date.
// It first marks any past-due pending items as overdue.
func (r *AssignmentRepository) GetByStudent(ctx context.Context, studentID int) ([]models.TestAssignment, error) {
	r.markOverdue(ctx)

	rows, err := r.pool.Query(ctx, `
		SELECT ta.id, ta.test_id, ta.assigned_by, ta.assigned_to, ta.due_date, ta.status, ta.assigned_at,
		       t.title, t.difficulty, t.time_limit_minutes, t.passing_score,
		       s.id, s.name
		FROM test_assignments ta
		JOIN tests t ON ta.test_id = t.id
		LEFT JOIN subjects s ON t.subject_id = s.id
		WHERE ta.assigned_to = $1
		ORDER BY ta.due_date ASC`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []models.TestAssignment
	for rows.Next() {
		var a models.TestAssignment
		a.Test = &models.Test{}
		var subjectID *int
		var subjectName *string

		if err := rows.Scan(
			&a.ID, &a.TestID, &a.AssignedBy, &a.AssignedTo, &a.DueDate, &a.Status, &a.AssignedAt,
			&a.Test.Title, &a.Test.Difficulty, &a.Test.TimeLimitMinutes, &a.Test.PassingScore,
			&subjectID, &subjectName,
		); err != nil {
			return nil, err
		}
		a.Test.ID = a.TestID
		if subjectID != nil {
			a.Test.Subject = &models.Subject{ID: *subjectID, Name: *subjectName}
		}
		assignments = append(assignments, a)
	}
	return assignments, rows.Err()
}

// GetByTeacher returns all assignments created by a teacher, including student and test info.
func (r *AssignmentRepository) GetByTeacher(ctx context.Context, teacherID int) ([]models.TestAssignment, error) {
	r.markOverdue(ctx)

	rows, err := r.pool.Query(ctx, `
		SELECT ta.id, ta.test_id, ta.assigned_by, ta.assigned_to, ta.due_date, ta.status, ta.assigned_at,
		       t.title, t.difficulty,
		       u.id, u.username, u.email
		FROM test_assignments ta
		JOIN tests t ON ta.test_id = t.id
		JOIN users u ON ta.assigned_to = u.id
		WHERE ta.assigned_by = $1
		ORDER BY ta.assigned_at DESC`, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []models.TestAssignment
	for rows.Next() {
		var a models.TestAssignment
		a.Test = &models.Test{}
		a.Student = &models.User{}

		if err := rows.Scan(
			&a.ID, &a.TestID, &a.AssignedBy, &a.AssignedTo, &a.DueDate, &a.Status, &a.AssignedAt,
			&a.Test.Title, &a.Test.Difficulty,
			&a.Student.ID, &a.Student.Username, &a.Student.Email,
		); err != nil {
			return nil, err
		}
		a.Test.ID = a.TestID
		assignments = append(assignments, a)
	}
	return assignments, rows.Err()
}

// MarkCompleted marks the pending assignment for a student/test pair as completed.
func (r *AssignmentRepository) MarkCompleted(ctx context.Context, studentID, testID int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE test_assignments SET status = 'completed'
		WHERE assigned_to = $1 AND test_id = $2 AND status IN ('pending', 'overdue')`,
		studentID, testID)
	return err
}

// CountPending returns the number of pending or overdue assignments for a student.
func (r *AssignmentRepository) CountPending(ctx context.Context, studentID int) int {
	r.markOverdue(ctx)
	var count int
	r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM test_assignments
		WHERE assigned_to = $1 AND status IN ('pending', 'overdue')`, studentID).Scan(&count)
	return count
}

// markOverdue silently updates past-due pending assignments to overdue.
func (r *AssignmentRepository) markOverdue(ctx context.Context) {
	r.pool.Exec(ctx, `
		UPDATE test_assignments SET status = 'overdue'
		WHERE status = 'pending' AND due_date < NOW()`)
}
