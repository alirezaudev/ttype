package tui

import (
	"strings"
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestCapsWarnReplacesTheStartHint(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60, Width: 50}
	s, _ := newHUDSession(t, cfg)

	m := NewTestModel(s, cfg, defaultTheme(), domain.VersionInfo{})
	capsOn := false
	m.capsProbe = func() bool { return capsOn }
	m.setSize(100, 40)

	view := stripANSI(m.View())
	if !strings.Contains(view, "start typing to begin") {
		t.Fatalf("view = %q, want the start hint", view)
	}

	capsOn = true
	warned := stripANSI(m.View())
	if !strings.Contains(warned, "Caps Lock?") {
		t.Fatalf("view = %q, want the caps warning", warned)
	}
	if strings.Contains(warned, "start typing to begin") {
		t.Fatal("the warning should take the hint line, not share it")
	}
}
