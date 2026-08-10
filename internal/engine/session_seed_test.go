package engine_test

import (
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
)

type seedRecorder struct {
	seeds []int64
}

func (r *seedRecorder) Generate(opts domain.GenerateOptions) (string, error) {
	r.seeds = append(r.seeds, opts.Seed)
	return "alpha beta gamma", nil
}

func newSeedSession(t *testing.T, seed int64) (*engine.Session, *seedRecorder) {
	t.Helper()

	source := &seedRecorder{}
	cfg := domain.TestConfig{
		Kind:     domain.TestKindTimed,
		Duration: domain.Duration(60),
		Seed:     seed,
	}
	s, err := engine.NewSession(cfg, source, engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	return s, source
}

func TestExplicitSeedSurvivesRestart(t *testing.T) {
	t.Parallel()

	s, source := newSeedSession(t, 42)
	if s.Seed() != 42 {
		t.Fatalf("seed = %d, want 42", s.Seed())
	}

	if err := s.Restart(); err != nil {
		t.Fatalf("Restart: %v", err)
	}
	if s.Seed() != 42 {
		t.Fatalf("seed after restart = %d, want 42", s.Seed())
	}

	for i, seed := range source.seeds {
		if seed != 42 {
			t.Fatalf("generate %d used seed %d, want 42", i, seed)
		}
	}
}

func TestUnseededSessionRollsAFreshSeed(t *testing.T) {
	t.Parallel()

	s, source := newSeedSession(t, 0)
	if s.Seed() == 0 {
		t.Fatal("an unseeded session must resolve to a real seed, not zero")
	}

	first := s.Seed()
	if err := s.Restart(); err != nil {
		t.Fatalf("Restart: %v", err)
	}
	if s.Seed() == first {
		t.Fatal("restarting an unseeded test should re-roll, or the words never change")
	}

	if len(source.seeds) != 2 {
		t.Fatalf("provider called %d times, want 2", len(source.seeds))
	}
	for i, seed := range source.seeds {
		if seed == 0 {
			t.Fatalf("generate %d received a zero seed", i)
		}
	}
}
