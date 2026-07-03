package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsSummary struct {
	TotalSeconds  int     `json:"totalSeconds"`
	TopLanguage   *string `json:"topLanguage"`
	TopProject    *string `json:"topProject"`
	DailyAverage  int     `json:"dailyAverage"`
	CurrentStreak int     `json:"currentStreak"`
}

// GetStatsSummary calculates total time, top language, top
// project, and daily average for a user within a date range.
// rangeFilter is a SQL WHERE clause fragment for the range.
func GetStatsSummary(ctx context.Context, pool *pgxpool.Pool, userID string, rangeSQL string) (*StatsSummary, error) {
	var s StatsSummary

	err := pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(duration_seconds), 0)
		FROM sessions
		WHERE user_id = $1 AND `+rangeSQL, userID).Scan(&s.TotalSeconds)
	if err != nil {
		return nil, err
	}

	err = pool.QueryRow(ctx, `
		SELECT language FROM sessions
		WHERE user_id = $1 AND `+rangeSQL+`
		GROUP BY language
		ORDER BY SUM(duration_seconds) DESC
		LIMIT 1
	`, userID).Scan(&s.TopLanguage)
	if err != nil && err.Error() != "no rows in result set" {
		return nil, err
	}

	err = pool.QueryRow(ctx, `
		SELECT project FROM sessions
		WHERE user_id = $1 AND `+rangeSQL+`
		GROUP BY project
		ORDER BY SUM(duration_seconds) DESC
		LIMIT 1
	`, userID).Scan(&s.TopProject)
	if err != nil && err.Error() != "no rows in result set" {
		return nil, err
	}

	s.DailyAverage = s.TotalSeconds // refine later
	s.CurrentStreak = 0             // built later, needs day-by-day logic

	return &s, nil
}

// RangeSQL converts a range string like "today", "week",
// "month", "all" into a SQL WHERE clause fragment.
func RangeSQL(rangeParam string) string {
	switch rangeParam {
	case "today":
		return "start_time >= CURRENT_DATE"
	case "week":
		return "start_time >= CURRENT_DATE - INTERVAL '7 days'"
	case "month":
		return "start_time >= CURRENT_DATE - INTERVAL '30 days'"
	default:
		return "1=1" // "all" — no filter
	}
}
