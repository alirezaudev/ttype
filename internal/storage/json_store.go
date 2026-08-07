package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
)

const settingFilename = "config.json"

type JSONStore struct {
	dirs       Dirs
	historyCap int
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
	path := filepath.Join(s.dirs.Config, settingFilename)
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
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
