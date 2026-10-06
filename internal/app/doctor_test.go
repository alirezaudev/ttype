package app

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestCheckLocaleWantsUTF8(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_CTYPE", "")
	t.Setenv("LANG", "en_US.UTF-8")
	if got := checkLocale(); !got.ok {
		t.Fatalf("locale check = %+v, want ok", got)
	}

	t.Setenv("LANG", "en_US.ISO-8859-1")
	got := checkLocale()
	if got.ok || got.hint == "" {
		t.Fatalf("locale check = %+v, want a failure with a hint", got)
	}
}

func TestCheckColorFollowsTerm(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "dumb")
	if got := checkColor(); got.ok {
		t.Fatalf("color check = %+v, want a failure on a dumb terminal", got)
	}

	t.Setenv("TERM", "xterm-256color")
	if got := checkColor(); !got.ok {
		t.Fatalf("color check = %+v, want ok", got)
	}
}

func TestPrintChecksCountsFailures(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	failed := printChecks(&buf, []checkResult{
		{name: "one", ok: true, note: "fine"},
		{name: "two", note: "broken", hint: "try this"},
	})

	if failed != 1 {
		t.Fatalf("failed = %d, want 1", failed)
	}
	out := buf.String()
	if !strings.Contains(out, "[ok  ] one") || !strings.Contains(out, "[fail] two") {
		t.Fatalf("output = %q", out)
	}
	if !strings.Contains(out, "try this") {
		t.Fatalf("hint missing from output: %q", out)
	}
}

func TestClipboardCheckFallsBackToOSC52(t *testing.T) {
	t.Parallel()

	noTools := func(string) (string, error) { return "", errors.New("not found") }
	env := func(vars map[string]string) func(string) string {
		return func(key string) string { return vars[key] }
	}

	if got := clipboardCheck(env(nil), noTools); !got.ok || !strings.Contains(got.note, "OSC 52") {
		t.Errorf("no tools: %+v, want ok through OSC 52", got)
	}
	if got := clipboardCheck(env(map[string]string{"TMUX": "x"}), noTools); !strings.Contains(got.note, "set-clipboard") {
		t.Errorf("tmux: %+v, want the set-clipboard hint", got)
	}
	tool := func(name string) (string, error) { return "/usr/bin/" + name, nil }
	if got := clipboardCheck(env(map[string]string{"SSH_CONNECTION": "1 2 3 4"}), tool); !strings.Contains(got.note, "OSC 52") {
		t.Errorf("over ssh: %+v, want OSC 52 even with a tool installed", got)
	}
}
