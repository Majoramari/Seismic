package models

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WakaTimeHeartbeat mirrors a single heartbeat entry from a WakaTime
// data dump export.
type WakaTimeHeartbeat struct {
	Entity          string     `json:"entity"`
	Type            string     `json:"type"`
	Project         string     `json:"project"`
	Language        string     `json:"language"`
	Editor          string     `json:"editor"`
	OperatingSystem string     `json:"operating_system"`
	Machine         string     `json:"machine"`
	Branch          string     `json:"branch"`
	Lines           *int       `json:"lines"`
	Cursorpos       *int       `json:"cursorpos"`
	Time            ImportTime `json:"time"`
}

type ImportTime string

func (t *ImportTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*t = ImportTime(strings.TrimSpace(s))
		return nil
	}

	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*t = ImportTime(n.String())
		return nil
	}

	return fmt.Errorf("invalid heartbeat time")
}

func (t ImportTime) UnixMilli() (int64, error) {
	value := strings.TrimSpace(string(t))
	if value == "" {
		return 0, fmt.Errorf("empty heartbeat time")
	}

	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed.UnixMilli(), nil
	}

	if n, err := strconv.ParseInt(value, 10, 64); err == nil {
		if absInt64(n) > 100_000_000_000 {
			return n, nil
		}
		return n * 1000, nil
	}

	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid heartbeat time %q", value)
	}
	if absFloat64(f) > 100_000_000_000 {
		return int64(f), nil
	}
	return int64(f * 1000), nil
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
	heartbeats = NormalizeImportedHeartbeats(heartbeats)
	inserted := 0

	for i := 0; i < len(heartbeats); i += batchSize {
		end := i + batchSize
		if end > len(heartbeats) {
			end = len(heartbeats)
		}
		chunk := heartbeats[i:end]

		batch := &pgx.Batch{}
		queued := 0
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
			queued++
		}

		br := pool.SendBatch(ctx, batch)
		for range queued {
			tag, err := br.Exec()
			if err != nil {
				br.Close()
				return inserted, err
			}
			inserted += int(tag.RowsAffected())
		}
		if err := br.Close(); err != nil {
			return inserted, err
		}

		if onProgress != nil {
			onProgress(inserted)
		}
	}

	return inserted, nil
}

func parseWakaTimeTime(t ImportTime) (int64, error) {
	return t.UnixMilli()
}

// FlattenImportFile combines both possible export shapes into one list.
func FlattenImportFile(f ImportFile) []WakaTimeHeartbeat {
	all := append([]WakaTimeHeartbeat{}, f.Heartbeats...)
	for _, day := range f.Days {
		all = append(all, day.Heartbeats...)
	}
	return all
}

type normalizedImportHeartbeat struct {
	heartbeat WakaTimeHeartbeat
	timeMs    int64
}

func NormalizeImportedHeartbeats(heartbeats []WakaTimeHeartbeat) []WakaTimeHeartbeat {
	items := make([]normalizedImportHeartbeat, 0, len(heartbeats))
	seen := make(map[string]struct{}, len(heartbeats))

	for _, hb := range heartbeats {
		hb = normalizeImportedHeartbeat(hb)

		timeMs, err := parseWakaTimeTime(hb.Time)
		if err != nil {
			continue
		}

		key := importHeartbeatKey(hb, timeMs)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		items = append(items, normalizedImportHeartbeat{
			heartbeat: hb,
			timeMs:    timeMs,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].timeMs != items[j].timeMs {
			return items[i].timeMs < items[j].timeMs
		}

		return compareImportHeartbeat(items[i].heartbeat, items[j].heartbeat) < 0
	})

	normalized := make([]WakaTimeHeartbeat, 0, len(items))
	for _, item := range items {
		normalized = append(normalized, item.heartbeat)
	}

	return normalized
}

func normalizeImportedHeartbeat(hb WakaTimeHeartbeat) WakaTimeHeartbeat {
	hb.Entity = strings.TrimSpace(hb.Entity)
	hb.Type = strings.TrimSpace(hb.Type)
	hb.Project = strings.TrimSpace(hb.Project)
	hb.Language = strings.TrimSpace(hb.Language)
	hb.Editor = strings.TrimSpace(hb.Editor)
	hb.OperatingSystem = strings.TrimSpace(hb.OperatingSystem)
	hb.Machine = strings.TrimSpace(hb.Machine)
	hb.Branch = strings.TrimSpace(hb.Branch)
	return hb
}

func importHeartbeatKey(hb WakaTimeHeartbeat, timeMs int64) string {
	return strings.Join([]string{
		strconv.FormatInt(timeMs, 10),
		hb.Entity,
		hb.Project,
		hb.Language,
		hb.Editor,
		hb.OperatingSystem,
		hb.Machine,
		hb.Branch,
	}, "\x00")
}

func compareImportHeartbeat(a, b WakaTimeHeartbeat) int {
	aFields := []string{a.Project, a.Language, a.Editor, a.OperatingSystem, a.Entity, a.Machine, a.Branch}
	bFields := []string{b.Project, b.Language, b.Editor, b.OperatingSystem, b.Entity, b.Machine, b.Branch}

	for i := range aFields {
		if aFields[i] < bFields[i] {
			return -1
		}
		if aFields[i] > bFields[i] {
			return 1
		}
	}

	return 0
}

func absInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

func absFloat64(n float64) float64 {
	if n < 0 {
		return -n
	}
	return n
}
