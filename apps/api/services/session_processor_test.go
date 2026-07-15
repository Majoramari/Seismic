package services

import (
	"testing"
	"time"
)

func TestBuildSessionsDoesNotDoubleCountEqualTimestamps(t *testing.T) {
	const base = int64(1714564800000)
	heartbeats := []rawHeartbeat{
		{
			UserID:   "user-1",
			Project:  "seismic",
			Language: "Go",
			Editor:   "VS Code",
			OS:       "Linux",
			TimeMs:   base,
		},
		{
			UserID:   "user-1",
			Project:  "seismic",
			Language: "TypeScript",
			Editor:   "VS Code",
			OS:       "Linux",
			TimeMs:   base,
		},
		{
			UserID:   "user-1",
			Project:  "seismic",
			Language: "TypeScript",
			Editor:   "VS Code",
			OS:       "Linux",
			TimeMs:   base + int64(heartbeatDuration/time.Millisecond),
		},
	}

	sessions := buildSessions(heartbeats)
	totalSeconds := 0
	for _, session := range sessions {
		totalSeconds += session.DurationSeconds
	}

	if totalSeconds != int((2 * heartbeatDuration).Seconds()) {
		t.Fatalf("totalSeconds = %d, want %d", totalSeconds, int((2 * heartbeatDuration).Seconds()))
	}
}
