package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// readJSON loads path into v; a file that doesn't parse is moved aside.
func (s *JSONStore) readJSON(path string, v any) (bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("read %s: %w", path, err)
	}
	parseErr := json.Unmarshal(data, v)
	if parseErr == nil {
		return true, nil
	}

	broken := fmt.Sprintf("%s.broken-%s", path, time.Now().Format("20060102-150405"))
	if err := os.Rename(path, broken); err != nil {
		// Another ttype moved it first.
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("parse %s: %w", path, parseErr)
	}

	s.mu.Lock()
	s.recovered = append(s.recovered, fmt.Sprintf("%s couldn't be read, so ttype started a new one; the old copy is %s", filepath.Base(path), broken))
	s.mu.Unlock()
	return false, nil
}

// Recovered lists the files moved aside since the store was opened.
func (s *JSONStore) Recovered() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.recovered...)
}

// CheckFiles parses every store file without touching it.
func CheckFiles(dirs Dirs) []error {
	var errs []error
	for _, path := range []string{
		filepath.Join(dirs.Config, settingFilename),
		filepath.Join(dirs.Data, historyFilename),
		filepath.Join(dirs.Data, bestsFilename),
	} {
		data, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err == nil {
			var v any
			err = json.Unmarshal(data, &v)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", path, err))
		}
	}
	return errs
}
