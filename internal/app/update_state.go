package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type updateState struct {
	LastCheck time.Time `json:"last_check"`
	Latest    string    `json:"latest,omitempty"`
}

func updateStatePath(dataDir string) string {
	return filepath.Join(dataDir, "update.json")
}

// A missing or broken state file only means checking again.
func loadUpdateState(path string) updateState {
	var state updateState
	data, err := os.ReadFile(path)
	if err != nil || json.Unmarshal(data, &state) != nil {
		return updateState{}
	}
	return state
}

func saveUpdateState(path string, state updateState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
