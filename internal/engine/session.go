package engine

import (
	"errors"
	"time"
	"unicode"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/stats"
)

// skipRune marks input positions abandoned by a space commit.
const skipRune = '\x00'

type TextSource interface {
	Generate(opts domain.GenerateOptions) (string, error)
}

type Session struct {
	config              domain.TestConfig
	target              string
	targetRunes         []rune
	input               []rune
	counts              domain.CharCounts
	state               domain.SessionState
	keystrokesCorrect   int
	keystrokesIncorrect int
	clock               Clock
	source              TextSource
	startedAt           time.Time
	endedAt             time.Time
	capsInversions      int
}

func NewSession(config domain.TestConfig, source TextSource, clock Clock) (*Session, error) {
	if clock == nil {
		clock = RealClock{}
	}
	if config.Kind == "" {
		config.Kind = domain.TestKindTimed
	}
	if config.Kind == domain.TestKindTimed && config.Duration <= 0 {
		return nil, errors.New("timed session requires a positive duration")
	}
	s := &Session{
		source: source,
		config: config,
		clock:  clock,
		state:  domain.SessionReady,
	}
	if err := s.loadTarget(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Session) Target() string             { return s.target }
func (s *Session) TargetRunes() []rune        { return s.targetRunes }
func (s *Session) Input() []rune              { return append([]rune(nil), s.input...) }
func (s *Session) Cursor() int                { return len(s.input) }
func (s *Session) Kind() domain.TestKind      { return s.config.Kind }
func (s *Session) Config() domain.TestConfig  { return s.config }
func (s *Session) Counts() domain.CharCounts  { return s.counts }
func (s *Session) State() domain.SessionState { return s.state }
func (s *Session) Keystrokes() (correct, incorrect int) {
	return s.keystrokesCorrect, s.keystrokesIncorrect
}

func (s *Session) Remaining() time.Duration {
	if s.config.Kind == domain.TestKindWords {
		return 0
	}

	left := time.Duration(s.config.Duration)*time.Second - s.Elapsed()
	if left < 0 {
		return 0
	}
	return left
}

func (s *Session) LiveStats() domain.LiveStats {
	live := stats.Live(s.Counts(), s.Elapsed())
	live.Accuracy = stats.Accuracy(s.keystrokesCorrect, s.keystrokesIncorrect)
	return live
}

func (s *Session) WordsProgress() (completed, total int) {
	if s.config.Kind != domain.TestKindWords {
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
	s.counts = domain.CharCounts{}
	s.keystrokesCorrect = 0
	s.keystrokesIncorrect = 0
	s.state = domain.SessionReady
	s.startedAt = time.Time{}
	s.endedAt = time.Time{}
	s.capsInversions = 0
	return s.loadTarget()
}

func (s *Session) loadTarget() error {
	wordLimit := 0
	if s.config.IsWordsMode() {
		wordLimit = s.config.WordCount
	}

	target, err := s.source.Generate(domain.GenerateOptions{
		Language:  s.config.Language,
		WordLimit: wordLimit,
	})
	if err != nil {
		return err
	}
	s.target = target
	s.targetRunes = []rune(target)
	return nil
}

func (s *Session) InputRune(r rune) {
	if s.state == domain.SessionFinished {
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
		s.keystrokesIncorrect++
		return
	}

	if s.state == domain.SessionReady {
		s.state = domain.SessionActive
		s.startedAt = s.clock.Now()
	}

	if skipping {
		s.skipCurrentWord(pos)
	} else {
		if pos < len(s.targetRunes) {
			s.updateCapsStreak(r, s.targetRunes[pos])
		}
		switch {
		case pos >= len(s.targetRunes):
			s.keystrokesIncorrect++
			s.counts.Extra++
		case r == s.targetRunes[pos]:
			s.keystrokesCorrect++
			s.counts.Correct++
		default:
			s.keystrokesIncorrect++
			s.counts.Incorrect++
		}
		s.input = append(s.input, r)
	}

	if s.config.Kind == domain.TestKindWords && len(s.input) >= len(s.targetRunes) {
		s.finish()
	}
}

func (s *Session) Backspace() bool {
	if s.state == domain.SessionFinished || len(s.input) == 0 {
		return false
	}

	pos := len(s.input)
	if s.input[pos-1] == skipRune {
		start := pos
		for start > 0 && s.input[start-1] == skipRune {
			start--
		}
		s.truncateInput(start)
		return true
	}

	s.truncateInput(pos - 1)
	return true
}

func (s *Session) DeleteWord() bool {
	if s.state == domain.SessionFinished || len(s.input) == 0 {
		return false
	}

	pos := len(s.input)
	start := wordStart(pos, s.input)
	if start != pos {
		s.truncateInput(start)
		return true
	}

	if s.input[pos-1] != ' ' {
		return false
	}

	prevStart := wordStart(pos-1, s.input)
	if s.typedCorrectly(prevStart, pos) {
		s.truncateInput(pos - 1)
		return true
	}

	s.truncateInput(prevStart)
	return true
}

func (s *Session) truncateInput(n int) {
	for i := n; i < len(s.input); i++ {
		switch {
		case s.input[i] == skipRune:
		case i >= len(s.targetRunes):
			s.counts.Extra--
		case s.input[i] == s.targetRunes[i]:
			s.counts.Correct--
		default:
			s.counts.Incorrect--
		}
	}
	s.input = s.input[:n]
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
	if s.state != domain.SessionActive {
		return false
	}

	if s.config.Kind == domain.TestKindWords {
		return false
	}

	if s.Elapsed() >= time.Duration(s.config.Duration)*time.Second {
		s.finish()
		return true
	}

	return false
}

func (s *Session) Finish() {
	if s.state == domain.SessionFinished {
		return
	}

	if s.state == domain.SessionReady {
		s.state = domain.SessionActive
		s.startedAt = s.clock.Now()
	}

	s.finish()
}

func (s *Session) Finished() bool {
	return s.state == domain.SessionFinished
}

func (s *Session) finish() {
	s.state = domain.SessionFinished
	s.endedAt = s.clock.Now()
}

func (s *Session) Elapsed() time.Duration {
	if s.startedAt.IsZero() {
		return 0
	}

	switch s.state {
	case domain.SessionActive:
		return s.clock.Now().Sub(s.startedAt)
	case domain.SessionFinished:
		return s.endedAt.Sub(s.startedAt)
	default:
		return 0
	}
}

func (s *Session) skipCurrentWord(pos int) {
	s.keystrokesIncorrect++

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
