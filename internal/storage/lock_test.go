package storage

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

// Two stores over the same directory stand in for two ttype processes
// finishing a run at the same moment: separate descriptors, so they contend
// through the lock file the way real processes do.
func TestConcurrentStoresKeepEveryResult(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dirs := Dirs{Data: filepath.Join(dir, "data"), Config: filepath.Join(dir, "config")}

	first, err := NewJSONStore(dirs)
	if err != nil {
		t.Fatalf("NewJSONStore: %v", err)
	}
	second, err := NewJSONStore(dirs)
	if err != nil {
		t.Fatalf("NewJSONStore: %v", err)
	}

	const runs = 20
	var wg sync.WaitGroup
	errs := make(chan error, runs*2)

	for i := 0; i < runs; i++ {
		for storeIndex, store := range []*JSONStore{first, second} {
			wg.Add(1)
			go func(store *JSONStore, id string) {
				defer wg.Done()
				result := domain.Result{
					ID:        id,
					Timestamp: time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
					Config:    domain.TestConfig{Kind: domain.TestKindTimed, Duration: 60},
					WPM:       60,
				}
				if _, err := store.SaveResult(result); err != nil {
					errs <- err
				}
			}(store, fmt.Sprintf("%d-%d", storeIndex, i))
		}
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("SaveResult: %v", err)
	}

	results, err := first.ListResults(0)
	if err != nil {
		t.Fatalf("ListResults: %v", err)
	}
	if len(results) != runs*2 {
		t.Fatalf("stored %d results, want %d — a concurrent save was lost", len(results), runs*2)
	}
}

func TestSettingsSurviveConcurrentSaves(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dirs := Dirs{Data: filepath.Join(dir, "data"), Config: filepath.Join(dir, "config")}

	stores := make([]*JSONStore, 4)
	for i := range stores {
		store, err := NewJSONStore(dirs)
		if err != nil {
			t.Fatalf("NewJSONStore: %v", err)
		}
		stores[i] = store
	}

	themes := []string{"default", "monokai", "dracula"}

	var wg sync.WaitGroup
	for i, store := range stores {
		wg.Add(1)
		go func(store *JSONStore, theme string) {
			defer wg.Done()
			settings := domain.DefaultSettings()
			settings.Theme = theme
			if err := store.SaveSettings(settings); err != nil {
				t.Errorf("SaveSettings: %v", err)
			}
		}(store, themes[i%len(themes)])
	}
	wg.Wait()

	settings, err := stores[0].LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings: %v", err)
	}
	if settings.Theme == "" {
		t.Fatal("settings file was left half written")
	}
}
