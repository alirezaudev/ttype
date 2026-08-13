package domain

import "time"

type ReplayEventKind byte

const (
	ReplayRune ReplayEventKind = iota
	ReplayBackspace
	ReplayDeleteWord
)

type ReplayEvent struct {
	Offset time.Duration
	Kind   ReplayEventKind
	Rune   rune
}

// Replay keeps the target verbatim; regenerating it from the seed breaks as
// soon as a downloaded language list changes under us.
type Replay struct {
	Target string
	Events []ReplayEvent
}
