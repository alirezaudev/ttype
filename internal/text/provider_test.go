package text

import (
	"strings"
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
)

func newTestProvider(t *testing.T) *Provider {
	t.Helper()

	p, err := NewProvider(t.TempDir())
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	return p
}

func TestGenerateWordLimitReturnsThatManyWords(t *testing.T) {
	t.Parallel()

	target, err := newTestProvider(t).Generate(domain.GenerateOptions{WordLimit: 25})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if got := len(strings.Fields(target)); got != 25 {
		t.Fatalf("words = %d, want 25", got)
	}
}

func TestGenerateWithoutLimitFillsTheBuffer(t *testing.T) {
	t.Parallel()

	target, err := newTestProvider(t).Generate(domain.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if len(target) < minTargetRunes {
		t.Fatalf("target = %d runes, want at least %d (a timed test must not run out of words)", len(target), minTargetRunes)
	}
	if strings.Contains(target, "  ") {
		t.Fatal("target should be single-spaced")
	}
}

func TestGeneratePlainTargetHasNoPunctuationOrNumbers(t *testing.T) {
	t.Parallel()

	target, err := newTestProvider(t).Generate(domain.GenerateOptions{WordLimit: 200})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if strings.ContainsAny(target, ",.0123456789") {
		t.Fatalf("plain target carries injected content: %q", target)
	}
}

func TestGenerateInjectsPunctuationAndNumbers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		wordLimit int
	}{
		{name: "words mode", wordLimit: 200},
		{name: "timed mode", wordLimit: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := newTestProvider(t)

			punctuated, err := p.Generate(domain.GenerateOptions{WordLimit: tc.wordLimit, Punctuation: true})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			if !strings.Contains(punctuated, ", ") && !strings.Contains(punctuated, ". ") {
				t.Fatalf("--punctuation produced no marks: %q", punctuated)
			}

			numbered, err := p.Generate(domain.GenerateOptions{WordLimit: tc.wordLimit, Numbers: true})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			if !strings.ContainsAny(numbered, "0123456789") {
				t.Fatalf("--numbers produced no digits: %q", numbered)
			}
		})
	}
}

func TestGeneratePunctuationKeepsTheWordCount(t *testing.T) {
	t.Parallel()

	target, err := newTestProvider(t).Generate(domain.GenerateOptions{WordLimit: 200, Punctuation: true, Numbers: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if got := len(strings.Fields(target)); got != 200 {
		t.Fatalf("words = %d, want 200", got)
	}
}

func TestGenerateIsDeterministicForASeed(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t)
	opts := domain.GenerateOptions{
		WordLimit:   120,
		Punctuation: true,
		Numbers:     true,
		Seed:        42,
	}

	first, err := p.Generate(opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	second, err := p.Generate(opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if first != second {
		t.Fatalf("same seed produced different targets:\n%q\n%q", first, second)
	}

	opts.Seed = 43
	other, err := p.Generate(opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if other == first {
		t.Fatal("a different seed should produce a different target")
	}
}

func TestGenerateRejectsAnOutOfRangeLimit(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t)

	if _, err := p.Generate(domain.GenerateOptions{WordLimit: -1}); err == nil {
		t.Fatal("a negative limit should fail")
	}
}

func TestGenerateLimitMayExceedTheItemCount(t *testing.T) {
	t.Parallel()

	target, err := newTestProvider(t).Generate(domain.GenerateOptions{
		Mode:      domain.TextModeSentences,
		WordLimit: 500,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if target == "" {
		t.Fatal("target should repeat items rather than come back empty")
	}
}

func TestGenerateSentencesMode(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t)

	target, err := p.Generate(domain.GenerateOptions{Mode: domain.TextModeSentences, WordLimit: 3})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(target, " ") || len(target) < 20 {
		t.Fatalf("sentences target looks wrong: %q", target)
	}

	injected, err := p.Generate(domain.GenerateOptions{
		Mode:        domain.TextModeSentences,
		WordLimit:   200,
		Punctuation: true,
		Numbers:     true,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if strings.ContainsAny(injected, "0123456789") {
		t.Fatal("--numbers must not inject into sentences mode")
	}
}

func TestGenerateRejectsAnUnknownMode(t *testing.T) {
	t.Parallel()

	if _, err := newTestProvider(t).Generate(domain.GenerateOptions{Mode: "cobol"}); err == nil {
		t.Fatal("a mode this build cannot generate should fail")
	}
}

func TestGenerateEveryRegisteredMode(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t)

	for _, mode := range domain.AllTextModes() {
		t.Run(string(mode), func(t *testing.T) {
			t.Parallel()

			target, err := p.Generate(domain.GenerateOptions{Mode: mode, WordLimit: 5})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			if strings.TrimSpace(target) == "" {
				t.Fatal("mode produced no text")
			}
			// Enter restarts and the engine drops control runes, so a newline can never be typed.
			if strings.Contains(target, "\n") {
				t.Fatalf("target has a newline nobody can type: %q", target)
			}
		})
	}
}

func TestGenerateRegexItemsCarryTheirComment(t *testing.T) {
	t.Parallel()

	target, err := newTestProvider(t).Generate(domain.GenerateOptions{
		Mode:      domain.TextModeRegex,
		WordLimit: 20,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if !strings.Contains(target, "# ") {
		t.Fatalf("regex items lost their paired comment: %q", target)
	}
}
