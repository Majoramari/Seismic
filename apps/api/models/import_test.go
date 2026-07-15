package models

import (
	"encoding/json"
	"testing"
)

func TestImportTimeUnixMilli(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{
			name:  "rfc3339",
			input: `"2024-05-01T12:00:00Z"`,
			want:  1714564800000,
		},
		{
			name:  "unix seconds",
			input: `1714564800`,
			want:  1714564800000,
		},
		{
			name:  "unix seconds float",
			input: `1714564800.5`,
			want:  1714564800500,
		},
		{
			name:  "unix milliseconds",
			input: `1714564800000`,
			want:  1714564800000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ImportTime
			if err := json.Unmarshal([]byte(tt.input), &got); err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}

			gotMs, err := got.UnixMilli()
			if err != nil {
				t.Fatalf("UnixMilli() error = %v", err)
			}
			if gotMs != tt.want {
				t.Fatalf("UnixMilli() = %d, want %d", gotMs, tt.want)
			}
		})
	}
}

func TestNormalizeImportedHeartbeatsDedupesAndSorts(t *testing.T) {
	heartbeats := []WakaTimeHeartbeat{
		{
			Entity:          " b.go ",
			Project:         " seismic ",
			Language:        " Go ",
			Editor:          " VS Code ",
			OperatingSystem: " Linux ",
			Time:            ImportTime("2024-05-01T12:00:30Z"),
		},
		{
			Entity:          "a.go",
			Project:         "seismic",
			Language:        "Go",
			Editor:          "VS Code",
			OperatingSystem: "Linux",
			Time:            ImportTime("2024-05-01T12:00:00Z"),
		},
		{
			Entity:          " a.go ",
			Project:         " seismic ",
			Language:        " Go ",
			Editor:          " VS Code ",
			OperatingSystem: " Linux ",
			Time:            ImportTime("2024-05-01T12:00:00Z"),
		},
		{
			Entity:   "bad.go",
			Project:  "seismic",
			Language: "Go",
			Editor:   "VS Code",
			Time:     ImportTime("not-a-time"),
		},
	}

	got := NormalizeImportedHeartbeats(heartbeats)
	if len(got) != 2 {
		t.Fatalf("len(NormalizeImportedHeartbeats()) = %d, want 2", len(got))
	}

	if got[0].Entity != "a.go" {
		t.Fatalf("first heartbeat entity = %q, want %q", got[0].Entity, "a.go")
	}
	if got[1].Entity != "b.go" {
		t.Fatalf("second heartbeat entity = %q, want %q", got[1].Entity, "b.go")
	}
	if got[1].Project != "seismic" || got[1].Language != "Go" || got[1].Editor != "VS Code" {
		t.Fatalf("heartbeat fields were not trimmed: %#v", got[1])
	}
}
