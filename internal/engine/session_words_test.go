package engine_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestWordsRecordWhatWasTyped(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "a letter is it", 60*time.Second)
	// Extras, a fixed typo, then a skip.
	typeString(s, "a lettter ix")
	s.Backspace()
	typeString(s, "s i ")

	want := []domain.WordResult{
		{Expected: "a"},
		{Expected: "letter", Typed: "lettter", Missed: true},
		{Expected: "is", Typed: "is", Missed: true, Corrected: true},
		{Expected: "it", Typed: "i", Missed: true},
	}
	if got := s.Words(); !reflect.DeepEqual(got, want) {
		t.Fatalf("words = %+v, want %+v", got, want)
	}
}

func TestWordsStopWhereTheRunDid(t *testing.T) {
	t.Parallel()

	s, clock := newTestSession(t, "one two three", 1*time.Second)
	typeString(s, "one tw")
	clock.Advance(2 * time.Second)
	s.Tick()

	result, err := s.Result()
	if err != nil {
		t.Fatalf("Result: %v", err)
	}
	want := []domain.WordResult{{Expected: "one"}, {Expected: "two"}}
	if !reflect.DeepEqual(result.Words, want) {
		t.Fatalf("words = %+v, want %+v", result.Words, want)
	}
}

func TestRestartForgetsMissedWords(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "cat dog", 60*time.Second)
	typeString(s, "cxt")
	if err := s.Restart(); err != nil {
		t.Fatalf("Restart: %v", err)
	}
	typeString(s, "cat")

	if got := s.Words(); len(got) != 1 || got[0].Missed {
		t.Fatalf("words = %+v, want one clean word", got)
	}
}
