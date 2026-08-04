package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
	"golang.org/x/sys/unix"
)

const settingFilename = "config.json"

type JSONStore struct {
	Dir string
}

func NewJSONStore(dir string) (*JSONStore, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, fmt.Errorf("config dir is empty")
	}
	dir = filepath.Clean(dir)
	if err := ensureDir(dir); err != nil {
		return nil, fmt.Errorf("config dir: %w", err)
	}
	return &JSONStore{Dir: dir}, nil
}

func (s *JSONStore) Path() string {
	return s.Dir
}

func NewDefaultStore() (*JSONStore, error) {
	var dir string
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dir = filepath.Join(home, "Library", "Preferences", "ttype")
	case "windows":
		base := os.Getenv("APPDATA")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}
			base = filepath.Join(home, "AppData", "Roaming")
		}

		dir = filepath.Join(base, "ttype")
	default:
		if base := os.Getenv("XDG_DATA_HOME"); base != "" {
			dir = filepath.Join(base, "ttype")
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}
			dir = filepath.Join(home, ".config", "ttype")
		}
	}

	return NewJSONStore(dir)
}

func (s *JSONStore) LoadSettings() (domain.Settings, error) {
	path := filepath.Join(s.Dir, settingFilename)
	settings := domain.DefaultSettings()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return settings, nil
		}
		return settings, fmt.Errorf("read settings: %w", err)
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		return domain.DefaultSettings(), fmt.Errorf("parse settings: %w", err)
	}

	return settings, nil
}

func (s *JSONStore) SaveSettings(settings domain.Settings) error {
	lock, err := acquireLock(s.Dir)
	if err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	defer lock.release()

	path := filepath.Join(s.Dir, settingFilename)
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o700)
}

type fileLock struct {
	f *os.File
}

func acquireLock(path string) (*fileLock, error) {
	f, err := os.OpenFile(filepath.Join(path, ".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("acquire lock file: %w", err)
	}

	return &fileLock{f: f}, nil
}

func (l *fileLock) release() {
	_ = unix.Flock(int(l.f.Fd()), unix.LOCK_UN)
	_ = l.f.Close()
}
