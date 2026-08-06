package tui

import (
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

func timedCfg() domain.TestConfig {
	return domain.TestConfig{Kind: domain.TestKindTimed, Duration: 60}
}

func TestSettingsPanelApplyReturnsConfig(t *testing.T) {
	t.Parallel()

	p := NewSettingsPanel(timedCfg(), defaultTheme())
	_, _, done, apply := p.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !done || !apply {
		t.Fatalf("enter: done=%v apply=%v, want true/true", done, apply)
	}
}

func TestSettingsPanelCancelDoesNotApply(t *testing.T) {
	t.Parallel()

	p := NewSettingsPanel(timedCfg(), defaultTheme())
	_, _, done, apply := p.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !done || apply {
		t.Fatalf("esc: done=%v apply=%v, want true/false", done, apply)
	}
}

func TestSettingsPanelKindToggle(t *testing.T) {
	t.Parallel()

	p := NewSettingsPanel(timedCfg(), defaultTheme())
	p, _, _, _ = p.Update(tea.KeyMsg{Type: tea.KeyRight})

	if p.cfg.Kind != domain.TestKindWords {
		t.Fatalf("kind = %v after right, want words", p.cfg.Kind)
	}
	if p.cfg.WordCount <= 0 {
		t.Fatal("word count should be defaulted when switching to words mode")
	}

	p, _, _, _ = p.Update(tea.KeyMsg{Type: tea.KeyRight})
	if p.cfg.Kind != domain.TestKindTimed {
		t.Fatalf("kind = %v after second right, want timed", p.cfg.Kind)
	}
}

func TestSettingsPanelDurationStep(t *testing.T) {
	t.Parallel()

	p := NewSettingsPanel(timedCfg(), defaultTheme())
	p, _, _, _ = p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p, _, _, _ = p.Update(tea.KeyMsg{Type: tea.KeyRight})

	if p.cfg.Duration != domain.Duration(75) {
		t.Fatalf("duration = %v, want 75s", p.cfg.Duration)
	}

	p, _, _, _ = p.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if p.cfg.Duration != domain.Duration(60) {
		t.Fatalf("duration = %v after left, want 60s", p.cfg.Duration)
	}
}

func TestSettingsPanelDurationFloor(t *testing.T) {
	t.Parallel()

	p := NewSettingsPanel(domain.TestConfig{Kind: domain.TestKindTimed, Duration: 15}, defaultTheme())
	p, _, _, _ = p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p, _, _, _ = p.Update(tea.KeyMsg{Type: tea.KeyLeft})

	if p.cfg.Duration < domain.Duration(15) {
		t.Fatalf("duration %v should not go below 15s", p.cfg.Duration)
	}
}

func TestSettingsPanelWidthFromAuto(t *testing.T) {
	t.Parallel()

	p := NewSettingsPanel(timedCfg(), defaultTheme())
	p.field = settingsWidth
	p, _, _, _ = p.Update(tea.KeyMsg{Type: tea.KeyRight})

	if p.cfg.Width != 60 {
		t.Fatalf("width = %d, want 60", p.cfg.Width)
	}
}

func TestSettingsPanelWidthLeftToAutoAtZero(t *testing.T) {
	t.Parallel()

	p := NewSettingsPanel(domain.TestConfig{Kind: domain.TestKindTimed, Duration: 60, Width: 60}, defaultTheme())
	p.field = settingsWidth
	for i := 0; i < 7; i++ {
		p, _, _, _ = p.Update(tea.KeyMsg{Type: tea.KeyLeft})
	}

	if p.cfg.Width != 0 {
		t.Fatalf("width = %d, want 0 (auto)", p.cfg.Width)
	}
}
