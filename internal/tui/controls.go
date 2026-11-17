package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type keyMap struct {
	Next, Previous, Pause, Details, Back, Quit key.Binding
	Down, Up, Top, Bottom, PageDown, PageUp    key.Binding
}

var controls = keyMap{
	Next:     key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("←/→ h/l", "move")),
	Previous: key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "previous")),
	Pause:    key.NewBinding(key.WithKeys("space"), key.WithHelp("Space", "pause")),
	Details:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "details")),
	Back:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "back")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↑/↓ j/k", "scroll")),
	Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Top:      key.NewBinding(key.WithKeys("home"), key.WithHelp("Home", "top")),
	Bottom:   key.NewBinding(key.WithKeys("end"), key.WithHelp("End", "bottom")),
	PageDown: key.NewBinding(key.WithKeys("pgdown", "ctrl+f"), key.WithHelp("PgDn", "page down")),
	PageUp:   key.NewBinding(key.WithKeys("pgup", "ctrl+b"), key.WithHelp("PgUp", "page up")),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Next, k.Pause, k.Details, k.Back, k.Quit, k.Down}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Previous, k.Next, k.Pause, k.Details, k.Back, k.Quit}, {k.Up, k.Down, k.PageUp, k.PageDown, k.Top, k.Bottom}}
}

func (m Model) helpView(width int) string {
	h := help.New()
	h.SetWidth(width)
	h.ShortSeparator = " · "
	h.Styles.ShortKey = ink(m.color())
	h.Styles.ShortDesc = muted
	h.Styles.ShortSeparator = muted
	keys := controls
	if m.paused {
		keys.Pause.SetHelp("Space", "resume")
	}
	if m.showDetails {
		keys.Next.SetEnabled(false)
		keys.Details.SetHelp("Enter", "back")
	}
	return h.View(keys)
}

func (m *Model) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case key.Matches(msg, controls.Quit):
		return tea.Quit
	case m.tooSmall():
		return nil
	case key.Matches(msg, controls.Pause):
		m.paused = !m.paused
	case key.Matches(msg, controls.Next):
		if !m.showDetails {
			m.move(1)
		}
	case key.Matches(msg, controls.Previous):
		if !m.showDetails {
			m.move(-1)
		}
	case key.Matches(msg, controls.Details):
		m.transition.from = ""
		m.showDetails = !m.showDetails
		m.viewport.GotoTop()
	case key.Matches(msg, controls.Back):
		m.transition.from = ""
		m.showDetails = false
		m.viewport.GotoTop()
	case key.Matches(msg, controls.Down):
		m.viewport.ScrollDown(1)
	case key.Matches(msg, controls.Up):
		m.viewport.ScrollUp(1)
	case key.Matches(msg, controls.Top):
		m.viewport.GotoTop()
	case key.Matches(msg, controls.Bottom):
		m.viewport.GotoBottom()
	case key.Matches(msg, controls.PageDown):
		m.viewport.PageDown()
	case key.Matches(msg, controls.PageUp):
		m.viewport.PageUp()
	}
	return nil
}
