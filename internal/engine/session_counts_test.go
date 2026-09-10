package engine_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

// recount rebuilds the counts from scratch. The session keeps them
// incrementally, so this is the oracle that says the bookkeeping still adds up.
func recount(input, target []rune) domain.CharCounts {
	var counts domain.CharCounts
	for i, r := range input {
		switch {
		case r == '\x00':
		case i >= len(target):
			counts.Extra++
		case r == target[i]:
			counts.Correct++
		default:
			counts.Incorrect++
		}
	}
	return counts
}

func TestIncrementalCountsMatchARecount(t *testing.T) {
	t.Parallel()

	const target = "the quick brown fox jumps over the lazy dog"
	s, _ := newTestSession(t, target, 300*time.Second)

	letters := []rune("thequickbrownfxjmpsvlazydg ")
	rng := rand.New(rand.NewSource(20260910))

	for i := 0; i < 5000; i++ {
		switch n := rng.Intn(10); {
		case n < 6:
			s.InputRune(letters[rng.Intn(len(letters))])
		case n < 8:
			s.Backspace()
		default:
			s.DeleteWord()
		}

		if got, want := s.Counts(), recount(s.Input(), []rune(target)); got != want {
			t.Fatalf("after %d ops counts = %+v, want %+v (input %q)", i+1, got, want, string(s.Input()))
		}
	}
}

func TestKeystrokeCountersNeverGoDown(t *testing.T) {
	t.Parallel()

	s, _ := newTestSession(t, "the cat sat", 60*time.Second)

	typeString(s, "the cXt")
	correct, incorrect := s.Keystrokes()

	s.Backspace()
	s.DeleteWord()
	s.Backspace()

	gotCorrect, gotIncorrect := s.Keystrokes()
	if gotCorrect != correct || gotIncorrect != incorrect {
		t.Fatalf("keystrokes = %d/%d after deleting, want %d/%d — correcting a mistake must not refund accuracy",
			gotCorrect, gotIncorrect, correct, incorrect)
	}
}
