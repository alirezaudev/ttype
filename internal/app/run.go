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

// RunOptions are the choices that shape a run beyond its test config.
type RunOptions struct {
	OutputJSON bool
	// CustomText, when set, is what gets typed instead of generated text.
	CustomText string
	// NoSave keeps the run out of history, bests and replays.
	NoSave bool
}

// RunTest starts an interactive typing test in the terminal.
func RunTest(cfg domain.TestConfig, store storage.Store, version string, opts RunOptions) error {
	outputJSON := opts.OutputJSON
	dirs, err := storage.DefaultDirs()
	if err != nil {
		return fmt.Errorf("resolve data dir: %w", err)
	}

	provider, err := text.NewProvider(dirs.Data)
	if err != nil {
		return fmt.Errorf("load text: %w", err)
	}

	var source engine.TextSource = provider
	if opts.CustomText != "" {
		source = newCustomSource(opts.CustomText, provider)
	}

	session, err := engine.NewSession(cfg, source, nil)
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}

	settings, err := store.LoadSettings()
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}

	var checkVersion func() domain.VersionInfo
	// Scripts and CI get a quiet run that never touches the network.
	if !outputJSON && os.Getenv("CI") == "" {
		checkVersion = func() domain.VersionInfo { return CheckVersion(version) }
	}

	model := tui.NewAppModel(tui.Options{
		// The welcome screen picks a mode, and custom text already has one.
		Welcome:      !settings.Onboarded && !outputJSON && opts.CustomText == "",
		QuitOnFinish: outputJSON,
		NoSave:       opts.NoSave,
		Version:      domain.VersionInfo{Local: version},
		CheckVersion: checkVersion,
		Config:       cfg,
		Provider:     source,
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
