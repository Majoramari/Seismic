package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Heartbeat represents a row in the heartbeats table.
type Heartbeat struct {
	File       string  `json:"file"`
	Project    string  `json:"project"`
	Language   string  `json:"language"`
	Editor     string  `json:"editor"`
	Branch     *string `json:"branch"`
	OS         *string `json:"os"`
	Machine    *string `json:"machine"`
	Lines      *int    `json:"lines"`
	CursorLine *int    `json:"cursorLine"`
	Timezone   *string `json:"timezone"`
	Keystrokes *int    `json:"keystrokes"`
	Time       int64   `json:"time"`
}

// InsertHeartbeat stores a raw heartbeat from an editor plugin.
func InsertHeartbeat(ctx context.Context, pool *pgxpool.Pool, userID string, h Heartbeat) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO heartbeats (
			user_id, file, project, language, editor, branch,
			os, machine, lines, cursor_line, timezone, keystrokes, time
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, userID, h.File, h.Project, h.Language, h.Editor, h.Branch,
		h.OS, h.Machine, h.Lines, h.CursorLine, h.Timezone, h.Keystrokes, h.Time)

	return err
}

// HasRecentDuplicate checks if the same user sent a heartbeat
// for the same file within the last 10 seconds.
func HasRecentDuplicate(ctx context.Context, pool *pgxpool.Pool, userID, file string, timeMs int64) (bool, error) {
	var exists bool

	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM heartbeats
			WHERE user_id = $1 AND file = $2
			 AND time BETWEEN $3 AND $4
		)
	`, userID, file, timeMs-10000, timeMs+10000).Scan(&exists)

	return exists, err
}
