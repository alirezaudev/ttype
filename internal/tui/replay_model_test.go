package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

func testReplay(t *testing.T) ReplayModel {
	t.Helper()

	result := domain.Result{
		Config: domain.TestConfig{Kind: domain.TestKindTimed, Duration: 15, Blind: true, Zen: true},
	}
	recording := domain.Replay{
		Target: "the cat sat",
		Events: []domain.ReplayEvent{
			{Offset: 100 * time.Millisecond, Kind: domain.ReplayRune, Rune: 't'},
			{Offset: 200 * time.Millisecond, Kind: domain.ReplayRune, Rune: 'h'},
			{Offset: 300 * time.Millisecond, Kind: domain.ReplayRune, Rune: 'x'},
			{Offset: 400 * time.Millisecond, Kind: domain.ReplayBackspace},
			{Offset: 500 * time.Millisecond, Kind: domain.ReplayRune, Rune: 'e'},
		},
	}

	m, err := NewReplayModel(result, recording, defaultTheme())
	if err != nil {
		t.Fatalf("NewReplayModel: %v", err)
	}
	m.setSize(80, 24)
	return m
}

func TestReplayFeedsEventsAtTheirOffsets(t *testing.T) {
	t.Parallel()

	m := testReplay(t)
	for i := 0; i < 12; i++ {
		next, _, done := m.Update(replayTickMsg(time.Now()))
		m = next
		if done {
			t.Fatal("playback ended early")
		}
	}

	if got := string(m.session.Input()); got != "the" {
		t.Fatalf("input = %q, want %q", got, "the")
	}
	if !m.done() {
		t.Fatal("playback should have consumed every event")
	}
}

func TestReplayForcesBlindAndZenOff(t *testing.T) {
	t.Parallel()

	m := testReplay(t)
	if m.cfg.Blind || m.cfg.Zen {
		t.Fatalf("replay config = blind:%v zen:%v, want both off", m.cfg.Blind, m.cfg.Zen)
	}
}

func TestReplayPauseSpeedAndBack(t *testing.T) {
	t.Parallel()

	m := testReplay(t)
	m, _, _ = m.Update(runeKey(' '))
	if !m.paused {
		t.Fatal("space did not pause playback")
	}

	before := m.elapsed
	m, _, _ = m.Update(replayTickMsg(time.Now()))
	if m.elapsed != before {
		t.Fatal("playback advanced while paused")
	}

	m, _, _ = m.Update(runeKey('4'))
	if m.speed != 4 {
		t.Fatalf("speed = %d, want 4", m.speed)
	}

	_, _, done := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !done {
		t.Fatal("esc should leave the replay")
	}
}

func TestReplayRestartRewinds(t *testing.T) {
	t.Parallel()

	m := testReplay(t)
	for i := 0; i < 12; i++ {
		m, _, _ = m.Update(replayTickMsg(time.Now()))
	}

	m, _, _ = m.Update(runeKey('r'))
	if m.next != 0 || m.elapsed != 0 {
		t.Fatalf("after restart next=%d elapsed=%v, want 0", m.next, m.elapsed)
	}
	if got := m.session.Input(); got != nil {
		t.Fatalf("input = %q, want empty", string(got))
	}
	if !strings.Contains(stripANSI(m.View()), "replay") {
		t.Fatal("replay view lost its header")
	}
}
