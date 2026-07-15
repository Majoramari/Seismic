package models

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsSummary struct {
	TotalSeconds    int     `json:"totalSeconds"`
	TopLanguage     *string `json:"topLanguage"`
	TopProject      *string `json:"topProject"`
	TopOS           *string `json:"topOS"`
	TopEditor       *string `json:"topEditor"`
	DailyAverage    int     `json:"dailyAverage"`
	CurrentStreak   int     `json:"currentStreak"`
	TotalKeystrokes int     `json:"totalKeystrokes"`
}

type LanguageStat struct {
	Language string `json:"language"`
	Seconds  int    `json:"seconds"`
}

type HeatmapDay struct {
	Date    string `json:"date"`
	Seconds int    `json:"seconds"`
}

type ProjectStat struct {
	Project string `json:"project"`
	Seconds int    `json:"seconds"`
}

type OSStat struct {
	OS      string `json:"os"`
	Seconds int    `json:"seconds"`
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
		SELECT LOWER(TRIM(language)) FROM sessions
		WHERE user_id = $1 AND `+rangeSQL+`
		GROUP BY LOWER(TRIM(language))
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

	heartbeatRangeSQL := strings.ReplaceAll(rangeSQL, "start_time", "received_at")

	err = pool.QueryRow(ctx, `
		SELECT COALESCE(NULLIF(os, ''), 'unknown') AS os
		FROM sessions
		WHERE user_id = $1 AND `+rangeSQL+`
		GROUP BY COALESCE(NULLIF(os, ''), 'unknown')
		ORDER BY SUM(duration_seconds) DESC
		LIMIT 1
	`, userID).Scan(&s.TopOS)
	if err != nil && err.Error() != "no rows in result set" {
		return nil, err
	}

	err = pool.QueryRow(ctx, `
		SELECT COALESCE(NULLIF(editor, ''), 'unknown') AS editor
		FROM sessions
		WHERE user_id = $1 AND `+rangeSQL+`
		GROUP BY COALESCE(NULLIF(editor, ''), 'unknown')
		ORDER BY SUM(duration_seconds) DESC
		LIMIT 1
	`, userID).Scan(&s.TopEditor)
	if err != nil && err.Error() != "no rows in result set" {
		return nil, err
	}

	err = pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(keystrokes), 0)
		FROM heartbeats
		WHERE user_id = $1 AND `+heartbeatRangeSQL, userID).Scan(&s.TotalKeystrokes)
	if err != nil {
		return nil, err
	}

	s.DailyAverage = s.TotalSeconds // refine later

	streak, err := GetCurrentStreak(ctx, pool, userID)
	if err != nil {
		return nil, err
	}
	s.CurrentStreak = streak

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

// GetLanguageBreakdown returns time spent per language.
func GetLanguageBreakdown(ctx context.Context, pool *pgxpool.Pool, userID string, rangeSQL string) ([]LanguageStat, error) {
	rows, err := pool.Query(ctx, `
		SELECT LOWER(TRIM(language)) AS language, SUM(duration_seconds) as seconds
		FROM sessions
		WHERE user_id = $1 AND `+rangeSQL+`
			AND LOWER(TRIM(language)) NOT IN ('textmate', 'unknown', 'log')
		GROUP BY LOWER(TRIM(language))
		HAVING SUM(duration_seconds) > 0
		ORDER BY seconds DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []LanguageStat
	for rows.Next() {
		var s LanguageStat
		if err := rows.Scan(&s.Language, &s.Seconds); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// GetHeatmap returns historical daily totals.
func GetHeatmap(ctx context.Context, pool *pgxpool.Pool, userID string) ([]HeatmapDay, error) {
	rows, err := pool.Query(ctx, `
		SELECT start_time::date as day, SUM(duration_seconds) as seconds
		FROM sessions
		WHERE user_id = $1
		GROUP BY day
		ORDER BY day ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var days []HeatmapDay
	for rows.Next() {
		var day time.Time
		var seconds int
		if err := rows.Scan(&day, &seconds); err != nil {
			return nil, err
		}
		days = append(days, HeatmapDay{Date: day.Format("2006-01-02"), Seconds: seconds})
	}
	return days, nil
}

// GetCurrentStreak returns how many consecutive days (ending
// today or yesterday) the user has coded at least once.
func GetCurrentStreak(ctx context.Context, pool *pgxpool.Pool, userID string) (int, error) {
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT start_time::date as day
		FROM sessions
		WHERE user_id = $1
		ORDER BY day DESC
	`, userID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var days []time.Time
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return 0, err
		}
		days = append(days, d)
	}

	if len(days) == 0 {
		return 0, nil
	}

	today := time.Now().Truncate(24 * time.Hour)
	streak := 0
	expected := today

	if !days[0].Equal(today) {
		expected = today.AddDate(0, 0, -1)
	}

	for _, d := range days {
		if d.Equal(expected) {
			streak++
			expected = expected.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return streak, nil
}

// GetDistinctLanguages returns every language a user has
// coded in, used to populate goal filter dropdowns.
func GetDistinctLanguages(ctx context.Context, pool *pgxpool.Pool, userID string) ([]string, error) {
	rows, err := pool.Query(ctx, `
		SELECT LOWER(TRIM(language)) AS language
		FROM sessions
		WHERE user_id = $1
			AND TRIM(language) <> ''
		GROUP BY LOWER(TRIM(language))
		ORDER BY language ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var languages []string
	for rows.Next() {
		var l string
		if err := rows.Scan(&l); err != nil {
			return nil, err
		}
		languages = append(languages, l)
	}
	return languages, nil
}

// GetDistinctProjects returns every project a user has
// coded in, used to populate goal filter dropdowns.
func GetDistinctProjects(ctx context.Context, pool *pgxpool.Pool, userID string) ([]string, error) {
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT project FROM sessions WHERE user_id = $1 ORDER BY project ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

// GetProjectBreakdown returns time spent per project.
func GetProjectBreakdown(ctx context.Context, pool *pgxpool.Pool, userID string, rangeSQL string) ([]ProjectStat, error) {
	rows, err := pool.Query(ctx, `
		SELECT project, SUM(duration_seconds) as seconds
		FROM sessions
		WHERE user_id = $1 AND `+rangeSQL+`
		GROUP BY project
		ORDER BY seconds DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []ProjectStat
	for rows.Next() {
		var s ProjectStat
		if err := rows.Scan(&s.Project, &s.Seconds); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

type EditorStat struct {
	Editor  string `json:"editor"`
	Seconds int    `json:"seconds"`
}

// GetEditorBreakdown returns time spent per editor.
func GetEditorBreakdown(ctx context.Context, pool *pgxpool.Pool, userID string, rangeSQL string) ([]EditorStat, error) {
	rows, err := pool.Query(ctx, `
		SELECT COALESCE(NULLIF(editor, ''), 'unknown') AS editor,
			SUM(duration_seconds)::int AS seconds
		FROM sessions
		WHERE user_id = $1 AND `+rangeSQL+`
		GROUP BY COALESCE(NULLIF(editor, ''), 'unknown')
		HAVING SUM(duration_seconds) > 0
		ORDER BY seconds DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []EditorStat
	for rows.Next() {
		var e EditorStat
		if err := rows.Scan(&e.Editor, &e.Seconds); err != nil {
			return nil, err
		}
		stats = append(stats, e)
	}
	return stats, nil
}

type TimelineDay struct {
	Date     string            `json:"date"`
	Seconds  int               `json:"seconds"`
	Projects []TimelineProject `json:"projects"`
}

type TimelineProject struct {
	Project string `json:"project"`
	Seconds int    `json:"seconds"`
}

// GetTimeline returns daily totals for the last N days,
// used for the project timeline bar chart.
func GetTimeline(ctx context.Context, pool *pgxpool.Pool, userID string, days int) ([]TimelineDay, error) {
	rows, err := pool.Query(ctx, `
		SELECT start_time::date AS day, project, SUM(duration_seconds)::int AS seconds
		FROM sessions
		WHERE user_id = $1
			AND start_time >= CURRENT_DATE - make_interval(days => $2)
		GROUP BY day, project
		HAVING SUM(duration_seconds) > 0
		ORDER BY day ASC, seconds DESC
	`, userID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	timelineByDate := make(map[string]*TimelineDay)
	var orderedDates []string

	for rows.Next() {
		var day time.Time
		var project string
		var seconds int
		if err := rows.Scan(&day, &project, &seconds); err != nil {
			return nil, err
		}

		date := day.Format("2006-01-02")
		timelineDay, exists := timelineByDate[date]
		if !exists {
			timelineDay = &TimelineDay{
				Date:     date,
				Projects: []TimelineProject{},
			}
			timelineByDate[date] = timelineDay
			orderedDates = append(orderedDates, date)
		}

		timelineDay.Seconds += seconds
		timelineDay.Projects = append(timelineDay.Projects, TimelineProject{
			Project: project,
			Seconds: seconds,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	timeline := make([]TimelineDay, 0, len(orderedDates))
	for _, date := range orderedDates {
		timeline = append(timeline, *timelineByDate[date])
	}

	return timeline, nil
}

func GetOSBreakdown(ctx context.Context, pool *pgxpool.Pool, userID string, rangeSQL string) ([]OSStat, error) {
	rows, err := pool.Query(ctx, `
		SELECT COALESCE(NULLIF(os, ''), 'unknown') AS os,
			SUM(duration_seconds)::int AS seconds
		FROM sessions
		WHERE user_id = $1 AND `+rangeSQL+`
		GROUP BY COALESCE(NULLIF(os, ''), 'unknown')
		HAVING SUM(duration_seconds) > 0
		ORDER BY seconds DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []OSStat
	for rows.Next() {
		var s OSStat
		if err := rows.Scan(&s.OS, &s.Seconds); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}
