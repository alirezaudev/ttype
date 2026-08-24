package app

import (
	"fmt"
	"os"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/storage"
	"github.com/alirezaudev/ttype/internal/tui"
	"golang.org/x/term"
)

type StatsOptions struct {
	Filter    domain.StatsFilter
	Trend     bool
	ExportCSV bool
}

func RunStats(store storage.Store, opts StatsOptions) error {
	if opts.ExportCSV {
		return store.ExportCSV(os.Stdout, opts.Filter)
	}

	summary, err := store.Summary(opts.Filter)
	if err != nil {
		return fmt.Errorf("load stats: %w", err)
	}

	var trend []float64
	if opts.Trend {
		if trend, err = store.Trend(opts.Filter); err != nil {
			return fmt.Errorf("load trend: %w", err)
		}
	}

	settings, err := store.LoadSettings()
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}

	fmt.Fprint(os.Stdout, tui.RenderStats(summary, trend, tui.ResolveTheme(settings.Theme), terminalWidth()))
	return nil
}

func terminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 72
	}
	return width
}
