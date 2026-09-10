package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Every binding in the app lives here, and the help lines under the test and
// result screens are generated from the same values so they cannot drift.
//
// One rule worth spelling out: nothing printable may be bound to an action
// that is reachable while a session is running. Terminals send no modifier
// bit with printable runes, so "shift+s" arrives as a plain 'S' and a
// shortcut is indistinguishable from typing it. Space is the input path
// itself and "?" is only dispatched before the first keystroke.

type appKeymap struct {
	Quit     key.Binding
	Back     key.Binding
	Exit     key.Binding
	Settings key.Binding
}

type testKeymap struct {
	Help       key.Binding
	Restart    key.Binding
	ToggleLive key.Binding
	DeleteWord key.Binding
	Backspace  key.Binding
	Skip       key.Binding
}

type resultsKeymap struct {
	Restart  key.Binding
	Copy     key.Binding
	Settings key.Binding
	Mode     key.Binding
	Language key.Binding
	Update   key.Binding
}

type pickerKeymap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	Confirm key.Binding
	Cancel  key.Binding
}

var appKeys = appKeymap{
	Quit: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
	Back: key.NewBinding(key.WithKeys("esc", "q", "Q"), key.WithHelp("esc/q", "back")),
	// The test screen is the root of the stack, so back leaves the program.
	Exit: key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc/ctrl+c", "quit")),
	// Reachable mid-test too, unlike the S mnemonic: a modifier chord can
	// never be mistaken for typed input.
	Settings: key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "settings")),
}

var testKeys = testKeymap{
	// Dispatched only before the first keystroke; mid-test "?" is typed input.
	Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Restart:    key.NewBinding(key.WithKeys("tab", "enter"), key.WithHelp("tab/enter", "restart")),
	ToggleLive: key.NewBinding(key.WithKeys("ctrl+o"), key.WithHelp("ctrl+o", "live stats")),
	// ctrl+h is what most terminals send for ctrl+backspace; the alt variants
	// cover option+backspace on macOS and alt+backspace on Linux.
	DeleteWord: key.NewBinding(
		key.WithKeys("ctrl+w", "ctrl+h", "alt+ctrl+h", "alt+backspace"),
		key.WithHelp("ctrl+bksp/ctrl+w", "delete word"),
	),
	Backspace: key.NewBinding(key.WithKeys("backspace"), key.WithHelp("backspace", "delete char")),
	Skip:      key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "next word")),
}

var resultsKeys = resultsKeymap{
	Restart:  key.NewBinding(key.WithKeys("tab", "enter", "r", "R"), key.WithHelp("tab/enter", "restart")),
	Copy:     key.NewBinding(key.WithKeys("C"), key.WithHelp("C", "copy result")),
	Settings: key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "settings")),
	Mode:     key.NewBinding(key.WithKeys("M"), key.WithHelp("M", "mode picker")),
	Language: key.NewBinding(key.WithKeys("L"), key.WithHelp("L", "language picker")),
	Update:   key.NewBinding(key.WithKeys("u", "U"), key.WithHelp("u", "update")),
}

var pickerKeys = pickerKeymap{
	Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/↓", "select")),
	Down:    key.NewBinding(key.WithKeys("down", "j")),
	Left:    key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/→", "adjust")),
	Right:   key.NewBinding(key.WithKeys("right", "l")),
	Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
	Cancel:  key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc/ctrl+c", "exit")),
}

func helpLine(bindings ...key.Binding) string {
	line := ""
	for _, b := range bindings {
		if line != "" {
			line += " - "
		}
		line += b.Help().Key + " " + b.Help().Desc
	}
	return line
}

func testHelpLine() string {
	return helpLine(testKeys.Help, testKeys.Restart, appKeys.Exit, testKeys.DeleteWord, testKeys.ToggleLive)
}

func resultsHelpLine() string {
	return helpLine(
		resultsKeys.Restart,
		appKeys.Quit,
		resultsKeys.Copy,
		resultsKeys.Settings,
		resultsKeys.Mode,
		resultsKeys.Language,
	)
}

func overlayRow(keyLabel, desc string) string {
	return fmt.Sprintf("  %-18s%s", keyLabel, desc)
}

func isQuitKey(msg tea.KeyMsg) bool {
	return key.Matches(msg, appKeys.Quit)
}

func isBackKey(msg tea.KeyMsg) bool {
	return key.Matches(msg, appKeys.Back)
}

// isEscKey singles esc out of Back where q and Q are typed input.
func isEscKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyEsc
}
