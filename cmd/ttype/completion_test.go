package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/alirezaudev/ttype/internal/storage"
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

func TestConfigAndStatsFlagsComplete(t *testing.T) {
	t.Parallel()

	root := newRootCmd()
	for _, tc := range []struct{ path, flag, want string }{
		{"config", "default-mode", "sql"},
		{"config", "theme", "dracula"},
		{"stats", "export", "csv"},
	} {
		cmd, _, err := root.Find([]string{tc.path})
		if err != nil {
			t.Fatalf("find %s: %v", tc.path, err)
		}
		values := completeFlag(t, cmd, tc.flag)
		if !slices.Contains(values, tc.want) {
			t.Errorf("%s --%s completions = %v, want %q", tc.path, tc.flag, values, tc.want)
		}
	}
}

func TestOutputFlagCompletes(t *testing.T) {
	t.Parallel()

	values := completeFlag(t, newRootCmd(), "output")
	if !slices.Contains(values, "json") {
		t.Fatalf("--output completions = %v, want json", values)
	}
}

func completeFlag(t *testing.T, cmd *cobra.Command, flag string) []string {
	t.Helper()

	complete, ok := cmd.GetFlagCompletionFunc(flag)
	if !ok {
		t.Fatalf("--%s has no completion func, so the shell completes filenames", flag)
	}
	values, directive := complete(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("--%s directive = %v, want NoFileComp", flag, directive)
	}
	return values
}

func TestLanguageCompletionListsTheCache(t *testing.T) {
	// The data dir is XDG on Linux but ~/Library on macOS, so point both at a
	// temp root and ask storage where that actually lands.
	dir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dir)
	t.Setenv("HOME", dir)

	dirs, err := storage.DefaultDirs()
	if err != nil {
		t.Fatalf("DefaultDirs: %v", err)
	}

	languages := filepath.Join(dirs.Data, "languages")
	if err := os.MkdirAll(languages, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	for _, name := range []string{"spanish.json", "english_1k.json", "_manifest.json", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(languages, name), []byte("{}"), 0o600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}

	got := completeFlag(t, newRootCmd(), "language")
	want := []string{"english_1k", "spanish"}
	if !slices.Equal(got, want) {
		t.Fatalf("--language suggests %q, want %q", got, want)
	}
}
