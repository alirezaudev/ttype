package domain

import "fmt"

type TextMode string

const (
	TextModeWords     TextMode = "words"
	TextModeSentences TextMode = "sentences"
	TextModeSQL       TextMode = "sql"
	TextModeGo        TextMode = "go"
)

func AllTextModes() []TextMode {
	return []TextMode{TextModeWords, TextModeSentences, TextModeSQL, TextModeGo}
}

func (m TextMode) CommitsWordsOnSpace() bool {
	return m == "" || m == TextModeWords || m == TextModeSentences
}

func ParseTextMode(s string) (TextMode, error) {
	for _, mode := range AllTextModes() {
		if TextMode(s) == mode {
			return mode, nil
		}
	}
	return "", fmt.Errorf("unknown text mode %q", s)
}
