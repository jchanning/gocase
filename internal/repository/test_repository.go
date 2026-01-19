package repository

import (
	"context"

	"my-app/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestRepository handles test database operations
type TestRepository struct {
	pool *pgxpool.Pool
}

// NewTestRepository creates a new test repository
func NewTestRepository(pool *pgxpool.Pool) *TestRepository {
	return &TestRepository{pool: pool}
}

// GetAll retrieves all tests
func (r *TestRepository) GetAll(ctx context.Context) ([]models.Test, error) {
	query := `
		SELECT t.id, t.title, t.description, t.subject_id, t.topic_id,
		       t.exam_standard, t.difficulty, t.time_limit_minutes,
		       t.passing_score, t.created_by, t.created_at, t.updated_at,
		       s.id, s.name, s.description
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
			&t.PassingScore, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
			&subjectID, &subjectName, &subjectDesc,
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

// GetByID retrieves a test by ID with all its questions and options
func (r *TestRepository) GetByID(ctx context.Context, id int) (*models.Test, error) {
	// Get test details
	test := &models.Test{}
	query := `
		SELECT t.id, t.title, t.description, t.subject_id, t.topic_id,
		       t.exam_standard, t.difficulty, t.time_limit_minutes,
		       t.passing_score, t.created_by, t.created_at, t.updated_at,
		       s.id, s.name, s.description
		FROM tests t
		LEFT JOIN subjects s ON t.subject_id = s.id
		WHERE t.id = $1`

	var subjectID *int
	var subjectName, subjectDesc *string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&test.ID, &test.Title, &test.Description, &test.SubjectID, &test.TopicID,
		&test.ExamStandard, &test.Difficulty, &test.TimeLimitMinutes,
		&test.PassingScore, &test.CreatedBy, &test.CreatedAt, &test.UpdatedAt,
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
		SELECT id, test_id, question_text, image_url, question_order, points, created_at
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
			&q.QuestionOrder, &q.Points, &q.CreatedAt)
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
		                   difficulty, time_limit_minutes, passing_score, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
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
		INSERT INTO questions (test_id, question_text, image_url, question_order, points)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query,
		question.TestID, question.QuestionText, question.ImageURL,
		question.QuestionOrder, question.Points,
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
		SELECT t.id, t.title, t.description, t.subject_id, t.topic_id,
		       t.exam_standard, t.difficulty, t.time_limit_minutes,
		       t.passing_score, t.created_by, t.created_at, t.updated_at,
		       s.id, s.name, s.description
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
			&t.PassingScore, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
			&subjectID, &subjectName, &subjectDesc,
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
