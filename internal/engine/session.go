package engine

import "time"

type Session struct {
	newTarget   func() (string, error)
	target      string
	targetRunes []rune
	input       []rune
	keystrokes  int
	correct     int
	incorrect   int
	duration    time.Duration
	clock       Clock
	startedAt   time.Time
	endedAt     time.Time
}

func (s *Session) Target() string      { return s.target }
func (s *Session) TargetRunes() []rune { return s.targetRunes }
func (s *Session) Input() []rune       { return append([]rune(nil), s.input...) }
func (s *Session) Cursor() int         { return len(s.input) }
func (s *Session) Remaining() time.Duration {
	left := s.duration - s.elapsed()
	if left < 0 {
		return 0
	}
	return left
}

func NewSession(newTarget func() (string, error), duration time.Duration, clock Clock) (*Session, error) {
	if clock == nil {
		clock = RealClock{}
	}
	s := &Session{
		newTarget: newTarget,
		duration:  duration,
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
	if len(s.input) == len(s.targetRunes) {
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

func (s *Session) Tick() bool {
	if s.Finished() {
		return true
	}

	if s.elapsed() >= s.duration {
		s.endedAt = s.clock.Now()
		return true
	}

	return false
}

func (s *Session) Finished() bool {
	return !s.endedAt.IsZero()
}

func (s *Session) WPM() float64 {
	minutes := s.elapsed().Minutes()
	if minutes == 0 {
		return 0
	}
	return (float64(s.correct) / 5.0) / minutes
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
