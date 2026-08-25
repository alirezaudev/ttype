package app

import (
	"bytes"
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
