package engine_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
)

type fakeSource struct {
	gen func() (string, error)
}

func (f fakeSource) Generate(domain.GenerateOptions) (string, error) {
	return f.gen()
}

func fixedSource(target string) fakeSource {
	return fakeSource{gen: func() (string, error) { return target, nil }}
}

func newTestSession(t *testing.T, target string, duration time.Duration) (*engine.Session, *engine.FakeClock) {
	t.Helper()

	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration(duration / time.Second)}
	s, err := engine.NewSession(cfg, fixedSource(target), clock)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	return s, clock
}

func newWordsSession(t *testing.T, target string) (*engine.Session, *engine.FakeClock) {
	t.Helper()

	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	cfg := domain.TestConfig{Kind: domain.TestKindWords, WordCount: countWordsInText([]rune(target))}
	s, err := engine.NewSession(cfg, fixedSource(target), clock)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	return s, clock
}

func countWordsInText(text []rune) int {
	if len(text) == 0 {
		return 0
	}

	n := 1
	for _, r := range text {
		if r == ' ' {
			n++
		}
	}
	return n
}

func typeString(s *engine.Session, text string) {
	for _, r := range text {
		s.InputRune(r)
	}
}

func TestSessionStartsOnFirstKeypress(t *testing.T) {
	t.Parallel()

	wantedDuration := 15 * time.Second
	s, clock := newTestSession(t, "abc", wantedDuration)

	if s.Remaining() != wantedDuration {
		t.Fatalf("remaining = %v, want %v", s.Remaining(), wantedDuration)
	}

	if s.Cursor() != 0 {
		t.Fatalf("cursor = %d, want 0", s.Cursor())
	}

	if s.Input() != nil {
		t.Fatalf("input = %v, want nil", s.Input())
	}

	typeString(s, "a")
	if s.Cursor() != 1 {
		t.Fatalf("cursor = %d, want 1", s.Cursor())
	}

	clock.Advance(2 * time.Second)
	if s.Remaining() != 13*time.Second {
		t.Fatalf("elapsed = %v, want 13s", s.Remaining())
	}

}

func TestSessionTickBeforeStartDoesNotFinish(t *testing.T) {
	t.Parallel()

	s, clock := newTestSession(t, "abc", 15*time.Second)

	clock.Advance(16 * time.Second)

	if s.Tick() {
		t.Fatal("Tick should not finish a session that has not started")
	}

	if s.Finished() {
		t.Fatal("session should not be finished before the first keystroke")
	}

	if got := s.Remaining(); got != 15*time.Second {
		t.Fatalf("Remaining() = %v, want 15s", got)
	}
}

func TestDeleteWordMidWordDeletesToWordStart(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "the cat sat", 60*time.Second)
	typeString(s, "the cat")

	if !s.DeleteWord() {
		t.Fatal("DeleteWord() = false, want true")
	}
	if got := string(s.Input()); got != "the " {
		t.Fatalf("input = %q, want %q", got, "the ")
	}
}

func TestDeleteWordCorrectPrevRemovesOnlySpace(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "the cat sat", 60*time.Second)
	typeString(s, "the ")

	if !s.DeleteWord() {
		t.Fatal("DeleteWord() = false, want true")
	}
	if got := string(s.Input()); got != "the" {
		t.Fatalf("input = %q, want %q", got, "the")
	}
}

func TestDeleteWordIncorrectPrevRemovesSpaceAndWord(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "the cat sat", 60*time.Second)
	typeString(s, "thX ")

	if !s.DeleteWord() {
		t.Fatal("DeleteWord() = false, want true")
	}
	if got := s.Input(); got != nil {
		t.Fatalf("input = %q, want empty", string(got))
	}
}

func TestDeleteWordAtStartReturnsFalse(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "the cat sat", 60*time.Second)

	if s.DeleteWord() {
		t.Fatal("DeleteWord() = true, want false")
	}
}

func TestSessionRestartResetsAndReloads(t *testing.T) {
	t.Parallel()

	calls := 0
	source := fakeSource{gen: func() (string, error) {
		calls++
		return fmt.Sprintf("run%d", calls), nil
	}}
	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	s, err := engine.NewSession(domain.TestConfig{Kind: domain.TestKindTimed, Duration: 15}, source, clock)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	typeString(s, "r")
	clock.Advance(2 * time.Second)

	if err := s.Restart(); err != nil {
		t.Fatalf("Restart: %v", err)
	}

	if s.Cursor() != 0 || s.Input() != nil {
		t.Fatalf("cursor=%d input=%v", s.Cursor(), s.Input())
	}
	if got := s.Target(); got != "run2" {
		t.Fatalf("Target() = %q, want %q", got, "run2")
	}
	if s.Remaining() != 15*time.Second {
		t.Fatalf("Remaining() = %v, want 15s", s.Remaining())
	}
	if calls != 2 {
		t.Fatalf("newTarget called %d times, want 2", calls)
	}
}

func TestSessionTimerExpiry(t *testing.T) {
	t.Parallel()

	s, clock := newTestSession(t, strings.Repeat("a", 200), 15*time.Second)
	typeString(s, "a")
	clock.Advance(16 * time.Second)

	if !s.Tick() {
		t.Fatal("Tick should finish session")
	}

	if !s.Finished() {
		t.Fatalf("Test is not finished")
	}

	if s.Remaining() != 0 {
		t.Fatalf("duration = %v", s.Remaining())
	}
}

func TestSessionWPMResult(t *testing.T) {
	t.Parallel()

	target := "ali"
	duration := 60 * time.Second
	s, clock := newTestSession(t, target, duration)

	typeString(s, "a")
	clock.Advance(2 * time.Second)
	s.Backspace()
	typeString(s, "a")
	clock.Advance(2 * time.Second)
	typeString(s, "l")
	clock.Advance(2 * time.Second)
	typeString(s, "i")

	live := s.LiveStats()
	if wpm := live.WPM; wpm != 8 {
		t.Fatalf("WPM = %v, want 8", wpm)
	}
	if got := live.RawWPM; got != 8 {
		t.Fatalf("RawWPM = %v, want 8", got)
	}
	if got := live.Accuracy; got != 100 {
		t.Fatalf("Accuracy = %v, want 100", got)
	}
}

func TestSessionCompletingTargetDoesNotFinish(t *testing.T) {
	t.Parallel()

	target := strings.Repeat("a", 200)
	s, _ := newTestSession(t, target, 15*time.Second)

	typeString(s, target)

	if s.Finished() {
		t.Fatalf("Finished() = %t, want false", s.Finished())
	}
}

func TestWordsSessionFinishesOnCompletingTarget(t *testing.T) {
	t.Parallel()

	s, _ := newWordsSession(t, "one two three")

	typeString(s, "one twX")
	s.Backspace()
	typeString(s, "o three")

	if !s.Finished() {
		t.Fatal("session should finish when the buffer covers the target")
	}
}

func TestWordsSessionFinishesEvenWithIncorrectChars(t *testing.T) {
	t.Parallel()

	s, _ := newWordsSession(t, "one two")

	typeString(s, "one twX")

	if !s.Finished() {
		t.Fatal("completion is positional; a wrong final char should still finish")
	}
}

func TestWordsSessionNeverTimesOut(t *testing.T) {
	t.Parallel()

	s, clock := newWordsSession(t, "one two three")

	typeString(s, "o")
	clock.Advance(time.Hour)

	if s.Tick() {
		t.Fatal("Tick should never finish a words session")
	}
	if s.Finished() {
		t.Fatal("words session should not time out")
	}
	if got := s.Remaining(); got != 0 {
		t.Fatalf("Remaining() = %v, want 0 (words mode has no countdown)", got)
	}
}

func TestFinishFreezesTimedSession(t *testing.T) {
	t.Parallel()

	target := strings.Repeat("a", 300)
	s, clock := newTestSession(t, target, 120*time.Second)

	typeString(s, target)
	clock.Advance(60 * time.Second)
	s.Finish()

	if !s.Finished() {
		t.Fatal("Finish should end the session")
	}
	if got := s.LiveStats().WPM; got != 60 {
		t.Fatalf("WPM = %v, want 60", got)
	}

	clock.Advance(30 * time.Second)
	if got := s.LiveStats().WPM; got != 60 {
		t.Fatalf("WPM = %v after Finish, want 60 (elapsed must freeze)", got)
	}
}

func TestEmptyKindDefaultsToTimed(t *testing.T) {
	t.Parallel()

	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	s, err := engine.NewSession(domain.TestConfig{Duration: 15}, fixedSource("abc"), clock)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	if got := s.Kind(); got != domain.TestKindTimed {
		t.Fatalf("Kind() = %q, want %q", got, domain.TestKindTimed)
	}

	typeString(s, "a")
	clock.Advance(16 * time.Second)

	if !s.Tick() {
		t.Fatal("empty kind should behave as a timed session")
	}
}

func TestTimedSessionRequiresDuration(t *testing.T) {
	t.Parallel()

	_, err := engine.NewSession(domain.TestConfig{Kind: domain.TestKindTimed}, fixedSource("abc"), nil)
	if err == nil {
		t.Fatal("timed session without a duration should error")
	}
}
