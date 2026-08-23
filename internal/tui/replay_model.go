package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type replayKeymap struct {
	Pause   key.Binding
	Speed1  key.Binding
	Speed2  key.Binding
	Speed4  key.Binding
	Restart key.Binding
	Back    key.Binding
}

var replayKeys = replayKeymap{
	Pause:   key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "pause")),
	Speed1:  key.NewBinding(key.WithKeys("1"), key.WithHelp("1/2/4", "speed")),
	Speed2:  key.NewBinding(key.WithKeys("2")),
	Speed4:  key.NewBinding(key.WithKeys("4")),
	Restart: key.NewBinding(key.WithKeys("r", "R"), key.WithHelp("r", "restart")),
	Back:    key.NewBinding(key.WithKeys("esc", "q", "Q"), key.WithHelp("esc/q", "back")),
}

const replayTick = 50 * time.Millisecond

type replayTickMsg time.Time

// ReplayModel drives a recorded session forward on a fake clock, so the stats
// it produces match the original run no matter how fast the frames arrive.
type ReplayModel struct {
	session  *engine.Session
	clock    *engine.FakeClock
	start    time.Time
	events   []domain.ReplayEvent
	next     int
	elapsed  time.Duration
	speed    int
	paused   bool
	cfg      domain.TestConfig
	theme    Theme
	renderer TestModel
	width    int
	height   int
}

func NewReplayModel(result domain.Result, recording domain.Replay, theme Theme) (ReplayModel, error) {
	// Blind and zen would hide the very thing you came to watch.
	cfg := result.Config
	cfg.Blind = false
	cfg.Zen = false

	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	session, err := engine.NewSession(cfg, engine.StaticText(recording.Target), clock)
	if err != nil {
		return ReplayModel{}, fmt.Errorf("replay: %w", err)
	}

	return ReplayModel{
		session:  session,
		clock:    clock,
		start:    clock.Now(),
		events:   recording.Events,
		speed:    1,
		cfg:      cfg,
		theme:    theme,
		renderer: NewTestModel(session, cfg, theme, domain.VersionInfo{}),
	}, nil
}

func (m *ReplayModel) setSize(width, height int) {
	m.width, m.height = width, height
	// Height 0 keeps the test renderer from pinning its own footer; the
	// replay screen owns the layout here.
	m.renderer.setSize(width, 0)
}

func (m ReplayModel) Init() tea.Cmd { return replayTickCmd() }

func replayTickCmd() tea.Cmd {
	return tea.Tick(replayTick, func(t time.Time) tea.Msg { return replayTickMsg(t) })
}

// Returns the model, a command and whether playback is over.
func (m ReplayModel) Update(msg tea.Msg) (ReplayModel, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case replayTickMsg:
		if !m.paused {
			m.advance(replayTick * time.Duration(m.speed))
		}
		return m, replayTickCmd(), false
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, replayKeys.Back):
			return m, nil, true
		case key.Matches(msg, replayKeys.Pause):
			m.paused = !m.paused
		case key.Matches(msg, replayKeys.Speed1):
			m.speed = 1
		case key.Matches(msg, replayKeys.Speed2):
			m.speed = 2
		case key.Matches(msg, replayKeys.Speed4):
			m.speed = 4
		case key.Matches(msg, replayKeys.Restart):
			m.next = 0
			m.elapsed = 0
			m.paused = false
			m.clock = engine.NewFakeClock(m.start)
			session, err := engine.NewSession(m.cfg, engine.StaticText(m.session.Target()), m.clock)
			if err == nil {
				m.session = session
				m.renderer = NewTestModel(session, m.cfg, m.theme, domain.VersionInfo{})
				m.renderer.setSize(m.width, 0)
			}
		}
	}
	return m, nil, false
}

// advance feeds every event whose offset has come due, setting the clock to
// the event's exact offset first so the session sees the original timings.
func (m *ReplayModel) advance(step time.Duration) {
	m.elapsed += step
	for m.next < len(m.events) && m.events[m.next].Offset <= m.elapsed {
		event := m.events[m.next]
		m.clock.Advance(m.start.Add(event.Offset).Sub(m.clock.Now()))
		switch event.Kind {
		case domain.ReplayRune:
			m.session.InputRune(event.Rune)
		case domain.ReplayBackspace:
			m.session.Backspace()
		case domain.ReplayDeleteWord:
			m.session.DeleteWord()
		}
		m.next++
	}
	m.clock.Advance(m.start.Add(m.elapsed).Sub(m.clock.Now()))
	m.session.Tick()
}

func (m ReplayModel) done() bool {
	return m.next >= len(m.events)
}

func (m ReplayModel) View() string {
	status := fmt.Sprintf("replay · %dx", m.speed)
	switch {
	case m.paused:
		status += " · paused"
	case m.done():
		status += " · end"
	}

	body := strings.Join([]string{
		m.theme.HUDTitle.Render(status),
		"",
		m.renderer.View(),
		"",
		m.theme.Help.Render(helpLine(replayKeys.Pause, replayKeys.Speed1, replayKeys.Restart, replayKeys.Back)),
	}, "\n")

	if m.width == 0 || m.height == 0 {
		return body
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}
