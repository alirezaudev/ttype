package engine_test

import (
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestSpaceMidWordSkipsToNextWord(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "abc def", 60*time.Second)
	typeString(s, "a ")

	if got := s.Cursor(); got != 4 {
		t.Fatalf("cursor = %d, want 4", got)
	}
	if got := string(s.Input()); got != "a\x00\x00\x00" {
		t.Fatalf("input = %q, want %q", got, "a\x00\x00\x00")
	}
	if got := s.Counts().Correct; got != 1 {
		t.Fatalf("correct = %d, want 1", got)
	}
	if got := s.Counts().Incorrect; got != 0 {
		t.Fatalf("incorrect = %d, want 0", got)
	}

	correct, incorrect := s.Keystrokes()
	if incorrect != 1 {
		t.Fatalf("incorrect keystrokes = %d, want 1", incorrect)
	}

	if correct != 1 {
		t.Fatalf("keystrokes = %d, want 1", correct)
	}
}

func TestSpaceOnUntouchedWordIsBlocked(t *testing.T) {
	t.Parallel()

	s, clock := newTestSession(t, "abc def", 60*time.Second)

	s.InputRune(' ')
	if got := s.Cursor(); got != 0 {
		t.Fatalf("cursor = %d, want 0", got)
	}

	clock.Advance(61 * time.Second)
	if s.Tick() || s.State() == domain.SessionFinished {
		t.Fatal("a blocked space must not start the session")
	}

	typeString(s, "abc ")
	s.InputRune(' ')
	if got := s.Cursor(); got != 4 {
		t.Fatalf("cursor = %d, want 4", got)
	}
}

func TestSpaceAtSeparatorIsPlainInput(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "abc def", 60*time.Second)
	typeString(s, "abc ")

	if got := string(s.Input()); got != "abc " {
		t.Fatalf("input = %q, want %q", got, "abc ")
	}
}

func TestSpaceSkipOnLastWordFinishesWordsSession(t *testing.T) {
	t.Parallel()

	s, _ := newWordsSession(t, "abc def")
	typeString(s, "abc d ")

	if s.State() != domain.SessionFinished {
		t.Fatal("skipping the last word should finish the test")
	}
	if got := s.Cursor(); got != 7 {
		t.Fatalf("cursor = %d, want 7", got)
	}
}

func TestBackspaceUndoesSkip(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "abc def", 60*time.Second)
	typeString(s, "a ")

	if !s.Backspace() {
		t.Fatal("Backspace() = false, want true")
	}
	if got := string(s.Input()); got != "a" {
		t.Fatalf("input = %q, want %q", got, "a")
	}

	typeString(s, "bc def")
	if got := string(s.Input()); got != "abc def" {
		t.Fatalf("input = %q, want %q", got, "abc def")
	}
}

func TestDeleteWordClearsSkippedWord(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "abc def", 60*time.Second)
	typeString(s, "a ")

	if !s.DeleteWord() {
		t.Fatal("DeleteWord() = false, want true")
	}
	if got := s.Input(); got != nil {
		t.Fatalf("input = %q, want empty", string(got))
	}
}

func TestDeleteWordAfterConsecutiveSkips(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "ab cd ef", 60*time.Second)
	typeString(s, "a c ")

	if !s.DeleteWord() {
		t.Fatal("DeleteWord() = false, want true")
	}
	if got := s.Input(); got != nil {
		t.Fatalf("input = %q, want empty", string(got))
	}
}

func TestExtraCharsAtWordEndDontSpill(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "cat dog", 60*time.Second)
	typeString(s, "catxx")

	if got := s.Cursor(); got != 3 {
		t.Fatalf("cursor = %d, want 3", got)
	}
	if got := string(s.Input()); got != "cat" {
		t.Fatalf("input = %q, want %q", got, "cat")
	}
	correct, incorrect := s.Keystrokes()
	if correct != 3 {
		t.Fatalf("correct keystrokes = %d, want 3", correct)
	}
	if incorrect != 2 {
		t.Fatalf("incorrect keystrokes = %d, want 2", incorrect)
	}
	if got := s.LiveStats().Accuracy; got != 60 {
		t.Fatalf("accuracy = %v, want 60", got)
	}

	typeString(s, " dog")
	if got := string(s.Input()); got != "cat dog" {
		t.Fatalf("input = %q, want %q", got, "cat dog")
	}
}

func TestExtraLettersAreKeptAfterTheWord(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "cat dog", 60*time.Second)
	typeString(s, "catxx")

	if got := string(s.ExtrasAt(3)); got != "xx" {
		t.Fatalf("extras = %q, want %q", got, "xx")
	}
	if got := s.Counts().Extra; got != 2 {
		t.Fatalf("extra = %d, want 2", got)
	}

	// Space keeps the extras.
	typeString(s, " d")
	if got := string(s.ExtrasAt(3)); got != "xx" {
		t.Fatalf("extras after space = %q, want %q", got, "xx")
	}

	// d, the space, then the extras.
	s.Backspace()
	s.Backspace()
	s.Backspace()
	if got := string(s.ExtrasAt(3)); got != "x" {
		t.Fatalf("extras = %q, want %q", got, "x")
	}
	if got := string(s.Input()); got != "cat" {
		t.Fatalf("input = %q, want %q", got, "cat")
	}
	s.Backspace()
	s.Backspace()
	if got := string(s.Input()); got != "ca" {
		t.Fatalf("input = %q, want %q", got, "ca")
	}
	if got := s.Counts().Extra; got != 0 {
		t.Fatalf("extra = %d, want 0", got)
	}
}

func TestDeleteWordTakesTheExtrasWithIt(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "cat dog", 60*time.Second)
	typeString(s, "catxx")
	s.DeleteWord()

	if s.Cursor() != 0 || len(s.ExtrasAt(3)) != 0 || s.Counts().Extra != 0 {
		t.Fatalf("cursor=%d extras=%q extra=%d", s.Cursor(), string(s.ExtrasAt(3)), s.Counts().Extra)
	}
}

func TestExtraLettersAreCapped(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "cat dog", 60*time.Second)
	typeString(s, "cat")
	for range 50 {
		s.InputRune('x')
	}

	if got := len(s.ExtrasAt(3)); got != 20 {
		t.Fatalf("extras = %d, want 20", got)
	}
	// Still wrong keys past the cap.
	if _, incorrect := s.Keystrokes(); incorrect != 50 {
		t.Fatalf("incorrect keystrokes = %d, want 50", incorrect)
	}
}

func TestRestartClearsSkip(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "abc def", 60*time.Second)
	typeString(s, "a ")

	if err := s.Restart(); err != nil {
		t.Fatalf("Restart: %v", err)
	}
	if s.Cursor() != 0 || s.Input() != nil {
		t.Fatalf("cursor=%d input=%q", s.Cursor(), string(s.Input()))
	}
}

func TestWordsProgressAdvancesPastSkippedWord(t *testing.T) {
	t.Parallel()
	s, _ := newWordsSession(t, "abc def")

	typeString(s, "a ")
	done, total := s.WordsProgress()
	if done != 1 || total != 2 {
		t.Fatalf("progress = %d/%d, want 1/2", done, total)
	}
}

func TestSkippedCountsMissedLetters(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "abcd efgh", 60*time.Second)
	typeString(s, "ab ")
	if got := s.Skipped(); got != 2 {
		t.Fatalf("skipped = %d, want 2", got)
	}

	typeString(s, "ef ")
	if got := s.Skipped(); got != 4 {
		t.Fatalf("skipped = %d, want 4", got)
	}

	s.Backspace()
	if got := s.Skipped(); got != 2 {
		t.Fatalf("skipped after undo = %d, want 2", got)
	}
}
