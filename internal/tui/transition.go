package tui

import (
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// A separate clock keeps page movement smooth without speeding up playback
// or the decorative background. Each transition lasts about 360 ms.
const transitionFrames = 12

type pageTransition struct {
	from      string
	direction int
	frame     int
	id        uint64
}

func (t pageTransition) active() bool { return t.from != "" && t.frame < transitionFrames }

type transitionTickMsg struct{ id uint64 }

func transitionTick(id uint64) tea.Cmd {
	return tea.Tick(30*time.Millisecond, func(time.Time) tea.Msg {
		return transitionTickMsg{id: id}
	})
}

func (m Model) transitionBodyView() string {
	body := m.bodyView()
	if m.options.NoAnimation || !m.transition.active() {
		return body
	}
	height, width := m.bodyHeight(), m.contentWidth()
	from := transitionLines(m.transition.from, width, height)
	to := transitionLines(body, width, height)
	progress := float64(m.transition.frame) / transitionFrames
	// Ease out so the incoming page settles gently into place.
	offset := int(math.Round((1 - math.Pow(1-progress, 3)) * float64(width)))
	for row := range from {
		if m.transition.direction > 0 {
			from[row] = transitionSlice(from[row], offset, width) + transitionSlice(to[row], 0, offset)
		} else {
			from[row] = transitionSlice(to[row], width-offset, width) + transitionSlice(from[row], 0, width-offset)
		}
	}
	return strings.Join(from, "\n")
}

func transitionSlice(line string, left, right int) string {
	part := ansi.Truncate(ansi.Cut(line, left, right), right-left, "")
	return part + strings.Repeat(" ", max(0, right-left-lipgloss.Width(part)))
}

// Clip by terminal cells, preserving ANSI styles and wide Unicode glyphs.
// Padding makes every frame occupy the same rectangle, including short pages.
func transitionLines(body string, width, height int) []string {
	lines := strings.Split(body, "\n")
	result := make([]string, height)
	for row := range result {
		if row < len(lines) {
			result[row] = ansi.Truncate(lines[row], width, "")
		}
		result[row] += strings.Repeat(" ", max(0, width-lipgloss.Width(result[row])))
	}
	return result
}
