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

//go:embed sql/snippets.txt
var sqlFile embed.FS

//go:embed go/snippets.txt
var goFile embed.FS

//go:embed backend/terms.txt
var backendFile embed.FS

func LoadWords() ([]string, error) {
	return loadLines(wordsFile, "words/en.txt")
}

func LoadSentences() ([]string, error) {
	return loadLines(sentencesFile, "sentences/en.txt")
}

func LoadSQL() ([]string, error) {
	return loadLines(sqlFile, "sql/snippets.txt")
}

func LoadGo() ([]string, error) {
	return loadLines(goFile, "go/snippets.txt")
}

func LoadBackend() ([]string, error) {
	return loadLines(backendFile, "backend/terms.txt")
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
