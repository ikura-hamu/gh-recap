package tui

import (
	"context"
	"fmt"
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
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
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
	delete(m.loaded, recap.KindCoauthors)
	m = update(m, tickMsg(m.lastTick.Add(4*time.Second)))
	m = update(m, recap.Demo(m.options, recap.KindCoauthors))
	m = update(m, tickMsg(m.lastTick.Add(2*time.Second)))
	if m.page != pageCalendar {
		t.Fatal("late co-authors should not delay autoplay")
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
	m = update(m, tickMsg(m.lastTick.Add(6*time.Second)))
	if !m.playbackFinished || m.page != pageSummary {
		t.Fatal("final page should stop")
	}
}

func TestVideoCreditsAppearAfterPlayback(t *testing.T) {
	m := readyModel()
	m.page = pageSummary
	m.options.NoAnimation = true
	m.options.Recording = true
	if strings.Contains(m.View().Content, videoInstallCommand) {
		t.Fatal("credits should wait for the final recap page to finish")
	}
	m = update(m, tickMsg(m.lastTick.Add(pageDuration)))
	screen := ansi.Strip(m.View().Content)
	if !strings.Contains(screen, videoCredit) || !strings.Contains(screen, videoInstallCommand) {
		t.Fatal("recording should end with the credit and install command")
	}
	m.options.Recording = false
	if strings.Contains(m.View().Content, videoInstallCommand) {
		t.Fatal("interactive playback should stay on the recap summary")
	}
}

func TestOverviewShowsYearBelowBanner(t *testing.T) {
	m := readyModel()
	content := m.content()
	if content.Metric != "2025" || content.Title != "" || !strings.Contains(content.Body, contributions) {
		t.Fatalf("overview should show the year below RECAP and contributions in the body: %+v", content)
	}
	m.loaded = nil
	if content := m.content(); content.Metric != "2025" || content.Title != "" {
		t.Fatalf("overview should show the year during collection: %+v", content)
	}
}

func TestPopularPRPage(t *testing.T) {
	m := readyModel()
	m.page = pagePopularPR
	pr := m.popularPR
	if pr == nil {
		t.Fatal("demo should provide a popular PR")
	}
	body := m.content().Body
	for _, text := range []string{pr.Repository, pr.Title, pr.Date, pr.Description, pr.Comments[0].Body, "@ren"} {
		if !strings.Contains(body, text) {
			t.Errorf("missing %q in popular PR page", text)
		}
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

func TestCommitWarningKeepsRecordingStatusAndLabelsPartialCounts(t *testing.T) {
	m := readyModel()
	m.options.Recording = true
	m.acceptResult(recap.Result{Kind: recap.KindCommits, Warnings: []string{"commit repository limit reached"}})
	if !m.failed[recap.KindCommits] || !m.failed[recap.KindCoauthors] {
		t.Fatal("commit truncation must mark commits and co-authors incomplete")
	}
	count := recap.Summarize(m.events).Kinds[recap.KindCommits]
	if report := m.markdownDetails(); !strings.Contains(report, fmt.Sprintf("| commits | %d recorded · incomplete |", count)) ||
		!strings.Contains(report, "commit repository limit reached") {
		t.Fatal("Markdown must include the partial commit count and the warning")
	}
	for _, state := range []string{playing, finished} {
		status := m.statusView()
		if !strings.Contains(status, state+" · 1×") || !strings.Contains(status, statusIncomplete) {
			t.Errorf("VHS should recognize the playback state even with warnings: %q", status)
		}
		m.playbackFinished = true
	}
}

func TestPlaybackSpeedKeys(t *testing.T) {
	m := readyModel()
	m = update(m, tickMsg(m.lastTick.Add(2*time.Second)))
	for _, step := range []struct {
		key   string
		speed float64
	}{
		{"j", 0.5}, {"j", 0.5}, {"k", 1}, {"k", 2}, {"k", 4}, {"k", 4}, {"j", 2},
	} {
		m = update(m, pressKey(step.key))
		if m.speed != step.speed || !strings.Contains(m.statusView(), fmt.Sprintf(" · %g× · ", step.speed)) {
			t.Fatalf("after %q, want %g× in model and status, got %g×", step.key, step.speed, m.speed)
		}
		if m.page != pageOverview || m.elapsed != 2*time.Second {
			t.Fatal("changing speed should preserve the page and its progress")
		}
	}
	m = update(m, tickMsg(m.lastTick.Add(2*time.Second)))
	if m.page != pageCalendar || m.speed != 2 {
		t.Fatal("at 2×, two real seconds should finish the remaining four seconds")
	}

	m = update(m, pressKey("enter"))
	m = update(m, pressKey("j"))
	if m.viewport.YOffset() != 1 || m.speed != 2 {
		t.Fatal("j should scroll down in details while preserving speed")
	}
	m = update(m, pressKey("k"))
	if m.viewport.YOffset() != 0 || m.speed != 2 {
		t.Fatal("k should scroll up in details while preserving speed")
	}
}

func TestSpaceFreezesVisiblePlayback(t *testing.T) {
	m := update(readyModel(), pressKey("l"))
	m = update(m, pressKey("space"))
	if m.transition.active() || !strings.Contains(ansi.Strip(m.statusView()), "PAUSED") {
		t.Fatal("space should settle the slide and show PAUSED")
	}
	paused := m.View().Content
	repeat := pressKey("space")
	repeat.IsRepeat = true
	m = update(m, repeat)
	m = update(m, tickMsg(m.lastTick.Add(20*time.Second)))
	if m.View().Content != paused {
		t.Fatal("paused playback and background should stay still")
	}
	m = update(m, pressKey("space"))
	m = update(m, tickMsg(m.lastTick.Add(time.Second)))
	if m.paused || m.elapsed != time.Second || m.sparkleFrame != 1 {
		t.Fatal("space should resume playback and background motion")
	}
}

func TestTopRanksLeadEachRow(t *testing.T) {
	m := readyModel()
	m.page = pageRepositories
	m.width, m.height = wideWidth, minimumHeight
	_, view := m.scrollViewport()
	for _, rank := range []string{"#1 ", "#2 ", "#3 "} {
		if !strings.Contains("\n"+ansi.Strip(view.View()), "\n"+rank) {
			t.Errorf("rank %s is not at the start of a visible row", strings.TrimSpace(rank))
		}
	}
}

func TestAllPagesAndSizes(t *testing.T) {
	m := readyModel()
	for _, size := range [][2]int{{minimumWidth, minimumHeight}, {wideWidth, minimumHeight}, {160, 40}, {120, 28}, {116, 24}, {100, 35}, {80, 28}, {40, 15}, {10, 5}} {
		m.width = size[0]
		m.height = size[1]
		for p := 0; p < pageCount; p++ {
			m.page = p
			content := m.content()
			if (content.Title == "" && content.Metric == "") || content.Body == "" {
				t.Fatalf("empty page %d", p)
			}
			view := m.View()
			if lipgloss.Width(view.Content) > m.width || lipgloss.Height(view.Content) > m.height {
				t.Fatalf("page %d exceeds terminal %dx%d: %dx%d", p, m.width, m.height, lipgloss.Width(view.Content), lipgloss.Height(view.Content))
			}
			if !m.tooSmall() && !strings.Contains(ansi.Strip(view.Content), "sources ·") {
				t.Errorf("page %d hides the footer at %dx%d", p, m.width, m.height)
			}
			if m.width == minimumWidth && m.height == minimumHeight {
				if m.wide() || m.scrollLayout().ViewportHeight < compactDisplayHeight {
					t.Errorf("page %d lacks a usable stacked viewport at minimum size", p)
				}
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
	for _, size := range [][2]int{{minimumWidth - 1, minimumHeight}, {minimumWidth, minimumHeight - 1}, {minimumWidth, minimumHeight}, {wideWidth - 1, minimumHeight}, {wideWidth, minimumHeight}, {10, 5}} {
		m := readyModel()
		m.width, m.height = size[0], size[1]
		for _, details := range []bool{false, true} {
			m.showDetails = details
			view := m.View()
			text := ansi.Strip(view.Content)
			wantWarning := size[0] < minimumWidth || size[1] < minimumHeight
			if strings.Contains(text, "Terminal too small") != wantWarning && size[0] >= 40 {
				t.Fatalf("wrong screen at %dx%d, details=%v", m.width, m.height, details)
			}
			if wantWarning && size[0] >= 40 && !strings.Contains(text, fmt.Sprintf("%d columns × %d rows", minimumWidth, minimumHeight)) {
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
	m = update(m, tea.WindowSizeMsg{Width: minimumWidth - 1, Height: minimumHeight})
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
	m = update(m, tea.WindowSizeMsg{Width: minimumWidth, Height: minimumHeight})
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
