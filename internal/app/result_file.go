package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

const resultFileVersion = 1

// ResultSource is the custom text and the file it came from.
type ResultSource struct {
	File string
	Text string
}

type resultFile struct {
	Version     int                 `json:"version"`
	Status      string              `json:"status"`
	StartedAt   string              `json:"started_at,omitempty"`
	DurationS   float64             `json:"duration_s"`
	WPM         float64             `json:"wpm"`
	Raw         float64             `json:"raw"`
	Accuracy    float64             `json:"accuracy"`
	Consistency float64             `json:"consistency"`
	Chars       resultChars         `json:"chars"`
	CharErrors  map[string]int      `json:"char_errors,omitempty"`
	Mode        domain.TextMode     `json:"mode,omitempty"`
	Language    string              `json:"language,omitempty"`
	Tag         string              `json:"tag,omitempty"`
	Source      *resultSource       `json:"source,omitempty"`
	Words       []domain.WordResult `json:"words,omitempty"`
}

type resultChars struct {
	Correct   int `json:"correct"`
	Incorrect int `json:"incorrect"`
	Extra     int `json:"extra"`
	Skipped   int `json:"skipped"`
}

type resultSource struct {
	File   string `json:"file,omitempty"`
	SHA256 string `json:"sha256"`
}

// ran is false when nothing was typed.
func newResultFile(r domain.Result, finished, ran bool, src ResultSource) resultFile {
	out := resultFile{Version: resultFileVersion, Status: "quit"}
	if !ran {
		return out
	}
	switch {
	case r.Failed:
		out.Status = "failed_min_wpm"
	case finished:
		out.Status = "completed"
	}
	out.StartedAt = r.Timestamp.Add(-r.Duration).Format(time.RFC3339)
	out.DurationS = math.Round(r.Duration.Seconds()*100) / 100
	out.WPM = r.WPM
	out.Raw = r.RawWPM
	out.Accuracy = r.Accuracy
	out.Consistency = r.Consistency
	out.Chars = resultChars{
		Correct:   r.Correct,
		Incorrect: r.Incorrect,
		Extra:     max(r.TotalChars-r.Correct-r.Incorrect, 0),
		Skipped:   r.Skipped,
	}
	out.CharErrors = r.CharErrors
	out.Mode = r.Config.TextMode
	out.Language = r.Config.Language
	out.Tag = r.Config.Tag
	out.Words = r.Words
	if src.Text != "" {
		// A hash, not the text, so private text stays private.
		sum := sha256.Sum256([]byte(strings.Join(strings.Fields(src.Text), " ")))
		out.Source = &resultSource{File: src.File, SHA256: hex.EncodeToString(sum[:])}
	}
	return out
}

func writeResultFile(path string, rf resultFile) error {
	data, err := json.MarshalIndent(rf, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write result file: %w", err)
	}
	return nil
}
