package repository

import (
	"context"
	"errors"

	"github.com/jchanning/gocase/internal/models"
)

var ErrReviewApprovalRequired = errors.New("test must be approved before it can be published")

// TestRepository handles test database operations
type TestRepository struct {
	pool dbQuerier
}

// NewTestRepository creates a new test repository
func NewTestRepository(pool dbQuerier) *TestRepository {
	return &TestRepository{pool: pool}
}

// GetAll retrieves all tests
func (r *TestRepository) GetAll(ctx context.Context) ([]models.Test, error) {
	query := `
		SELECT t.id, t.title, COALESCE(t.description, ''), t.subject_id, t.topic_id,
		       t.exam_standard, t.difficulty, t.time_limit_minutes,
		       t.passing_score, t.published, t.review_status, t.reviewed_by, t.reviewed_at,
		       t.review_notes, t.submitted_for_review_at, t.notes_filename, t.created_by, t.created_at, t.updated_at,
		       s.id, s.name, s.description,
		       (SELECT COUNT(*) FROM questions q WHERE q.test_id = t.id) AS question_count
		FROM tests t
		LEFT JOIN subjects s ON t.subject_id = s.id
		ORDER BY t.created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []models.Test
	for rows.Next() {
		var t models.Test
		var subjectID *int
		var subjectName, subjectDesc *string

		err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.SubjectID, &t.TopicID,
			&t.ExamStandard, &t.Difficulty, &t.TimeLimitMinutes,
			&t.PassingScore, &t.Published, &t.ReviewStatus, &t.ReviewedBy, &t.ReviewedAt,
			&t.ReviewNotes, &t.SubmittedForReviewAt, &t.NotesFilename, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
			&subjectID, &subjectName, &subjectDesc,
			&t.QuestionCount,
		)
		if err != nil {
			return nil, err
		}

		if subjectID != nil {
			t.Subject = &models.Subject{
				ID:          *subjectID,
				Name:        *subjectName,
				Description: *subjectDesc,
			}
		}

		tests = append(tests, t)
	}

	return tests, rows.Err()
}

// Update updates an existing test
func (r *TestRepository) Update(ctx context.Context, test *models.Test) error {
	query := `
		UPDATE tests
		SET title = $1, description = $2, subject_id = $3, topic_id = $4,
		    exam_standard = $5, difficulty = $6, time_limit_minutes = $7,
		    passing_score = $8, review_status = 'draft', reviewed_by = NULL,
		    reviewed_at = NULL, review_notes = NULL, submitted_for_review_at = NULL,
		    published = false, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9
		RETURNING updated_at`

	return r.pool.QueryRow(ctx, query,
		test.Title, test.Description, test.SubjectID, test.TopicID,
		test.ExamStandard, test.Difficulty, test.TimeLimitMinutes,
		test.PassingScore, test.ID,
	).Scan(&test.UpdatedAt)
}

// PublishTest publishes a test making it available to students
func (r *TestRepository) PublishTest(ctx context.Context, testID int) error {
	var reviewStatus string
	if err := r.pool.QueryRow(ctx, `SELECT review_status FROM tests WHERE id = $1`, testID).Scan(&reviewStatus); err != nil {
		return err
	}
	if reviewStatus != "approved" {
		return ErrReviewApprovalRequired
	}
	query := `UPDATE tests SET published = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, testID)
	return err
}

// UnpublishTest unpublishes a test
func (r *TestRepository) UnpublishTest(ctx context.Context, testID int) error {
	query := `UPDATE tests SET published = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, testID)
	return err
}

// SubmitForReview marks a test as awaiting human review and records an audit event.
func (r *TestRepository) SubmitForReview(ctx context.Context, testID, actorID int) error {
	if _, err := r.pool.Exec(ctx, `
		UPDATE tests
		SET review_status = 'pending_review', submitted_for_review_at = CURRENT_TIMESTAMP,
		    reviewed_by = NULL, reviewed_at = NULL, review_notes = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`, testID); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO test_review_events (test_id, actor_id, decision) VALUES ($1, $2, 'submitted')`, testID, actorID)
	return err
}

// ApproveReview approves a test for publication and records reviewer metadata.
func (r *TestRepository) ApproveReview(ctx context.Context, testID, reviewerID int, notes string) error {
	if _, err := r.pool.Exec(ctx, `
		UPDATE tests
		SET review_status = 'approved', reviewed_by = $2, reviewed_at = CURRENT_TIMESTAMP,
		    review_notes = NULLIF($3, ''), updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`, testID, reviewerID, notes); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO test_review_events (test_id, actor_id, decision, notes) VALUES ($1, $2, 'approved', NULLIF($3, ''))`, testID, reviewerID, notes)
	return err
}

// RequestChanges moves a test back out of the approval state and records reviewer metadata.
func (r *TestRepository) RequestChanges(ctx context.Context, testID, reviewerID int, notes string) error {
	if _, err := r.pool.Exec(ctx, `
		UPDATE tests
		SET review_status = 'changes_requested', reviewed_by = $2, reviewed_at = CURRENT_TIMESTAMP,
		    review_notes = NULLIF($3, ''), published = false, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`, testID, reviewerID, notes); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO test_review_events (test_id, actor_id, decision, notes) VALUES ($1, $2, 'changes_requested', NULLIF($3, ''))`, testID, reviewerID, notes)
	return err
}

// DeleteTest deletes a test and all its questions
func (r *TestRepository) DeleteTest(ctx context.Context, testID int) error {
	query := `DELETE FROM tests WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, testID)
	return err
}

// GetRecommendation finds a published test in the same subject at the specified difficulty
// that the given user has not yet completed. Returns nil, pgx.ErrNoRows when none is found.
func (r *TestRepository) GetRecommendation(ctx context.Context, subjectID int, difficulty string, excludeTestID, userID int) (*models.Test, error) {
	var t models.Test
	var subjectIDOut *int
	var subjectName, subjectDesc *string

	err := r.pool.QueryRow(ctx, `
		SELECT t.id, t.title, COALESCE(t.description, ''), t.subject_id, t.difficulty,
		       t.time_limit_minutes, t.passing_score,
		       s.id, s.name, s.description,
		       (SELECT COUNT(*) FROM questions q WHERE q.test_id = t.id) AS question_count
		FROM tests t
		LEFT JOIN subjects s ON t.subject_id = s.id
		WHERE t.subject_id = $1
		  AND LOWER(t.difficulty) = LOWER($2)
		  AND t.published = true
		  AND t.id != $3
		  AND NOT EXISTS (
		    SELECT 1 FROM test_attempts ta
		    WHERE ta.test_id = t.id AND ta.user_id = $4 AND ta.status = 'completed'
		  )
		ORDER BY t.created_at DESC
		LIMIT 1`,
		subjectID, difficulty, excludeTestID, userID,
	).Scan(
		&t.ID, &t.Title, &t.Description, &t.SubjectID, &t.Difficulty,
		&t.TimeLimitMinutes, &t.PassingScore,
		&subjectIDOut, &subjectName, &subjectDesc,
		&t.QuestionCount,
	)
	if err != nil {
		return nil, err
	}

	if subjectIDOut != nil {
		t.Subject = &models.Subject{ID: *subjectIDOut, Name: *subjectName, Description: *subjectDesc}
	}
	return &t, nil
}

// UpdateTestNotes updates the notes filename for a test
func (r *TestRepository) UpdateTestNotes(ctx context.Context, testID int, notesFilename *string) error {
	query := `UPDATE tests SET notes_filename = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, notesFilename, testID)
	return err
}

// DeleteQuestion deletes a question
func (r *TestRepository) DeleteQuestion(ctx context.Context, questionID int) error {
	query := `DELETE FROM questions WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, questionID)
	return err
}

// UpdateQuestion updates a question
func (r *TestRepository) UpdateQuestion(ctx context.Context, question *models.Question) error {
	query := `
		UPDATE questions
		SET question_text = $1, image_url = $2, points = $3, question_order = $4, explanation = $5
		WHERE id = $6`

	_, err := r.pool.Exec(ctx, query,
		question.QuestionText, question.ImageURL, question.Points, question.QuestionOrder, question.Explanation, question.ID)
	return err
}

// UpdateAnswerOption updates an answer option
func (r *TestRepository) UpdateAnswerOption(ctx context.Context, option *models.AnswerOption) error {
	query := `
		UPDATE answer_options
		SET option_text = $1, is_correct = $2
		WHERE id = $3`

	_, err := r.pool.Exec(ctx, query, option.OptionText, option.IsCorrect, option.ID)
	return err
}

// DeleteAnswerOption deletes an answer option
func (r *TestRepository) DeleteAnswerOption(ctx context.Context, optionID int) error {
	query := `DELETE FROM answer_options WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, optionID)
	return err
}

// GetByID retrieves a test by ID with all its questions and options
func (r *TestRepository) GetByID(ctx context.Context, id int) (*models.Test, error) {
	// Get test details
	test := &models.Test{}
	query := `
		SELECT t.id, t.title, COALESCE(t.description, ''), t.subject_id, t.topic_id,
		       t.exam_standard, t.difficulty, t.time_limit_minutes,
		       t.passing_score, t.published, t.review_status, t.reviewed_by, t.reviewed_at,
		       t.review_notes, t.submitted_for_review_at, t.notes_filename, t.created_by, t.created_at, t.updated_at,
		       s.id, s.name, s.description
		FROM tests t
		LEFT JOIN subjects s ON t.subject_id = s.id
		WHERE t.id = $1`

	var subjectID *int
	var subjectName, subjectDesc *string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&test.ID, &test.Title, &test.Description, &test.SubjectID, &test.TopicID,
		&test.ExamStandard, &test.Difficulty, &test.TimeLimitMinutes,
		&test.PassingScore, &test.Published, &test.ReviewStatus, &test.ReviewedBy, &test.ReviewedAt,
		&test.ReviewNotes, &test.SubmittedForReviewAt, &test.NotesFilename, &test.CreatedBy, &test.CreatedAt, &test.UpdatedAt,
		&subjectID, &subjectName, &subjectDesc,
	)
	if err != nil {
		return nil, err
	}

	if subjectID != nil {
		test.Subject = &models.Subject{
			ID:          *subjectID,
			Name:        *subjectName,
			Description: *subjectDesc,
		}
	}

	// Get questions and their options
	questions, err := r.getQuestionsByTestID(ctx, id)
	if err != nil {
		return nil, err
	}
	test.Questions = questions

	return test, nil
}

// getQuestionsByTestID retrieves all questions for a test
func (r *TestRepository) getQuestionsByTestID(ctx context.Context, testID int) ([]models.Question, error) {
	query := `
		SELECT id, test_id, question_text, image_url, question_order, points, COALESCE(explanation, ''), created_at
		FROM questions
		WHERE test_id = $1
		ORDER BY question_order`

	rows, err := r.pool.Query(ctx, query, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []models.Question
	for rows.Next() {
		var q models.Question
		err := rows.Scan(&q.ID, &q.TestID, &q.QuestionText, &q.ImageURL,
			&q.QuestionOrder, &q.Points, &q.Explanation, &q.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Get options for this question
		options, err := r.getOptionsByQuestionID(ctx, q.ID)
		if err != nil {
			return nil, err
		}
		q.Options = options

		questions = append(questions, q)
	}

	return questions, rows.Err()
}

// getOptionsByQuestionID retrieves all answer options for a question
func (r *TestRepository) getOptionsByQuestionID(ctx context.Context, questionID int) ([]models.AnswerOption, error) {
	query := `
		SELECT id, question_id, option_text, is_correct, option_order, created_at
		FROM answer_options
		WHERE question_id = $1
		ORDER BY option_order`

	rows, err := r.pool.Query(ctx, query, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []models.AnswerOption
	for rows.Next() {
		var opt models.AnswerOption
		err := rows.Scan(&opt.ID, &opt.QuestionID, &opt.OptionText,
			&opt.IsCorrect, &opt.OptionOrder, &opt.CreatedAt)
		if err != nil {
			return nil, err
		}
		options = append(options, opt)
	}

	return options, rows.Err()
}

// Create creates a new test (used by admin/teacher)
func (r *TestRepository) Create(ctx context.Context, test *models.Test) error {
	query := `
		INSERT INTO tests (title, description, subject_id, topic_id, exam_standard,
		                   difficulty, time_limit_minutes, passing_score, review_status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'draft', $9)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		test.Title, test.Description, test.SubjectID, test.TopicID,
		test.ExamStandard, test.Difficulty, test.TimeLimitMinutes,
		test.PassingScore, test.CreatedBy,
	).Scan(&test.ID, &test.CreatedAt, &test.UpdatedAt)
}

// CreateQuestion creates a new question
func (r *TestRepository) CreateQuestion(ctx context.Context, question *models.Question) error {
	query := `
		INSERT INTO questions (test_id, question_text, image_url, question_order, points, explanation)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query,
		question.TestID, question.QuestionText, question.ImageURL,
		question.QuestionOrder, question.Points, question.Explanation,
	).Scan(&question.ID, &question.CreatedAt)
}

// CreateAnswerOption creates a new answer option
func (r *TestRepository) CreateAnswerOption(ctx context.Context, option *models.AnswerOption) error {
	query := `
		INSERT INTO answer_options (question_id, option_text, is_correct, option_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query,
		option.QuestionID, option.OptionText, option.IsCorrect, option.OptionOrder,
	).Scan(&option.ID, &option.CreatedAt)
}

// GetSubjects retrieves all subjects
func (r *TestRepository) GetSubjects(ctx context.Context) ([]models.Subject, error) {
	query := `SELECT id, name, description, created_at FROM subjects ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subjects []models.Subject
	for rows.Next() {
		var s models.Subject
		err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, s)
	}

	return subjects, rows.Err()
}

// GetOrCreateSubject gets or creates a subject by name
func (r *TestRepository) GetOrCreateSubject(ctx context.Context, name, description string) (int, error) {
	var id int

	// Try to get existing subject
	err := r.pool.QueryRow(ctx, "SELECT id FROM subjects WHERE name = $1", name).Scan(&id)
	if err == nil {
		return id, nil
	}

	// Create new subject
	err = r.pool.QueryRow(ctx,
		"INSERT INTO subjects (name, description) VALUES ($1, $2) RETURNING id",
		name, description,
	).Scan(&id)

	return id, err
}

// GetOrCreateTopic gets or creates a topic by name and subject
func (r *TestRepository) GetOrCreateTopic(ctx context.Context, subjectID int, name, description string) (int, error) {
	var id int

	// Try to get existing topic
	err := r.pool.QueryRow(ctx,
		"SELECT id FROM topics WHERE subject_id = $1 AND name = $2",
		subjectID, name,
	).Scan(&id)
	if err == nil {
		return id, nil
	}

	// Create new topic
	err = r.pool.QueryRow(ctx,
		"INSERT INTO topics (subject_id, name, description) VALUES ($1, $2, $3) RETURNING id",
		subjectID, name, description,
	).Scan(&id)

	return id, err
}

// GetByCreator retrieves all tests created by a specific user
func (r *TestRepository) GetByCreator(ctx context.Context, userID int) ([]models.Test, error) {
	query := `
		SELECT t.id, t.title, COALESCE(t.description, ''), t.subject_id, t.topic_id,
		       t.exam_standard, t.difficulty, t.time_limit_minutes,
		       t.passing_score, t.published, t.review_status, t.reviewed_by, t.reviewed_at,
		       t.review_notes, t.submitted_for_review_at, t.notes_filename, t.created_by, t.created_at, t.updated_at,
		       s.id, s.name, s.description,
		       (SELECT COUNT(*) FROM questions q WHERE q.test_id = t.id) AS question_count
		FROM tests t
		LEFT JOIN subjects s ON t.subject_id = s.id
		WHERE t.created_by = $1
		ORDER BY t.created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []models.Test
	for rows.Next() {
		var t models.Test
		var subjectID *int
		var subjectName, subjectDesc *string

		err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.SubjectID, &t.TopicID,
			&t.ExamStandard, &t.Difficulty, &t.TimeLimitMinutes,
			&t.PassingScore, &t.Published, &t.ReviewStatus, &t.ReviewedBy, &t.ReviewedAt,
			&t.ReviewNotes, &t.SubmittedForReviewAt, &t.NotesFilename, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
			&subjectID, &subjectName, &subjectDesc,
			&t.QuestionCount,
		)
		if err != nil {
			return nil, err
		}

		if subjectID != nil {
			t.Subject = &models.Subject{
				ID:          *subjectID,
				Name:        *subjectName,
				Description: *subjectDesc,
			}
		}

		tests = append(tests, t)
	}

	return tests, rows.Err()
}

// GetPendingReview retrieves all tests currently awaiting peer review, oldest first.
func (r *TestRepository) GetPendingReview(ctx context.Context) ([]models.Test, error) {
	query := `
		SELECT t.id, t.title, COALESCE(t.description, ''), t.subject_id, t.topic_id,
		       t.exam_standard, t.difficulty, t.time_limit_minutes,
		       t.passing_score, t.published, t.review_status, t.reviewed_by, t.reviewed_at,
		       t.review_notes, t.submitted_for_review_at, t.notes_filename, t.created_by, t.created_at, t.updated_at,
		       s.id, s.name, s.description,
		       (SELECT COUNT(*) FROM questions q WHERE q.test_id = t.id) AS question_count
		FROM tests t
		LEFT JOIN subjects s ON t.subject_id = s.id
		WHERE t.review_status = 'pending_review'
		ORDER BY t.submitted_for_review_at ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []models.Test
	for rows.Next() {
		var t models.Test
		var subjectID *int
		var subjectName, subjectDesc *string

		err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.SubjectID, &t.TopicID,
			&t.ExamStandard, &t.Difficulty, &t.TimeLimitMinutes,
			&t.PassingScore, &t.Published, &t.ReviewStatus, &t.ReviewedBy, &t.ReviewedAt,
			&t.ReviewNotes, &t.SubmittedForReviewAt, &t.NotesFilename, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
			&subjectID, &subjectName, &subjectDesc,
			&t.QuestionCount,
		)
		if err != nil {
			return nil, err
		}

		if subjectID != nil {
			t.Subject = &models.Subject{
				ID:          *subjectID,
				Name:        *subjectName,
				Description: *subjectDesc,
			}
		}

		tests = append(tests, t)
	}

	return tests, rows.Err()
}
