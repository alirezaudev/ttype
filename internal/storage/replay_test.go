package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestReplayRoundTrip(t *testing.T) {
	t.Parallel()

	s := testStore(t)
	want := domain.Replay{
		Target: "the cat sat",
		Events: []domain.ReplayEvent{
			{Offset: 0, Kind: domain.ReplayRune, Rune: 't'},
			{Offset: 140 * time.Millisecond, Kind: domain.ReplayRune, Rune: 'h'},
			{Offset: 500 * time.Millisecond, Kind: domain.ReplayBackspace},
			{Offset: 900 * time.Millisecond, Kind: domain.ReplayDeleteWord},
			{Offset: 2100 * time.Millisecond, Kind: domain.ReplayRune, Rune: 'é'},
		},
	}

	if err := s.SaveReplay("abc123", want); err != nil {
		t.Fatalf("SaveReplay: %v", err)
	}
	got, err := s.LoadReplay("abc123")
	if err != nil {
		t.Fatalf("LoadReplay: %v", err)
	}

	if got.Target != want.Target {
		t.Fatalf("target = %q, want %q", got.Target, want.Target)
	}
	if len(got.Events) != len(want.Events) {
		t.Fatalf("events = %d, want %d", len(got.Events), len(want.Events))
	}
	for i, ev := range want.Events {
		if got.Events[i] != ev {
			t.Fatalf("event %d = %+v, want %+v", i, got.Events[i], ev)
		}
	}
}

func TestLoadReplayMissingAndCorrupt(t *testing.T) {
	t.Parallel()

	s := testStore(t)
	if _, err := s.LoadReplay("nothinghere"); !errors.Is(err, ErrNoReplay) {
		t.Fatalf("LoadReplay err = %v, want ErrNoReplay", err)
	}

	if err := ensureDir(s.replaysDir()); err != nil {
		t.Fatalf("ensureDir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(s.replaysDir(), "junk.bin"), []byte("hello"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := s.LoadReplay("junk"); err == nil {
		t.Fatal("LoadReplay on garbage = nil error, want failure")
	}
}

func TestSaveReplayPrunesOldest(t *testing.T) {
	t.Parallel()

	s := testStore(t)
	replay := domain.Replay{
		Target: "a",
		Events: []domain.ReplayEvent{{Kind: domain.ReplayRune, Rune: 'a'}},
	}

	base := time.Now().Add(-time.Hour)
	for i := 0; i < maxReplayFiles+5; i++ {
		id := fmt.Sprintf("run%03d", i)
		if err := s.SaveReplay(id, replay); err != nil {
			t.Fatalf("SaveReplay %s: %v", id, err)
		}
		stamp := base.Add(time.Duration(i) * time.Minute)
		if err := os.Chtimes(filepath.Join(s.replaysDir(), id+".bin"), stamp, stamp); err != nil {
			t.Fatalf("chtimes: %v", err)
		}
	}

	entries, err := os.ReadDir(s.replaysDir())
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != maxReplayFiles {
		t.Fatalf("kept %d replays, want %d", len(entries), maxReplayFiles)
	}
	if _, err := s.LoadReplay("run000"); !errors.Is(err, ErrNoReplay) {
		t.Fatalf("oldest replay still present: %v", err)
	}
}
