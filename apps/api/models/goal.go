package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Goal struct {
	ID               string    `json:"id"`
	Scope            string    `json:"scope"`
	ScopeValue       *string   `json:"scopeValue"`
	Period           string    `json:"period"`
	TargetSeconds    int       `json:"targetSeconds"`
	RemindersEnabled bool      `json:"remindersEnabled"`
	ProgressSeconds  int       `json:"progressSeconds"`
	Percentage       int       `json:"percentage"`
	Active           bool      `json:"active"`
	CreatedAt        time.Time `json:"createdAt"`
}

type CreateGoalInput struct {
	Scope            string  `json:"scope"`
	ScopeValue       *string `json:"scopeValue"`
	Period           string  `json:"period"`
	TargetSeconds    int     `json:"targetSeconds"`
	RemindersEnabled bool    `json:"remindersEnabled"`
}

func CreateGoal(ctx context.Context, pool *pgxpool.Pool, userID string, input CreateGoalInput) (*Goal, error) {
	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM goals WHERE user_id = $1 AND active = true`, userID).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count >= 5 {
		return nil, fmt.Errorf("maximum of 5 active goals reached")
	}

	var g Goal
	err = pool.QueryRow(ctx, `
		INSERT INTO goals (user_id, scope, scope_value, period, target_seconds, reminders_enabled)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, scope, scope_value, period, target_seconds, reminders_enabled, active, created_at
	`, userID, input.Scope, input.ScopeValue, input.Period, input.TargetSeconds, input.RemindersEnabled).Scan(
		&g.ID, &g.Scope, &g.ScopeValue, &g.Period, &g.TargetSeconds, &g.RemindersEnabled, &g.Active, &g.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func UpdateGoal(ctx context.Context, pool *pgxpool.Pool, userID, goalID string, input CreateGoalInput) (*Goal, error) {
	var g Goal
	err := pool.QueryRow(ctx, `
		UPDATE goals
		SET scope = $1, scope_value = $2, period = $3, target_seconds = $4, reminders_enabled = $5
		WHERE id = $6 AND user_id = $7
		RETURNING id, scope, scope_value, period, target_seconds, reminders_enabled, active, created_at
	`, input.Scope, input.ScopeValue, input.Period, input.TargetSeconds, input.RemindersEnabled, goalID, userID).Scan(
		&g.ID, &g.Scope, &g.ScopeValue, &g.Period, &g.TargetSeconds, &g.RemindersEnabled, &g.Active, &g.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func GetActiveGoalsWithProgress(ctx context.Context, pool *pgxpool.Pool, userID string) ([]Goal, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, scope, scope_value, period, target_seconds, reminders_enabled, active, created_at
		FROM goals
		WHERE user_id = $1 AND active = true
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []Goal
	for rows.Next() {
		var g Goal
		if err := rows.Scan(&g.ID, &g.Scope, &g.ScopeValue, &g.Period, &g.TargetSeconds, &g.RemindersEnabled, &g.Active, &g.CreatedAt); err != nil {
			return nil, err
		}
		goals = append(goals, g)
	}

	for i := range goals {
		progress, err := getGoalProgress(ctx, pool, userID, goals[i])
		if err != nil {
			return nil, err
		}
		goals[i].ProgressSeconds = progress
		if goals[i].TargetSeconds > 0 {
			goals[i].Percentage = (progress * 100) / goals[i].TargetSeconds
			if goals[i].Percentage > 100 {
				goals[i].Percentage = 100
			}
		}
	}

	return goals, nil
}

func getGoalProgress(ctx context.Context, pool *pgxpool.Pool, userID string, g Goal) (int, error) {
	periodSQL := "start_time >= CURRENT_DATE"
	if g.Period == "weekly" {
		periodSQL = "start_time >= date_trunc('week', CURRENT_DATE)"
	} else if g.Period == "monthly" {
		periodSQL = "start_time >= date_trunc('month', CURRENT_DATE)"
	}

	scopeSQL := ""
	args := []any{userID}

	switch g.Scope {
	case "language":
		if g.ScopeValue == nil {
			return 0, nil
		}
		scopeSQL = "AND LOWER(TRIM(language)) = LOWER(TRIM($2))"
		args = append(args, *g.ScopeValue)
	case "project":
		if g.ScopeValue == nil {
			return 0, nil
		}
		scopeSQL = "AND LOWER(TRIM(project)) = LOWER(TRIM($2))"
		args = append(args, *g.ScopeValue)
	}

	var seconds int
	query := `
		SELECT COALESCE(SUM(duration_seconds), 0)
		FROM sessions
		WHERE user_id = $1 AND ` + periodSQL + ` ` + scopeSQL

	err := pool.QueryRow(ctx, query, args...).Scan(&seconds)
	return seconds, err
}

func DeleteGoal(ctx context.Context, pool *pgxpool.Pool, userID, goalID string) error {
	_, err := pool.Exec(ctx, `
		DELETE FROM goals WHERE id = $1 AND user_id = $2
	`, goalID, userID)
	return err
}
