package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestRenderResultContainsStats(t *testing.T) {
	result := domain.Result{
		WPM:       120.50,
		RawWPM:    135.00,
		Accuracy:  94.50,
		Incorrect: 7,
		Duration:  32 * time.Second,
		Config:    domain.TestConfig{Kind: domain.TestKindTimed, Duration: 60},
	}

	out := renderResult(result, defaultTheme(), 0, 0, statusNotice{})

	for _, want := range []string{"120.50", "135.00", "94.50%", "7", "0:32", "60s", "timed", "Test Complete"} {
		if !strings.Contains(out, want) {
			t.Errorf("renderResult() missing %q\nfull output:\n%s", want, out)
		}
	}
}

func TestRenderResultWordsSubtitle(t *testing.T) {
	result := domain.Result{
		Config: domain.TestConfig{Kind: domain.TestKindWords, WordCount: 25},
	}

	out := renderResult(result, defaultTheme(), 0, 0, statusNotice{})

	for _, want := range []string{"25 words", "words"} {
		if !strings.Contains(out, want) {
			t.Errorf("renderResult() missing %q", want)
		}
	}
}

func TestRenderResultCentered(t *testing.T) {
	result := domain.Result{
		Config: domain.TestConfig{Kind: domain.TestKindTimed, Duration: 60},
	}

	out := renderResult(result, defaultTheme(), 120, 40, statusNotice{})

	lines := strings.Split(out, "\n")
	if len(lines) != 40 {
		t.Errorf("centered output height = %d, want 40", len(lines))
	}
}
