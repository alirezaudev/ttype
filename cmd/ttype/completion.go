package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/storage"
	"github.com/alirezaudev/ttype/internal/tui"
	"github.com/spf13/cobra"
)

func registerCompletions(root *cobra.Command) {
	_ = root.RegisterFlagCompletionFunc("mode", suggestValues(textModeNames()...))
	_ = root.RegisterFlagCompletionFunc("theme", suggestValues(tui.ThemeNames()...))
	_ = root.RegisterFlagCompletionFunc("time", suggestValues("15", "30", "60", "120"))
	_ = root.RegisterFlagCompletionFunc("words", suggestValues("10", "25", "50", "100"))

	_ = root.RegisterFlagCompletionFunc("language", suggestLanguages)

	if stats, _, err := root.Find([]string{"stats"}); err == nil {
		_ = stats.RegisterFlagCompletionFunc("mode", suggestValues(textModeNames()...))
		_ = stats.RegisterFlagCompletionFunc("export", suggestValues("csv"))
	}
	if config, _, err := root.Find([]string{"config"}); err == nil {
		_ = config.RegisterFlagCompletionFunc("default-mode", suggestValues(textModeNames()...))
		_ = config.RegisterFlagCompletionFunc("theme", suggestValues(tui.ThemeNames()...))
		_ = config.RegisterFlagCompletionFunc("default-time", suggestValues("15", "30", "60", "120"))
		_ = config.RegisterFlagCompletionFunc("default-words", suggestValues("10", "25", "50", "100"))
		_ = config.RegisterFlagCompletionFunc("default-language", suggestLanguages)
	}
	_ = root.RegisterFlagCompletionFunc("output", suggestValues("json"))
}

// Only languages already on disk are suggested: reaching for the network
// while the shell waits for completions would feel broken.
func suggestLanguages(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	dirs, err := storage.DefaultDirs()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	entries, err := os.ReadDir(filepath.Join(dirs.Data, "languages"))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var ids []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".json") || strings.HasPrefix(name, "_") {
			continue
		}
		ids = append(ids, strings.TrimSuffix(name, ".json"))
	}
	return ids, cobra.ShellCompDirectiveNoFileComp
}

func textModeNames() []string {
	modes := domain.AllTextModes()
	out := make([]string, len(modes))
	for i, mode := range modes {
		out[i] = string(mode)
	}
	return out
}

func suggestValues(values ...string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return values, cobra.ShellCompDirectiveNoFileComp
	}
}
