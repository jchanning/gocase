package models

import "time"

// User represents a user in the system (student, teacher, or admin)
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose password hash in JSON
	Username     string    `json:"username"`
	Role         string    `json:"role"` // student, teacher, admin
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Subject represents a subject category
type Subject struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Topic represents a specific topic within a subject
type Topic struct {
	ID          int       `json:"id"`
	SubjectID   int       `json:"subject_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Test represents a complete test/exam
type Test struct {
	ID               int       `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	SubjectID        *int      `json:"subject_id"`
	TopicID          *int      `json:"topic_id"`
	ExamStandard     string    `json:"exam_standard"` // GCSE, A-Level, Primary, Secondary
	Difficulty       string    `json:"difficulty"`    // Easy, Medium, Hard
	TimeLimitMinutes int       `json:"time_limit_minutes"`
	PassingScore     int       `json:"passing_score"`
	Published        bool      `json:"published"`
	NotesFilename    *string   `json:"notes_filename"`
	CreatedBy        *int      `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Related data (not in DB, populated via joins)
	Subject   *Subject   `json:"subject,omitempty"`
	Topic     *Topic     `json:"topic,omitempty"`
	Questions []Question `json:"questions,omitempty"`
}

// Question represents a single question in a test
type Question struct {
	ID            int       `json:"id"`
	TestID        int       `json:"test_id"`
	QuestionText  string    `json:"question_text"`
	ImageURL      *string   `json:"image_url"`
	QuestionOrder int       `json:"question_order"`
	Points        int       `json:"points"`
	CreatedAt     time.Time `json:"created_at"`

	// Related data
	Options []AnswerOption `json:"options,omitempty"`
}

// AnswerOption represents one of four possible answers
type AnswerOption struct {
	ID          int       `json:"id"`
	QuestionID  int       `json:"question_id"`
	OptionText  string    `json:"option_text"`
	IsCorrect   bool      `json:"is_correct,omitempty"` // Only shown to teachers/admins or after test
	OptionOrder int       `json:"option_order"`
	CreatedAt   time.Time `json:"created_at"`
}

// TestAttempt represents a student's attempt at a test
type TestAttempt struct {
	ID               int        `json:"id"`
	UserID           int        `json:"user_id"`
	TestID           int        `json:"test_id"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	Score            *int       `json:"score"`
	TotalPoints      *int       `json:"total_points"`
	TimeTakenSeconds *int       `json:"time_taken_seconds"`
	Status           string     `json:"status"` // in_progress, completed, abandoned
	CreatedAt        time.Time  `json:"created_at"`

	// Related data
	Test    *Test           `json:"test,omitempty"`
	Answers []StudentAnswer `json:"answers,omitempty"`
	User    *User           `json:"user,omitempty"`
}

// StudentAnswer represents a student's answer to a question
type StudentAnswer struct {
	ID               int       `json:"id"`
	AttemptID        int       `json:"attempt_id"`
	QuestionID       int       `json:"question_id"`
	SelectedOptionID *int      `json:"selected_option_id"`
	IsCorrect        *bool     `json:"is_correct"`
	AnsweredAt       time.Time `json:"answered_at"`

	// Related data
	Question       *Question     `json:"question,omitempty"`
	SelectedOption *AnswerOption `json:"selected_option,omitempty"`
}

// Achievement represents a badge or achievement
type Achievement struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	BadgeIcon     string    `json:"badge_icon"`
	CriteriaType  string    `json:"criteria_type"`
	CriteriaValue int       `json:"criteria_value"`
	PointsAwarded int       `json:"points_awarded"`
	CreatedAt     time.Time `json:"created_at"`
}

// UserAchievement represents an achievement earned by a user
type UserAchievement struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	AchievementID int       `json:"achievement_id"`
	EarnedAt      time.Time `json:"earned_at"`

	// Related data
	Achievement *Achievement `json:"achievement,omitempty"`
}

// UserStats represents a user's overall statistics
type UserStats struct {
	UserID         int       `json:"user_id"`
	TotalPoints    int       `json:"total_points"`
	TestsCompleted int       `json:"tests_completed"`
	TestsPassed    int       `json:"tests_passed"`
	CurrentStreak  int       `json:"current_streak"`
	BestStreak     int       `json:"best_streak"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TestUpload represents the structure for uploading tests via JSON/CSV
type TestUpload struct {
	Title            string           `json:"title"`
	Description      string           `json:"description"`
	Subject          string           `json:"subject"`
	Topic            string           `json:"topic"`
	ExamStandard     string           `json:"exam_standard"`
	Difficulty       string           `json:"difficulty"`
	TimeLimitMinutes int              `json:"time_limit_minutes"`
	PassingScore     int              `json:"passing_score"`
	Questions        []QuestionUpload `json:"questions"`
}

// QuestionUpload represents a question for upload
type QuestionUpload struct {
	QuestionText string   `json:"question_text"`
	ImageURL     string   `json:"image_url,omitempty"`
	Points       int      `json:"points"`
	Options      []string `json:"options"`       // Array of 4 options
	CorrectIndex int      `json:"correct_index"` // 0-3, which option is correct
}
