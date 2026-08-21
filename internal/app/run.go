package app

import (
	"fmt"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
	"github.com/alirezaudev/ttype/internal/storage"
	"github.com/alirezaudev/ttype/internal/text"
	"github.com/alirezaudev/ttype/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

// OpenStore opens the default JSON store.
func OpenStore() (storage.Store, error) {
	store, err := storage.NewDefaultStore()
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}
	return store, nil
}

// RunTest starts an interactive typing test in the terminal.
func RunTest(cfg domain.TestConfig, store storage.Store) error {
	dirs, err := storage.DefaultDirs()
	if err != nil {
		return fmt.Errorf("resolve data dir: %w", err)
	}

	provider, err := text.NewProvider(dirs.Data)
	if err != nil {
		return fmt.Errorf("load text: %w", err)
	}

	session, err := engine.NewSession(cfg, provider, nil)
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}

	model := tui.NewAppModel(tui.Options{
		Config:   cfg,
		Provider: provider,
		Cache:    provider.LanguageCache(),
		Store:    store,
		Session:  session,
	})
	if _, err := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}
