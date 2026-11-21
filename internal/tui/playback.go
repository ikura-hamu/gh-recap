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
	pagePopularPR
	pageReviews
	pageRepositories
	pageLanguages
	pageSummary
	pageCount
)

const pageDuration = 6 * time.Second

type pageDefinition struct {
	Banner string
	// An empty source means this page combines all activity sources.
	Source string
}

var pageDefinitions = [pageCount]pageDefinition{
	pageOverview:     {Banner: "RECAP"},
	pageCalendar:     {Banner: "ACTIVITY"},
	pageMonths:       {Banner: "RHYTHM"},
	pageStreak:       {Banner: "STREAK"},
	pageCommits:      {Banner: "COMMITS", Source: recap.KindCommits},
	pageCoauthors:    {Banner: "COAUTHORS", Source: recap.KindCoauthors},
	pagePullRequests: {Banner: "PULL REQS", Source: recap.KindPullRequests},
	pagePopularPR:    {Banner: "POPULAR PR", Source: recap.KindPullRequests},
	pageReviews:      {Banner: "REVIEWS", Source: recap.KindReviews},
	pageRepositories: {Banner: "REPOS"},
	pageLanguages:    {Banner: "LANGUAGES"},
	pageSummary:      {Banner: "ENCORE"},
}

func (m *Model) changeSpeed(factor float64) {
	m.speed = min(4, max(0.5, m.speed*factor))
}

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
	if !m.options.NoAnimation && !m.paused && !m.tooSmall() && m.login != "" && m.identityError == nil {
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
		m.elapsed += time.Duration(float64(now.Sub(m.lastTick)) * m.speed)
		if m.elapsed >= pageDuration {
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
