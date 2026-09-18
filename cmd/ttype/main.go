package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/alirezaudev/ttype/internal/app"
	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	// Only the release pipeline stamps a bare version; make stamps git describe.
	release bool
)

// Pseudo-versions and "(devel)" are not releases, so they stay "dev".
var releaseVersion = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

func main() {
	if runtime.GOOS == "windows" {
		app.RemoveOldBinary()
	}
	release = releaseVersion.MatchString("v" + version)
	version = resolveVersion(version, debug.ReadBuildInfo)
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// go install skips the release ldflags but still stamps the module version.
func resolveVersion(stamped string, readBuildInfo func() (*debug.BuildInfo, bool)) string {
	if stamped != "dev" {
		return stamped
	}
	info, ok := readBuildInfo()
	if !ok || !releaseVersion.MatchString(info.Main.Version) {
		return stamped
	}
	return strings.TrimPrefix(info.Main.Version, "v")
}

func newRootCmd() *cobra.Command {
	flags := &testCLIFlags{}

	cmd := &cobra.Command{
		Use:   "ttype",
		Short: "Terminal typing practice",
		Long:  "A terminal-first typing test. Use --time for timed tests or --words for word count tests.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Flags parsed fine, so the usage text would only bury the error;
			// main prints it once.
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			store, err := app.OpenStore()
			if err != nil {
				return err
			}

			if flags.output != "" && flags.output != "json" {
				return fmt.Errorf("unknown output format %q (only json)", flags.output)
			}

			cfg, err := resolveTestConfig(cmd, store, flags)
			if err != nil {
				return err
			}
			cfg.Tag = flags.tag

			text, err := readCustomText(cmd, flags, stdinPiped())
			if err != nil {
				return err
			}
			if text != "" {
				cfg = app.CustomConfig(cfg, text, cmd.Flags().Changed("words"), cmd.Flags().Changed("time"))
			}

			source := app.ResultSource{Text: text}
			if flags.file != "-" {
				source.File = flags.file
			}
			return app.RunTest(cfg, store, app.Build{Version: version, Release: release}, app.RunOptions{
				OutputJSON: flags.output == "json",
				CustomText: text,
				NoSave:     flags.noSave,
				ResultFile: flags.resultFile,
				Source:     source,
			})
		},
	}

	cmd.Version = version
	bindTestFlags(cmd, flags)

	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newHistoryCmd())
	cmd.AddCommand(newStatsCmd())
	cmd.AddCommand(newLanguagesCmd())
	cmd.AddCommand(newDoctorCmd())
	cmd.AddCommand(newClearCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newUninstallCmd())

	registerCompletions(cmd)

	return cmd
}

func newConfigCmd() *cobra.Command {
	var (
		duration, wordCount, width, minWPM int
		mode, language, theme, update      string
		punctuation, numbers, blind, zen   bool
	)

	cmd := &cobra.Command{
		Use:               "config",
		Short:             "Show or change the saved defaults",
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := app.OpenStore()
			if err != nil {
				return err
			}

			changes := app.ConfigChanges{}
			flags := cmd.Flags()
			if flags.Changed("default-time") {
				changes.Duration = &duration
			}
			if flags.Changed("default-words") {
				changes.WordCount = &wordCount
			}
			if flags.Changed("default-mode") {
				changes.Mode = &mode
			}
			if flags.Changed("default-language") {
				changes.Language = &language
			}
			if flags.Changed("theme") {
				changes.Theme = &theme
			}
			if flags.Changed("default-width") {
				changes.Width = &width
			}
			if flags.Changed("default-min-wpm") {
				changes.MinWPM = &minWPM
			}
			if flags.Changed("default-punctuation") {
				changes.Punctuation = &punctuation
			}
			if flags.Changed("default-numbers") {
				changes.Numbers = &numbers
			}
			if flags.Changed("default-blind") {
				changes.Blind = &blind
			}
			if flags.Changed("default-zen") {
				changes.Zen = &zen
			}
			if flags.Changed("update") {
				changes.Update = &update
			}
			return app.RunConfig(store, changes)
		},
	}

	cmd.Flags().IntVar(&duration, "default-time", 0, "Default timed test duration in seconds")
	cmd.Flags().IntVar(&wordCount, "default-words", 0, "Default word count test size")
	cmd.Flags().StringVar(&mode, "default-mode", "", "Default text mode")
	cmd.Flags().StringVar(&language, "default-language", "", "Default language id")
	cmd.Flags().StringVar(&theme, "theme", "", "Default color theme")
	cmd.Flags().IntVar(&width, "default-width", 0, "Default typing area width (0 = auto)")
	cmd.Flags().IntVar(&minWPM, "default-min-wpm", 0, "Default minimum WPM (0 = off)")
	cmd.Flags().BoolVar(&punctuation, "default-punctuation", false, "Inject punctuation by default")
	cmd.Flags().BoolVar(&numbers, "default-numbers", false, "Inject numbers by default")
	cmd.Flags().BoolVar(&blind, "default-blind", false, "Start in blind mode by default")
	cmd.Flags().BoolVar(&zen, "default-zen", false, "Start in zen mode by default")
	cmd.Flags().StringVar(&update, "update", "", "New releases: auto installs them in the background, notify only tells you, off never checks")

	return cmd
}

func newHistoryCmd() *cobra.Command {
	var last, limit int
	var plain bool
	var tag string

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
			return app.RunHistory(store, n, plain, tag)
		},
	}
	cmd.Flags().IntVar(&last, "last", 10, "Number of recent results to show")
	cmd.Flags().IntVar(&limit, "limit", 0, "Alias for --last")
	cmd.Flags().BoolVar(&plain, "plain", false, "Print a plain table instead of the interactive list")
	cmd.Flags().StringVar(&tag, "tag", "", "Only runs with this tag")

	return cmd
}

func newStatsCmd() *cobra.Command {
	var mode, export, tag string
	var excludeFailed, trend bool

	cmd := &cobra.Command{
		Use:               "stats",
		Short:             "View statistics and personal bests",
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := app.OpenStore()
			if err != nil {
				return err
			}

			filter := domain.StatsFilter{ExcludeFailed: excludeFailed, Tag: tag}
			if cmd.Flags().Changed("mode") {
				textMode, err := domain.ParseTextMode(mode)
				if err != nil {
					return err
				}
				filter.TextMode = &textMode
			}

			if export != "" && export != "csv" {
				return fmt.Errorf("unknown export format %q (only csv)", export)
			}
			return app.RunStats(store, app.StatsOptions{
				Filter:    filter,
				Trend:     trend,
				ExportCSV: export == "csv",
			})
		},
	}
	cmd.Flags().StringVar(&mode, "mode", "", "Filter stats by text mode")
	cmd.Flags().StringVar(&tag, "tag", "", "Filter stats by tag")
	cmd.Flags().BoolVar(&excludeFailed, "exclude-failed", false, "Exclude failed tests from stats")
	cmd.Flags().BoolVar(&trend, "trend", false, "Show a sparkline of wpm over time")
	cmd.Flags().StringVar(&export, "export", "", "Export the history instead of printing stats (csv)")

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

func newUninstallCmd() *cobra.Command {
	var purge, yes bool

	cmd := &cobra.Command{
		Use:               "uninstall",
		Short:             "Remove ttype from this machine",
		ValidArgsFunction: cobra.NoFileCompletions,
		// The error says what to do; usage would bury it, and main prints it.
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return app.RunUninstall(os.Stdout, os.Stdin, app.Build{Version: version, Release: release}, purge, yes)
		},
	}
	cmd.Flags().BoolVar(&purge, "purge", false, "Also delete your settings and history")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")

	return cmd
}

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "update",
		Short:             "Update ttype to the latest release",
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		SilenceErrors:     true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return app.RunUpdate(os.Stdout, app.Build{Version: version, Release: release})
		},
	}
}

func newClearCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:       "clear [history|languages|all]",
		Short:     "Delete saved history or downloaded languages",
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: []string{"history", "languages", "all"},
		RunE: func(_ *cobra.Command, args []string) error {
			target := app.ClearHistory
			if len(args) == 1 {
				parsed, err := app.ParseClearTarget(args[0])
				if err != nil {
					return err
				}
				target = parsed
			}

			store, err := app.OpenStore()
			if err != nil {
				return err
			}
			return app.RunClear(store, target, yes, os.Stdout, os.Stdin)
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")

	return cmd
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "doctor",
		Short:             "Check that the terminal and environment are set up",
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(_ *cobra.Command, _ []string) error {
			store, err := app.OpenStore()
			if err != nil {
				return err
			}
			return app.RunDoctor(store)
		},
	}
}

// readCustomText reads text piped in, from --file or from --text. The flags
// that generate text make no sense next to it, so they are refused rather
// than quietly ignored.
func readCustomText(cmd *cobra.Command, f *testCLIFlags, piped bool) (string, error) {
	text, err := app.ReadCustomText(app.CustomInput{
		Stdin:      os.Stdin,
		StdinPiped: piped,
		File:       f.file,
		Text:       f.text,
		TextSet:    cmd.Flags().Changed("text"),
	})
	if err != nil || text == "" {
		return "", err
	}
	for _, name := range []string{"mode", "language", "punctuation", "numbers"} {
		if cmd.Flags().Changed(name) {
			return "", fmt.Errorf("--%s generates text, so it can't be used with your own", name)
		}
	}
	return text, nil
}

// A terminal, /dev/null or no stdin at all is not text to type.
func stdinPiped() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeNamedPipe != 0 || info.Mode().IsRegular()
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
	output      string
	file        string
	text        string
	noSave      bool
	resultFile  string
	tag         string
	allowSkip   bool
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
	cmd.Flags().StringVar(&f.output, "output", "", "Print the result instead of showing it (json)")
	cmd.Flags().StringVar(&f.file, "file", "", "Type the text of a file (- for stdin)")
	cmd.Flags().StringVar(&f.text, "text", "", "Type this text")
	cmd.Flags().BoolVar(&f.noSave, "no-save", false, "Keep this run out of history, bests and replays")
	cmd.Flags().StringVar(&f.resultFile, "result-file", "", "Write the last run to this file as JSON on exit")
	cmd.Flags().StringVar(&f.tag, "tag", "", "Label this run, to filter history and stats by it")

	// Skipping is part of how space works now, so the old flag does nothing.
	cmd.Flags().BoolVar(&f.allowSkip, "allow-skip", false, "Deprecated, has no effect")
	_ = cmd.Flags().MarkDeprecated("allow-skip", "space already commits the word")
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
