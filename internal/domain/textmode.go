package domain

import "fmt"

type TextMode string

const (
	TextModeWords     TextMode = "words"
	TextModeSentences TextMode = "sentences"
	TextModeSQL       TextMode = "sql"
	TextModeGo        TextMode = "go"
	TextModeBackend   TextMode = "backend"
	TextModePython    TextMode = "python"
	TextModeShell     TextMode = "shell"
	TextModeRegex     TextMode = "regex"
	// TextModeCustom is text the user brought. It is not in AllTextModes: it
	// can't be picked or saved as a default, since there would be no text.
	TextModeCustom TextMode = "custom"
)

func AllTextModes() []TextMode {
	return []TextMode{
		TextModeWords, TextModeSentences, TextModeSQL, TextModeGo,
		TextModeBackend, TextModePython, TextModeShell, TextModeRegex,
	}
}

func (m TextMode) CommitsWordsOnSpace() bool {
	return m == "" || m == TextModeWords || m == TextModeSentences || m == TextModeCustom
}

func ParseTextMode(s string) (TextMode, error) {
	for _, mode := range append(AllTextModes(), TextModeCustom) {
		if TextMode(s) == mode {
			return mode, nil
		}
	}
	return "", fmt.Errorf("unknown text mode %q", s)
}
