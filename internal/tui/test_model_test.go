package tui

import (
	"testing"
	"time"
)

func TestFormatClock(t *testing.T) {
	t.Parallel()

	if got := formatClock(10 * time.Second); got != "0:10" {
		t.Fatalf("formatClock(10s) = %q, want 0:10", got)
	}
	if got := formatClock(9900 * time.Millisecond); got != "0:10" {
		t.Fatalf("formatClock(9.9s) = %q, want 0:10", got)
	}
	if got := formatClock(9400 * time.Millisecond); got != "0:09" {
		t.Fatalf("formatClock(9.4s) = %q, want 0:09", got)
	}
	if got := formatClock(90 * time.Second); got != "1:30" {
		t.Fatalf("formatClock(90s) = %q, want 1:30", got)
	}
	if got := formatClock(-time.Second); got != "0:00" {
		t.Fatalf("formatClock(-1s) = %q, want 0:00", got)
	}
}
