package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func loadedPicker(t *testing.T, current string, ids ...string) LanguagePicker {
	t.Helper()

	p := NewLanguagePicker(nil, current, defaultTheme())
	p, _, _, _ = p.Update(LanguagesLoadedMsg{IDs: ids})
	return p
}

func typeFilter(p LanguagePicker, filter string) LanguagePicker {
	for _, r := range filter {
		p, _, _, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return p
}

func TestLanguagePickerStartsOnTheCurrentLanguage(t *testing.T) {
	t.Parallel()

	p := loadedPicker(t, "french", "french", "spanish")

	if got := p.Selected(); got != "french" {
		t.Fatalf("selected = %q, want french", got)
	}
}

func TestLanguagePickerFilterNarrowsTheList(t *testing.T) {
	t.Parallel()

	p := loadedPicker(t, "", "french", "spanish", "swedish")
	p = typeFilter(p, "sp")

	if len(p.filtered) != 1 {
		t.Fatalf("filtered = %v, want just spanish", p.filtered)
	}
	if got := p.Selected(); got != "spanish" {
		t.Fatalf("selected = %q, want spanish", got)
	}
}

func TestLanguagePickerBackspaceWidensTheList(t *testing.T) {
	t.Parallel()

	p := loadedPicker(t, "", "french", "spanish")
	p = typeFilter(p, "fre")
	p, _, _, _ = p.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	if p.filter != "fr" {
		t.Fatalf("filter = %q, want fr", p.filter)
	}
	if len(p.filtered) != 1 || p.filtered[0] != "french" {
		t.Fatalf("filtered = %v, want just french", p.filtered)
	}
}

func TestLanguagePickerEnterApplies(t *testing.T) {
	t.Parallel()

	p := loadedPicker(t, "", "french", "spanish")
	p, _, _, _ = p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p, _, done, apply := p.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !done || !apply {
		t.Fatalf("enter: done=%v apply=%v, want true/true", done, apply)
	}
	if got := p.Selected(); got != "french" {
		t.Fatalf("selected = %q, want french", got)
	}
}

func TestLanguagePickerEscDoesNotApply(t *testing.T) {
	t.Parallel()

	p := loadedPicker(t, "", "french")
	_, _, done, apply := p.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !done || apply {
		t.Fatalf("esc: done=%v apply=%v, want true/false", done, apply)
	}
}

func TestLanguagePickerBuiltInComesFirst(t *testing.T) {
	t.Parallel()

	p := loadedPicker(t, "", "french")

	if p.filtered[0] != "" {
		t.Fatalf("first entry = %q, want the built-in list", p.filtered[0])
	}
}
