package storage

import "github.com/alirezaudev/ttype/internal/domain"

type Store interface {
	Path() string
	LoadSettings() (domain.Settings, error)
	SaveSettings(domain.Settings) error
}
