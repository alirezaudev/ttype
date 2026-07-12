package assets

import (
	"embed"
	"strings"
)

//go:embed words/en.txt
var wordsFile embed.FS

func LoadWords() ([]string, error) {
	data, err := wordsFile.ReadFile("words/en.txt")
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

	return lines, nil
}
