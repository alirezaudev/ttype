package assets

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed words/en.txt
var wordsFile embed.FS

//go:embed sentences/en.txt
var sentencesFile embed.FS

func LoadWords() ([]string, error) {
	return loadLines(wordsFile, "words/en.txt")
}

func LoadSentences() ([]string, error) {
	return loadLines(sentencesFile, "sentences/en.txt")
}

func loadLines(fs embed.FS, path string) ([]string, error) {
	data, err := fs.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("no lines in %s", path)
	}
	return lines, nil
}
