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
// or the decorative background. Each transition lasts about 240 ms.
const transitionFrames = 8

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
	from := transitionRows(m.transition.from, height)
	to := transitionRows(body, height)
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

// transitionSlice clips and pads each row when it is composed into the frame.
func transitionRows(body string, height int) []string {
	lines := strings.Split(body, "\n")
	if len(lines) < height {
		lines = append(lines, make([]string, height-len(lines))...)
	}
	return lines[:height]
}
