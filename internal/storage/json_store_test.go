package storage

import (
	"path/filepath"
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
)

func testStore(t *testing.T) *JSONStore {
	t.Helper()

	dir := t.TempDir()
	s, err := NewJSONStore(Dirs{Data: filepath.Join(dir, "data"), Config: filepath.Join(dir, "config")})
	if err != nil {
		t.Fatalf("NewJSONStore: %v", err)
	}
	return s
}

func TestJSONStoreSettingsRoundTrip(t *testing.T) {
	t.Parallel()

	s := testStore(t)

	settings := domain.Settings{
		DefaultDuration:  domain.Duration30,
		DefaultWordCount: 25,
		DefaultWidth:     80,
		Theme:            "dracula",
	}
	if err := s.SaveSettings(settings); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}

	got, err := s.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings: %v", err)
	}
	if got != settings {
		t.Fatalf("settings = %+v, want %+v", got, settings)
	}
}

func TestJSONStoreMissingSettingsReturnsDefaults(t *testing.T) {
	t.Parallel()

	s := testStore(t)

	got, err := s.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings: %v", err)
	}
	want := domain.DefaultSettings()
	if got != want {
		t.Fatalf("settings = %+v, want %+v", got, want)
	}
}
