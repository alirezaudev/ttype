package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestResultFileStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		result        domain.Result
		finished, ran bool
		want          string
	}{
		{"completed", domain.Result{}, true, true, "completed"},
		{"quit", domain.Result{}, false, true, "quit"},
		{"nothing typed", domain.Result{}, false, false, "quit"},
		{"min wpm", domain.Result{Failed: true}, true, true, "failed_min_wpm"},
	}
	for _, tc := range cases {
		if got := newResultFile(tc.result, tc.finished, tc.ran, ResultSource{}).Status; got != tc.want {
			t.Errorf("%s: status = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestResultFileDescribesTheRun(t *testing.T) {
	t.Parallel()

	end := time.Date(2026, 9, 18, 10, 0, 30, 0, time.UTC)
	result := domain.Result{
		Timestamp:  end,
		Duration:   30 * time.Second,
		Config:     domain.TestConfig{TextMode: domain.TextModeCustom},
		WPM:        50,
		Correct:    10,
		Incorrect:  2,
		TotalChars: 15,
		Words:      []domain.WordResult{{Expected: "letter", Typed: "lettter", Missed: true}},
	}
	rf := newResultFile(result, true, true, ResultSource{File: "typing.txt", Text: "I  think"})

	if rf.StartedAt != "2026-09-18T10:00:00Z" || rf.DurationS != 30 {
		t.Fatalf("started %s, lasted %v", rf.StartedAt, rf.DurationS)
	}
	if rf.Chars.Extra != 3 {
		t.Fatalf("extra = %d, want 3", rf.Chars.Extra)
	}
	// Whitespace doesn't change the hash.
	want := newResultFile(result, true, true, ResultSource{Text: "I think"}).Source.SHA256
	if rf.Source == nil || rf.Source.File != "typing.txt" || rf.Source.SHA256 != want {
		t.Fatalf("source = %+v", rf.Source)
	}

	path := filepath.Join(t.TempDir(), "run.json")
	if err := writeResultFile(path, rf); err != nil {
		t.Fatalf("writeResultFile: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var back map[string]any
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if back["version"] != float64(1) || !strings.Contains(string(data), `"typed": "lettter"`) {
		t.Fatalf("file = %s", data)
	}
}
