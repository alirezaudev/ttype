package app

import (
	"fmt"

	"github.com/alirezaudev/ttype/internal/storage"
)

func OpenStore() (storage.Store, error) {
	store, err := storage.NewDefaultStore()
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}
	return store, nil
}
