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

func TestGenerateRejectsAnOutOfRangeLimit(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t)

	if _, err := p.Generate(domain.GenerateOptions{WordLimit: -1}); err == nil {
		t.Fatal("a negative limit should fail")
	}
	if _, err := p.Generate(domain.GenerateOptions{WordLimit: 1 << 20}); err == nil {
		t.Fatal("a limit larger than the word list should fail")
	}
}
