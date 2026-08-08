package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

func FormatShareLine(result domain.Result) string {
	length := fmt.Sprintf("%ds", result.Config.Duration.Seconds())
	if result.Config.IsWordsMode() {
		length = fmt.Sprintf("%dw", result.Config.WordCount)
	}
	return fmt.Sprintf("ttype — %.0f WPM / %.1f%% accuracy / %s", result.WPM, result.Accuracy, length)
}

func copyToClipboard(text string) error {
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

func copyResultCmd(result domain.Result) tea.Cmd {
	return func() tea.Msg {
		return clipboardCopiedMsg{err: copyToClipboardFn(FormatShareLine(result))}
	}
}
