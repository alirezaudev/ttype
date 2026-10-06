package tui

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/aymanbagabas/go-osc52/v2"
	tea "github.com/charmbracelet/bubbletea"
)

func FormatShareLine(result domain.Result) string {
	length := fmt.Sprintf("%ds", result.Config.Duration.Seconds())
	if result.Config.IsWordsMode() {
		length = fmt.Sprintf("%dw", result.Config.WordCount)
	}
	return fmt.Sprintf("ttype — %.0f WPM / %.1f%% accuracy / %s", result.WPM, result.Accuracy, length)
}

// copyToClipboard tries the local tools, then asks the terminal with OSC 52.
// Over SSH the tools would copy on the server, so it goes straight to OSC 52.
func copyToClipboard(text string, terminal io.Writer) error {
	if os.Getenv("SSH_CONNECTION") == "" && os.Getenv("SSH_TTY") == "" {
		if err := copyWithTool(text); err == nil {
			return nil
		}
	}
	if terminal == nil {
		return errors.New("no terminal to copy through")
	}
	return writeOSC52(terminal, text, os.Getenv)
}

func writeOSC52(w io.Writer, text string, getenv func(string) string) error {
	seq := osc52.New(text)
	switch {
	case getenv("TMUX") != "":
		seq = seq.Tmux()
	case getenv("STY") != "":
		seq = seq.Screen()
	}
	_, err := seq.WriteTo(w)
	return err
}

func copyWithTool(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		}
	default:
		return fmt.Errorf("clipboard not supported on %s", runtime.GOOS)
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

var copyToClipboardFn = copyToClipboard

type clipboardCopiedMsg struct{ err error }

func copyResultCmd(result domain.Result, terminal io.Writer) tea.Cmd {
	return func() tea.Msg {
		return clipboardCopiedMsg{err: copyToClipboardFn(FormatShareLine(result), terminal)}
	}
}
