package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
