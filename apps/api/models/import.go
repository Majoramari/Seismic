package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WakaTimeHeartbeat mirrors a single heartbeat entry from a WakaTime
// data dump export.
type WakaTimeHeartbeat struct {
	Entity          string `json:"entity"`
	Type            string `json:"type"`
	Project         string `json:"project"`
	Language        string `json:"language"`
	Editor          string `json:"editor"`
	OperatingSystem string `json:"operating_system"`
	Machine         string `json:"machine"`
	Branch          string `json:"branch"`
	Lines           *int   `json:"lines"`
	Cursorpos       *int   `json:"cursorpos"`
	Time            string `json:"time"` // RFC3339
}

type ImportExportInfo struct {
	TotalHeartbeats      int `json:"total_heartbeats"`
	TotalDurationSeconds int `json:"total_duration_seconds"`
}

type ImportFile struct {
	ExportInfo ImportExportInfo    `json:"export_info"`
	Heartbeats []WakaTimeHeartbeat `json:"heartbeats"`
	Days       []struct {
		Heartbeats []WakaTimeHeartbeat `json:"heartbeats"`
	} `json:"days"`
}

// InsertImportedHeartbeats bulk-inserts heartbeats from an external
// import (WakaTime, etc). They land in the same heartbeats table as
// live editor pings, marked unprocessed so the normal session
// processor picks them up on its next run.
func InsertImportedHeartbeats(ctx context.Context, pool *pgxpool.Pool, userID string, heartbeats []WakaTimeHeartbeat, onProgress func(inserted int)) (int, error) {
	const batchSize = 500
	inserted := 0

	for i := 0; i < len(heartbeats); i += batchSize {
		end := i + batchSize
		if end > len(heartbeats) {
			end = len(heartbeats)
		}
		chunk := heartbeats[i:end]

		batch := &pgx.Batch{}
		for _, hb := range chunk {
			t, err := parseWakaTimeTime(hb.Time)
			if err != nil {
				continue
			}
			lines := 0
			if hb.Lines != nil {
				lines = *hb.Lines
			}
			cursor := 0
			if hb.Cursorpos != nil {
				cursor = *hb.Cursorpos
			}
			batch.Queue(`
				INSERT INTO heartbeats
					(user_id, file, project, language, editor, branch, os, machine, lines, cursor_line, time)
				SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
				WHERE NOT EXISTS (
					SELECT 1 FROM heartbeats
					WHERE user_id = $1
						AND file = $2
						AND project = $3
						AND language = $4
						AND editor = $5
						AND time = $11
				)
			`, userID, hb.Entity, hb.Project, hb.Language, hb.Editor, hb.Branch,
				hb.OperatingSystem, hb.Machine, lines, cursor, t)
		}

		br := pool.SendBatch(ctx, batch)
		for range chunk {
			if tag, err := br.Exec(); err == nil {
				inserted += int(tag.RowsAffected())
			}
		}
		br.Close()

		if onProgress != nil {
			onProgress(inserted)
		}
	}

	return inserted, nil
}

func parseWakaTimeTime(s string) (int64, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0, err
	}
	return t.UnixMilli(), nil
}

// FlattenImportFile combines both possible export shapes into one list.
func FlattenImportFile(f ImportFile) []WakaTimeHeartbeat {
	all := append([]WakaTimeHeartbeat{}, f.Heartbeats...)
	for _, day := range f.Days {
		all = append(all, day.Heartbeats...)
	}
	return all
}
