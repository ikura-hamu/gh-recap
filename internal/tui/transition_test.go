package tui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestHorizontalTransition(t *testing.T) {
	for _, direction := range []int{1, -1} {
		m := readyModel()
		m.page = pageCalendar
		m.transition = pageTransition{
			from: strings.Repeat("X", m.contentWidth()), direction: direction, frame: 3, id: 1,
		}
		line := ansi.Strip(strings.Split(m.transitionBodyView(), "\n")[0])
		if direction > 0 && !strings.HasPrefix(line, "XXX") {
			t.Fatal("next page should enter from the right")
		}
		if direction < 0 && !strings.HasSuffix(line, "XXX") {
			t.Fatal("previous page should enter from the left")
		}
		for frame := 0; frame < transitionFrames; frame++ {
			m.transition.frame = frame
			body := m.transitionBodyView()
			if lipgloss.Width(body) != m.contentWidth() || lipgloss.Height(body) != m.bodyHeight() {
				t.Fatalf("frame %d changed body dimensions", frame)
			}
		}
	}
}

func TestTransitionLifecycle(t *testing.T) {
	m := readyModel()
	next, cmd := m.Update(pressKey("l"))
	m = next.(Model)
	if cmd == nil || !m.transition.active() || m.transition.direction != 1 {
		t.Fatal("manual navigation should schedule a transition")
	}
	oldID := m.transition.id
	before := m.transitionBodyView()
	m = update(m, pressKey("h"))
	if m.transition.from != before || m.transition.direction != -1 {
		t.Fatal("rapid navigation should start from the currently visible frame")
	}
	m = update(m, transitionTickMsg{id: oldID})
	if m.transition.frame != 0 {
		t.Fatal("stale ticks should not advance a new transition")
	}
	for range transitionFrames {
		m = update(m, transitionTickMsg{id: m.transition.id})
	}
	if m.transition.active() || m.transition.from != "" || m.transitionBodyView() != m.bodyView() {
		t.Fatal("completed transition should release the old page")
	}
	m = update(m, tickMsg(m.lastTick.Add(m.duration()+time.Second)))
	if !m.transition.active() {
		t.Fatal("autoplay should also animate page changes")
	}
	m = update(m, tea.WindowSizeMsg{Width: 140, Height: 30})
	if m.transition.active() {
		t.Fatal("resize should discard old page dimensions")
	}
	m = update(m, pressKey("l"))
	m = update(m, pressKey("enter"))
	if m.transition.active() {
		t.Fatal("opening details should finish the page transition")
	}
	m = readyModel()
	m.options.NoAnimation = true
	m = update(m, pressKey("l"))
	if m.transition.active() || m.transitionBodyView() != m.bodyView() {
		t.Fatal("no-animation should change pages immediately")
	}
}

func TestTransitionSliceWideCharacters(t *testing.T) {
	line := ink("#FF78C6").Render("あいうえお")
	for left := 0; left <= 10; left++ {
		part := transitionSlice(line, left, 10)
		if lipgloss.Width(part) != 10-left {
			t.Fatalf("slice at cell %d has incorrect width", left)
		}
	}
}
