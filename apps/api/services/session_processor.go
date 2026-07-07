package services

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const sessionGap = 2 * time.Minute
const maxSessionLength = 6 * time.Hour

type rawHeartbeat struct {
	ID       string
	UserID   string
	Project  string
	Language string
	TimeMs   int64
}

// ProcessSessions groups unprocessed heartbeats into sessions
// and stores them. Meant to run periodically in the background.
func ProcessSessions(ctx context.Context, pool *pgxpool.Pool) error {
	rows, err := pool.Query(ctx, `
		SELECT id, user_id, project, language, time
		FROM heartbeats
		WHERE processed = false
		ORDER BY user_id, project, time ASC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var heartbeats []rawHeartbeat
	for rows.Next() {
		var h rawHeartbeat
		if err := rows.Scan(&h.ID, &h.UserID, &h.Project, &h.Language, &h.TimeMs); err != nil {
			return err
		}
		heartbeats = append(heartbeats, h)
	}

	sessions := buildSessions(heartbeats)

	for _, s := range sessions {
		_, err := pool.Exec(ctx, `
			INSERT INTO sessions (user_id, project, language, start_time, end_time, duration_seconds)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, s.UserID, s.Project, s.Language, s.Start, s.End, s.DurationSeconds)
		if err != nil {
			return err
		}
	}

	ids := make([]string, len(heartbeats))
	for i, h := range heartbeats {
		ids[i] = h.ID
	}
	if len(ids) > 0 {
		_, err = pool.Exec(ctx, `UPDATE heartbeats SET processed = true WHERE id = ANY($1)`, ids)
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
	Start           time.Time
	End             time.Time
	DurationSeconds int
}

// buildSessions groups a time-ordered list of heartbeats into sessions,
// splitting whenever the gap exceeds sessionGap.
func buildSessions(heartbeats []rawHeartbeat) []builtSession {
	var sessions []builtSession
	if len(heartbeats) == 0 {
		return sessions
	}

	var current *builtSession

	for _, h := range heartbeats {
		t := time.UnixMilli(h.TimeMs)

		startsNew := current == nil ||
			current.UserID != h.UserID ||
			current.Project != h.Project ||
			t.Sub(current.End) > sessionGap

		if startsNew {
			if current != nil {
				sessions = append(sessions, *current)
			}
			current = &builtSession{
				UserID:   h.UserID,
				Project:  h.Project,
				Language: h.Language,
				Start:    t,
				End:      t,
			}
			continue
		}

		current.End = t
		current.Language = h.Language // last language seen in the session
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
