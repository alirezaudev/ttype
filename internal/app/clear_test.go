package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alirezaudev/ttype/internal/storage"
)

func clearStore(t *testing.T) *storage.JSONStore {
	t.Helper()

	dir := t.TempDir()
	store, err := storage.NewJSONStore(storage.Dirs{
		Data:   filepath.Join(dir, "data"),
		Config: filepath.Join(dir, "config"),
	})
	if err != nil {
		t.Fatalf("NewJSONStore: %v", err)
	}

	data := store.Paths().Data
	for _, name := range []string{"history.json", "bests.json"} {
		if err := os.WriteFile(filepath.Join(data, name), []byte("[]"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	for _, name := range []string{"replays", "languages"} {
		if err := os.MkdirAll(filepath.Join(data, name), 0o700); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
	}
	return store
}

func TestClearHistoryKeepsLanguages(t *testing.T) {
	t.Parallel()

	store := clearStore(t)
	data := store.Paths().Data

	var out bytes.Buffer
	if err := RunClear(store, ClearHistory, true, &out, strings.NewReader("")); err != nil {
		t.Fatalf("RunClear: %v", err)
	}

	for _, name := range []string{"history.json", "bests.json", "replays"} {
		if _, err := os.Stat(filepath.Join(data, name)); !os.IsNotExist(err) {
			t.Errorf("%s survived the clear", name)
		}
	}
	if _, err := os.Stat(filepath.Join(data, "languages")); err != nil {
		t.Error("clearing history should leave downloaded languages alone")
	}
}

func TestClearAllRemovesEverything(t *testing.T) {
	t.Parallel()

	store := clearStore(t)
	data := store.Paths().Data

	var out bytes.Buffer
	if err := RunClear(store, ClearAll, true, &out, strings.NewReader("")); err != nil {
		t.Fatalf("RunClear: %v", err)
	}

	for _, name := range []string{"history.json", "bests.json", "replays", "languages"} {
		if _, err := os.Stat(filepath.Join(data, name)); !os.IsNotExist(err) {
			t.Errorf("%s survived the clear", name)
		}
	}
}

func TestClearStopsWithoutConfirmation(t *testing.T) {
	t.Parallel()

	store := clearStore(t)

	var out bytes.Buffer
	if err := RunClear(store, ClearAll, false, &out, strings.NewReader("n\n")); err != nil {
		t.Fatalf("RunClear: %v", err)
	}

	if _, err := os.Stat(filepath.Join(store.Paths().Data, "history.json")); err != nil {
		t.Error("history was deleted despite the answer being no")
	}
	if !strings.Contains(out.String(), "Nothing was deleted") {
		t.Errorf("output = %q", out.String())
	}
}

func TestParseClearTarget(t *testing.T) {
	t.Parallel()

	if _, err := ParseClearTarget("everything"); err == nil {
		t.Fatal("an unknown target should be rejected")
	}
	if got, err := ParseClearTarget("languages"); err != nil || got != ClearLanguages {
		t.Fatalf("got %q/%v", got, err)
	}
}
