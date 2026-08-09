package main

import (
	"github.com/alirezaudev/ttype/internal/tui"
	"github.com/spf13/cobra"
)

func registerCompletions(root *cobra.Command) {
	_ = root.RegisterFlagCompletionFunc("theme", suggestValues(tui.ThemeNames()...))
	_ = root.RegisterFlagCompletionFunc("time", suggestValues("15", "30", "60", "120"))
	_ = root.RegisterFlagCompletionFunc("words", suggestValues("10", "25", "50", "100"))

	_ = root.RegisterFlagCompletionFunc("language", suggestValues())
}

func suggestValues(values ...string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return values, cobra.ShellCompDirectiveNoFileComp
	}
}
