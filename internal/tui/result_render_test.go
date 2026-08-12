package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/storage"
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

	out := renderResult(result, storage.PBUpdate{}, defaultTheme(), 0, 0, statusNotice{})

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

	out := renderResult(result, storage.PBUpdate{}, defaultTheme(), 0, 0, statusNotice{})

	for _, want := range []string{"25 words", "words"} {
		if !strings.Contains(out, want) {
			t.Errorf("renderResult() missing %q", want)
		}
	}
}

func TestRenderResultChartFitsSmallTerminal(t *testing.T) {
	t.Setenv("LC_ALL", "C.UTF-8")

	result := domain.Result{
		Config:        domain.TestConfig{Kind: domain.TestKindTimed, Duration: 30},
		WPMHistory:    []float64{0, 20, 40, 55, 60, 58},
		RawWPMHistory: []float64{30, 50, 70, 65, 60},
		ErrorHistory:  []int{0, 1, 0, 2, 0},
	}
	pb := storage.PBUpdate{IsNew: true, Label: "words/30s", PrevWPM: 51}

	out := renderResult(result, pb, defaultTheme(), 80, 24, errorNotice("result not saved"))

	if lines := strings.Split(out, "\n"); len(lines) != 24 {
		t.Fatalf("height = %d, want 24:\n%s", len(lines), out)
	}
	if !strings.Contains(stripANSI(out), "wpm over time") {
		t.Fatalf("chart missing:\n%s", out)
	}
}

func TestRenderResultCentered(t *testing.T) {
	result := domain.Result{
		Config: domain.TestConfig{Kind: domain.TestKindTimed, Duration: 60},
	}

	out := renderResult(result, storage.PBUpdate{}, defaultTheme(), 120, 40, statusNotice{})

	lines := strings.Split(out, "\n")
	if len(lines) != 40 {
		t.Errorf("centered output height = %d, want 40", len(lines))
	}
}
