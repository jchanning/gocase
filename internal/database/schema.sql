-- GoCaSE Database Schema
-- Multiple Choice Question Test Application

-- User Roles: student, teacher, admin
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    username VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'student' CHECK (role IN ('student', 'teacher', 'admin')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Subjects (Math, Science, History, etc.)
CREATE TABLE IF NOT EXISTS subjects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Topics within subjects (Algebra, Biology, World War II, etc.)
CREATE TABLE IF NOT EXISTS topics (
    id SERIAL PRIMARY KEY,
    subject_id INTEGER REFERENCES subjects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(subject_id, name)
);

-- Tests/Exams
CREATE TABLE IF NOT EXISTS tests (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    subject_id INTEGER REFERENCES subjects(id) ON DELETE SET NULL,
    topic_id INTEGER REFERENCES topics(id) ON DELETE SET NULL,
    exam_standard VARCHAR(50) NOT NULL CHECK (exam_standard IN ('GCSE', 'IGCSE', 'A-Level', 'Primary', 'Secondary')),
    difficulty VARCHAR(20) NOT NULL CHECK (difficulty IN ('Easy', 'Medium', 'Hard')),
    time_limit_minutes INTEGER NOT NULL DEFAULT 10,
    passing_score INTEGER NOT NULL DEFAULT 60,
    published BOOLEAN DEFAULT FALSE,
    review_status VARCHAR(30) NOT NULL DEFAULT 'draft' CHECK (review_status IN ('draft', 'pending_review', 'approved', 'changes_requested')),
    reviewed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP,
    review_notes TEXT,
    submitted_for_review_at TIMESTAMP,
    notes_filename VARCHAR(500),
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE tests ADD COLUMN IF NOT EXISTS review_status VARCHAR(30) NOT NULL DEFAULT 'draft';
ALTER TABLE tests ADD COLUMN IF NOT EXISTS reviewed_by INTEGER REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE tests ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMP;
ALTER TABLE tests ADD COLUMN IF NOT EXISTS review_notes TEXT;
ALTER TABLE tests ADD COLUMN IF NOT EXISTS submitted_for_review_at TIMESTAMP;
ALTER TABLE tests DROP CONSTRAINT IF EXISTS tests_review_status_check;
ALTER TABLE tests ADD CONSTRAINT tests_review_status_check CHECK (review_status IN ('draft', 'pending_review', 'approved', 'changes_requested'));
UPDATE tests
SET review_status = 'approved', review_notes = COALESCE(review_notes, 'Legacy published test auto-approved during review workflow rollout.')
WHERE published = TRUE AND (review_status IS NULL OR review_status = 'draft');

CREATE TABLE IF NOT EXISTS test_review_events (
    id SERIAL PRIMARY KEY,
    test_id INTEGER NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    actor_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    decision VARCHAR(30) NOT NULL CHECK (decision IN ('submitted', 'approved', 'changes_requested', 'reset_to_draft')),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS test_feedback_issues (
    id SERIAL PRIMARY KEY,
    test_id INTEGER NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    question_id INTEGER NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    attempt_id INTEGER NOT NULL REFERENCES test_attempts(id) ON DELETE CASCADE,
    reported_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    issue_type VARCHAR(40) NOT NULL CHECK (issue_type IN ('incorrect_answer', 'unclear_explanation', 'question_text_issue', 'other')),
    student_comment TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'in_review', 'resolved', 'dismissed')),
    review_response TEXT,
    reviewed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Questions
CREATE TABLE IF NOT EXISTS questions (
    id SERIAL PRIMARY KEY,
    test_id INTEGER REFERENCES tests(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    image_url VARCHAR(500),
    question_order INTEGER NOT NULL,
    points INTEGER NOT NULL DEFAULT 1,
    explanation TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(test_id, question_order)
);
-- Migrate existing installations
ALTER TABLE questions ADD COLUMN IF NOT EXISTS explanation TEXT;

-- Answer Options (4 per question)
CREATE TABLE IF NOT EXISTS answer_options (
    id SERIAL PRIMARY KEY,
    question_id INTEGER REFERENCES questions(id) ON DELETE CASCADE,
    option_text TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    option_order INTEGER NOT NULL CHECK (option_order BETWEEN 1 AND 4),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(question_id, option_order)
);

-- Student Test Attempts
CREATE TABLE IF NOT EXISTS test_attempts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    test_id INTEGER REFERENCES tests(id) ON DELETE CASCADE,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    score INTEGER,
    total_points INTEGER,
    time_taken_seconds INTEGER,
    status VARCHAR(20) DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'completed', 'abandoned')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Student Answers
CREATE TABLE IF NOT EXISTS student_answers (
    id SERIAL PRIMARY KEY,
    attempt_id INTEGER REFERENCES test_attempts(id) ON DELETE CASCADE,
    question_id INTEGER REFERENCES questions(id) ON DELETE CASCADE,
    selected_option_id INTEGER REFERENCES answer_options(id) ON DELETE SET NULL,
    is_correct BOOLEAN,
    answered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(attempt_id, question_id)
);

-- Achievements/Badges
CREATE TABLE IF NOT EXISTS achievements (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    badge_icon VARCHAR(50),
    criteria_type VARCHAR(50) NOT NULL CHECK (criteria_type IN ('tests_completed', 'perfect_score', 'streak', 'high_score', 'subject_master')),
    criteria_value INTEGER NOT NULL,
    points_awarded INTEGER NOT NULL DEFAULT 10,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User Achievements
CREATE TABLE IF NOT EXISTS user_achievements (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    achievement_id INTEGER REFERENCES achievements(id) ON DELETE CASCADE,
    earned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, achievement_id)
);

-- Test Assignments (teacher assigns a test to a student with a deadline)
CREATE TABLE IF NOT EXISTS test_assignments (
    id SERIAL PRIMARY KEY,
    test_id INTEGER NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    assigned_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    assigned_to INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    due_date TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'overdue')),
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_assignments_assigned_to ON test_assignments(assigned_to);
CREATE INDEX IF NOT EXISTS idx_assignments_test ON test_assignments(test_id);
CREATE INDEX IF NOT EXISTS idx_assignments_assigned_by ON test_assignments(assigned_by);

-- User Points/Scores
CREATE TABLE IF NOT EXISTS user_stats (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    total_points INTEGER DEFAULT 0,
    tests_completed INTEGER DEFAULT 0,
    tests_passed INTEGER DEFAULT 0,
    current_streak INTEGER DEFAULT 0,
    best_streak INTEGER DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_topics_subject ON topics(subject_id);
CREATE INDEX IF NOT EXISTS idx_tests_subject ON tests(subject_id);
CREATE INDEX IF NOT EXISTS idx_tests_topic ON tests(topic_id);
CREATE INDEX IF NOT EXISTS idx_tests_review_status ON tests(review_status);
CREATE INDEX IF NOT EXISTS idx_questions_test ON questions(test_id);
CREATE INDEX IF NOT EXISTS idx_answer_options_question ON answer_options(question_id);
CREATE INDEX IF NOT EXISTS idx_test_attempts_user ON test_attempts(user_id);
CREATE INDEX IF NOT EXISTS idx_test_attempts_test ON test_attempts(test_id);
CREATE INDEX IF NOT EXISTS idx_student_answers_attempt ON student_answers(attempt_id);

-- ============================================================
-- Syllabus & Revision Planner
-- ============================================================

-- Syllabi: an authoritative topic list for a subject at an exam level
CREATE TABLE IF NOT EXISTS syllabi (
    id SERIAL PRIMARY KEY,
    subject_id INTEGER REFERENCES subjects(id) ON DELETE SET NULL,
    exam_standard VARCHAR(50) NOT NULL CHECK (exam_standard IN ('GCSE', 'IGCSE', 'A-Level', 'Primary', 'Secondary')),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sections group topics within a syllabus
CREATE TABLE IF NOT EXISTS syllabus_sections (
    id SERIAL PRIMARY KEY,
    syllabus_id INTEGER NOT NULL REFERENCES syllabi(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    section_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Topics are individual curriculum entries within a section
CREATE TABLE IF NOT EXISTS syllabus_topics (
    id SERIAL PRIMARY KEY,
    syllabus_id INTEGER NOT NULL REFERENCES syllabi(id) ON DELETE CASCADE,
    section_id INTEGER REFERENCES syllabus_sections(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    estimated_hours REAL NOT NULL DEFAULT 1.0,
    topic_order INTEGER NOT NULL DEFAULT 0,
    notes_content TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Links syllabus topics to existing tests
CREATE TABLE IF NOT EXISTS syllabus_topic_tests (
    syllabus_topic_id INTEGER NOT NULL REFERENCES syllabus_topics(id) ON DELETE CASCADE,
    test_id INTEGER NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (syllabus_topic_id, test_id)
);

-- A student's revision plan for a specific syllabus
CREATE TABLE IF NOT EXISTS revision_plans (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    syllabus_id INTEGER NOT NULL REFERENCES syllabi(id) ON DELETE CASCADE,
    exam_date DATE NOT NULL,
    hours_per_day REAL NOT NULL DEFAULT 2.0,
    study_days TEXT NOT NULL DEFAULT '[1,2,3,4,5]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, syllabus_id)
);

-- Individual scheduled study sessions generated from a revision plan
CREATE TABLE IF NOT EXISTS revision_sessions (
    id SERIAL PRIMARY KEY,
    plan_id INTEGER NOT NULL REFERENCES revision_plans(id) ON DELETE CASCADE,
    session_date DATE NOT NULL,
    syllabus_topic_id INTEGER NOT NULL REFERENCES syllabus_topics(id) ON DELETE CASCADE,
    hours_allocated REAL NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'completed', 'skipped')),
    notes TEXT,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_syllabi_subject ON syllabi(subject_id);
CREATE INDEX IF NOT EXISTS idx_syllabi_published ON syllabi(is_published);
CREATE INDEX IF NOT EXISTS idx_syllabus_sections_syllabus ON syllabus_sections(syllabus_id);
CREATE INDEX IF NOT EXISTS idx_syllabus_topics_syllabus ON syllabus_topics(syllabus_id);
CREATE INDEX IF NOT EXISTS idx_syllabus_topics_section ON syllabus_topics(section_id);
CREATE INDEX IF NOT EXISTS idx_revision_plans_user ON revision_plans(user_id);
CREATE INDEX IF NOT EXISTS idx_revision_sessions_plan ON revision_sessions(plan_id);
CREATE INDEX IF NOT EXISTS idx_revision_sessions_date ON revision_sessions(session_date);
CREATE INDEX IF NOT EXISTS idx_user_achievements_user ON user_achievements(user_id);
CREATE INDEX IF NOT EXISTS idx_test_review_events_test ON test_review_events(test_id);
CREATE INDEX IF NOT EXISTS idx_feedback_issues_status ON test_feedback_issues(status);
CREATE INDEX IF NOT EXISTS idx_feedback_issues_test ON test_feedback_issues(test_id);

-- Insert default achievements
INSERT INTO achievements (name, description, badge_icon, criteria_type, criteria_value, points_awarded) VALUES
    ('First Steps', 'Complete your first test', '🎯', 'tests_completed', 1, 10),
    ('Perfect Score', 'Score 100% on any test', '⭐', 'perfect_score', 100, 50),
    ('Test Master', 'Complete 10 tests', '🏆', 'tests_completed', 10, 100),
    ('Quick Learner', 'Score above 90% on 5 tests', '🌟', 'tests_completed', 5, 75),
    ('Streak Champion', 'Maintain a 5-day study streak', '🔥', 'streak', 5, 50)
ON CONFLICT (name) DO NOTHING;

-- Insert default subjects
INSERT INTO subjects (name, description) VALUES
    ('Mathematics', 'Numbers, algebra, geometry, and more'),
    ('Science', 'Biology, chemistry, physics'),
    ('History', 'World history, civilizations, and events'),
    ('English', 'Grammar, literature, and writing'),
    ('Geography', 'Countries, continents, and natural features')
ON CONFLICT (name) DO NOTHING;

-- Insert default admin user (GoCaSEAdmin)
INSERT INTO users (email, username, password_hash, role) VALUES
    ('john.channing@gmail.com', 'GoCaSEAdmin', '$2a$10$TdIl8d8fqY6uhKZTkswGsOHaZ8WDpa74zOVfZI0ZDVIVToYiJQicK', 'admin')
ON CONFLICT (email) DO NOTHING;
