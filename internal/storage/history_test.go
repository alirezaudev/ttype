package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestSaveResultRoundTrip(t *testing.T) {
	t.Parallel()

	s := testStore(t)
	want := domain.Result{
		ID:                  "abc123",
		Timestamp:           time.Date(2026, 8, 7, 10, 30, 0, 0, time.UTC),
		Config:              domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60, Theme: "dracula", Language: "spanish", Width: 80},
		WPM:                 92.5,
		RawWPM:              98.25,
		Accuracy:            96.5,
		Correct:             270,
		Incorrect:           10,
		KeystrokesCorrect:   280,
		KeystrokesIncorrect: 12,
		TotalChars:          280,
		Duration:            60 * time.Second,
	}

	if _, err := s.SaveResult(want); err != nil {
		t.Fatalf("SaveResult: %v", err)
	}

	got, err := s.ListResults(0)
	if err != nil {
		t.Fatalf("ListResults: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("results = %d, want 1", len(got))
	}
	if !reflect.DeepEqual(got[0], want) {
		t.Fatalf("result = %+v, want %+v", got[0], want)
	}
}

func TestListResultsTolerantOfMissingFields(t *testing.T) {
	t.Parallel()

	s := testStore(t)
	fixture := `[{"id":"old","timestamp":"2026-01-01T00:00:00Z","wpm":50}]`
	if err := os.WriteFile(filepath.Join(s.Paths().Data, historyFilename), []byte(fixture), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	got, err := s.ListResults(0)
	if err != nil {
		t.Fatalf("ListResults: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("results = %d, want 1", len(got))
	}
	if got[0].ID != "old" || got[0].WPM != 50 {
		t.Fatalf("result = %+v, want the stored id and wpm", got[0])
	}
	if got[0].Accuracy != 0 || got[0].TotalChars != 0 {
		t.Fatalf("result = %+v, want zero values for the absent fields", got[0])
	}
}

func TestSaveResultEvictsOldest(t *testing.T) {
	t.Parallel()

	s := testStore(t)
	s.SetHistoryCap(3)

	for i := 1; i <= 5; i++ {
		result := domain.Result{
			ID:        fmt.Sprintf("run%d", i),
			Timestamp: time.Now().UTC(),
			Config:    domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60},
		}
		if _, err := s.SaveResult(result); err != nil {
			t.Fatalf("SaveResult: %v", err)
		}
	}

	got, err := s.ListResults(0)
	if err != nil {
		t.Fatalf("ListResults: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("results = %d, want 3", len(got))
	}
	if got[0].ID != "run5" || got[2].ID != "run3" {
		t.Fatalf("results = %s…%s, want run5…run3 (newest first)", got[0].ID, got[2].ID)
	}
}
