package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestGetByStudent_MarksOverdueAndLoadsAssignments(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgx mock pool: %v", err)
	}
	defer mock.Close()

	repo := NewAssignmentRepository(mock)
	dueDate := time.Now().Add(24 * time.Hour)
	assignedAt := time.Now()
	assignedBy := 3
	subjectID := 4
	subjectName := "Math"

	overdueQuery := regexp.QuoteMeta(`
		UPDATE test_assignments SET status = 'overdue'
		WHERE status = 'pending' AND due_date < NOW()`)
	selectQuery := regexp.QuoteMeta(`
		SELECT ta.id, ta.test_id, ta.assigned_by, ta.assigned_to, ta.due_date, ta.status, ta.assigned_at,
		       t.title, t.difficulty, t.time_limit_minutes, t.passing_score,
		       s.id, s.name
		FROM test_assignments ta
		JOIN tests t ON ta.test_id = t.id
		LEFT JOIN subjects s ON t.subject_id = s.id
		WHERE ta.assigned_to = $1
		ORDER BY ta.due_date ASC`)

	mock.ExpectExec(overdueQuery).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	rows := pgxmock.NewRows([]string{"id", "test_id", "assigned_by", "assigned_to", "due_date", "status", "assigned_at", "title", "difficulty", "time_limit_minutes", "passing_score", "subject_id", "subject_name"}).
		AddRow(5, 9, &assignedBy, 12, dueDate, "pending", assignedAt, "Trig Test", "Medium", 30, 70, &subjectID, &subjectName)
	mock.ExpectQuery(selectQuery).WithArgs(12).WillReturnRows(rows)

	assignments, err := repo.GetByStudent(context.Background(), 12)
	if err != nil {
		t.Fatalf("GetByStudent returned error: %v", err)
	}
	if len(assignments) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(assignments))
	}
	if assignments[0].Test == nil || assignments[0].Test.Subject == nil || assignments[0].Test.Subject.Name != "Math" {
		t.Fatalf("expected nested test subject to be populated, got %#v", assignments[0].Test)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet pgx expectations: %v", err)
	}
}

func TestMarkCompleted_UpdatesPendingAndOverdueAssignments(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgx mock pool: %v", err)
	}
	defer mock.Close()

	repo := NewAssignmentRepository(mock)
	query := regexp.QuoteMeta(`
		UPDATE test_assignments SET status = 'completed'
		WHERE assigned_to = $1 AND test_id = $2 AND status IN ('pending', 'overdue')`)
	mock.ExpectExec(query).WithArgs(14, 8).WillReturnResult(pgxmock.NewResult("UPDATE", 2))

	if err := repo.MarkCompleted(context.Background(), 14, 8); err != nil {
		t.Fatalf("MarkCompleted returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet pgx expectations: %v", err)
	}
}

func TestCountPending_MarksOverdueBeforeCounting(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgx mock pool: %v", err)
	}
	defer mock.Close()

	repo := NewAssignmentRepository(mock)
	overdueQuery := regexp.QuoteMeta(`
		UPDATE test_assignments SET status = 'overdue'
		WHERE status = 'pending' AND due_date < NOW()`)
	countQuery := regexp.QuoteMeta(`
		SELECT COUNT(*) FROM test_assignments
		WHERE assigned_to = $1 AND status IN ('pending', 'overdue')`)

	mock.ExpectExec(overdueQuery).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectQuery(countQuery).WithArgs(6).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(4))

	count := repo.CountPending(context.Background(), 6)
	if count != 4 {
		t.Fatalf("expected 4 pending assignments, got %d", count)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet pgx expectations: %v", err)
	}
}
