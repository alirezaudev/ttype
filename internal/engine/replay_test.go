package engine_test

import (
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
)

func TestEventsSkipIgnoredInput(t *testing.T) {
	t.Parallel()

	s, _ := newWordsSession(t, "the cat")
	s.InputRune(' ')
	s.Backspace()
	s.InputRune('\t')
	if got := s.Events(); len(got) != 0 {
		t.Fatalf("events = %+v, want none", got)
	}

	s.InputRune('t')
	s.Backspace()
	s.InputRune('t')
	s.DeleteWord()
	kinds := []domain.ReplayEventKind{
		domain.ReplayRune, domain.ReplayBackspace, domain.ReplayRune, domain.ReplayDeleteWord,
	}
	events := s.Events()
	if len(events) != len(kinds) {
		t.Fatalf("events = %+v, want %d", events, len(kinds))
	}
	for i, want := range kinds {
		if events[i].Kind != want {
			t.Fatalf("event %d kind = %d, want %d", i, events[i].Kind, want)
		}
	}
}

// A recorded run fed back at its own offsets must land on the same result —
// that is the whole promise of the sidecars.
func TestReplayReproducesTheResult(t *testing.T) {
	t.Parallel()

	const target = "the cat sat on the mat"
	s, clock := newTestSession(t, target, 10*time.Second)
	for _, r := range "the cst " {
		s.InputRune(r)
		clock.Advance(120 * time.Millisecond)
	}
	s.Backspace()
	s.DeleteWord()
	clock.Advance(300 * time.Millisecond)
	for _, r := range "cat sat" {
		s.InputRune(r)
		clock.Advance(90 * time.Millisecond)
	}
	clock.Advance(10 * time.Second)
	s.Tick()

	want, err := s.Result()
	if err != nil {
		t.Fatalf("Result: %v", err)
	}

	playClock := engine.NewFakeClock(time.Date(2026, 3, 3, 8, 0, 0, 0, time.UTC))
	play, err := engine.NewSession(s.Config(), engine.StaticText(target), playClock)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	start := playClock.Now()
	for _, ev := range s.Events() {
		playClock.Advance(start.Add(ev.Offset).Sub(playClock.Now()))
		switch ev.Kind {
		case domain.ReplayRune:
			play.InputRune(ev.Rune)
		case domain.ReplayBackspace:
			play.Backspace()
		case domain.ReplayDeleteWord:
			play.DeleteWord()
		}
	}
	playClock.Advance(want.Duration - playClock.Now().Sub(start))
	play.Tick()

	got, err := play.Result()
	if err != nil {
		t.Fatalf("Result: %v", err)
	}
	if string(play.Input()) != string(s.Input()) {
		t.Fatalf("input = %q, want %q", string(play.Input()), string(s.Input()))
	}
	if got.WPM != want.WPM || got.RawWPM != want.RawWPM || got.Accuracy != want.Accuracy {
		t.Fatalf("stats = %.2f/%.2f/%.2f, want %.2f/%.2f/%.2f",
			got.WPM, got.RawWPM, got.Accuracy, want.WPM, want.RawWPM, want.Accuracy)
	}
	if got.Consistency != want.Consistency {
		t.Fatalf("consistency = %.2f, want %.2f", got.Consistency, want.Consistency)
	}
}
