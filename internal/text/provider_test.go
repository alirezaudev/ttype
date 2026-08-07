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
