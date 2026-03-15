package repository

import (
	"context"
	"strings"

	"github.com/jchanning/gocase/internal/models"
)

// FeedbackRepository handles student issue reporting and staff review workflows.
type FeedbackRepository struct {
	pool dbQuerier
}

// NewFeedbackRepository creates a new feedback repository.
func NewFeedbackRepository(pool dbQuerier) *FeedbackRepository {
	return &FeedbackRepository{pool: pool}
}

// CreateIssue stores a new student-reported issue.
func (r *FeedbackRepository) CreateIssue(ctx context.Context, issue *models.TestFeedbackIssue) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO test_feedback_issues (
			test_id, question_id, attempt_id, reported_by, issue_type, student_comment
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, status, created_at, updated_at`,
		issue.TestID, issue.QuestionID, issue.AttemptID, issue.ReportedBy, issue.IssueType, issue.StudentComment,
	).Scan(&issue.ID, &issue.Status, &issue.CreatedAt, &issue.UpdatedAt)
}

// ListIssues returns reported issues, optionally filtered by status.
func (r *FeedbackRepository) ListIssues(ctx context.Context, status string) ([]models.TestFeedbackIssue, error) {
	baseQuery := `
		SELECT i.id, i.test_id, i.question_id, i.attempt_id, i.reported_by,
		       i.issue_type, i.student_comment, i.status, i.review_response,
		       i.reviewed_by, i.reviewed_at, i.created_at, i.updated_at,
		       t.title, q.question_text,
		       reporter.username,
		       COALESCE(reviewer.username, '')
		FROM test_feedback_issues i
		JOIN tests t ON t.id = i.test_id
		JOIN questions q ON q.id = i.question_id
		JOIN users reporter ON reporter.id = i.reported_by
		LEFT JOIN users reviewer ON reviewer.id = i.reviewed_by`

	args := []any{}
	if strings.TrimSpace(status) != "" && status != "all" {
		baseQuery += ` WHERE i.status = $1`
		args = append(args, status)
	}
	baseQuery += ` ORDER BY CASE i.status WHEN 'open' THEN 1 WHEN 'in_review' THEN 2 WHEN 'resolved' THEN 3 ELSE 4 END, i.created_at DESC`

	rows, err := r.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := []models.TestFeedbackIssue{}
	for rows.Next() {
		var issue models.TestFeedbackIssue
		var testTitle string
		var questionText string
		if err := rows.Scan(
			&issue.ID, &issue.TestID, &issue.QuestionID, &issue.AttemptID, &issue.ReportedBy,
			&issue.IssueType, &issue.StudentComment, &issue.Status, &issue.ReviewResponse,
			&issue.ReviewedBy, &issue.ReviewedAt, &issue.CreatedAt, &issue.UpdatedAt,
			&testTitle, &questionText, &issue.ReportedByName, &issue.ReviewedByName,
		); err != nil {
			return nil, err
		}
		issue.Test = &models.Test{ID: issue.TestID, Title: testTitle}
		issue.Question = &models.Question{ID: issue.QuestionID, QuestionText: questionText}
		issues = append(issues, issue)
	}

	return issues, rows.Err()
}

// UpdateIssueReview updates staff response/status for a reported issue.
func (r *FeedbackRepository) UpdateIssueReview(ctx context.Context, issueID int, status, response string, reviewerID int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE test_feedback_issues
		SET status = $2,
		    review_response = NULLIF($3, ''),
		    reviewed_by = $4,
		    reviewed_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`, issueID, status, response, reviewerID)
	return err
}
