package storage

import "github.com/alirezaudev/ttype/internal/domain"

const maxHistoryEntries = 1000

type Store interface {
	Paths() Dirs
	LoadSettings() (domain.Settings, error)
	SaveSettings(domain.Settings) error
	SaveResult(domain.Result) (PBUpdate, error)
	ListResults(limit int) ([]domain.Result, error)
	LoadBests() (domain.PersonalBests, error)
	Summary(domain.StatsFilter) (domain.StatsSummary, error)
}

// Replays stay off Store so a store that only keeps results is still usable;
// callers reach them through a type assertion.
type ReplayStore interface {
	SaveReplay(id string, replay domain.Replay) error
	LoadReplay(id string) (domain.Replay, error)
}
