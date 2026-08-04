package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	ThemeDefault = "default"
	ThemeMonokai = "monokai"
	ThemeDracula = "dracula"
)

var themeNames = []string{ThemeDefault, ThemeMonokai, ThemeDracula}

type Theme struct {
	Correct   lipgloss.Style
	Incorrect lipgloss.Style
	Pending   lipgloss.Style
	Cursor    lipgloss.Style
	HUD       lipgloss.Style
	HUDValue  lipgloss.Style
	HUDWPM    lipgloss.Style
	HUDRaw    lipgloss.Style
	HUDAcc    lipgloss.Style
	HUDErr    lipgloss.Style
	HUDTime   lipgloss.Style
	HUDTitle  lipgloss.Style
	HUDMode   lipgloss.Style
	Help      lipgloss.Style
	Finished  lipgloss.Style
	Footer    lipgloss.Style

	Border       lipgloss.Style
	FlashBg      lipgloss.Style
	SelectedItem lipgloss.Style
	CapsWarn     lipgloss.Style
}

func ParseThemeName(name string) (string, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return ThemeDefault, nil
	}

	for _, t := range themeNames {
		if name == t {
			return name, nil
		}
	}

	return "", fmt.Errorf("unknown theme %q (available themes: %s)", name, strings.Join(themeNames, ", "))
}

func ThemeNames() []string {
	out := make([]string, len(themeNames))
	copy(out, themeNames)
	return out
}

func ResolveTheme(name string) Theme {
	switch name {
	case ThemeMonokai:
		return monokaiTheme()
	case ThemeDracula:
		return draculaTheme()
	default:
		return defaultTheme()
	}
}

func defaultTheme() Theme {
	return Theme{
		Correct: lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		Incorrect: lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Background(lipgloss.Color("236")),
		Pending: lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		Cursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("240")).
			Underline(true),
		HUD:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")),
		HUDValue: lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
		HUDWPM:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")),
		HUDRaw:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")),
		HUDAcc:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")),
		HUDErr:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")),
		HUDTime:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3")),
		HUDTitle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")),
		HUDMode:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")),
		Help:     lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
		Finished: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3")),
		Footer:   lipgloss.NewStyle().Foreground(lipgloss.Color("243")),

		Border:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		FlashBg: lipgloss.NewStyle().Background(lipgloss.Color("236")),
		SelectedItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")),
		CapsWarn: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("3")),
	}
}

func monokaiTheme() Theme {
	return Theme{
		Correct:   lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E22E")),
		Incorrect: lipgloss.NewStyle().Foreground(lipgloss.Color("#F92672")).Background(lipgloss.Color("#3E3D32")),
		Pending:   lipgloss.NewStyle().Foreground(lipgloss.Color("#75715E")),
		Cursor:    lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")).Background(lipgloss.Color("#49483E")).Underline(true),
		HUD:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#66D9EF")),
		HUDValue:  lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")),
		HUDWPM:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#66D9EF")),
		HUDRaw:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AE81FF")),
		HUDAcc:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A6E22E")),
		HUDErr:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F92672")),
		HUDTime:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E6DB74")),
		HUDTitle:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F8F8F2")),
		HUDMode:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FD971F")),
		Help:      lipgloss.NewStyle().Foreground(lipgloss.Color("#75715E")),
		Finished:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E6DB74")),
		Footer:    lipgloss.NewStyle().Foreground(lipgloss.Color("#75715E")),

		Border:  lipgloss.NewStyle().Foreground(lipgloss.Color("#75715E")),
		FlashBg: lipgloss.NewStyle().Background(lipgloss.Color("#3E3D32")),
		SelectedItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			Background(lipgloss.Color("#49483E")),
		CapsWarn: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#272822")).
			Background(lipgloss.Color("#E6DB74")),
	}
}

func draculaTheme() Theme {
	return Theme{
		Correct:   lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")),
		Incorrect: lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Background(lipgloss.Color("#44475A")),
		Pending:   lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
		Cursor:    lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")).Background(lipgloss.Color("#44475A")).Underline(true),
		HUD:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8BE9FD")),
		HUDValue:  lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")),
		HUDWPM:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8BE9FD")),
		HUDRaw:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9")),
		HUDAcc:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50FA7B")),
		HUDErr:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5555")),
		HUDTime:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F1FA8C")),
		HUDTitle:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF79C6")),
		HUDMode:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C")),
		Help:      lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
		Finished:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C")),
		Footer:    lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),

		Border:  lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
		FlashBg: lipgloss.NewStyle().Background(lipgloss.Color("#44475A")),
		SelectedItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			Background(lipgloss.Color("#BD93F9")),
		CapsWarn: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#282A36")).
			Background(lipgloss.Color("#F1FA8C")),
	}
}

func (t Theme) HUDStatLabel(label string) lipgloss.Style {
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "wpm":
		return t.HUDWPM
	case "raw":
		return t.HUDRaw
	case "acc":
		return t.HUDAcc
	case "err":
		return t.HUDErr
	case "timed", "words", "duration", "length", "test":
		return t.HUDTime
	default:
		return t.HUD
	}
}

func DefaultTheme() Theme {
	return ResolveTheme(ThemeDefault)
}
