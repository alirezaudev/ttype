package domain

// VersionInfo is what the footer needs to know about releases.
type VersionInfo struct {
	Local           string `json:"local"`
	Latest          string `json:"latest,omitempty"`
	UpdateAvailable bool   `json:"update_available,omitempty"`
}
