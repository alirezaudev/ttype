package app

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestCleanCustomText(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"hello\r\nworld\n":                       "hello\nworld",
		"a\tb   c\n\n\nd":                        "a b c\n\n\nd",
		"\x1b[31mred\x1b[0m and \x1b[1;32mgreen": "red and green",
		"“quoted” ‘single’ – dash — em…":         `"quoted" 'single' - dash - em...`,
		"  \u00a0padded\u200b\ufeff  ":           "padded",
		"bell\x07 and nul\x00":                   "bell and nul",
		"سلام\u200cدنیا":                         "سلام\u200cدنیا",
		"I think":                                "I think",
	}
	for raw, want := range tests {
		if got := CleanCustomText(raw); got != want {
			t.Errorf("CleanCustomText(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestReadCustomText(t *testing.T) {
	t.Parallel()

	file := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(file, []byte("from a\nfile\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	stdin := func(s string) *strings.Reader { return strings.NewReader(s) }

	tests := []struct {
		name    string
		in      CustomInput
		want    string
		wantErr string
	}{
		{"nothing given", CustomInput{Stdin: stdin("ignored")}, "", ""},
		{"piped", CustomInput{Stdin: stdin("piped\ttext\n"), StdinPiped: true}, "piped text", ""},
		{"file", CustomInput{File: file}, "from a\nfile", ""},
		{"file dash", CustomInput{Stdin: stdin("dash"), StdinPiped: true, File: "-"}, "dash", ""},
		{"text", CustomInput{Text: "  given  text ", TextSet: true}, "given text", ""},
		{"piped and text", CustomInput{Stdin: stdin("x"), StdinPiped: true, Text: "y", TextSet: true}, "", "one way"},
		{"piped and file", CustomInput{Stdin: stdin("x"), StdinPiped: true, File: file}, "", "one way"},
		{"file and text", CustomInput{File: file, Text: "y", TextSet: true}, "", "one way"},
		{"empty pipe", CustomInput{Stdin: stdin(" \n\n"), StdinPiped: true}, "", "stdin has no text"},
		{"empty text", CustomInput{Text: "", TextSet: true}, "", "--text has no text"},
		{"missing file", CustomInput{File: filepath.Join(t.TempDir(), "nope")}, "", "read "},
		{"too much", CustomInput{Stdin: stdin(strings.Repeat("a ", maxCustomText)), StdinPiped: true}, "", "MiB"},
	}

	for _, test := range tests {
		got, err := ReadCustomText(test.in)
		if test.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Errorf("%s: err = %v, want one mentioning %q", test.name, err, test.wantErr)
			}
			continue
		}
		if err != nil || got != test.want {
			t.Errorf("%s: ReadCustomText = %q, %v; want %q", test.name, got, err, test.want)
		}
	}
}

type modeSource struct{}

func (modeSource) Generate(opts domain.GenerateOptions) (string, error) {
	return "generated " + string(opts.Mode), nil
}

func TestCustomSourceHandsOutTheText(t *testing.T) {
	t.Parallel()

	source := newCustomSource("one two three four", modeSource{})

	tests := []struct {
		opts domain.GenerateOptions
		want string
	}{
		{domain.GenerateOptions{Mode: domain.TextModeCustom}, "one two three four"},
		{domain.GenerateOptions{Mode: domain.TextModeCustom, WordLimit: 2}, "one two"},
		{domain.GenerateOptions{Mode: domain.TextModeCustom, WordLimit: 9}, "one two three four"},
		// Picking another mode mid-session gets generated text again.
		{domain.GenerateOptions{Mode: domain.TextModeGo}, "generated go"},
	}
	for _, test := range tests {
		if got, _ := source.Generate(test.opts); got != test.want {
			t.Errorf("Generate(%+v) = %q, want %q", test.opts, got, test.want)
		}
	}
}

func TestCustomConfig(t *testing.T) {
	t.Parallel()

	base := domain.TestConfig{
		Kind: domain.TestKindTimed, Duration: domain.Duration60, TextMode: domain.TextModeGo,
		Language: "spanish", Punctuation: true, Numbers: true, Blind: true,
	}
	text := "one two three four five"

	whole := CustomConfig(base, text, false, false)
	if whole.TextMode != domain.TextModeCustom || whole.Kind != domain.TestKindWords || whole.WordCount != 5 {
		t.Fatalf("default = %+v, want all 5 words as a word test", whole)
	}
	if whole.Language != "" || whole.Punctuation || whole.Numbers || !whole.Blind {
		t.Fatalf("default = %+v, want the text options cleared and the rest kept", whole)
	}

	words := base
	words.Kind, words.WordCount = domain.TestKindWords, 3
	if got := CustomConfig(words, text, true, false); got.WordCount != 3 {
		t.Fatalf("--words 3 = %+v, want 3 words", got)
	}
	words.WordCount = 50
	if got := CustomConfig(words, text, true, false); got.WordCount != 5 {
		t.Fatalf("--words 50 = %+v, want the 5 words there are", got)
	}

	if got := CustomConfig(base, text, false, true); got.Kind != domain.TestKindTimed || got.Duration != domain.Duration60 {
		t.Fatalf("--time 60 = %+v, want a timed test", got)
	}
}

func TestCustomCannotBeAModeOrADefault(t *testing.T) {
	t.Parallel()

	if _, err := ConfigFromFlags(TestFlags{TimeSec: 30, Mode: "custom"}); err == nil {
		t.Fatal("--mode custom without text should fail")
	}
	mode := "custom"
	if _, err := applyConfigChanges(domain.Settings{}, ConfigChanges{Mode: &mode}); err == nil {
		t.Fatal("custom should not be accepted as the default mode")
	}
}

func TestCustomSourceKnowsEachWordsLine(t *testing.T) {
	t.Parallel()

	source := newCustomSource(CleanCustomText("I think\n\nit  is\n"), modeSource{})
	opts := domain.GenerateOptions{Mode: domain.TextModeCustom}

	text, err := source.Generate(opts)
	if err != nil || text != "I think it is" {
		t.Fatalf("Generate = %q, %v", text, err)
	}
	if got := source.WordLines(opts); !slices.Equal(got, []int{1, 1, 3, 3}) {
		t.Fatalf("lines = %v, want [1 1 3 3]", got)
	}
	if got := source.WordLines(domain.GenerateOptions{Mode: domain.TextModeWords}); got != nil {
		t.Fatalf("lines for generated text = %v, want none", got)
	}
}
