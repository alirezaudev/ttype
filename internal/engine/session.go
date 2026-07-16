package engine

import (
	"errors"
	"math"
	"time"
)

type TestKind string

const (
	TestKindTimed TestKind = "timed"
	TestKindWords TestKind = "words"
)

type Config struct {
	Kind     TestKind
	Duration time.Duration
}

type Session struct {
	newTarget   func() (string, error)
	config      Config
	target      string
	targetRunes []rune
	input       []rune
	keystrokes  int
	correct     int
	incorrect   int
	clock       Clock
	startedAt   time.Time
	endedAt     time.Time
}

func (s *Session) Target() string      { return s.target }
func (s *Session) TargetRunes() []rune { return s.targetRunes }
func (s *Session) Input() []rune       { return append([]rune(nil), s.input...) }
func (s *Session) Cursor() int         { return len(s.input) }
func (s *Session) Kind() TestKind      { return s.config.Kind }
func (s *Session) Remaining() time.Duration {
	if s.config.Kind == TestKindWords {
		return 0
	}

	left := s.config.Duration - s.elapsed()
	if left < 0 {
		return 0
	}
	return left
}

func NewSession(newTarget func() (string, error), config Config, clock Clock) (*Session, error) {
	if clock == nil {
		clock = RealClock{}
	}
	if config.Kind == "" {
		config.Kind = TestKindTimed
	}
	if config.Kind == TestKindTimed && config.Duration <= 0 {
		return nil, errors.New("timed session requires a positive duration")
	}
	s := &Session{
		newTarget: newTarget,
		config:    config,
		clock:     clock,
	}
	if err := s.loadTarget(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Session) Restart() error {
	s.input = s.input[:0]
	s.keystrokes = 0
	s.correct = 0
	s.incorrect = 0
	s.startedAt = time.Time{}
	s.endedAt = time.Time{}
	return s.loadTarget()
}

func (s *Session) loadTarget() error {
	text, err := s.newTarget()
	if err != nil {
		return err
	}
	s.target = text
	s.targetRunes = []rune(text)
	return nil
}

func (s *Session) InputRune(r rune) {
	if s.Finished() || len(s.input) >= len(s.targetRunes) {
		return
	}

	pos := len(s.input)
	if s.startedAt.IsZero() {
		s.startedAt = s.clock.Now()
	}

	s.keystrokes++
	switch {
	case r == s.targetRunes[pos]:
		s.correct++
	default:
		s.incorrect++
	}
	s.input = append(s.input, r)

	if s.config.Kind == TestKindWords && len(s.input) >= len(s.targetRunes) {
		s.endedAt = s.clock.Now()
	}
}

func (s *Session) Backspace() bool {
	if s.Finished() || len(s.input) == 0 {
		return false
	}

	s.input = s.input[:len(s.input)-1]
	return true
}

func (s *Session) DeleteWord() bool {
	if s.Finished() || len(s.input) == 0 {
		return false
	}

	pos := len(s.input)
	start := wordStart(pos, s.input)
	if start != pos {
		s.input = s.input[:start]
		return true
	}

	if s.input[pos-1] != ' ' {
		return false
	}

	prevStart := wordStart(pos-1, s.input)
	if s.typedCorrectly(prevStart, pos) {
		s.input = s.input[:pos-1]
		return true
	}

	s.input = s.input[:prevStart]
	return true
}

func wordStart(pos int, input []rune) int {
	for i := pos - 1; i >= 0; i-- {
		if input[i] == ' ' {
			return i + 1
		}
	}
	return 0
}

func (s *Session) typedCorrectly(start, end int) bool {
	if end > len(s.targetRunes) {
		return false
	}
	for i := start; i < end; i++ {
		if s.input[i] != s.targetRunes[i] {
			return false
		}
	}
	return true
}

func (s *Session) Tick() bool {
	if s.Finished() {
		return true
	}

	if s.config.Kind == TestKindWords {
		return false
	}

	if s.elapsed() >= s.config.Duration {
		s.endedAt = s.clock.Now()
		return true
	}

	return false
}

func (s *Session) Finish() {
	if s.Finished() {
		return
	}

	s.endedAt = s.clock.Now()
}

func (s *Session) Finished() bool {
	return !s.endedAt.IsZero()
}

func (s *Session) WPM() float64 {
	minutes := s.elapsed().Minutes()
	if minutes == 0 {
		return 0
	}
	wpm := (float64(s.correctChars()) / 5.0) / minutes
	return math.Round(wpm*100) / 100
}

func (s *Session) elapsed() time.Duration {
	if s.startedAt.IsZero() {
		return 0
	}
	if s.endedAt.IsZero() {
		return s.clock.Now().Sub(s.startedAt)
	}
	return s.endedAt.Sub(s.startedAt)
}

func (s *Session) correctChars() int {
	n := 0
	for i, r := range s.input {
		if r == s.targetRunes[i] {
			n++
		}
	}
	return n
}
