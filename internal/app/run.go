package app

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
	"github.com/alirezaudev/ttype/internal/storage"
	"github.com/alirezaudev/ttype/internal/text"
	"github.com/alirezaudev/ttype/internal/text/langcache"
	"github.com/alirezaudev/ttype/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
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
	// ResultFile gets the last run as JSON when ttype exits.
	ResultFile string
	Source     ResultSource
}

// RunTest starts an interactive typing test in the terminal.
func RunTest(cfg domain.TestConfig, store storage.Store, build Build, opts RunOptions) error {
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
	warning := ""
	// A saved language whose list is gone can't come back offline, and that
	// must not stop ttype from starting. Scripts get the error instead.
	if err != nil && cfg.Language != "" && !outputJSON {
		warning = fmt.Sprintf("couldn't load %s, so this is english", langcache.DisplayName(cfg.Language))
		cfg.Language = ""
		session, err = engine.NewSession(cfg, source, nil)
	}
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}

	settings, err := store.LoadSettings()
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}

	var (
		checkVersion  func() domain.VersionInfo
		installUpdate func(latest string, asked bool) error
	)
	// Scripts and CI get a quiet run that never touches the network.
	mode := UpdateModeFor(settings.Update, os.Getenv("TTYPE_UPDATE"))
	if !outputJSON && os.Getenv("CI") == "" && mode != domain.UpdateOff {
		checkVersion = func() domain.VersionInfo { return CheckVersion(build, mode) }
		installUpdate = func(latest string, asked bool) error { return InstallUpdate(build, latest, asked) }
	}

	model := tui.NewAppModel(tui.Options{
		// The welcome screen picks a mode, and custom text already has one.
		Welcome:       !settings.Onboarded && !outputJSON && opts.CustomText == "",
		QuitOnFinish:  outputJSON,
		NoSave:        opts.NoSave,
		Version:       domain.VersionInfo{Local: build.Version},
		CheckVersion:  checkVersion,
		InstallUpdate: installUpdate,
		Config:        cfg,
		Provider:      source,
		Cache:         provider.LanguageCache(),
		Store:         store,
		Session:       session,
		Warning:       warning,
	})
	options := []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion()}
	// With stdout redirected (--output json > run.json), the test is drawn on
	// the terminal itself, so stdout only gets the JSON.
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		if tty, err := os.OpenFile(terminalPath(), os.O_WRONLY, 0); err == nil {
			defer tty.Close()
			options = append(options, tea.WithOutput(tty))
		}
	}
	final, err := tea.NewProgram(model, options...).Run()
	if err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	if opts.ResultFile != "" {
		result, finished, ran := final.(tui.AppModel).LastRun()
		if err := writeResultFile(opts.ResultFile, newResultFile(result, finished, ran, opts.Source)); err != nil {
			return err
		}
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

func terminalPath() string {
	if runtime.GOOS == "windows" {
		return "CONOUT$"
	}
	return "/dev/tty"
}
