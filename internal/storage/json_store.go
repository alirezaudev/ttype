package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alirezaudev/ttype/internal/domain"
)

const settingFilename = "config.json"

type JSONStore struct {
	dirs       Dirs
	historyCap int

	mu        sync.Mutex
	recovered []string
}

func NewJSONStore(dirs Dirs) (*JSONStore, error) {
	dirs.Data = strings.TrimSpace(dirs.Data)
	dirs.Config = strings.TrimSpace(dirs.Config)
	if dirs.Data == "" || dirs.Config == "" {
		return nil, fmt.Errorf("store dirs are empty")
	}

	dirs.Data = filepath.Clean(dirs.Data)
	dirs.Config = filepath.Clean(dirs.Config)
	if err := ensureDir(dirs.Data); err != nil {
		return nil, fmt.Errorf("data dir: %w", err)
	}
	if err := ensureDir(dirs.Config); err != nil {
		return nil, fmt.Errorf("config dir: %w", err)
	}

	return &JSONStore{dirs: dirs}, nil
}

func (s *JSONStore) Paths() Dirs {
	return s.dirs
}

func NewDefaultStore() (*JSONStore, error) {
	dirs, err := DefaultDirs()
	if err != nil {
		return nil, err
	}

	return NewJSONStore(dirs)
}

func (s *JSONStore) LoadSettings() (domain.Settings, error) {
	settings := domain.DefaultSettings()
	ok, err := s.readJSON(filepath.Join(s.dirs.Config, settingFilename), &settings)
	if !ok {
		return domain.DefaultSettings(), err
	}
	return settings, nil
}

func (s *JSONStore) SaveSettings(settings domain.Settings) error {
	lock, err := acquireLock(s.dirs.Config)
	if err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	defer lock.release()

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	return writeFile(filepath.Join(s.dirs.Config, settingFilename), data)
}

func (s *JSONStore) SetHistoryCap(n int) {
	s.historyCap = n
}

func writeFile(path string, data []byte) error {
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	// Synced first, so a crash can't leave an empty file behind the rename.
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
