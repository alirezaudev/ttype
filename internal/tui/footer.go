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
	TermsURL   = "https://github.com/alirezaudev/ttype#readme"
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
		{"about", TermsURL},
	}
}

// The links row is centered with the theme name and version on the right, so
// the layout has to be computed once and shared by the renderer and the
// hit-test — otherwise clicks land next to what they look like they hit.
type footerLayout struct {
	links       []footerLink
	showTheme   bool
	showVersion bool
	rowWidth    int
	offset      int
	leftWidth   int
	gap         int
	total       int
}

// The row must never wrap: a wrapped footer changes the section height and
// every click then lands one row off. So drop content until it fits — links
// go from the right, then the theme name, and the version label survives
// last because it carries the update notice.
func footerLayoutFor(cfg domain.TestConfig, ver domain.VersionInfo, viewWidth int) footerLayout {
	rowWidth := max(viewWidth, 1)
	all := footerLinks()

	for _, candidate := range []struct {
		links int
		theme bool
	}{
		{len(all), true}, {3, true}, {2, true}, {1, true}, {1, false}, {0, false},
	} {
		l := measureFooter(cfg, ver, rowWidth, all[:candidate.links], candidate.theme)
		if l.total <= rowWidth {
			return l
		}
	}
	return footerLayout{rowWidth: rowWidth}
}

func measureFooter(cfg domain.TestConfig, ver domain.VersionInfo, rowWidth int, links []footerLink, showTheme bool) footerLayout {
	labels := make([]string, 0, len(links))
	for _, link := range links {
		labels = append(labels, link.label)
	}
	left := strings.Join(labels, "  ")

	right := versionLabel(ver)
	if showTheme {
		right = cfg.Theme + "  " + right
	}

	gap := 0
	if left != "" {
		gap = 4
		if len(left)+len(right)+gap < 40 {
			gap = 8
		}
	}

	leftWidth := runewidth.StringWidth(left)
	total := leftWidth + gap + runewidth.StringWidth(right)
	offset := (rowWidth - total) / 2
	if offset < 0 {
		offset = 0
	}

	return footerLayout{
		links:       links,
		showTheme:   showTheme,
		showVersion: true,
		rowWidth:    rowWidth,
		offset:      offset,
		leftWidth:   leftWidth,
		gap:         gap,
		total:       total,
	}
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
	switch {
	case ver.Installed != "":
		arrow := " → "
		if !unicodeCapable() {
			arrow = " -> "
		}
		label += arrow + ver.Installed
	case ver.UpdateAvailable:
		label += " new"
	case ver.JustUpdated:
		label += " updated"
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
	row := lipgloss.NewStyle().Width(l.rowWidth).Align(lipgloss.Center)
	if !l.showVersion {
		return row.Render("")
	}

	left := make([]string, 0, len(l.links))
	for _, link := range l.links {
		left = append(left, hyperlink(link.url, theme.Footer.Underline(true).Render(link.label)))
	}

	versionStyle := theme.Footer.Underline(true)
	if ver.UpdateAvailable || ver.Installed != "" || ver.JustUpdated {
		versionStyle = theme.Finished.Underline(true)
	}
	right := hyperlink(GitHubURL+"/releases", versionStyle.Render(versionLabel(ver)))
	if l.showTheme {
		right = theme.Footer.Render(cfg.Theme) + "  " + right
	}

	line := strings.Join(left, "  ")
	if line != "" {
		line += strings.Repeat(" ", l.gap)
	}
	return row.Render(line + right)
}

func renderFooterSection(theme Theme, cfg domain.TestConfig, ver domain.VersionInfo, width int) string {
	if cfg.Zen {
		return ""
	}
	return lipgloss.JoinVertical(
		lipgloss.Center,
		renderFooter(theme, cfg, ver, width),
		"",
		footerHelp(theme, width),
	)
}

// The hint degrades too, or it wraps where the links row no longer does.
func footerHelp(theme Theme, width int) string {
	rowWidth := max(width, 1)
	text := ""
	for _, candidate := range []string{"click footer links - u update", "u update"} {
		if runewidth.StringWidth(candidate) <= rowWidth {
			text = candidate
			break
		}
	}
	return lipgloss.NewStyle().Width(rowWidth).Align(lipgloss.Center).Render(theme.Help.Render(text))
}

// footerURLAt maps a click to a link. The links row is the third from the
// bottom, under the blank line and the hint.
func footerURLAt(x, y, viewWidth, viewHeight int, cfg domain.TestConfig, ver domain.VersionInfo) (string, bool) {
	if cfg.Zen || viewHeight < 3 || y != viewHeight-3 || x < 0 {
		return "", false
	}

	l := footerLayoutFor(cfg, ver, viewWidth)
	col := l.offset
	for _, link := range l.links {
		w := runewidth.StringWidth(link.label)
		if x >= col && x < col+w {
			return link.url, true
		}
		col += w + 2
	}

	if !l.showVersion {
		return "", false
	}
	versionStart := l.offset + l.leftWidth + l.gap
	if l.showTheme {
		versionStart += runewidth.StringWidth(cfg.Theme) + 2
	}
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
