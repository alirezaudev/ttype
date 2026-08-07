package app

import (
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestConfigFromFlags(t *testing.T) {
	t.Parallel()

	timed, err := ConfigFromFlags(TestFlags{TimeSec: 30})
	if err != nil {
		t.Fatalf("ConfigFromFlags: %v", err)
	}
	if timed.Kind != domain.TestKindTimed || timed.Duration != domain.Duration30 {
		t.Fatalf("cfg = %+v, want a 30s timed test", timed)
	}

	words, err := ConfigFromFlags(TestFlags{TimeSec: 30, WordCount: 25})
	if err != nil {
		t.Fatalf("ConfigFromFlags: %v", err)
	}
	if words.Kind != domain.TestKindWords || words.WordCount != 25 {
		t.Fatalf("cfg = %+v, want a 25 word test", words)
	}

	if _, err := ConfigFromFlags(TestFlags{TimeSec: 30, Theme: "nope"}); err == nil {
		t.Fatal("an unknown theme should fail")
	}
}
