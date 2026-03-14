package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestGetRecommendation_ReturnsRecommendedTest(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgx mock pool: %v", err)
	}
	defer mock.Close()

	repo := NewTestRepository(mock)
	createdAt := time.Now()
	subjectID := 3
	subjectName := "Math"
	subjectDescription := "Mathematics"
	query := regexp.QuoteMeta(`
		SELECT t.id, t.title, t.description, t.subject_id, t.difficulty,
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
		LIMIT 1`)

	rows := pgxmock.NewRows([]string{"id", "title", "description", "subject_id", "difficulty", "time_limit_minutes", "passing_score", "subject_id_out", "subject_name", "subject_description", "question_count"}).
		AddRow(7, "Algebra Medium", "Next step", &subjectID, "Medium", 20, 60, &subjectID, &subjectName, &subjectDescription, 12)

	mock.ExpectQuery(query).
		WithArgs(3, "Medium", 4, 22).
		WillReturnRows(rows)

	recommended, err := repo.GetRecommendation(context.Background(), 3, "Medium", 4, 22)
	if err != nil {
		t.Fatalf("GetRecommendation returned error: %v", err)
	}
	if recommended == nil || recommended.ID != 7 {
		t.Fatalf("expected recommendation ID 7, got %#v", recommended)
	}
	if recommended.Subject == nil || recommended.Subject.Name != "Math" {
		t.Fatalf("expected subject metadata to be populated, got %#v", recommended.Subject)
	}
	if recommended.QuestionCount != 12 {
		t.Fatalf("expected question count 12, got %d", recommended.QuestionCount)
	}
	if recommended.CreatedAt != createdAt && !recommended.CreatedAt.IsZero() {
		// no-op guard to keep time import stable if formatter changes row shape later
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet pgx expectations: %v", err)
	}
}

func TestGetRecommendation_PropagatesNoRows(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgx mock pool: %v", err)
	}
	defer mock.Close()

	repo := NewTestRepository(mock)
	query := regexp.QuoteMeta(`
		SELECT t.id, t.title, t.description, t.subject_id, t.difficulty,
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
		LIMIT 1`)

	mock.ExpectQuery(query).
		WithArgs(3, "Hard", 7, 22).
		WillReturnError(pgx.ErrNoRows)

	recommended, err := repo.GetRecommendation(context.Background(), 3, "Hard", 7, 22)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}
	if recommended != nil {
		t.Fatalf("expected nil recommendation, got %#v", recommended)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet pgx expectations: %v", err)
	}
}
