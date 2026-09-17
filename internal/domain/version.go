package domain

import "fmt"

// VersionInfo is what the footer needs to know about releases.
type VersionInfo struct {
	Local           string `json:"local"`
	Latest          string `json:"latest,omitempty"`
	UpdateAvailable bool   `json:"update_available,omitempty"`

	// CanInstall is set when ttype may replace itself with Latest.
	CanInstall bool `json:"-"`
	// AutoInstall is set when it should do that on its own, in the background.
	AutoInstall bool `json:"-"`
	// Command is how to update when ttype can't do it itself.
	Command string `json:"-"`
	// Installed is the version put in place during this session; it runs
	// from the next launch.
	Installed string `json:"-"`
	// JustUpdated is set on the first launch after an update in the background.
	JustUpdated bool `json:"-"`
}

// UpdateMode says what ttype does about a new release.
type UpdateMode string

const (
	UpdateAuto   UpdateMode = "auto"
	UpdateNotify UpdateMode = "notify"
	UpdateOff    UpdateMode = "off"
)

func ParseUpdateMode(s string) (UpdateMode, error) {
	switch mode := UpdateMode(s); mode {
	case UpdateAuto, UpdateNotify, UpdateOff:
		return mode, nil
	}
	return "", fmt.Errorf("unknown update mode %q (auto, notify or off)", s)
}
