package engine_test

import (
	"testing"
	"time"
)

func TestCapsLockBasic(t *testing.T) {
	s, _ := newTestSession(t, "hello world", 60*time.Second)

	typeString(s, "H")
	if s.CapsLockSuspected() {
		t.Fatal("one inversion shouldn't trigger yet")
	}

	typeString(s, "E")
	if !s.CapsLockSuspected() {
		t.Fatal("two inversions should trigger")
	}

	typeString(s, "l") // correct case resets
	if s.CapsLockSuspected() {
		t.Fatal("correct case letter should clear it")
	}
}

func TestCapsLockBackspaceKeepsWarning(t *testing.T) {
	s, _ := newTestSession(t, "hello world", 60*time.Second)

	typeString(s, "HE")
	s.Backspace()
	s.Backspace()

	if !s.CapsLockSuspected() {
		t.Fatal("backspace should not clear the warning")
	}
}

func TestCapsLockRestartResets(t *testing.T) {
	s, _ := newTestSession(t, "hello world", 60*time.Second)

	typeString(s, "HE")
	_ = s.Restart()

	if s.CapsLockSuspected() {
		t.Fatal("restart should clear the warning")
	}
}
