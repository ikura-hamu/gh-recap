package tui

import (
	"context"
	"slices"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/ikura-hamu/gh-recap/internal/recap"
)

func (m Model) CollectionWarnings() []string {
	return slices.Clone(m.warnings)
}

type viewerMsg struct {
	login string
	err   error
}

type tickMsg time.Time

// Model owns collection state, playback state, and the reusable Charm components.
type Model struct {
	ctx     context.Context
	options recap.Options
	api     recap.API

	login           string
	identityError   error
	events          []recap.Event
	coauthors       map[string]int
	popularPR       *recap.PopularPR
	popularPRFailed bool
	loaded          map[string]bool
	failed          map[string]bool
	warnings        []string

	page             int
	width            int
	height           int
	transitionFrame  int
	transition       pageTransition
	sparkleFrame     uint64
	paused           bool
	showDetails      bool
	playbackFinished bool
	speed            float64
	elapsed          time.Duration
	lastTick         time.Time

	loading       spinner.Model
	viewport      viewport.Model
	viewportCache *cachedViewport
}

func New(ctx context.Context, options recap.Options, api recap.API) Model {
	return Model{
		ctx:      ctx,
		options:  options,
		api:      api,
		loaded:   make(map[string]bool),
		failed:   make(map[string]bool),
		speed:    1,
		width:    80,
		height:   28,
		viewport: viewport.New(),
		loading:  spinner.New(spinner.WithSpinner(spinner.MiniDot)),
	}
}

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) Init() tea.Cmd {
	var spin tea.Cmd
	if !m.options.NoAnimation {
		spin = m.loading.Tick
	}
	return tea.Batch(tick(), spin, func() tea.Msg { return m.loadViewer() })
}

func (m Model) loadViewer() viewerMsg {
	if m.options.Demo {
		return viewerMsg{login: "you"}
	}
	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()
	login, err := recap.Viewer(ctx, m.api)
	return viewerMsg{login, err}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var command tea.Cmd
	previousTransition := m.transition.id
	switch msg := msg.(type) {
	case transitionTickMsg:
		if !m.transition.active() || msg.id != m.transition.id {
			return m, nil
		}
		m.transition.frame++
		if !m.transition.active() {
			m.transition.from = ""
			return m, nil
		}
		return m, transitionTick(m.transition.id)
	case spinner.TickMsg:
		if m.options.NoAnimation || len(m.loaded) == len(recap.Sources) || m.identityError != nil {
			return m, nil
		}
		var cmd tea.Cmd
		m.loading, cmd = m.loading.Update(msg)
		return m, cmd
	case viewerMsg:
		m.viewportCache = nil
		m.login = msg.login
		m.identityError = msg.err
		if msg.err != nil {
			return m, nil
		}
		commands := make([]tea.Cmd, 0, len(recap.Kinds))
		for _, kind := range recap.Kinds {
			commands = append(commands, func() tea.Msg { return m.loadActivity(kind) })
		}
		return m, tea.Batch(commands...)
	case recap.Result:
		m.viewportCache = nil
		m.acceptResult(msg)
		if !m.paused {
			m.elapsed = 0
		}
		if msg.Kind == recap.KindCommits {
			command = func() tea.Msg { return m.loadCoauthors(msg.Events) }
		}
	case tea.WindowSizeMsg:
		m.transition.from = ""
		wasSmall := m.tooSmall()
		m.width = msg.Width
		m.height = msg.Height
		if wasSmall || m.tooSmall() {
			m.lastTick = time.Time{}
		}
	case tea.KeyPressMsg:
		m.syncViewport()
		if cmd := m.handleKey(msg); cmd != nil {
			return m, cmd
		}
	case tickMsg:
		if !m.options.NoAnimation && !m.paused {
			m.sparkleFrame++
		}
		m.advancePlayback(time.Time(msg))
		command = tick()
	}
	m.syncViewport()
	if m.transition.active() && m.transition.id != previousTransition {
		command = tea.Batch(command, transitionTick(m.transition.id))
	}
	return m, command
}

func (m *Model) acceptResult(result recap.Result) {
	m.loaded[result.Kind] = true
	m.events = append(m.events, result.Events...)
	if result.Kind == recap.KindCoauthors {
		m.coauthors = result.Coauthors
	}
	if result.Kind == recap.KindPullRequests {
		m.popularPR = result.PopularPR
		m.popularPRFailed = result.PopularPRErr != nil
	}
	m.warnings = append(m.warnings, result.Warnings...)
	if result.Err != nil {
		m.failed[result.Kind] = true
		m.warnings = append(m.warnings, result.Kind+": "+result.Err.Error())
	}
	if result.Kind == recap.KindCommits && (result.Err != nil || len(result.Warnings) > 0) {
		m.failed[recap.KindCoauthors] = true
	}
}

func (m Model) loadCoauthors(events []recap.Event) recap.Result {
	if m.options.Demo {
		return recap.Demo(m.options, recap.KindCoauthors)
	}
	repositories := make([]recap.Repo, 0, len(events))
	for _, event := range events {
		repositories = append(repositories, event.Repo)
	}
	ctx, cancel := context.WithTimeout(m.ctx, 2*time.Minute)
	defer cancel()
	return recap.FetchCoauthors(ctx, m.api, m.options, m.login, repositories)
}

func (m Model) loadActivity(kind string) recap.Result {
	if m.options.Demo {
		return recap.Demo(m.options, kind)
	}
	ctx, cancel := context.WithTimeout(m.ctx, 2*time.Minute)
	defer cancel()
	return recap.Fetch(ctx, m.api, m.options, kind)
}
