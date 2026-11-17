package tui

import (
	"time"

	"github.com/ikura-hamu/gh-recap/internal/recap"
)

const (
	pageOverview = iota
	pageCalendar
	pageMonths
	pageStreak
	pageCommits
	pageCoauthors
	pagePullRequests
	pageReviews
	pageRepositories
	pageLanguages
	pageSummary
	pageCount
)

type pageDefinition struct {
	Banner   string
	Duration time.Duration
	// An empty source means this page combines all activity sources.
	Source string
}

var pageDefinitions = [pageCount]pageDefinition{
	pageOverview:     {Banner: "RECAP", Duration: 8 * time.Second},
	pageCalendar:     {Banner: "ACTIVITY", Duration: 12 * time.Second},
	pageMonths:       {Banner: "RHYTHM", Duration: 8 * time.Second},
	pageStreak:       {Banner: "STREAK", Duration: 8 * time.Second},
	pageCommits:      {Banner: "COMMITS", Duration: 8 * time.Second, Source: recap.KindCommits},
	pageCoauthors:    {Banner: "COAUTHORS", Duration: 12 * time.Second, Source: recap.KindCoauthors},
	pagePullRequests: {Banner: "PULL REQS", Duration: 8 * time.Second, Source: recap.KindPullRequests},
	pageReviews:      {Banner: "REVIEWS", Duration: 8 * time.Second, Source: recap.KindReviews},
	pageRepositories: {Banner: "REPOS", Duration: 12 * time.Second},
	pageLanguages:    {Banner: "LANGUAGES", Duration: 12 * time.Second},
	pageSummary:      {Banner: "ENCORE", Duration: 8 * time.Second},
}

func (m Model) duration() time.Duration { return pageDefinitions[m.page].Duration }

func (m Model) ready() bool {
	source := pageDefinitions[m.page].Source
	if source != "" {
		return m.loaded[source]
	}
	for _, kind := range recap.Kinds {
		if !m.loaded[kind] {
			return false
		}
	}
	return true
}

func (m *Model) move(delta int) {
	next := m.page + delta
	if next < 0 || next >= pageCount {
		return
	}
	if !m.options.NoAnimation && !m.tooSmall() && m.login != "" && m.identityError == nil {
		from := m.transitionBodyView()
		m.transition = pageTransition{from: from, direction: delta, id: m.transition.id + 1}
	}
	m.page = next
	m.elapsed = 0
	m.viewport.GotoTop()
	m.playbackFinished = false
	m.transitionFrame = 0
}

func (m *Model) advancePlayback(now time.Time) {
	canAdvance := !m.tooSmall() && !m.lastTick.IsZero() && !m.paused && !m.showDetails && !m.playbackFinished && m.ready() && m.identityError == nil
	if canAdvance {
		m.elapsed += now.Sub(m.lastTick)
		if m.elapsed >= m.duration() {
			if m.page == pageSummary {
				m.playbackFinished = true
			} else {
				m.move(1)
			}
		}
	}
	m.lastTick = now
	m.transitionFrame++
}
