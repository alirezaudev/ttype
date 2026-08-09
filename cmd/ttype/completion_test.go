package main

import (
	"slices"
	"testing"

	"github.com/alirezaudev/ttype/internal/tui"
	"github.com/spf13/cobra"
)

func TestFlagCompletionsNeverOfferFiles(t *testing.T) {
	t.Parallel()

	root := newRootCmd()

	cases := []struct {
		flag string
		want []string
	}{
		{flag: "theme", want: tui.ThemeNames()},
		{flag: "time", want: []string{"15", "30", "60", "120"}},
		{flag: "words", want: []string{"10", "25", "50", "100"}},
		{flag: "language", want: nil},
	}

	for _, tc := range cases {
		t.Run(tc.flag, func(t *testing.T) {
			complete, ok := root.GetFlagCompletionFunc(tc.flag)
			if !ok {
				t.Fatalf("--%s has no completion func, so the shell completes filenames", tc.flag)
			}

			got, directive := complete(root, nil, "")
			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Fatalf("--%s directive = %v, want NoFileComp", tc.flag, directive)
			}
			if !slices.Equal(got, tc.want) {
				t.Fatalf("--%s suggests %q, want %q", tc.flag, got, tc.want)
			}
		})
	}
}

func TestSubcommandsWithoutArgsCompleteNoFiles(t *testing.T) {
	t.Parallel()

	root := newRootCmd()

	for _, path := range [][]string{{"history"}, {"languages"}, {"languages", "download"}} {
		cmd, _, err := root.Find(path)
		if err != nil {
			t.Fatalf("find %v: %v", path, err)
		}
		if cmd.ValidArgsFunction == nil {
			t.Fatalf("%q takes no positional args but completes filenames", cmd.CommandPath())
		}
	}
}
