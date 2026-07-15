package services

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const sessionGap = 15 * time.Minute
const heartbeatDuration = 30 * time.Second
const maxSessionLength = 6 * time.Hour

type rawHeartbeat struct {
	ID       string
	UserID   string
	Project  string
	Language string
	Editor   string
	OS       string
	TimeMs   int64
}

// ProcessSessions groups unprocessed heartbeats into sessions
// and stores them. Meant to run periodically in the background.
func ProcessSessions(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := processSessionsTx(ctx, tx, nil); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ReprocessUserSessions rebuilds a user's sessions from their raw
// heartbeats. Use after imports when session grouping rules may have
// changed or when newly inserted heartbeats should merge with older ones.
func ReprocessUserSessions(ctx context.Context, pool *pgxpool.Pool, userID string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE heartbeats SET processed = false WHERE user_id = $1`, userID); err != nil {
		return err
	}
	if err := processSessionsTx(ctx, tx, &userID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func processSessionsTx(ctx context.Context, tx pgx.Tx, userID *string) error {
	args := []any{}
	userFilter := ""
	if userID != nil {
		args = append(args, *userID)
		userFilter = "AND user_id = $1"
	}

	rows, err := tx.Query(ctx, `
		SELECT
			id,
			user_id,
			project,
			language,
			COALESCE(NULLIF(editor, ''), 'unknown'),
			COALESCE(NULLIF(os, ''), 'unknown'),
			time
		FROM heartbeats
		WHERE processed = false `+userFilter+`
		ORDER BY
			user_id,
			time ASC,
			project,
			language,
			COALESCE(NULLIF(editor, ''), 'unknown'),
			COALESCE(NULLIF(os, ''), 'unknown'),
			file,
			id
		FOR UPDATE SKIP LOCKED
	`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var heartbeats []rawHeartbeat
	for rows.Next() {
		var h rawHeartbeat
		if err := rows.Scan(&h.ID, &h.UserID, &h.Project, &h.Language, &h.Editor, &h.OS, &h.TimeMs); err != nil {
			return err
		}
		heartbeats = append(heartbeats, h)
	}

	sessions := buildSessions(heartbeats)

	seenUsers := make(map[string]bool)
	for _, s := range sessions {
		if seenUsers[s.UserID] {
			err = insertSessionTx(ctx, tx, s)
		} else {
			err = saveSessionTx(ctx, tx, s)
			seenUsers[s.UserID] = true
		}
		if err != nil {
			return err
		}
	}

	ids := make([]string, len(heartbeats))
	for i, h := range heartbeats {
		ids[i] = h.ID
	}
	if len(ids) > 0 {
		_, err = tx.Exec(ctx, `UPDATE heartbeats SET processed = true WHERE id::text = ANY($1)`, ids)
		if err != nil {
			return err
		}
	}

	log.Printf("Session processor: created %d sessions from %d heartbeats\n", len(sessions), len(heartbeats))
	return nil
}

type builtSession struct {
	UserID          string
	Project         string
	Language        string
	Editor          string
	OS              string
	Start           time.Time
	End             time.Time
	DurationSeconds int
}

type existingSession struct {
	ID       string
	Project  string
	Language string
	Editor   string
	OS       string
	Start    time.Time
	End      time.Time
}

func saveSessionTx(ctx context.Context, tx pgx.Tx, s builtSession) error {
	latest, err := latestSessionTx(ctx, tx, s.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return insertSessionTx(ctx, tx, s)
		}
		return err
	}

	if latest.Project == s.Project &&
		latest.Language == s.Language &&
		latest.Editor == s.Editor &&
		latest.OS == s.OS &&
		!s.Start.Before(latest.Start) &&
		s.Start.Sub(latest.End.Add(-heartbeatDuration)) <= sessionGap {
		end := latest.End
		if s.End.After(end) {
			end = s.End
		}

		duration := end.Sub(latest.Start)
		if duration > maxSessionLength {
			duration = maxSessionLength
		}

		_, err := tx.Exec(ctx, `
			UPDATE sessions
			SET end_time = $1, duration_seconds = $2
			WHERE id = $3
		`, end, int(duration.Seconds()), latest.ID)
		return err
	}

	return insertSessionTx(ctx, tx, s)
}

func latestSessionTx(ctx context.Context, tx pgx.Tx, userID string) (existingSession, error) {
	var s existingSession
	err := tx.QueryRow(ctx, `
		SELECT id, project, language, editor, os, start_time, end_time
		FROM sessions
		WHERE user_id = $1
		ORDER BY end_time DESC
		LIMIT 1
	`, userID).Scan(&s.ID, &s.Project, &s.Language, &s.Editor, &s.OS, &s.Start, &s.End)
	return s, err
}

func insertSessionTx(ctx context.Context, tx pgx.Tx, s builtSession) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO sessions (user_id, project, language, editor, os, start_time, end_time, duration_seconds)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT DO NOTHING
	`, s.UserID, s.Project, s.Language, s.Editor, s.OS, s.Start, s.End, s.DurationSeconds)
	return err
}

// buildSessions groups a time-ordered list of heartbeats into sessions,
// splitting whenever the project, language, or activity window changes.
func buildSessions(heartbeats []rawHeartbeat) []builtSession {
	var sessions []builtSession
	if len(heartbeats) == 0 {
		return sessions
	}

	var current *builtSession

	for i, h := range heartbeats {
		t := time.UnixMilli(h.TimeMs)
		end := t.Add(heartbeatDuration)
		if i+1 < len(heartbeats) && heartbeats[i+1].UserID == h.UserID {
			next := time.UnixMilli(heartbeats[i+1].TimeMs)
			if !next.Before(t) && next.Sub(t) <= sessionGap {
				end = next
			}
		}
		if !end.After(t) {
			continue
		}

		startsNew := current == nil ||
			current.UserID != h.UserID ||
			current.Project != h.Project ||
			current.Language != h.Language ||
			current.Editor != h.Editor ||
			current.OS != h.OS ||
			t.After(current.End)

		if startsNew {
			if current != nil {
				sessions = append(sessions, *current)
			}
			current = &builtSession{
				UserID:   h.UserID,
				Project:  h.Project,
				Language: h.Language,
				Editor:   h.Editor,
				OS:       h.OS,
				Start:    t,
				End:      end,
			}
			continue
		}

		if end.After(current.End) {
			current.End = end
		}
	}

	if current != nil {
		sessions = append(sessions, *current)
	}

	for i := range sessions {
		duration := sessions[i].End.Sub(sessions[i].Start)
		if duration > maxSessionLength {
			duration = maxSessionLength
		}
		sessions[i].DurationSeconds = int(duration.Seconds())
	}

	return sessions
}
