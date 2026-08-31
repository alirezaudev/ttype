package app

import (
	"encoding/json"
	"fmt"
	"os"

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
func RunTest(cfg domain.TestConfig, store storage.Store, version string, outputJSON bool) error {
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

	settings, err := store.LoadSettings()
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}

	model := tui.NewAppModel(tui.Options{
		Welcome:      !settings.Onboarded && !outputJSON,
		QuitOnFinish: outputJSON,
		Version:      domain.VersionInfo{Local: version},
		CheckVersion: func() domain.VersionInfo { return CheckVersion(version) },
		Config:       cfg,
		Provider:     provider,
		Cache:        provider.LanguageCache(),
		Store:        store,
		Session:      session,
	})
	final, err := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
	if err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	if !outputJSON {
		return nil
	}
	result, ok := final.(tui.AppModel).Result()
	if !ok {
		return nil
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
