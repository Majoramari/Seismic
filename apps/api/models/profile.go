package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ── Profile response types ────────────────────────────────────────────────────

// HeatmapCell is one cell in the 52×7 activity grid sent to the frontend.
type HeatmapCell struct {
	Date  string `json:"date"`
	Level int    `json:"level"` // 0-4
}

// ProfileInfoField represents one personal-info item and whether it's filled.
type ProfileInfoField struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Completed bool   `json:"completed"`
}

// Achievement is the denormalised view of user_achievements JOIN achievement_types.
type Achievement struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	BadgeClass  string `json:"badgeClass"`
	EarnedAt    string `json:"earnedAt"` // formatted as "Jan 02, 2006"
}

// ActivityLogItem is one entry in the recent-activity feed.
type ActivityLogItem struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
	At   string `json:"at"` // formatted as RFC3339
}

// ProfileResponse is the complete payload returned by GET /api/profile.
type ProfileResponse struct {
	// Identity
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	AvatarURL string `json:"avatarUrl"`

	// Bio fields (nullable in DB → empty string when nil)
	Role       string `json:"role"`
	Location   string `json:"location"`
	University string `json:"university"`
	Bio        string `json:"bio"`
	TimeZone   string `json:"timeZone"`
	Website    string `json:"website"`
	Gender     string `json:"gender"`
	Languages  []string `json:"languages"`

	// Date strings
	JoinDate   string `json:"joinDate"`   // "Jan 2006"
	LastActive string `json:"lastActive"` // "Online now" or relative
	MemberFor  string `json:"memberFor"`  // "2 years", "5 months", etc.

	// Coding metrics (from user_profile_stats)
	TotalCodingSeconds int64 `json:"totalCodingSeconds"`
	TotalActiveDays    int   `json:"totalActiveDays"`
	CurrentStreak      int   `json:"currentStreak"`
	MaxStreak          int   `json:"maxStreak"`

	// Problem solving (placeholder – user_problem_stats may have no row)
	Solved        int `json:"solved"`
	TotalProblems int `json:"totalProblems"`
	Attempting    int `json:"attempting"`

	// Personal info completion (computed in handler)
	CompletionPercent int                `json:"completionPercent"`
	InfoFields        []ProfileInfoField `json:"infoFields"`

	// Rich data
	Heatmap        [][]HeatmapCell   `json:"heatmap"`
	RecentActivity []ActivityLogItem `json:"recentActivity"`
	Achievements   []Achievement     `json:"achievements"`
}


// UserProfileStats caches aggregate coding data for the profile page
type UserProfileStats struct {
	UserID             string    `json:"userId"`
	TotalCodingSeconds int64     `json:"totalCodingSeconds"`
	TotalActiveDays    int       `json:"totalActiveDays"`
	CurrentStreak      int       `json:"currentStreak"`
	MaxStreak          int       `json:"maxStreak"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// GetProfileStats retrieves cached profile stats for a user.
func GetProfileStats(ctx context.Context, pool *pgxpool.Pool, userID string) (*UserProfileStats, error) {
	var s UserProfileStats
	err := pool.QueryRow(ctx, `
		SELECT user_id, total_coding_seconds, total_active_days, current_streak, max_streak, updated_at
		FROM user_profile_stats
		WHERE user_id = $1
	`, userID).Scan(&s.UserID, &s.TotalCodingSeconds, &s.TotalActiveDays, &s.CurrentStreak, &s.MaxStreak, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateProfileStats recalculates aggregate stats from the sessions table
// and upserts them into user_profile_stats for the given user.
func UpdateProfileStats(ctx context.Context, pool *pgxpool.Pool, userID string) error {
	var totalSeconds int64
	_ = pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(duration_seconds), 0) FROM sessions WHERE user_id = $1
	`, userID).Scan(&totalSeconds)

	var totalDays int
	_ = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT start_time::date) FROM sessions WHERE user_id = $1
	`, userID).Scan(&totalDays)

	currentStreak, _ := GetCurrentStreak(ctx, pool, userID)

	var prevMax int
	_ = pool.QueryRow(ctx, `
		SELECT COALESCE(max_streak, 0) FROM user_profile_stats WHERE user_id = $1
	`, userID).Scan(&prevMax)

	maxStreak := prevMax
	if currentStreak > maxStreak {
		maxStreak = currentStreak
	}

	_, err := pool.Exec(ctx, `
		INSERT INTO user_profile_stats (user_id, total_coding_seconds, total_active_days, current_streak, max_streak, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			total_coding_seconds = EXCLUDED.total_coding_seconds,
			total_active_days = EXCLUDED.total_active_days,
			current_streak = EXCLUDED.current_streak,
			max_streak = GREATEST(user_profile_stats.max_streak, EXCLUDED.current_streak),
			updated_at = EXCLUDED.updated_at
	`, userID, totalSeconds, totalDays, currentStreak, maxStreak)

	return err
}

// AchievementType represents a badge that can be earned
type AchievementType struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	BadgeClass  string `json:"badgeClass"`
}

// UserAchievement tracks which user earned which badge and when
type UserAchievement struct {
	ID                string    `json:"id"`
	UserID            string    `json:"userId"`
	AchievementTypeID string    `json:"achievementTypeId"`
	EarnedAt          time.Time `json:"earnedAt"`
}

// ActivityLog represents an entry in the user's recent activity feed
type ActivityLog struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Kind      string    `json:"kind"`
	Text      string    `json:"text"`
	Metadata  any       `json:"metadata,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// UserProblemStats is a placeholder for future problem-solving features
type UserProblemStats struct {
	UserID             string    `json:"userId"`
	SolvedCount        int       `json:"solvedCount"`
	TotalProblems      int       `json:"totalProblems"`
	AttemptingCount    int       `json:"attemptingCount"`
	Rating             int       `json:"rating"`
	ContributionPoints int       `json:"contributionPoints"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// GetProblemStats fetches placeholder problem-solving stats.
// Returns zero values (no error) if the user has no row yet.
// Do NOT insert a row here; this table is a placeholder.
func GetProblemStats(ctx context.Context, pool *pgxpool.Pool, userID string) (*UserProblemStats, error) {
	var s UserProblemStats
	err := pool.QueryRow(ctx, `
		SELECT user_id, solved_count, total_problems, attempting_count, rating, contribution_points, updated_at
		FROM user_problem_stats
		WHERE user_id = $1
	`, userID).Scan(&s.UserID, &s.SolvedCount, &s.TotalProblems, &s.AttemptingCount, &s.Rating, &s.ContributionPoints, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return &UserProblemStats{}, nil // placeholder: return zeros, no insert
	}
	if err != nil {
		return &UserProblemStats{}, nil // silently degrade; profile still loads
	}
	return &s, nil
}

// GetProfileAchievements returns all achievements earned by a user,
// joining user_achievements with achievement_types for full metadata.
func GetProfileAchievements(ctx context.Context, pool *pgxpool.Pool, userID string) ([]Achievement, error) {
	rows, err := pool.Query(ctx, `
		SELECT at.key, at.title, at.description, at.badge_class, ua.earned_at
		FROM user_achievements ua
		JOIN achievement_types at ON at.id = ua.achievement_type_id
		WHERE ua.user_id = $1
		ORDER BY ua.earned_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []Achievement
	for rows.Next() {
		var a Achievement
		var earnedAt time.Time
		if err := rows.Scan(&a.Key, &a.Title, &a.Description, &a.BadgeClass, &earnedAt); err != nil {
			return nil, err
		}
		a.EarnedAt = earnedAt.Format("Jan 02, 2006")
		achievements = append(achievements, a)
	}
	if achievements == nil {
		achievements = []Achievement{}
	}
	return achievements, nil
}

// GetRecentActivity returns the last `limit` activity log entries for a user.
func GetRecentActivity(ctx context.Context, pool *pgxpool.Pool, userID string, limit int) ([]ActivityLogItem, error) {
	rows, err := pool.Query(ctx, `
		SELECT kind, text, created_at
		FROM activity_log
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ActivityLogItem
	for rows.Next() {
		var item ActivityLogItem
		var createdAt time.Time
		if err := rows.Scan(&item.Kind, &item.Text, &createdAt); err != nil {
			return nil, err
		}
		item.At = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	if items == nil {
		items = []ActivityLogItem{}
	}
	return items, nil
}

// UpdateProfileRequest represents the allowed fields for updating a user profile.
// Pointers differentiate between a field not sent (nil) and explicitly emptied ("").
type UpdateProfileRequest struct {
	FirstName  *string   `json:"firstName"`
	Bio        *string   `json:"bio"`
	Location   *string   `json:"location"`
	Role       *string   `json:"role"`
	University *string   `json:"university"`
	Website    *string   `json:"website"`
	Gender     *string   `json:"gender"`
	Languages  *[]string `json:"languages"`
	TimeZone   *string   `json:"timeZone"`
}

// UpdateUserProfile explicitly updates only the whitelisted profile fields.
func UpdateUserProfile(ctx context.Context, pool *pgxpool.Pool, userID string, req *UpdateProfileRequest) error {
	setIdx := 1
	var args []interface{}
	query := "UPDATE users SET "

	// Explicit Whitelist construction: Column names are hardcoded
	if req.FirstName != nil {
		query += fmt.Sprintf("first_name = $%d, ", setIdx)
		args = append(args, *req.FirstName)
		setIdx++
	}
	if req.Bio != nil {
		query += fmt.Sprintf("bio = $%d, ", setIdx)
		args = append(args, *req.Bio)
		setIdx++
	}
	if req.Location != nil {
		query += fmt.Sprintf("location = $%d, ", setIdx)
		args = append(args, *req.Location)
		setIdx++
	}
	if req.Role != nil {
		query += fmt.Sprintf("role = $%d, ", setIdx)
		args = append(args, *req.Role)
		setIdx++
	}
	if req.University != nil {
		query += fmt.Sprintf("university = $%d, ", setIdx)
		args = append(args, *req.University)
		setIdx++
	}
	if req.Website != nil {
		query += fmt.Sprintf("website = $%d, ", setIdx)
		args = append(args, *req.Website)
		setIdx++
	}
	if req.Gender != nil {
		query += fmt.Sprintf("gender = $%d, ", setIdx)
		args = append(args, *req.Gender)
		setIdx++
	}
	if req.TimeZone != nil {
		query += fmt.Sprintf("time_zone = $%d, ", setIdx)
		args = append(args, *req.TimeZone)
		setIdx++
	}
	if req.Languages != nil {
		query += fmt.Sprintf("languages = $%d, ", setIdx)
		args = append(args, *req.Languages)
		setIdx++
	}

	// If no valid update fields were provided, simply return
	if setIdx == 1 {
		return nil
	}

	// Remove the trailing comma and space
	query = query[:len(query)-2]

	query += fmt.Sprintf(" WHERE id = $%d", setIdx)
	args = append(args, userID)

	_, err := pool.Exec(ctx, query, args...)
	return err
}
