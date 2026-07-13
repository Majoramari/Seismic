package services

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/majoramari/seismic/apps/api/models"
)

// CheckAndAwardBadges runs after session processing and awards
// any badges a user has newly earned. Safe to call repeatedly —
// uses ON CONFLICT DO NOTHING so it never awards duplicates.
func CheckAndAwardBadges(ctx context.Context, pool *pgxpool.Pool) error {
	rows, err := pool.Query(ctx, `SELECT DISTINCT id FROM users WHERE deleted_at IS NULL`)
	if err != nil {
		return err
	}
	var userIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		userIDs = append(userIDs, id)
	}
	rows.Close()

	for _, userID := range userIDs {
		if err := checkUserBadges(ctx, pool, userID); err != nil {
			log.Println("badge check error for user", userID, ":", err)
		}
	}
	return nil
}

func checkUserBadges(ctx context.Context, pool *pgxpool.Pool, userID string) error {
	award := func(badgeType string) error {
		_, err := pool.Exec(ctx, `
			INSERT INTO badges (user_id, badge_type)
			VALUES ($1, $2)
			ON CONFLICT (user_id, badge_type) DO NOTHING
		`, userID, badgeType)
		return err
	}

	// first_heartbeat: has any session at all
	var hasSession bool
	err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sessions WHERE user_id = $1)`, userID).Scan(&hasSession)
	if err != nil {
		return err
	}
	if hasSession {
		if err := award("first_heartbeat"); err != nil {
			return err
		}
	}

	// century: 100+ total hours
	var totalSeconds int
	err = pool.QueryRow(ctx, `SELECT COALESCE(SUM(duration_seconds), 0) FROM sessions WHERE user_id = $1`, userID).Scan(&totalSeconds)
	if err != nil {
		return err
	}
	if totalSeconds >= 360000 {
		if err := award("century"); err != nil {
			return err
		}
	}

	// polyglot: 5+ distinct languages
	var languageCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(DISTINCT language) FROM sessions WHERE user_id = $1`, userID).Scan(&languageCount)
	if err != nil {
		return err
	}
	if languageCount >= 5 {
		if err := award("polyglot"); err != nil {
			return err
		}
	}

	// night_owl: sessions starting after 2am on 5+ distinct days
	var nightOwlDays int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT start_time::date)
		FROM sessions
		WHERE user_id = $1 AND EXTRACT(HOUR FROM start_time) BETWEEN 0 AND 4
	`, userID).Scan(&nightOwlDays)
	if err != nil {
		return err
	}
	if nightOwlDays >= 5 {
		if err := award("night_owl"); err != nil {
			return err
		}
	}

	// early_bird: sessions starting before 6am on 5+ distinct days
	var earlyBirdDays int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT start_time::date)
		FROM sessions
		WHERE user_id = $1 AND EXTRACT(HOUR FROM start_time) BETWEEN 4 AND 6
	`, userID).Scan(&earlyBirdDays)
	if err != nil {
		return err
	}
	if earlyBirdDays >= 5 {
		if err := award("early_bird"); err != nil {
			return err
		}
	}

	// week_streak / month_streak: reuse existing streak calc
	streak, err := models.GetCurrentStreak(ctx, pool, userID)
	if err != nil {
		return err
	}
	if streak >= 7 {
		if err := award("week_streak"); err != nil {
			return err
		}
	}
	if streak >= 30 {
		if err := award("month_streak"); err != nil {
			return err
		}
	}

	return nil
}
