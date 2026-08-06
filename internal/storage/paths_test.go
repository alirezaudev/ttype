package storage

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultDirsFollowsXDG(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skip("xdg lookup is unix only")
	}

	data := t.TempDir()
	config := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	t.Setenv("XDG_CONFIG_HOME", config)

	dirs, err := DefaultDirs()
	if err != nil {
		t.Fatalf("DefaultDirs: %v", err)
	}

	if want := filepath.Join(data, "ttype"); dirs.Data != want {
		t.Fatalf("data = %q, want %q", dirs.Data, want)
	}
	if want := filepath.Join(config, "ttype"); dirs.Config != want {
		t.Fatalf("config = %q, want %q", dirs.Config, want)
	}
}

func TestDefaultDirsKeepsDataOutOfConfig(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skip("xdg lookup is unix only")
	}

	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")

	dirs, err := DefaultDirs()
	if err != nil {
		t.Fatalf("DefaultDirs: %v", err)
	}

	if dirs.Data == dirs.Config {
		t.Fatalf("data and config resolved to the same dir: %q", dirs.Data)
	}
}
