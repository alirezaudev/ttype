package engine_test

import (
	"testing"
	"time"
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
	if got := s.Correct(); got != 1 {
		t.Fatalf("correct = %d, want 1", got)
	}
	if got := s.Incorrect(); got != 1 {
		t.Fatalf("incorrect = %d, want 1", got)
	}
	if got := s.Keystrokes(); got != 2 {
		t.Fatalf("keystrokes = %d, want 2", got)
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
	if s.Tick() || s.Finished() {
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

	if !s.Finished() {
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
	if got := s.Keystrokes(); got != 5 {
		t.Fatalf("keystrokes = %d, want 5", got)
	}
	if got := s.Correct(); got != 3 {
		t.Fatalf("correct = %d, want 3", got)
	}
	if got := s.Incorrect(); got != 2 {
		t.Fatalf("incorrect = %d, want 2", got)
	}
	if got := s.Accuracy(); got != 60 {
		t.Fatalf("accuracy = %v, want 60", got)
	}

	typeString(s, " dog")
	if got := string(s.Input()); got != "cat dog" {
		t.Fatalf("input = %q, want %q", got, "cat dog")
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
