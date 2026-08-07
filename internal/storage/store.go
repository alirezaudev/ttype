package storage

import "github.com/alirezaudev/ttype/internal/domain"

const maxHistoryEntries = 1000

type Store interface {
	Paths() Dirs
	LoadSettings() (domain.Settings, error)
	SaveSettings(domain.Settings) error
	SaveResult(domain.Result) error
	ListResults(limit int) ([]domain.Result, error)
}
