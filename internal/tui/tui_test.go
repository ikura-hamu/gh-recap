package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ikura-hamu/gh-recap/internal/recap"
)

func pressKey(s string) tea.KeyPressMsg {
	if s == "space" {
		return tea.KeyPressMsg{Code: tea.KeySpace}
	}
	if s == "enter" {
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	}
	return tea.KeyPressMsg{Code: rune(s[0]), Text: s}
}

func update(m Model, msg tea.Msg) Model {
	next, _ := m.Update(msg)
	return next.(Model)
}
func readyModel() Model {
	m := New(context.Background(), recap.Options{Year: 2025, Demo: true}, nil)
	m.login = "you"
	m.width, m.height = 120, 28
	for _, k := range recap.Sources {
		m = update(m, recap.Demo(m.options, k))
	}
	m.lastTick = time.Now()
	return m
}

func TestPlayback(t *testing.T) {
	m := readyModel()
	m = update(m, tickMsg(m.lastTick.Add(8*time.Second)))
	if m.page != pageCalendar {
		t.Fatal("autoplay did not advance")
	}
	m = update(m, pressKey("space"))
	m = update(m, tickMsg(m.lastTick.Add(20*time.Second)))
	if m.page != pageCalendar {
		t.Fatal("pause ignored")
	}
	m = update(m, pressKey("l"))
	if m.page != pageMonths || m.elapsed != 0 {
		t.Fatal("manual move should reset timer")
	}
	m = update(m, pressKey("space"))
	m = update(m, pressKey("enter"))
	m = update(m, tickMsg(m.lastTick.Add(20*time.Second)))
	if m.page != pageMonths || !m.showDetails {
		t.Fatal("detail should suspend playback")
	}
	m.showDetails = false
	m.page = pageSummary
	m.elapsed = 0
	m = update(m, tickMsg(m.lastTick.Add(8*time.Second)))
	if !m.playbackFinished || m.page != pageSummary {
		t.Fatal("final page should stop")
	}
}

func TestWaitingForData(t *testing.T) {
	m := readyModel()
	m.loaded = map[string]bool{}
	m = update(m, tickMsg(m.lastTick.Add(time.Minute)))
	if m.page != pageOverview || m.elapsed != 0 {
		t.Fatal("loading consumed playback time")
	}
}

func TestAllPagesAndSizes(t *testing.T) {
	m := readyModel()
	for _, size := range [][2]int{{160, 40}, {120, 28}, {116, 24}, {100, 35}, {80, 28}, {40, 15}, {10, 5}} {
		m.width = size[0]
		m.height = size[1]
		for p := 0; p < pageCount; p++ {
			m.page = p
			content := m.content()
			if content.Title == "" || content.Body == "" {
				t.Fatalf("empty page %d", p)
			}
			view := m.View()
			if lipgloss.Width(view.Content) > m.width || lipgloss.Height(view.Content) > m.height {
				t.Fatalf("page %d exceeds terminal %dx%d: %dx%d", p, m.width, m.height, lipgloss.Width(view.Content), lipgloss.Height(view.Content))
			}
		}
	}
	m.showDetails = true
	content := m.content()
	if !strings.Contains(content.Body, "DAILY ACTIVITY") {
		t.Fatal("missing details")
	}
}

func TestScrollStopsAtBottom(t *testing.T) {
	for _, detail := range []bool{false, true} {
		name := "highlight"
		if detail {
			name = "details"
		}
		t.Run(name, func(t *testing.T) {
			m := readyModel()
			m.paused = true
			m.showDetails = detail
			m.page = pageCoauthors
			m.width = 116
			m.height = 24
			_, m.viewport = m.scrollViewport()
			for i := 0; i < 1000 && !m.viewport.AtBottom(); i++ {
				m = update(m, tea.KeyPressMsg{Code: tea.KeyDown})
			}
			if !m.viewport.AtBottom() || m.viewport.YOffset() == 0 {
				t.Fatal("expected scrollable content to reach bottom")
			}
			bottom := m.viewport.YOffset()
			for i := 0; i < 30; i++ {
				m = update(m, tea.KeyPressMsg{Code: tea.KeyDown})
			}
			if m.viewport.YOffset() != bottom {
				t.Fatal("down at bottom accumulated hidden scroll")
			}
			before := m.View().Content
			m = update(m, tea.KeyPressMsg{Code: tea.KeyUp})
			if m.viewport.YOffset() != bottom-1 || m.View().Content == before {
				t.Fatal("one up key should immediately scroll back")
			}
		})
	}
}

func TestScrollClampsAfterResize(t *testing.T) {
	m := readyModel()
	m.paused = true
	m.showDetails = true
	m.width = 116
	m.height = 24
	m = update(m, tea.KeyPressMsg{Code: tea.KeyEnd})
	before := m.viewport.YOffset()
	m = update(m, tea.WindowSizeMsg{Width: 140, Height: 60})
	if m.viewport.YOffset() >= before || !m.viewport.AtBottom() {
		t.Fatal("resizing should clamp the stored offset to the new bottom")
	}
	bottom := m.viewport.YOffset()
	m = update(m, tea.KeyPressMsg{Code: tea.KeyUp})
	if m.viewport.YOffset() != bottom-1 {
		t.Fatal("scroll after resize should react immediately")
	}
}

func TestCoauthorsPageAndCollection(t *testing.T) {
	m := readyModel()
	delete(m.loaded, recap.KindCoauthors)
	m.coauthors = nil
	if !m.ready() {
		t.Fatal("supplementary collection should not block overview playback")
	}
	m.page = pageCoauthors
	if m.ready() || !strings.Contains(m.content().Body, "Loading") || m.content().Metric != "" {
		t.Fatal("coauthors page should wait for its own result")
	}
	// Commit collection must schedule coauthor collection, rather than relying
	// on the page being visited or requiring another viewer query.
	next, command := m.Update(recap.Demo(m.options, recap.KindCommits))
	if command == nil {
		t.Fatal("commit result did not start coauthor collection")
	}
	m = update(next.(Model), command())
	content := m.content()
	if !m.ready() || content.Metric != "5" || !strings.Contains(content.Body, "@aki") {
		t.Fatalf("coauthor result did not reach the page: %+v", content)
	}
	if strings.Index(content.Body, "@aki") > strings.Index(content.Body, "@haru") {
		t.Fatal("coauthors should be ordered by descending commit count")
	}
	m.failed[recap.KindCoauthors] = true
	if !strings.Contains(m.content().Title, "incomplete") {
		t.Fatal("partial coauthor results should be labeled")
	}
	m.coauthors = nil
	if !strings.Contains(m.content().Body, "No co-authored commits") {
		t.Fatal("missing empty coauthors state")
	}
	m.showDetails = true
	if !strings.Contains(m.content().Body, "CO-AUTHORS") {
		t.Fatal("coauthors should also be available in details")
	}
}

func TestMinimumTerminalSize(t *testing.T) {
	for _, size := range [][2]int{{115, 24}, {116, 23}, {80, 24}, {10, 5}, {116, 24}, {120, 28}} {
		m := readyModel()
		m.width, m.height = size[0], size[1]
		for _, details := range []bool{false, true} {
			m.showDetails = details
			view := m.View()
			text := ansi.Strip(view.Content)
			wantWarning := size[0] < 116 || size[1] < 24
			if strings.Contains(text, "Terminal too small") != wantWarning && size[0] >= 40 {
				t.Fatalf("wrong screen at %dx%d, details=%v", m.width, m.height, details)
			}
			if wantWarning && size[0] >= 40 && !strings.Contains(text, "116 columns × 24 rows") {
				t.Fatal("warning should explain the minimum dimensions")
			}
			if lipgloss.Width(view.Content) > m.width || lipgloss.Height(view.Content) > m.height {
				t.Fatalf("warning exceeds terminal dimensions at %dx%d", m.width, m.height)
			}
		}
	}
}

func TestSmallTerminalSuspendsPlayback(t *testing.T) {
	m := readyModel()
	now := m.lastTick
	m.elapsed = 2 * time.Second
	m = update(m, tea.WindowSizeMsg{Width: 115, Height: 24})
	m = update(m, tickMsg(now.Add(time.Minute)))
	if m.page != pageOverview || m.elapsed != 2*time.Second {
		t.Fatal("hidden recap should not advance")
	}
	m = update(m, pressKey("l"))
	m = update(m, pressKey("enter"))
	if m.page != pageOverview || m.showDetails {
		t.Fatal("hidden navigation should be ignored")
	}
	_, quit := m.Update(pressKey("q"))
	if quit == nil {
		t.Fatal("quit should still work on a small terminal")
	}
	m = update(m, tea.WindowSizeMsg{Width: 116, Height: 24})
	if strings.Contains(ansi.Strip(m.View().Content), "Terminal too small") {
		t.Fatal("resizing to the minimum should restore the recap")
	}
	m = update(m, tickMsg(now.Add(2*time.Minute)))
	if m.elapsed != 2*time.Second {
		t.Fatal("resizing should not consume time spent on the warning")
	}
	m = update(m, tickMsg(now.Add(2*time.Minute+time.Second)))
	if m.elapsed != 3*time.Second {
		t.Fatal("playback should resume after resizing")
	}
}
