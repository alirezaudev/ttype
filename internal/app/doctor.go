package app

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alirezaudev/ttype/internal/storage"
	"golang.org/x/term"
)

type checkResult struct {
	name string
	ok   bool
	note string
	hint string
}

func RunDoctor(store storage.Store) error {
	checks := []checkResult{
		checkTerminal(),
		checkColor(),
		checkLocale(),
		checkClipboard(),
		checkDataDir(store.Paths()),
	}

	failed := printChecks(os.Stdout, checks)
	if failed > 0 {
		return fmt.Errorf("%d check(s) need attention", failed)
	}
	return nil
}

func printChecks(out io.Writer, checks []checkResult) int {
	failed := 0
	for _, check := range checks {
		mark := "ok  "
		if !check.ok {
			mark = "fail"
			failed++
		}
		fmt.Fprintf(out, "  [%s] %-12s %s\n", mark, check.name, check.note)
		if !check.ok && check.hint != "" {
			fmt.Fprintf(out, "         %s\n", check.hint)
		}
	}
	return failed
}

func checkTerminal() checkResult {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return checkResult{
			name: "terminal", note: "not a terminal",
			hint: "run ttype directly rather than through a pipe",
		}
	}

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return checkResult{name: "terminal", note: "size unknown", hint: err.Error()}
	}
	note := fmt.Sprintf("%dx%d", width, height)
	if width < 40 || height < 12 {
		return checkResult{
			name: "terminal", note: note,
			hint: "ttype wants at least 40x12; the layout degrades below that",
		}
	}
	return checkResult{name: "terminal", ok: true, note: note}
}

func checkColor() checkResult {
	term := os.Getenv("TERM")
	switch {
	case term == "" || term == "dumb":
		return checkResult{
			name: "color", note: "no color support",
			hint: "set TERM (for example xterm-256color)",
		}
	case os.Getenv("NO_COLOR") != "":
		return checkResult{name: "color", ok: true, note: "disabled by NO_COLOR"}
	default:
		return checkResult{name: "color", ok: true, note: term}
	}
}

func checkLocale() checkResult {
	for _, key := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		value := os.Getenv(key)
		if value == "" {
			continue
		}
		upper := strings.ToUpper(value)
		if strings.Contains(upper, "UTF-8") || strings.Contains(upper, "UTF8") {
			return checkResult{name: "locale", ok: true, note: value}
		}
		return checkResult{
			name: "locale", note: value,
			hint: "set a UTF-8 locale for the chart and big digits",
		}
	}
	return checkResult{name: "locale", ok: true, note: "unset, assuming UTF-8"}
}

func checkClipboard() checkResult {
	var candidates []string
	switch runtime.GOOS {
	case "darwin":
		candidates = []string{"pbcopy"}
	case "windows":
		candidates = []string{"clip"}
	default:
		candidates = []string{"wl-copy", "xclip", "xsel"}
	}

	for _, name := range candidates {
		if _, err := exec.LookPath(name); err == nil {
			return checkResult{name: "clipboard", ok: true, note: name}
		}
	}
	return checkResult{
		name: "clipboard", note: "no clipboard tool",
		hint: "install one of: " + strings.Join(candidates, ", "),
	}
}

func checkDataDir(dirs storage.Dirs) checkResult {
	probe := filepath.Join(dirs.Data, ".doctor")
	if err := os.WriteFile(probe, []byte("ok"), 0o600); err != nil {
		return checkResult{name: "data dir", note: dirs.Data, hint: err.Error()}
	}
	_ = os.Remove(probe)
	return checkResult{name: "data dir", ok: true, note: dirs.Data}
}
