package engine_test

import (
	"reflect"
	"testing"
	"time"
)

func TestRawWPMHistoryIsInstantaneous(t *testing.T) {
	t.Parallel()

	s, clock := newTestSession(t, "abcdefghij", 60*time.Second)

	typeString(s, "ab")
	clock.Advance(2500 * time.Millisecond)
	typeString(s, "cd")
	clock.Advance(500 * time.Millisecond)
	s.Finish()

	want := []float64{24, 0, 24}
	if got := s.RawWPMHistory(); !reflect.DeepEqual(got, want) {
		t.Fatalf("raw history = %v, want %v", got, want)
	}
	if got := s.ErrorHistory(); !reflect.DeepEqual(got, []int{0, 0, 0}) {
		t.Fatalf("error history = %v, want [0 0 0]", got)
	}
}

func TestShortLastSecondIsDropped(t *testing.T) {
	t.Parallel()

	s, clock := newTestSession(t, "abcdefghij", 60*time.Second)

	typeString(s, "ab")
	clock.Advance(1200 * time.Millisecond)
	typeString(s, "cd")
	s.Finish()

	if got := len(s.RawWPMHistory()); got != 1 {
		t.Fatalf("samples = %d, want 1", got)
	}
}

func TestRestartClearsTheSeries(t *testing.T) {
	t.Parallel()

	s, clock := newTestSession(t, "abcdefghij", 60*time.Second)

	typeString(s, "ab")
	clock.Advance(time.Second)
	s.Finish()

	if err := s.Restart(); err != nil {
		t.Fatalf("Restart: %v", err)
	}
	if got := s.RawWPMHistory(); len(got) != 0 {
		t.Fatalf("raw history = %v, want empty", got)
	}
	if got := s.WPMHistory(); len(got) != 0 {
		t.Fatalf("wpm history = %v, want empty", got)
	}
}
