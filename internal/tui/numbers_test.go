package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ikura-hamu/gh-recap/internal/recap"
)

func TestCompactCount(t *testing.T) {
	for _, tc := range []struct {
		count int
		want  string
	}{
		{0, "0"}, {9, "9"}, {123, "123"}, {9999, "9999"},
		{10000, "10.0K"}, {12345, "12.3K"}, {99999, "99.9K"},
		{100000, "100K"}, {999999, "999K"},
		{1000000, "1.00M"}, {1234567, "1.23M"}, {9999999, "9.99M"},
		{10000000, "10.0M"}, {99999999, "99.9M"},
		{100000000, "100M"}, {999999999, "999M"},
		{1000000000, "999M+"},
	} {
		if got := compactCount(tc.count); got != tc.want {
			t.Errorf("compactCount(%d) = %q, want %q", tc.count, got, tc.want)
		}
	}
}

func TestMetricGlyphs(t *testing.T) {
	for _, r := range "0123456789KM.+" {
		glyph, ok := glyphs[r]
		if !ok {
			t.Fatalf("missing glyph %q", r)
		}
		width := len([]rune(glyph[0]))
		for _, row := range glyph {
			if len([]rune(row)) != width {
				t.Fatalf("inconsistent glyph width for %q", r)
			}
		}
		if !strings.Contains(display(string(r)), "█") {
			t.Fatalf("empty glyph %q", r)
		}
	}
	m := readyModel()
	m.height = 40
	if lipgloss.Height(m.metricView("10.0K", 40)) != 5 {
		t.Fatal("expected five-row large metric")
	}
	if got := ansi.Strip(m.metricView("10.0K", 5)); got != "10.0K" {
		t.Fatalf("narrow terminal should fall back to compact text: %q", got)
	}
}

func TestKindMetricStates(t *testing.T) {
	m := readyModel()
	m.page = pageCommits
	if content := m.content(); content.Metric == "" || content.Title != "commits" {
		t.Fatal("loaded commits should have a large metric and label")
	}
	m.failed[recap.KindCommits] = true
	if content := m.content(); content.Metric == "" || !strings.Contains(content.Title, "incomplete") {
		t.Fatal("partial data should retain its incomplete label")
	}
	delete(m.loaded, recap.KindCommits)
	if content := m.content(); content.Metric != "" || !strings.Contains(content.Title, "Loading") {
		t.Fatal("loading should not display a misleading zero metric")
	}
	m.showDetails = true
	if m.content().Metric != "" {
		t.Fatal("details should retain ordinary numeric text")
	}
}

func TestEncoreMetricsFitMinimumWideTerminal(t *testing.T) {
	m := readyModel()
	m.page = pageSummary
	m.width, m.height = wideWidth, minimumHeight
	_, view := m.scrollViewport()
	visible := ansi.Strip(view.View())
	for _, label := range []string{"commits", "pull requests", "reviews", "issues", "new", "repositories", "active days", "longest streak", "Enter to explore"} {
		if !strings.Contains(visible, label) {
			t.Errorf("missing %q in the visible ENCORE panel", label)
		}
	}
	if !strings.ContainsAny(visible, "█▀▄") || view.TotalLineCount() > view.Height() {
		t.Fatal("large ENCORE metrics should fit without scrolling")
	}
	lines := strings.Split(visible, "\n")
	for i, line := range lines[:len(lines)-1] {
		if strings.Contains(line, "commits") && strings.TrimSpace(lines[i+1]) != "" {
			t.Fatal("ENCORE should leave a blank line between metric rows")
		}
	}
}
