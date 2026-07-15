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
	TimeMs   int64
}

// ProcessSessions groups unprocessed heartbeats into sessions
// and stores them. Meant to run periodically in the background.
func ProcessSessions(ctx context.Context, pool *pgxpool.Pool) error {
	rows, err := pool.Query(ctx, `
		SELECT id, user_id, project, language, time
		FROM heartbeats
		WHERE processed = false
		ORDER BY user_id, time ASC, project, language
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
		if err := saveSession(ctx, pool, s); err != nil {
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

type existingSession struct {
	ID       string
	Project  string
	Language string
	Start    time.Time
	End      time.Time
}

func saveSession(ctx context.Context, pool *pgxpool.Pool, s builtSession) error {
	latest, err := latestSession(ctx, pool, s.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return insertSession(ctx, pool, s)
		}
		return err
	}

	if latest.Project == s.Project &&
		latest.Language == s.Language &&
		s.Start.Sub(latest.End.Add(-heartbeatDuration)) <= sessionGap {
		end := latest.End
		if s.End.After(end) {
			end = s.End
		}

		duration := end.Sub(latest.Start)
		if duration > maxSessionLength {
			duration = maxSessionLength
		}

		_, err := pool.Exec(ctx, `
			UPDATE sessions
			SET end_time = $1, duration_seconds = $2
			WHERE id = $3
		`, end, int(duration.Seconds()), latest.ID)
		return err
	}

	return insertSession(ctx, pool, s)
}

func latestSession(ctx context.Context, pool *pgxpool.Pool, userID string) (existingSession, error) {
	var s existingSession
	err := pool.QueryRow(ctx, `
		SELECT id, project, language, start_time, end_time
		FROM sessions
		WHERE user_id = $1
		ORDER BY end_time DESC
		LIMIT 1
	`, userID).Scan(&s.ID, &s.Project, &s.Language, &s.Start, &s.End)
	return s, err
}

func insertSession(ctx context.Context, pool *pgxpool.Pool, s builtSession) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO sessions (user_id, project, language, start_time, end_time, duration_seconds)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, s.UserID, s.Project, s.Language, s.Start, s.End, s.DurationSeconds)
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

	for _, h := range heartbeats {
		t := time.UnixMilli(h.TimeMs)
		end := t.Add(heartbeatDuration)

		startsNew := current == nil ||
			current.UserID != h.UserID ||
			current.Project != h.Project ||
			current.Language != h.Language ||
			t.Sub(current.End.Add(-heartbeatDuration)) > sessionGap

		if startsNew {
			if current != nil {
				sessions = append(sessions, *current)
			}
			current = &builtSession{
				UserID:   h.UserID,
				Project:  h.Project,
				Language: h.Language,
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
