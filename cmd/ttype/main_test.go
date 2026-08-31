package main

import (
	"io"
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/spf13/cobra"
)

type fakeStore struct {
	settings domain.Settings
}

func (f fakeStore) LoadSettings() (domain.Settings, error) {
	return f.settings, nil
}

func TestResolveTestConfig(t *testing.T) {
	t.Parallel()

	store := fakeStore{settings: domain.Settings{
		DefaultDuration: domain.Duration30,
		DefaultWidth:    80,
		Theme:           "dracula",
		Language:        "spanish",
	}}

	cmd := &cobra.Command{}
	flags := &testCLIFlags{}
	bindTestFlags(cmd, flags)

	saved, err := resolveTestConfig(cmd, store, flags)
	if err != nil {
		t.Fatalf("resolveTestConfig: %v", err)
	}
	want := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration30, Width: 80, Theme: "dracula", Language: "spanish", TextMode: domain.TextModeWords}
	if saved != want {
		t.Fatalf("cfg = %+v, want %+v", saved, want)
	}

	if err := cmd.Flags().Set("words", "25"); err != nil {
		t.Fatalf("set words: %v", err)
	}
	if err := cmd.Flags().Set("language", "french"); err != nil {
		t.Fatalf("set language: %v", err)
	}

	overridden, err := resolveTestConfig(cmd, store, flags)
	if err != nil {
		t.Fatalf("resolveTestConfig: %v", err)
	}
	if overridden.Kind != domain.TestKindWords || overridden.WordCount != 25 {
		t.Fatalf("cfg = %+v, want a 25 word test", overridden)
	}
	if overridden.Language != "french" {
		t.Fatalf("language = %q, want the flag to win", overridden.Language)
	}
	if overridden.Theme != "dracula" {
		t.Fatalf("theme = %q, want the saved setting to survive", overridden.Theme)
	}
}

func TestOutputRejectsUnknownFormat(t *testing.T) {
	t.Parallel()

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--output", "yaml"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	if err := cmd.Execute(); err == nil {
		t.Fatal("an unknown output format should be rejected")
	}
}

func TestAllowSkipIsDeprecatedAndHidden(t *testing.T) {
	t.Parallel()

	flag := newRootCmd().Flags().Lookup("allow-skip")
	if flag == nil {
		t.Fatal("--allow-skip should still parse for old scripts")
	}
	if flag.Deprecated == "" || !flag.Hidden {
		t.Fatalf("--allow-skip should be deprecated and hidden, got %+v", flag)
	}
}
