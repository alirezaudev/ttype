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
	dirs, err := DefaultDirs()
	if err != nil {
		return nil, err
	}

	return NewJSONStore(dirs.Config)
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
