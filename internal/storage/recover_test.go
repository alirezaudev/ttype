package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestBrokenFilesAreMovedAside(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path func(*JSONStore) string
		load func(*JSONStore) error
	}{
		{"config", func(s *JSONStore) string { return filepath.Join(s.dirs.Config, settingFilename) }, func(s *JSONStore) error {
			got, err := s.LoadSettings()
			if err == nil && got != domain.DefaultSettings() {
				t.Errorf("settings = %+v, want defaults", got)
			}
			return err
		}},
		{"history", func(s *JSONStore) string { return filepath.Join(s.dirs.Data, historyFilename) }, func(s *JSONStore) error {
			got, err := s.ListResults(0)
			if err == nil && len(got) != 0 {
				t.Errorf("history has %d runs, want 0", len(got))
			}
			return err
		}},
		{"bests", func(s *JSONStore) string { return s.bestsPath() }, func(s *JSONStore) error {
			got, err := s.LoadBests()
			if err == nil && (got.BestWPM != 0 || len(got.ByConfig) != 0) {
				t.Errorf("bests = %+v, want empty", got)
			}
			return err
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			s := testStore(t)
			path := test.path(s)
			if err := os.WriteFile(path, []byte(`{"theme": "dracula",`), 0o600); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}
			if errs := CheckFiles(s.Paths()); len(errs) != 1 || !strings.Contains(errs[0].Error(), filepath.Base(path)) {
				t.Fatalf("CheckFiles = %v, want one error naming %s", errs, filepath.Base(path))
			}

			for i := 0; i < 2; i++ {
				if err := test.load(s); err != nil {
					t.Fatalf("load %d: %v", i+1, err)
				}
			}

			copies, _ := filepath.Glob(path + ".broken-*")
			if len(copies) != 1 {
				t.Fatalf("broken copies = %v, want one", copies)
			}
			if got := s.Recovered(); len(got) != 1 || !strings.Contains(got[0], copies[0]) {
				t.Fatalf("Recovered = %v, want one notice naming %s", got, copies[0])
			}
			if errs := CheckFiles(s.Paths()); len(errs) != 0 {
				t.Fatalf("CheckFiles after recovery = %v", errs)
			}
		})
	}
}

func TestSaveResultAfterBrokenFiles(t *testing.T) {
	t.Parallel()

	s := testStore(t)
	for _, path := range []string{filepath.Join(s.dirs.Data, historyFilename), s.bestsPath()} {
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}

	result := domain.Result{ID: "r1", Timestamp: time.Now(), WPM: 80, Accuracy: 97,
		Config: domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration30}}
	if _, err := s.SaveResult(result); err != nil {
		t.Fatalf("SaveResult: %v", err)
	}
	got, err := s.ListResults(0)
	if err != nil || len(got) != 1 {
		t.Fatalf("ListResults = %d runs, %v; want 1", len(got), err)
	}
	if bests, _ := s.LoadBests(); bests.BestWPM != 80 {
		t.Fatalf("best wpm = %v, want 80", bests.BestWPM)
	}
}
