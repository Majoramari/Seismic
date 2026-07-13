package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type goalReminderRow struct {
	GoalID        string
	UserID        string
	Email         string
	Scope         string
	ScopeValue    *string
	Period        string
	TargetSeconds int
}

// CheckGoalReminders runs periodically. It finds goals that
// are close to their deadline, not yet complete, and haven't
// been reminded yet this period, then sends an email nudge.
func CheckGoalReminders(ctx context.Context, pool *pgxpool.Pool, emailCfg EmailConfig) error {
	rows, err := pool.Query(ctx, `
		SELECT g.id, g.user_id, u.email, g.scope, g.scope_value, g.period, g.target_seconds
		FROM goals g
		JOIN users u ON u.id = g.user_id
		WHERE g.active = true
		  AND g.reminders_enabled = true
		  AND u.deleted_at IS NULL
		  AND (
		    g.last_reminded_at IS NULL
		    OR (g.period = 'daily' AND g.last_reminded_at < CURRENT_DATE)
		    OR (g.period = 'weekly' AND g.last_reminded_at < date_trunc('week', CURRENT_DATE))
		    OR (g.period = 'monthly' AND g.last_reminded_at < date_trunc('month', CURRENT_DATE))
		  )
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var goals []goalReminderRow
	for rows.Next() {
		var g goalReminderRow
		if err := rows.Scan(&g.GoalID, &g.UserID, &g.Email, &g.Scope, &g.ScopeValue, &g.Period, &g.TargetSeconds); err != nil {
			return err
		}
		goals = append(goals, g)
	}

	sentCount := 0
	for _, g := range goals {
		if !isNearDeadline(g.Period) {
			continue
		}

		progress, err := getGoalProgressPublic(ctx, pool, g.UserID, g.Scope, g.ScopeValue, g.Period)
		if err != nil {
			log.Println("goal reminder progress error:", err)
			continue
		}

		if progress >= g.TargetSeconds {
			continue // already done, no need to nudge
		}

		label := goalLabel(g.Scope, g.ScopeValue)
		err = SendGoalReminderEmail(
			emailCfg,
			g.Email,
			label,
			formatDuration(progress),
			formatDuration(g.TargetSeconds),
			g.Period,
		)
		if err != nil {
			log.Println("failed to send goal reminder:", err)
			continue
		}

		_, err = pool.Exec(ctx, `UPDATE goals SET last_reminded_at = now() WHERE id = $1`, g.GoalID)
		if err != nil {
			log.Println("failed to update last_reminded_at:", err)
		}

		sentCount++
	}

	log.Printf("Goal reminders: sent %d of %d checked\n", sentCount, len(goals))
	return nil
}

// isNearDeadline reports whether the current time falls
// within the "reminder window" close to the end of the
// goal's period — the last 2 hours of a day, the last day
// of a week, or the last 2 days of a month.
func isNearDeadline(period string) bool {
	now := time.Now()

	switch period {
	case "daily":
		hoursLeft := 24 - now.Hour()
		return hoursLeft <= 2

	case "weekly":
		// Go's Weekday: Sunday=0 ... Saturday=6.
		// Treat the week as ending Sunday night.
		daysLeft := 7 - int(now.Weekday())
		if now.Weekday() == time.Sunday {
			daysLeft = 0
		}
		return daysLeft <= 1

	case "monthly":
		firstOfNextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
		daysLeft := int(firstOfNextMonth.Sub(now).Hours() / 24)
		return daysLeft <= 2

	default:
		return false
	}
}

func goalLabel(scope string, scopeValue *string) string {
	if scope == "language" && scopeValue != nil {
		return fmt.Sprintf("Language: %s", *scopeValue)
	}
	if scope == "project" && scopeValue != nil {
		return fmt.Sprintf("Project: %s", *scopeValue)
	}
	return "All programming activity"
}

func formatDuration(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func getGoalProgressPublic(ctx context.Context, pool *pgxpool.Pool, userID, scope string, scopeValue *string, period string) (int, error) {
	periodSQL := "start_time >= CURRENT_DATE"
	if period == "weekly" {
		periodSQL = "start_time >= date_trunc('week', CURRENT_DATE)"
	} else if period == "monthly" {
		periodSQL = "start_time >= date_trunc('month', CURRENT_DATE)"
	}

	scopeSQL := ""
	args := []any{userID}

	if scope == "language" && scopeValue != nil {
		scopeSQL = "AND language = $2"
		args = append(args, *scopeValue)
	} else if scope == "project" && scopeValue != nil {
		scopeSQL = "AND project = $2"
		args = append(args, *scopeValue)
	}

	var seconds int
	query := `SELECT COALESCE(SUM(duration_seconds), 0) FROM sessions WHERE user_id = $1 AND ` + periodSQL + ` ` + scopeSQL
	err := pool.QueryRow(ctx, query, args...).Scan(&seconds)
	return seconds, err
}
