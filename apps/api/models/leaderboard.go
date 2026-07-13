package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LeaderboardResult struct {
	Entries  []LeaderboardEntry `json:"entries"`
	YourRank *int               `json:"yourRank"`
}

type LeaderboardEntry struct {
	Rank         int     `json:"rank"`
	Username     string  `json:"username"`
	Seconds      int     `json:"seconds"`
	TopLanguage  string  `json:"topLanguage"`
	IsYou        bool    `json:"isYou"`
	ProfileImage *string `json:"profileImage"`
	Streak       int     `json:"streak"`
}

func GetLeaderboard(ctx context.Context, pool *pgxpool.Pool, rangeSQL string, limit int, currentUserID string) (*LeaderboardResult, error) {
	rows, err := pool.Query(ctx, `
		SELECT
			u.id,
			u.username,
			u.avatar_url,
			SUM(s.duration_seconds) as total_seconds,
			(
				SELECT language FROM sessions s2
				WHERE s2.user_id = u.id
				GROUP BY language
				ORDER BY SUM(duration_seconds) DESC
				LIMIT 1
			) as top_language
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		LEFT JOIN privacy_settings p ON p.user_id = u.id
		WHERE `+rangeSQL+`
			AND (p.hide_leaderboard IS NULL OR p.hide_leaderboard = false)
			AND (p.profile_public IS NULL OR p.profile_public = true)
			AND u.deleted_at IS NULL
		GROUP BY u.id, u.username, u.avatar_url
		ORDER BY total_seconds DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type rawRow struct {
		userID string
		entry  LeaderboardEntry
	}
	var rawRows []rawRow
	rank := 1
	var yourRank *int

	for rows.Next() {
		var e LeaderboardEntry
		var userID string
		var topLang *string

		if err := rows.Scan(&userID, &e.Username, &e.ProfileImage, &e.Seconds, &topLang); err != nil {
			return nil, err
		}

		if topLang != nil {
			e.TopLanguage = *topLang
		}

		e.Rank = rank
		e.IsYou = userID == currentUserID

		if e.IsYou {
			r := rank
			yourRank = &r
		}

		rawRows = append(rawRows, rawRow{userID: userID, entry: e})
		rank++
	}

	result := &LeaderboardResult{YourRank: yourRank}

	cutoff := len(rawRows)
	if cutoff > limit {
		cutoff = limit
	}

	entries := make([]LeaderboardEntry, 0, cutoff)
	for i := 0; i < cutoff; i++ {
		streak, err := GetCurrentStreak(ctx, pool, rawRows[i].userID)
		if err != nil {
			return nil, err
		}
		rawRows[i].entry.Streak = streak
		entries = append(entries, rawRows[i].entry)
	}

	result.Entries = entries

	return result, nil
}
