package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestRenderResultContainsStats(t *testing.T) {
	snap := resultSnapshot{
		wpm:      120.50,
		rawWpm:   135.00,
		accuracy: 94.50,
		errors:   7,
		elapsed:  32 * time.Second,
		cfg:      domain.TestConfig{Kind: domain.TestKindTimed, Duration: 60},
	}

	out := renderResult(snap, defaultTheme(), 0, 0)

	for _, want := range []string{"120.50", "135.00", "94.50%", "7", "0:32", "60s", "timed", "Test Complete"} {
		if !strings.Contains(out, want) {
			t.Errorf("renderResult() missing %q\nfull output:\n%s", want, out)
		}
	}
}

func TestRenderResultWordsSubtitle(t *testing.T) {
	snap := resultSnapshot{
		cfg: domain.TestConfig{Kind: domain.TestKindWords, WordCount: 25},
	}

	out := renderResult(snap, defaultTheme(), 0, 0)

	for _, want := range []string{"25 words", "words"} {
		if !strings.Contains(out, want) {
			t.Errorf("renderResult() missing %q", want)
		}
	}
}

func TestRenderResultCentered(t *testing.T) {
	snap := resultSnapshot{
		cfg: domain.TestConfig{Kind: domain.TestKindTimed, Duration: 60},
	}

	out := renderResult(snap, defaultTheme(), 120, 40)

	lines := strings.Split(out, "\n")
	if len(lines) != 40 {
		t.Errorf("centered output height = %d, want 40", len(lines))
	}
}
