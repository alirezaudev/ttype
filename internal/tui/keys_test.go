package tui

import (
	"strings"
	"testing"
	"unicode"

	"github.com/charmbracelet/bubbles/key"
)

// Space is the input path itself; everything else on the test screen has to
// carry a modifier, or it fires in the middle of a word.
func TestTestScreenBindsNoPrintableRune(t *testing.T) {
	t.Parallel()

	bindings := map[string]key.Binding{
		"restart":     testKeys.Restart,
		"delete word": testKeys.DeleteWord,
		"backspace":   testKeys.Backspace,
	}

	for name, binding := range bindings {
		for _, k := range binding.Keys() {
			if len([]rune(k)) != 1 {
				continue
			}
			r := []rune(k)[0]
			if unicode.IsPrint(r) && r != ' ' {
				t.Errorf("%s binds the printable rune %q", name, k)
			}
		}
	}
}

func TestEveryBindingHasKeysAndHelp(t *testing.T) {
	t.Parallel()

	bindings := []key.Binding{
		appKeys.Quit, appKeys.Back, appKeys.Exit,
		testKeys.Help, testKeys.Restart, testKeys.ToggleLive, testKeys.DeleteWord, testKeys.Backspace, testKeys.Skip,
		resultsKeys.Restart, resultsKeys.Copy, resultsKeys.Settings,
		resultsKeys.Mode, resultsKeys.Language, resultsKeys.Update,
		pickerKeys.Up, pickerKeys.Left, pickerKeys.Confirm, pickerKeys.Cancel,
	}

	for _, b := range bindings {
		if len(b.Keys()) == 0 {
			t.Errorf("binding %+v has no keys", b.Help())
		}
		if b.Help().Key == "" || b.Help().Desc == "" {
			t.Errorf("binding %v has no help text", b.Keys())
		}
	}
}

func TestHelpLinesComeFromTheBindings(t *testing.T) {
	t.Parallel()

	if got, want := testHelpLine(), "? help - tab/enter restart - esc/ctrl+c quit - ctrl+bksp/ctrl+w delete word - ctrl+o live stats"; got != want {
		t.Fatalf("testHelpLine() = %q, want %q", got, want)
	}
	if got := resultsHelpLine(); !strings.Contains(got, "C copy result") {
		t.Fatalf("resultsHelpLine() = %q, want the copy binding", got)
	}
}
