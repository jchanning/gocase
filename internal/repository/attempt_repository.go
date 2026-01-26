package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"my-app/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AttemptRepository handles test attempt database operations
type AttemptRepository struct {
	pool *pgxpool.Pool
}

// StreakStats represents current and best streak values.
type StreakStats struct {
	Current int
	Best    int
}

// AttemptSearchFilter describes optional filters for querying attempts.
type AttemptSearchFilter struct {
	UserID      *int
	StudentName string
	TestName    string
	DateFrom    *time.Time
	DateTo      *time.Time
	ScoreMin    *int
	ScoreMax    *int
}

// NewAttemptRepository creates a new attempt repository
func NewAttemptRepository(pool *pgxpool.Pool) *AttemptRepository {
	return &AttemptRepository{pool: pool}
}

// Create creates a new test attempt
func (r *AttemptRepository) Create(ctx context.Context, attempt *models.TestAttempt) error {
	query := `
		INSERT INTO test_attempts (user_id, test_id, started_at, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query,
		attempt.UserID, attempt.TestID, attempt.StartedAt, attempt.Status,
	).Scan(&attempt.ID, &attempt.CreatedAt)
}

// GetByID retrieves an attempt by ID
func (r *AttemptRepository) GetByID(ctx context.Context, id int) (*models.TestAttempt, error) {
	attempt := &models.TestAttempt{}
	query := `
		SELECT id, user_id, test_id, started_at, completed_at, score,
		       total_points, time_taken_seconds, status, created_at
		FROM test_attempts
		WHERE id = $1`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&attempt.ID, &attempt.UserID, &attempt.TestID, &attempt.StartedAt,
		&attempt.CompletedAt, &attempt.Score, &attempt.TotalPoints,
		&attempt.TimeTakenSeconds, &attempt.Status, &attempt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return attempt, nil
}

// Complete marks an attempt as completed with score
func (r *AttemptRepository) Complete(ctx context.Context, attemptID, score, totalPoints int) error {
	query := `
		UPDATE test_attempts
		SET completed_at = $2, score = $3, total_points = $4,
		    time_taken_seconds = EXTRACT(EPOCH FROM ($2 - started_at))::INTEGER,
		    status = 'completed'
		WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, attemptID, time.Now(), score, totalPoints)
	return err
}

// SaveAnswer saves a student's answer to a question
func (r *AttemptRepository) SaveAnswer(ctx context.Context, answer *models.StudentAnswer) error {
	query := `
		INSERT INTO student_answers (attempt_id, question_id, selected_option_id, is_correct)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (attempt_id, question_id)
		DO UPDATE SET selected_option_id = $3, is_correct = $4, answered_at = CURRENT_TIMESTAMP
		RETURNING id, answered_at`

	return r.pool.QueryRow(ctx, query,
		answer.AttemptID, answer.QuestionID, answer.SelectedOptionID, answer.IsCorrect,
	).Scan(&answer.ID, &answer.AnsweredAt)
}

// GetAnswersByAttemptID retrieves all answers for an attempt
func (r *AttemptRepository) GetAnswersByAttemptID(ctx context.Context, attemptID int) ([]models.StudentAnswer, error) {
	query := `
		SELECT sa.id, sa.attempt_id, sa.question_id, sa.selected_option_id,
		       sa.is_correct, sa.answered_at
		FROM student_answers sa
		WHERE sa.attempt_id = $1
		ORDER BY sa.question_id`

	rows, err := r.pool.Query(ctx, query, attemptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []models.StudentAnswer
	for rows.Next() {
		var a models.StudentAnswer
		err := rows.Scan(&a.ID, &a.AttemptID, &a.QuestionID,
			&a.SelectedOptionID, &a.IsCorrect, &a.AnsweredAt)
		if err != nil {
			return nil, err
		}
		answers = append(answers, a)
	}

	return answers, rows.Err()
}

// GetUserAttempts retrieves all test attempts for a user
func (r *AttemptRepository) GetUserAttempts(ctx context.Context, userID int, limit int) ([]models.TestAttempt, error) {
	query := `
		SELECT ta.id, ta.user_id, ta.test_id, ta.started_at, ta.completed_at,
		       ta.score, ta.total_points, ta.time_taken_seconds, ta.status, ta.created_at,
		       t.id, t.title, t.description, t.subject_id, t.topic_id,
		       t.exam_standard, t.difficulty, t.time_limit_minutes, t.passing_score,
		       t.created_by, t.created_at, t.updated_at,
		       s.id, s.name, s.description,
		       tp.id, tp.name, tp.description
		FROM test_attempts ta
		JOIN tests t ON ta.test_id = t.id
		LEFT JOIN subjects s ON t.subject_id = s.id
		LEFT JOIN topics tp ON t.topic_id = tp.id
		WHERE ta.user_id = $1
		ORDER BY ta.created_at DESC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attempts []models.TestAttempt
	for rows.Next() {
		var a models.TestAttempt
		a.Test = &models.Test{}

		// Variables for nullable subject and topic
		var subjectID, topicID *int
		var subjectName, subjectDesc, topicName, topicDesc *string

		err := rows.Scan(
			&a.ID, &a.UserID, &a.TestID, &a.StartedAt, &a.CompletedAt,
			&a.Score, &a.TotalPoints, &a.TimeTakenSeconds, &a.Status, &a.CreatedAt,
			&a.Test.ID, &a.Test.Title, &a.Test.Description, &a.Test.SubjectID,
			&a.Test.TopicID, &a.Test.ExamStandard, &a.Test.Difficulty,
			&a.Test.TimeLimitMinutes, &a.Test.PassingScore, &a.Test.CreatedBy,
			&a.Test.CreatedAt, &a.Test.UpdatedAt,
			&subjectID, &subjectName, &subjectDesc,
			&topicID, &topicName, &topicDesc,
		)
		if err != nil {
			return nil, err
		}

		// Populate subject if present
		if subjectID != nil && subjectName != nil {
			a.Test.Subject = &models.Subject{
				ID:          *subjectID,
				Name:        *subjectName,
				Description: *subjectDesc,
			}
		}

		// Populate topic if present
		if topicID != nil && topicName != nil {
			a.Test.Topic = &models.Topic{
				ID:          *topicID,
				Name:        *topicName,
				Description: *topicDesc,
			}
		}

		attempts = append(attempts, a)
	}

	return attempts, rows.Err()
}

// GetUserStreakStats calculates the current and best streak for a user based on completed attempts.
// Streak is counted in whole days; completing at least one test on a day extends the streak.
func (r *AttemptRepository) GetUserStreakStats(ctx context.Context, userID int) (StreakStats, error) {
	stats := StreakStats{}

	// Fetch distinct completion dates (UTC) in descending order
	query := `
		SELECT DISTINCT DATE(completed_at) as day
		FROM test_attempts
		WHERE user_id = $1 AND status = 'completed' AND completed_at IS NOT NULL
		ORDER BY day DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	var prevDay *time.Time
	for rows.Next() {
		var day time.Time
		if err := rows.Scan(&day); err != nil {
			return stats, err
		}

		if prevDay == nil {
			// First day initializes streaks
			stats.Current = 1
			stats.Best = 1
		} else {
			// Check if this day is consecutive to previous day
			diff := prevDay.Sub(day).Hours() / 24
			if diff == 1 {
				stats.Current++
				if stats.Current > stats.Best {
					stats.Best = stats.Current
				}
			} else {
				// Streak broken; only update best if needed
				if stats.Current > stats.Best {
					stats.Best = stats.Current
				}
				// Reset current streak to 1 (this day counts as new streak)
				stats.Current = 1
			}
		}

		prevDay = &day
	}

	// Ensure best is at least current if only one day or all consecutive
	if stats.Best < stats.Current {
		stats.Best = stats.Current
	}

	return stats, rows.Err()
}

// GetUserTestStats retrieves statistics for a user's performance on tests
func (r *AttemptRepository) GetUserTestStats(ctx context.Context, userID int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get total attempts
	var totalAttempts int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM test_attempts WHERE user_id = $1 AND status = 'completed'",
		userID,
	).Scan(&totalAttempts)
	if err != nil {
		return nil, err
	}
	stats["total_attempts"] = totalAttempts

	// Get average score
	var avgScore float64
	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(AVG(CAST(score AS FLOAT) / NULLIF(total_points, 0) * 100), 0)
		 FROM test_attempts
		 WHERE user_id = $1 AND status = 'completed' AND total_points > 0`,
		userID,
	).Scan(&avgScore)
	if err != nil {
		return nil, err
	}
	stats["average_score"] = avgScore

	// Get recent improvement trend (last 5 vs previous 5)
	query := `
		WITH recent AS (
			SELECT AVG(CAST(score AS FLOAT) / NULLIF(total_points, 0) * 100) as avg_recent
			FROM (
				SELECT score, total_points
				FROM test_attempts
				WHERE user_id = $1 AND status = 'completed' AND total_points > 0
				ORDER BY completed_at DESC
				LIMIT 5
			) r
		),
		previous AS (
			SELECT AVG(CAST(score AS FLOAT) / NULLIF(total_points, 0) * 100) as avg_previous
			FROM (
				SELECT score, total_points
				FROM test_attempts
				WHERE user_id = $1 AND status = 'completed' AND total_points > 0
				ORDER BY completed_at DESC
				LIMIT 10 OFFSET 5
			) p
		)
		SELECT COALESCE(recent.avg_recent, 0), COALESCE(previous.avg_previous, 0)
		FROM recent, previous`

	var recentAvg, previousAvg float64
	err = r.pool.QueryRow(ctx, query, userID).Scan(&recentAvg, &previousAvg)
	if err != nil {
		return nil, err
	}
	stats["recent_average"] = recentAvg
	stats["improvement"] = recentAvg - previousAvg

	return stats, nil
}

// GetByTestID retrieves all attempts for a specific test
func (r *AttemptRepository) GetByTestID(ctx context.Context, testID int) ([]models.TestAttempt, error) {
	query := `
		SELECT id, user_id, test_id, started_at, completed_at, score,
		       total_points, time_taken_seconds, status, created_at
		FROM test_attempts
		WHERE test_id = $1
		ORDER BY started_at DESC`

	rows, err := r.pool.Query(ctx, query, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attempts []models.TestAttempt
	for rows.Next() {
		var attempt models.TestAttempt
		err := rows.Scan(
			&attempt.ID, &attempt.UserID, &attempt.TestID, &attempt.StartedAt,
			&attempt.CompletedAt, &attempt.Score, &attempt.TotalPoints,
			&attempt.TimeTakenSeconds, &attempt.Status, &attempt.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, attempt)
	}

	return attempts, rows.Err()
}

// SearchAttempts returns attempts filtered by the provided criteria and includes user/test metadata.
func (r *AttemptRepository) SearchAttempts(ctx context.Context, filter AttemptSearchFilter) ([]models.TestAttempt, error) {
	var (
		builder strings.Builder
		args    []interface{}
	)

	builder.WriteString(`SELECT ta.id, ta.user_id, ta.test_id, ta.started_at, ta.completed_at, ta.score,
	       ta.total_points, ta.time_taken_seconds, ta.status, ta.created_at,
	       u.id, u.username, u.email, u.role,
	       t.id, t.title, t.exam_standard, t.difficulty
	FROM test_attempts ta
	JOIN users u ON ta.user_id = u.id
	JOIN tests t ON ta.test_id = t.id
	WHERE 1=1`)

	if filter.UserID != nil {
		args = append(args, *filter.UserID)
		builder.WriteString(fmt.Sprintf(" AND ta.user_id = $%d", len(args)))
	}

	if filter.StudentName != "" {
		args = append(args, "%"+filter.StudentName+"%")
		builder.WriteString(fmt.Sprintf(" AND u.username ILIKE $%d", len(args)))
	}

	if filter.TestName != "" {
		args = append(args, "%"+filter.TestName+"%")
		builder.WriteString(fmt.Sprintf(" AND t.title ILIKE $%d", len(args)))
	}

	if filter.DateFrom != nil {
		args = append(args, *filter.DateFrom)
		builder.WriteString(fmt.Sprintf(" AND COALESCE(ta.completed_at, ta.started_at) >= $%d", len(args)))
	}

	if filter.DateTo != nil {
		args = append(args, *filter.DateTo)
		builder.WriteString(fmt.Sprintf(" AND COALESCE(ta.completed_at, ta.started_at) <= $%d", len(args)))
	}

	if filter.ScoreMin != nil {
		args = append(args, *filter.ScoreMin)
		builder.WriteString(fmt.Sprintf(" AND ta.score >= $%d", len(args)))
	}

	if filter.ScoreMax != nil {
		args = append(args, *filter.ScoreMax)
		builder.WriteString(fmt.Sprintf(" AND ta.score <= $%d", len(args)))
	}

	builder.WriteString(" ORDER BY COALESCE(ta.completed_at, ta.started_at) DESC")

	rows, err := r.pool.Query(ctx, builder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attempts []models.TestAttempt
	for rows.Next() {
		var attempt models.TestAttempt
		attempt.User = &models.User{}
		attempt.Test = &models.Test{}

		err := rows.Scan(
			&attempt.ID, &attempt.UserID, &attempt.TestID, &attempt.StartedAt, &attempt.CompletedAt,
			&attempt.Score, &attempt.TotalPoints, &attempt.TimeTakenSeconds, &attempt.Status, &attempt.CreatedAt,
			&attempt.User.ID, &attempt.User.Username, &attempt.User.Email, &attempt.User.Role,
			&attempt.Test.ID, &attempt.Test.Title, &attempt.Test.ExamStandard, &attempt.Test.Difficulty,
		)
		if err != nil {
			return nil, err
		}

		attempts = append(attempts, attempt)
	}

	return attempts, rows.Err()
}
