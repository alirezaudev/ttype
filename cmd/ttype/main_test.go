package main

import (
	"io"
	"runtime/debug"
	"strings"
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

func TestResolveVersion(t *testing.T) {
	t.Parallel()

	buildInfo := func(version string, ok bool) func() (*debug.BuildInfo, bool) {
		return func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{Main: debug.Module{Version: version}}, ok
		}
	}

	tests := []struct {
		name    string
		stamped string
		read    func() (*debug.BuildInfo, bool)
		want    string
	}{
		{"release ldflags win", "1.4.0", buildInfo("v9.9.9", true), "1.4.0"},
		{"go install of a tag", "dev", buildInfo("v1.2.0", true), "1.2.0"},
		{"local build", "dev", buildInfo("(devel)", true), "dev"},
		{"pseudo-version", "dev", buildInfo("v1.0.1-0.20260913120000-bd2b72b1c3d4+dirty", true), "dev"},
		{"no build info", "dev", buildInfo("", false), "dev"},
	}

	for _, test := range tests {
		if got := resolveVersion(test.stamped, test.read); got != test.want {
			t.Errorf("%s: resolveVersion(%q) = %q, want %q", test.name, test.stamped, got, test.want)
		}
	}
}

func TestCustomTextRefusesFlagsThatGenerateText(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		{"--text", "hello", "--mode", "go"},
		{"--text", "hello", "--language", "spanish"},
		{"--text", "hello", "--punctuation"},
		{"--text", "hello", "--numbers"},
	} {
		cmd := &cobra.Command{}
		flags := &testCLIFlags{}
		bindTestFlags(cmd, flags)
		if err := cmd.ParseFlags(args); err != nil {
			t.Fatalf("ParseFlags(%v): %v", args, err)
		}
		if _, err := readCustomText(cmd, flags, false); err == nil || !strings.Contains(err.Error(), args[2]) {
			t.Errorf("%v: err = %v, want it to name %s", args, err, args[2])
		}
	}

	cmd := &cobra.Command{}
	flags := &testCLIFlags{}
	bindTestFlags(cmd, flags)
	if err := cmd.ParseFlags([]string{"--text", "hello  world", "--blind", "--words", "1"}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	if text, err := readCustomText(cmd, flags, false); err != nil || text != "hello world" {
		t.Fatalf("readCustomText = %q, %v; want the text, with --blind and --words allowed", text, err)
	}
}
