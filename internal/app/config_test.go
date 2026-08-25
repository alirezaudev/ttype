package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/storage"
)

func TestApplyConfigChangesOnlyTouchesWhatWasPassed(t *testing.T) {
	t.Parallel()

	settings := domain.Settings{
		DefaultDuration: domain.Duration60,
		DefaultMode:     domain.TextModeWords,
		Theme:           "default",
		Punctuation:     true,
	}

	mode := "sql"
	got, err := applyConfigChanges(settings, ConfigChanges{Mode: &mode})
	if err != nil {
		t.Fatalf("applyConfigChanges: %v", err)
	}

	if got.DefaultMode != domain.TextModeSQL {
		t.Fatalf("mode = %q, want sql", got.DefaultMode)
	}
	if !got.Punctuation || got.Theme != "default" || got.DefaultDuration != domain.Duration60 {
		t.Fatalf("unrelated defaults changed: %+v", got)
	}
}

func TestSettingATimeClearsTheWordCount(t *testing.T) {
	t.Parallel()

	seconds := 30
	got, err := applyConfigChanges(domain.Settings{DefaultWordCount: 25}, ConfigChanges{Duration: &seconds})
	if err != nil {
		t.Fatalf("applyConfigChanges: %v", err)
	}
	if got.DefaultWordCount != 0 || got.DefaultDuration != domain.Duration30 {
		t.Fatalf("settings = %+v, want a 30s timed default", got)
	}
}

func TestApplyConfigChangesRejectsBadValues(t *testing.T) {
	t.Parallel()

	mode := "brainfuck"
	if _, err := applyConfigChanges(domain.Settings{}, ConfigChanges{Mode: &mode}); err == nil {
		t.Fatal("an unknown mode should be rejected")
	}

	theme := "neon"
	if _, err := applyConfigChanges(domain.Settings{}, ConfigChanges{Theme: &theme}); err == nil {
		t.Fatal("an unknown theme should be rejected")
	}
}

func TestPrintSettingsShowsThePaths(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	printSettings(&buf, domain.DefaultSettings(), storage.Dirs{Config: "/tmp/cfg", Data: "/tmp/data"})

	for _, want := range []string{"60s", "english (built-in)", "auto", "off", "/tmp/cfg", "/tmp/data"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("config output missing %q:\n%s", want, buf.String())
		}
	}
}
