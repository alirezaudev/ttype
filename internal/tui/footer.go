package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const (
	GitHubURL  = "https://github.com/alirezaudev/ttype"
	ContactURL = "https://github.com/alirezaudev/ttype/issues"
	SupportURL = "https://github.com/sponsors/alirezaudev"
	TermsURL   = "https://github.com/alirezaudev/ttype/blob/main/LICENSE"
)

type footerLink struct {
	label string
	url   string
}

func footerLinks() []footerLink {
	return []footerLink{
		{"contact", ContactURL},
		{"support", SupportURL},
		{"github", GitHubURL},
		{"terms", TermsURL},
	}
}

// The links row is centered with the theme name and version on the right, so
// the layout has to be computed once and shared by the renderer and the
// hit-test — otherwise clicks land next to what they look like they hit.
type footerLayout struct {
	rowWidth  int
	offset    int
	leftWidth int
	gap       int
}

func footerLayoutFor(cfg domain.TestConfig, ver domain.VersionInfo, viewWidth int) footerLayout {
	rowWidth := max(viewWidth, 1)

	labels := make([]string, 0, len(footerLinks()))
	for _, link := range footerLinks() {
		labels = append(labels, link.label)
	}
	left := strings.Join(labels, "  ")
	right := cfg.Theme + "  " + versionLabel(ver)

	gap := 4
	if len(left)+len(right)+gap < 40 {
		gap = 8
	}

	leftWidth := runewidth.StringWidth(left)
	offset := (rowWidth - leftWidth - gap - runewidth.StringWidth(right)) / 2
	if offset < 0 {
		offset = 0
	}
	return footerLayout{rowWidth: rowWidth, offset: offset, leftWidth: leftWidth, gap: gap}
}

func versionLabel(ver domain.VersionInfo) string {
	local := ver.Local
	if local == "" {
		local = "dev"
	}
	label := "⎇ " + local
	if !unicodeCapable() {
		label = "v" + local
	}
	if ver.UpdateAvailable {
		label += " new"
	}
	return label
}

// OSC 8 makes the label ctrl/cmd-clickable in terminals that support it.
func hyperlink(url, text string) string {
	if url == "" || text == "" {
		return text
	}
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, text)
}

func renderFooter(theme Theme, cfg domain.TestConfig, ver domain.VersionInfo, width int) string {
	l := footerLayoutFor(cfg, ver, width)

	left := make([]string, 0, len(footerLinks()))
	for _, link := range footerLinks() {
		left = append(left, hyperlink(link.url, theme.Footer.Underline(true).Render(link.label)))
	}

	versionStyle := theme.Footer.Underline(true)
	if ver.UpdateAvailable {
		versionStyle = theme.Finished.Underline(true)
	}
	right := theme.Footer.Render(cfg.Theme) + "  " +
		hyperlink(GitHubURL+"/releases", versionStyle.Render(versionLabel(ver)))

	line := strings.Join(left, "  ") + strings.Repeat(" ", l.gap) + right
	return lipgloss.NewStyle().Width(l.rowWidth).Align(lipgloss.Center).Render(line)
}

func renderFooterSection(theme Theme, cfg domain.TestConfig, ver domain.VersionInfo, width int) string {
	if cfg.Zen {
		return ""
	}
	return lipgloss.JoinVertical(
		lipgloss.Center,
		renderFooter(theme, cfg, ver, width),
		"",
		lipgloss.NewStyle().Width(max(width, 1)).Align(lipgloss.Center).
			Render(theme.Help.Render("click footer links - u update")),
	)
}

// footerURLAt maps a click to a link. The links row is the third from the
// bottom, under the blank line and the hint.
func footerURLAt(x, y, viewWidth, viewHeight int, cfg domain.TestConfig, ver domain.VersionInfo) (string, bool) {
	if cfg.Zen || viewHeight < 3 || y != viewHeight-3 || x < 0 {
		return "", false
	}

	l := footerLayoutFor(cfg, ver, viewWidth)
	col := l.offset
	for _, link := range footerLinks() {
		w := runewidth.StringWidth(link.label)
		if x >= col && x < col+w {
			return link.url, true
		}
		col += w + 2
	}

	versionStart := l.offset + l.leftWidth + l.gap + runewidth.StringWidth(cfg.Theme) + 2
	if x >= versionStart && x < versionStart+runewidth.StringWidth(versionLabel(ver)) {
		return GitHubURL + "/releases", true
	}
	return "", false
}

func footerClickCmd(msg tea.MouseMsg, width, height int, cfg domain.TestConfig, ver domain.VersionInfo) tea.Cmd {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return nil
	}
	url, ok := footerURLAt(msg.X, msg.Y, width, height, cfg, ver)
	if !ok {
		return nil
	}
	return openURLCmd(url)
}

// Swapped in tests so a click never actually launches a browser.
var openBrowserFn = openBrowser

type browserOpenedMsg struct{ err error }

func openURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		return browserOpenedMsg{err: openBrowserFn(url)}
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

// composeWithBottomFooter centers the body in the space above the footer and
// pins the footer to the last rows.
func composeWithBottomFooter(body, footer string, width, height int) string {
	if width == 0 || height == 0 {
		return body
	}
	if footer == "" {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, body)
	}

	footerHeight := lipgloss.Height(footer)
	top := lipgloss.Place(width, max(height-footerHeight, 1), lipgloss.Center, lipgloss.Center, body)
	return top + "\n" + footer
}
