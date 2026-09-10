package tui

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// Every theme has to fill every slot, or some element silently renders
// unstyled in one theme and nobody notices until a screenshot.
func TestEveryThemeFillsEverySlot(t *testing.T) {
	t.Parallel()

	for _, name := range ThemeNames() {
		theme := reflect.ValueOf(ResolveTheme(name))
		for i := 0; i < theme.NumField(); i++ {
			style, ok := theme.Field(i).Interface().(lipgloss.Style)
			if !ok {
				continue
			}
			if !styled(style) {
				t.Errorf("theme %q leaves %s unstyled", name, theme.Type().Field(i).Name)
			}
		}
	}
}

func styled(style lipgloss.Style) bool {
	noColor := lipgloss.NoColor{}
	return style.GetForeground() != lipgloss.TerminalColor(noColor) ||
		style.GetBackground() != lipgloss.TerminalColor(noColor) ||
		style.GetBold() || style.GetUnderline() || style.GetItalic()
}

// Colours belong to the theme file alone; a hard-coded one somewhere else
// ignores whichever theme the user picked.
func TestColorsLiveOnlyInTheThemeFile(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || name == "theme.go" {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		source, err := os.ReadFile(filepath.Join(".", name))
		if err != nil {
			t.Fatalf("ReadFile %s: %v", name, err)
		}
		if strings.Contains(string(source), "lipgloss.Color(") {
			t.Errorf("%s hard-codes a colour; add a slot to the theme instead", name)
		}
	}
}

func TestParseThemeName(t *testing.T) {
	t.Parallel()

	if got, err := ParseThemeName(""); err != nil || got != ThemeDefault {
		t.Fatalf("ParseThemeName(\"\") = %q/%v, want the default theme", got, err)
	}
	if got, err := ParseThemeName("  DRACULA "); err != nil || got != ThemeDracula {
		t.Fatalf("ParseThemeName = %q/%v, want dracula", got, err)
	}
	if _, err := ParseThemeName("neon"); err == nil {
		t.Fatal("an unknown theme should be rejected")
	}
}
