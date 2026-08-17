package main

import (
	"fmt"
	"os"

	"github.com/alirezaudev/ttype/internal/app"
	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	flags := &testCLIFlags{}

	cmd := &cobra.Command{
		Use:   "ttype",
		Short: "Terminal typing practice",
		Long:  "A terminal-first typing test. Use --time for timed tests or --words for word count tests.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := app.OpenStore()
			if err != nil {
				return err
			}

			cfg, err := resolveTestConfig(cmd, store, flags)
			if err != nil {
				return err
			}
			return app.RunTest(cfg, store)
		},
	}

	cmd.Version = version
	bindTestFlags(cmd, flags)

	cmd.AddCommand(newHistoryCmd())
	cmd.AddCommand(newStatsCmd())
	cmd.AddCommand(newLanguagesCmd())

	registerCompletions(cmd)

	return cmd
}

func newHistoryCmd() *cobra.Command {
	var last, limit int
	var plain bool

	cmd := &cobra.Command{
		Use:               "history",
		Short:             "View past test results (Enter watches a replay)",
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := app.OpenStore()
			if err != nil {
				return err
			}

			n := last
			if cmd.Flags().Changed("limit") && !cmd.Flags().Changed("last") {
				n = limit
			}
			if n <= 0 {
				n = 10
			}
			return app.RunHistory(store, n, plain)
		},
	}
	cmd.Flags().IntVar(&last, "last", 10, "Number of recent results to show")
	cmd.Flags().IntVar(&limit, "limit", 0, "Alias for --last")
	cmd.Flags().BoolVar(&plain, "plain", false, "Print a plain table instead of the interactive list")

	return cmd
}

func newStatsCmd() *cobra.Command {
	var mode string
	var excludeFailed bool

	cmd := &cobra.Command{
		Use:               "stats",
		Short:             "View statistics and personal bests",
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := app.OpenStore()
			if err != nil {
				return err
			}

			filter := domain.StatsFilter{ExcludeFailed: excludeFailed}
			if cmd.Flags().Changed("mode") {
				textMode, err := domain.ParseTextMode(mode)
				if err != nil {
					return err
				}
				filter.TextMode = &textMode
			}
			return app.RunStats(store, filter)
		},
	}
	cmd.Flags().StringVar(&mode, "mode", "", "Filter stats by text mode")
	cmd.Flags().BoolVar(&excludeFailed, "exclude-failed", false, "Exclude failed tests from stats")

	return cmd
}

func newLanguagesCmd() *cobra.Command {
	var yes bool
	var jobs int

	cmd := &cobra.Command{
		Use:               "languages",
		Short:             "Manage downloadable language lists",
		ValidArgsFunction: cobra.NoFileCompletions,
	}

	download := &cobra.Command{
		Use:               "download",
		Short:             "Download all language lists to local cache",
		Long:              "Fetches every available language list once and stores it under the ttype data directory. Shows estimated download size and asks for confirmation unless --yes is passed.",
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(_ *cobra.Command, _ []string) error {
			return app.RunDownloadLanguages(os.Stdout, os.Stdin, yes, jobs)
		},
	}
	download.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	download.Flags().IntVarP(&jobs, "jobs", "j", 8, "Number of parallel downloads")

	cmd.AddCommand(download)

	return cmd
}

type testCLIFlags struct {
	timeSec     int
	wordCount   int
	language    string
	theme       string
	width       int
	mode        string
	punctuation bool
	numbers     bool
	blind       bool
	zen         bool
	minWPM      int
	seed        int64
}

func bindTestFlags(cmd *cobra.Command, f *testCLIFlags) {
	cmd.Flags().IntVar(&f.timeSec, "time", 0, "Timed test duration in seconds")
	cmd.Flags().IntVar(&f.wordCount, "words", 0, "Word count test (e.g. 25)")
	cmd.Flags().StringVar(&f.language, "language", "", "language id (e.g. spanish, english_1k); downloads once and caches locally")
	cmd.Flags().StringVar(&f.mode, "mode", "", "Text mode")
	cmd.Flags().StringVar(&f.theme, "theme", "", "Color theme")
	cmd.Flags().IntVar(&f.width, "width", 0, "Typing area width in characters")
	cmd.Flags().BoolVar(&f.punctuation, "punctuation", false, "Inject punctuation between words")
	cmd.Flags().BoolVar(&f.numbers, "numbers", false, "Inject numbers into word tests")
	cmd.Flags().BoolVar(&f.blind, "blind", false, "Hide correctness while typing")
	cmd.Flags().BoolVar(&f.zen, "zen", false, "Words only: no header, no hints")
	cmd.Flags().IntVar(&f.minWPM, "min-wpm", 0, "Fail the test if WPM drops below this")
	cmd.Flags().Int64Var(&f.seed, "seed", 0, "Random seed for word generation")
}

func resolveTestConfig(cmd *cobra.Command, store interface {
	LoadSettings() (domain.Settings, error)
}, f *testCLIFlags) (domain.TestConfig, error) {
	settings, err := store.LoadSettings()
	if err != nil {
		return domain.TestConfig{}, err
	}

	timeSec := f.timeSec
	if !cmd.Flags().Changed("time") || timeSec == 0 {
		timeSec = settings.DefaultDuration.Seconds()
	}
	if timeSec <= 0 {
		timeSec = domain.Duration60.Seconds()
	}

	wordCount := f.wordCount
	if !cmd.Flags().Changed("words") {
		wordCount = settings.DefaultWordCount
	}

	language := f.language
	if !cmd.Flags().Changed("language") {
		language = settings.Language
	}

	theme := f.theme
	if !cmd.Flags().Changed("theme") || theme == "" {
		theme = settings.Theme
	}

	width := f.width
	if !cmd.Flags().Changed("width") {
		width = settings.DefaultWidth
	}

	mode := f.mode
	if !cmd.Flags().Changed("mode") {
		mode = string(settings.DefaultMode)
	}

	punctuation := f.punctuation
	if !cmd.Flags().Changed("punctuation") {
		punctuation = settings.Punctuation
	}

	numbers := f.numbers
	if !cmd.Flags().Changed("numbers") {
		numbers = settings.Numbers
	}

	blind := f.blind
	if !cmd.Flags().Changed("blind") {
		blind = settings.Blind
	}

	zen := f.zen
	if !cmd.Flags().Changed("zen") {
		zen = settings.Zen
	}

	minWPM := f.minWPM
	if !cmd.Flags().Changed("min-wpm") {
		minWPM = settings.DefaultMinWPM
	}

	return app.ConfigFromFlags(app.TestFlags{
		TimeSec:     timeSec,
		WordCount:   wordCount,
		Language:    language,
		Theme:       theme,
		Width:       width,
		Mode:        mode,
		Punctuation: punctuation,
		Numbers:     numbers,
		Blind:       blind,
		Zen:         zen,
		MinWPM:      minWPM,
		Seed:        f.seed,
	})
}
