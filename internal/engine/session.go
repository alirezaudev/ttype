package engine

import (
	"errors"
	"time"
	"unicode"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/stats"
)

type TestKind string

const (
	TestKindTimed TestKind = "timed"
	TestKindWords TestKind = "words"
)

// skipRune marks input positions abandoned by a space commit.
const skipRune = '\x00'

type Config struct {
	Kind      TestKind
	Duration  time.Duration
	WordCount int
	Width     int
	Theme     string
}

type Session struct {
	newTarget      func() (string, error)
	config         Config
	target         string
	targetRunes    []rune
	input          []rune
	keystrokes     int
	correct        int
	incorrect      int
	clock          Clock
	startedAt      time.Time
	endedAt        time.Time
	capsInversions int
}

func (s *Session) Target() string      { return s.target }
func (s *Session) TargetRunes() []rune { return s.targetRunes }
func (s *Session) Input() []rune       { return append([]rune(nil), s.input...) }
func (s *Session) Cursor() int         { return len(s.input) }
func (s *Session) Kind() TestKind      { return s.config.Kind }
func (s *Session) Keystrokes() int     { return s.keystrokes }
func (s *Session) Correct() int        { return s.correct }
func (s *Session) Incorrect() int      { return s.incorrect }
func (s *Session) Started() bool       { return !s.startedAt.IsZero() }

func (s *Session) Remaining() time.Duration {
	if s.config.Kind == TestKindWords {
		return 0
	}

	left := s.config.Duration - s.Elapsed()
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

func (s *Session) WordsProgress() (completed, total int) {
	if s.config.Kind != TestKindWords {
		return 0, 0
	}

	total = s.config.WordCount
	completed = countCompletedWords(s.input, s.targetRunes)
	if completed > total {
		completed = total
	}
	return completed, total
}

func countCompletedWords(input []rune, target []rune) int {
	if len(target) == 0 {
		return 0
	}
	if len(input) >= len(target) {
		return countWordsInText(target)
	}

	specs := 0
	for i := range input {
		if target[i] == ' ' {
			specs++
		}
	}

	return specs
}

func countWordsInText(text []rune) int {
	if len(text) == 0 {
		return 0
	}

	n := 1
	for _, r := range text {
		if r == ' ' {
			n++
		}
	}
	return n
}

func (s *Session) CapsLockSuspected() bool {
	return s.capsInversions >= 2
}

func (s *Session) updateCapsStreak(typed, expected rune) {
	if !hasCase(typed) || !hasCase(expected) {
		return
	}
	if unicode.IsUpper(typed) != unicode.IsUpper(expected) {
		s.capsInversions++
	} else {
		s.capsInversions = 0
	}
}

func hasCase(r rune) bool {
	return unicode.IsUpper(r) || unicode.IsLower(r)
}

func (s *Session) Restart() error {
	s.input = s.input[:0]
	s.keystrokes = 0
	s.correct = 0
	s.incorrect = 0
	s.startedAt = time.Time{}
	s.endedAt = time.Time{}
	s.capsInversions = 0
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
	if unicode.IsControl(r) {
		return
	}

	pos := len(s.input)
	skipping := false
	if r == ' ' && pos < len(s.targetRunes) && s.targetRunes[pos] != ' ' {
		if pos == wordStartAt(s.targetRunes, pos) {
			return
		}
		skipping = true
	}

	if r != ' ' && pos < len(s.targetRunes) && s.targetRunes[pos] == ' ' {
		s.incorrect++
		s.keystrokes++
		return
	}

	if s.startedAt.IsZero() {
		s.startedAt = s.clock.Now()
	}

	if skipping {
		s.skipCurrentWord(pos)
	} else {
		s.updateCapsStreak(r, s.targetRunes[pos])
		s.keystrokes++
		switch {
		case r == s.targetRunes[pos]:
			s.correct++
		default:
			s.incorrect++
		}
		s.input = append(s.input, r)
	}

	if s.config.Kind == TestKindWords && len(s.input) >= len(s.targetRunes) {
		s.endedAt = s.clock.Now()
	}
}

func (s *Session) Backspace() bool {
	if s.Finished() || len(s.input) == 0 {
		return false
	}

	pos := len(s.input)
	if s.input[pos-1] == skipRune {
		start := pos
		for start > 0 && s.input[start-1] == skipRune {
			start--
		}
		s.input = s.input[:start]
		return true
	}

	s.input = s.input[:pos-1]
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

func wordStartAt(target []rune, pos int) int {
	if pos > len(target) {
		pos = len(target)
	}
	start := pos
	for start > 0 && target[start-1] != ' ' {
		start--
	}
	return start
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

	if s.Elapsed() >= s.config.Duration {
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
	return stats.WPM(s.correctChars(), s.Elapsed())
}

func (s *Session) Accuracy() float64 {
	return stats.Accuracy(s.correct, s.incorrect)
}

func (s *Session) RawWPM() float64 {
	return stats.RawWPM(s.rawBufferCounts(), s.Elapsed())
}

func (s *Session) rawBufferCounts() domain.CharCounts {
	var counts domain.CharCounts
	for i, r := range s.input {
		switch {
		case r == skipRune:
		case i >= len(s.targetRunes):
			counts.Extra++
		case r == s.targetRunes[i]:
			counts.Correct++
		default:
			counts.Incorrect++
		}
	}

	return counts
}

func (s *Session) Elapsed() time.Duration {
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

func (s *Session) skipCurrentWord(pos int) {
	s.incorrect++
	s.keystrokes++

	end := wordEndAt(s.targetRunes, pos)
	for i := pos; i < end; i++ {
		s.input = append(s.input, skipRune)
	}
	if end < len(s.targetRunes) && s.targetRunes[end] == ' ' {
		s.input = append(s.input, skipRune)
	}
}

func wordEndAt(target []rune, start int) int {
	end := start
	for end < len(target) && target[end] != ' ' {
		end++
	}
	return end
}
