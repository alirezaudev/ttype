package domain

import "fmt"

type TextMode string

const (
	TextModeWords     TextMode = "words"
	TextModeSentences TextMode = "sentences"
)

func AllTextModes() []TextMode {
	return []TextMode{TextModeWords, TextModeSentences}
}

func ParseTextMode(s string) (TextMode, error) {
	for _, mode := range AllTextModes() {
		if TextMode(s) == mode {
			return mode, nil
		}
	}
	return "", fmt.Errorf("unknown text mode %q", s)
}
