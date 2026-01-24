package repository

import (
	"context"

	"my-app/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository handles user database operations
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (email, password_hash, username, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query, user.Email, user.PasswordHash, user.Username, user.Role).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, password_hash, username, role, created_at, updated_at
		FROM users
		WHERE email = $1`

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Username,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, password_hash, username, role, created_at, updated_at
		FROM users
		WHERE id = $1`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Username,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// InitializeUserStats creates initial stats for a new user
func (r *UserRepository) InitializeUserStats(ctx context.Context, userID int) error {
	query := `
		INSERT INTO user_stats (user_id, total_points, tests_completed, tests_passed, current_streak, best_streak)
		VALUES ($1, 0, 0, 0, 0, 0)
		ON CONFLICT (user_id) DO NOTHING`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// GetUserStats retrieves user statistics
func (r *UserRepository) GetUserStats(ctx context.Context, userID int) (*models.UserStats, error) {
	stats := &models.UserStats{}
	query := `
		SELECT user_id, total_points, tests_completed, tests_passed, current_streak, best_streak, updated_at
		FROM user_stats
		WHERE user_id = $1`

	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&stats.UserID, &stats.TotalPoints, &stats.TestsCompleted,
		&stats.TestsPassed, &stats.CurrentStreak, &stats.BestStreak, &stats.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// UpdateUserStats updates user statistics
func (r *UserRepository) UpdateUserStats(ctx context.Context, stats *models.UserStats) error {
	query := `
		UPDATE user_stats
		SET total_points = $2, tests_completed = $3, tests_passed = $4,
		    current_streak = $5, best_streak = $6, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1`

	_, err := r.pool.Exec(ctx, query,
		stats.UserID, stats.TotalPoints, stats.TestsCompleted,
		stats.TestsPassed, stats.CurrentStreak, stats.BestStreak,
	)
	return err
}

// GetUserAchievements retrieves all achievements earned by a user
func (r *UserRepository) GetUserAchievements(ctx context.Context, userID int) ([]models.UserAchievement, error) {
	query := `
		SELECT ua.id, ua.user_id, ua.achievement_id, ua.earned_at,
		       a.id, a.name, a.description, a.badge_icon, a.criteria_type,
		       a.criteria_value, a.points_awarded, a.created_at
		FROM user_achievements ua
		JOIN achievements a ON ua.achievement_id = a.id
		WHERE ua.user_id = $1
		ORDER BY ua.earned_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []models.UserAchievement
	for rows.Next() {
		var ua models.UserAchievement
		ua.Achievement = &models.Achievement{}

		err := rows.Scan(
			&ua.ID, &ua.UserID, &ua.AchievementID, &ua.EarnedAt,
			&ua.Achievement.ID, &ua.Achievement.Name, &ua.Achievement.Description,
			&ua.Achievement.BadgeIcon, &ua.Achievement.CriteriaType,
			&ua.Achievement.CriteriaValue, &ua.Achievement.PointsAwarded,
			&ua.Achievement.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		achievements = append(achievements, ua)
	}

	return achievements, rows.Err()
}

// AwardAchievement awards an achievement to a user
func (r *UserRepository) AwardAchievement(ctx context.Context, userID, achievementID int) error {
	query := `
		INSERT INTO user_achievements (user_id, achievement_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, achievement_id) DO NOTHING`

	_, err := r.pool.Exec(ctx, query, userID, achievementID)
	return err
}

// HasAchievement checks if a user has a specific achievement
func (r *UserRepository) HasAchievement(ctx context.Context, userID, achievementID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM user_achievements WHERE user_id = $1 AND achievement_id = $2)`
	err := r.pool.QueryRow(ctx, query, userID, achievementID).Scan(&exists)
	return exists, err
}

// GetAllUsers retrieves all users
func (r *UserRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT id, email, username, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.Role, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

// GetUsersByRole retrieves users by role
func (r *UserRepository) GetUsersByRole(ctx context.Context, role string) ([]models.User, error) {
	query := `
		SELECT id, email, username, role, created_at, updated_at
		FROM users
		WHERE role = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.Role, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

// UpdateUserRole updates a user's role
func (r *UserRepository) UpdateUserRole(ctx context.Context, userID int, role string) error {
	query := `UPDATE users SET role = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, role, userID)
	return err
}

// UpdatePasswordHash updates a user's password hash
func (r *UserRepository) UpdatePasswordHash(ctx context.Context, userID int, newHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, newHash, userID)
	return err
}

// DeleteUser deletes a user by ID
func (r *UserRepository) DeleteUser(ctx context.Context, userID int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}
