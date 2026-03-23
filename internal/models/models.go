package models

import "time"

// Valid difficulty levels
var ValidDifficulties = []string{"Easy", "Medium", "Hard"}

// Valid exam standards
var ValidExamStandards = []string{"Primary", "Secondary", "GCSE", "IGCSE", "A-Level"}

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
	ID                   int        `json:"id"`
	Title                string     `json:"title"`
	Description          string     `json:"description"`
	SubjectID            *int       `json:"subject_id"`
	TopicID              *int       `json:"topic_id"`
	ExamStandard         string     `json:"exam_standard"` // GCSE, A-Level, Primary, Secondary
	Difficulty           string     `json:"difficulty"`    // Easy, Medium, Hard
	TimeLimitMinutes     int        `json:"time_limit_minutes"`
	PassingScore         int        `json:"passing_score"`
	Published            bool       `json:"published"`
	ReviewStatus         string     `json:"review_status"`
	ReviewedBy           *int       `json:"reviewed_by"`
	ReviewedAt           *time.Time `json:"reviewed_at"`
	ReviewNotes          *string    `json:"review_notes"`
	SubmittedForReviewAt *time.Time `json:"submitted_for_review_at"`
	NotesFilename        *string    `json:"notes_filename"`
	CreatedBy            *int       `json:"created_by"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	// Related data (not in DB, populated via joins)
	Subject       *Subject   `json:"subject,omitempty"`
	Topic         *Topic     `json:"topic,omitempty"`
	Questions     []Question `json:"questions,omitempty"`
	QuestionCount int        `json:"question_count,omitempty"`
}

// TestReviewEvent records an audit event in the content review workflow.
type TestReviewEvent struct {
	ID        int       `json:"id"`
	TestID    int       `json:"test_id"`
	ActorID   *int      `json:"actor_id"`
	Decision  string    `json:"decision"`
	Notes     *string   `json:"notes"`
	CreatedAt time.Time `json:"created_at"`

	Actor *User `json:"actor,omitempty"`
}

// TestFeedbackIssue captures a student-reported question or explanation issue.
type TestFeedbackIssue struct {
	ID             int        `json:"id"`
	TestID         int        `json:"test_id"`
	QuestionID     int        `json:"question_id"`
	AttemptID      int        `json:"attempt_id"`
	ReportedBy     int        `json:"reported_by"`
	IssueType      string     `json:"issue_type"`
	StudentComment string     `json:"student_comment"`
	Status         string     `json:"status"`
	ReviewResponse *string    `json:"review_response"`
	ReviewedBy     *int       `json:"reviewed_by"`
	ReviewedAt     *time.Time `json:"reviewed_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	Test           *Test     `json:"test,omitempty"`
	Question       *Question `json:"question,omitempty"`
	Reporter       *User     `json:"reporter,omitempty"`
	Reviewer       *User     `json:"reviewer,omitempty"`
	ReportedByName string    `json:"reported_by_name,omitempty"`
	ReviewedByName string    `json:"reviewed_by_name,omitempty"`
}

// Question represents a single question in a test
type Question struct {
	ID            int       `json:"id"`
	TestID        int       `json:"test_id"`
	QuestionText  string    `json:"question_text"`
	ImageURL      *string   `json:"image_url"`
	QuestionOrder int       `json:"question_order"`
	Points        int       `json:"points"`
	Explanation   string    `json:"explanation"`
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

// TestAssignment represents a test assigned by a teacher to a student with a deadline.
type TestAssignment struct {
	ID         int       `json:"id"`
	TestID     int       `json:"test_id"`
	AssignedBy *int      `json:"assigned_by"`
	AssignedTo int       `json:"assigned_to"`
	DueDate    time.Time `json:"due_date"`
	Status     string    `json:"status"` // pending, completed, overdue
	AssignedAt time.Time `json:"assigned_at"`

	// Related data
	Test     *Test `json:"test,omitempty"`
	Student  *User `json:"student,omitempty"`
	Assigner *User `json:"assigner,omitempty"`
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
	NotesFilename    string           `json:"notes_filename,omitempty"`
}

// QuestionUpload represents a question for upload
type QuestionUpload struct {
	QuestionText string   `json:"question_text"`
	ImageURL     string   `json:"image_url,omitempty"`
	Points       int      `json:"points"`
	Options      []string `json:"options"`       // Array of 4 options
	CorrectIndex int      `json:"correct_index"` // 0-3, which option is correct
	Explanation  string   `json:"explanation,omitempty"`
}

// ----------------------------------------------------------------
// Syllabus & Revision Planner
// ----------------------------------------------------------------

// Syllabus represents an authoritative topic list for a subject at an exam level.
type Syllabus struct {
	ID           int       `json:"id"`
	SubjectID    *int      `json:"subject_id"`
	ExamStandard string    `json:"exam_standard"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	IsPublished  bool      `json:"is_published"`
	CreatedBy    *int      `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Related
	Subject    *Subject          `json:"subject,omitempty"`
	Sections   []SyllabusSection `json:"sections,omitempty"`
	TopicCount int               `json:"topic_count,omitempty"`
}

// SyllabusSection is a grouping of topics within a syllabus.
type SyllabusSection struct {
	ID           int       `json:"id"`
	SyllabusID   int       `json:"syllabus_id"`
	Title        string    `json:"title"`
	SectionOrder int       `json:"section_order"`
	CreatedAt    time.Time `json:"created_at"`

	Topics []SyllabusTopic `json:"topics,omitempty"`
}

// SyllabusTopic is a single curriculum entry within a syllabus section.
type SyllabusTopic struct {
	ID             int       `json:"id"`
	SyllabusID     int       `json:"syllabus_id"`
	SectionID      *int      `json:"section_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	EstimatedHours float64   `json:"estimated_hours"`
	TopicOrder     int       `json:"topic_order"`
	NotesContent   string    `json:"notes_content"`
	CreatedAt      time.Time `json:"created_at"`

	Tests []Test `json:"tests,omitempty"`
}

// RevisionPlan is a student's study schedule built from a syllabus.
type RevisionPlan struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	SyllabusID  int       `json:"syllabus_id"`
	ExamDate    time.Time `json:"exam_date"`
	HoursPerDay float64   `json:"hours_per_day"`
	StudyDays   []int     `json:"study_days"` // 0=Sun…6=Sat, stored as JSON TEXT in DB
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Related / computed
	Syllabus       *Syllabus         `json:"syllabus,omitempty"`
	Sessions       []RevisionSession `json:"sessions,omitempty"`
	DaysUntilExam  int               `json:"days_until_exam"`
	TotalSessions  int               `json:"total_sessions"`
	CompletedCount int               `json:"completed_count"`
	Progress       int               `json:"progress"` // 0-100, percent of sessions completed
}

// RevisionSession is a single scheduled study block within a revision plan.
type RevisionSession struct {
	ID              int        `json:"id"`
	PlanID          int        `json:"plan_id"`
	SessionDate     time.Time  `json:"session_date"`
	SyllabusTopicID int        `json:"syllabus_topic_id"`
	HoursAllocated  float64    `json:"hours_allocated"`
	Status          string     `json:"status"` // scheduled, completed, skipped
	Notes           string     `json:"notes"`
	CompletedAt     *time.Time `json:"completed_at"`
	CreatedAt       time.Time  `json:"created_at"`

	Topic *SyllabusTopic `json:"topic,omitempty"`
}
