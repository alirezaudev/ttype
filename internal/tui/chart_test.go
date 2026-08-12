package tui

import (
	"strings"
	"testing"
)

func TestRenderSparklineASCIIRamp(t *testing.T) {
	t.Parallel()

	out := renderSparkline([]float64{0, 50, 100})
	if out != " -#" {
		t.Fatalf("sparkline = %q, want %q", out, " -#")
	}
	if renderSparkline(nil) != "" {
		t.Fatalf("empty input should render nothing")
	}

	got := downsampleSeries([]float64{0, 10, 20, 30, 40, 50}, 3)
	if len(got) != 3 || got[0] != 0 || got[2] != 50 {
		t.Fatalf("downsampled = %v, want 3 points from 0 to 50", got)
	}
	if out := downsampleSeries([]float64{1, 2}, 10); len(out) != 2 {
		t.Fatalf("short input should pass through, got %v", out)
	}
}

func TestRenderResultChartBraille(t *testing.T) {
	t.Setenv("LC_ALL", "C.UTF-8")

	out := renderResultChart(
		[]float64{0, 30, 60, 60},
		[]float64{20, 40, 70, 80},
		[]int{0, 2, 0, 0},
		defaultTheme(), 40, 6,
	)

	want := "wpm over time · peak 60\n" +
		"80┤                     ⣀⣀⣤⣤⣤⣤⣶⣶⣶⣶⣶⣶⣾⣿⣿⣿\n" +
		"  │            ×⣀⣠⣤⣴⣶⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿\n" +
		"40┤⣀⣀⣠⣤⣤⣤⣶⣶⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿\n" +
		" 0┤⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿\n" +
		"  └0s─────────────────────────────────3s"
	if got := stripANSI(out); got != want {
		t.Fatalf("chart shape changed:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderResultChartFallsBackToSparkline(t *testing.T) {
	history := []float64{20, 40, 60}

	cases := []struct {
		name          string
		locale        string
		width, height int
	}{
		{"narrow", "C.UTF-8", 30, 10},
		{"short", "C.UTF-8", 60, 3},
		{"ascii locale", "C", 60, 10},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("LC_ALL", tc.locale)

			out := stripANSI(renderResultChart(history, nil, nil, defaultTheme(), tc.width, tc.height))
			lines := strings.Split(out, "\n")
			if len(lines) != 2 || lines[0] != "wpm over time · peak 60" {
				t.Fatalf("chart = %q, want header + sparkline", out)
			}
			if containsBraille(out) {
				t.Fatalf("expected ASCII sparkline, got braille: %q", out)
			}
		})
	}
}

func TestRenderResultChartNothingToPlot(t *testing.T) {
	t.Parallel()

	theme := defaultTheme()
	if out := renderResultChart(nil, nil, nil, theme, 60, 10); out != "" {
		t.Fatalf("empty history rendered %q", out)
	}
	if out := renderResultChart([]float64{50}, []float64{60}, []int{1}, theme, 60, 10); out != "" {
		t.Fatalf("single sample rendered %q", out)
	}
	if out := renderResultChart([]float64{50, 60}, nil, nil, theme, 60, 1); out != "" {
		t.Fatalf("height 1 rendered %q", out)
	}
}

func containsBraille(s string) bool {
	return strings.ContainsFunc(s, func(r rune) bool { return r >= 0x2800 && r <= 0x28FF })
}
