package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
)

// maxCustomText caps what is read, so piping a huge file in can't eat memory
// before the test even starts.
const maxCustomText = 1 << 20

// CustomInput is where the text for a custom test can come from.
type CustomInput struct {
	Stdin      io.Reader
	StdinPiped bool
	File       string // "-" reads stdin
	Text       string
	TextSet    bool
}

// ReadCustomText returns the cleaned text, or "" when none was given. Giving
// it more than one way is an error rather than a guess.
func ReadCustomText(in CustomInput) (string, error) {
	fromStdin := in.File == "-" || (in.StdinPiped && in.File == "")
	sources := 0
	for _, given := range []bool{fromStdin, in.File != "" && in.File != "-", in.TextSet} {
		if given {
			sources++
		}
	}
	if sources > 1 || (in.StdinPiped && in.File != "" && in.File != "-") {
		return "", errors.New("give the text one way: piped in, --file or --text")
	}

	var (
		raw  string
		from string
		err  error
	)
	switch {
	case in.TextSet:
		raw, from = in.Text, "--text"
	case fromStdin:
		raw, err = readLimited(in.Stdin)
		from = "stdin"
	case in.File != "":
		from = in.File
		var f *os.File
		if f, err = os.Open(in.File); err == nil {
			raw, err = readLimited(f)
			f.Close()
		}
	default:
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read %s: %w", from, err)
	}

	text := CleanCustomText(raw)
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("%s has no text to type", from)
	}
	return text, nil
}

func readLimited(r io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxCustomText+1))
	if err != nil {
		return "", err
	}
	if len(data) > maxCustomText {
		return "", fmt.Errorf("more than %d MiB of text", maxCustomText>>20)
	}
	return string(data), nil
}

var escapeSequence = regexp.MustCompile(`\x1b(\[[0-?]*[ -/]*[@-~]|\][^\x07\x1b]*(\x07|\x1b\\)|[@-Z\\-_])`)

// Characters keyboards don't have, mapped to the ones they do.
var typeable = strings.NewReplacer(
	"“", `"`, "”", `"`, "„", `"`, "«", `"`, "»", `"`,
	"‘", "'", "’", "'", "‚", "'",
	"–", "-", "—", "-", "−", "-",
	"…", "...",
	"\u00a0", " ", "\u200b", "", "\ufeff", "",
)

// CleanCustomText makes pasted or piped text typeable: colour codes and
// control characters go, typographic punctuation becomes ASCII, and
// whitespace collapses to single spaces. Line breaks stay, for line numbers.
func CleanCustomText(raw string) string {
	text := escapeSequence.ReplaceAllString(raw, "")
	text = typeable.Replace(text)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return ' '
			}
			if unicode.IsControl(r) {
				return -1
			}
			return r
		}, line)
		lines[i] = strings.Join(strings.Fields(line), " ")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

// customSource hands out the custom text, and leaves every other mode to the
// usual provider, so picking a mode mid-session still works.
type customSource struct {
	words    []string
	lines    []int
	fallback engine.TextSource
}

func newCustomSource(text string, fallback engine.TextSource) customSource {
	s := customSource{fallback: fallback}
	for i, line := range strings.Split(text, "\n") {
		for _, word := range strings.Fields(line) {
			s.words = append(s.words, word)
			s.lines = append(s.lines, i+1)
		}
	}
	return s
}

// WordLines gives the line each word came from.
func (s customSource) WordLines(opts domain.GenerateOptions) []int {
	if opts.Mode != domain.TextModeCustom {
		return nil
	}
	return s.lines
}

func (s customSource) Generate(opts domain.GenerateOptions) (string, error) {
	if opts.Mode != domain.TextModeCustom {
		return s.fallback.Generate(opts)
	}
	words := s.words
	if opts.WordLimit > 0 && opts.WordLimit < len(words) {
		words = words[:opts.WordLimit]
	}
	return strings.Join(words, " "), nil
}

// CustomConfig turns a config into one for the given text: the whole text
// unless --words or --time says otherwise, and none of the options that
// generate text.
func CustomConfig(cfg domain.TestConfig, text string, wordsSet, timeSet bool) domain.TestConfig {
	cfg.TextMode = domain.TextModeCustom
	cfg.Language = ""
	cfg.Punctuation = false
	cfg.Numbers = false

	total := len(strings.Fields(text))
	switch {
	case timeSet:
		cfg.Kind = domain.TestKindTimed
		cfg.WordCount = 0
	case wordsSet && cfg.WordCount > 0:
		cfg.Kind = domain.TestKindWords
		cfg.WordCount = min(cfg.WordCount, total)
	default:
		cfg.Kind = domain.TestKindWords
		cfg.WordCount = total
	}
	return cfg
}
