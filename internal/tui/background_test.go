package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

func TestScreenBackgroundFillsTerminal(t *testing.T) {
	for _, size := range [][2]int{{10, 5}, {116, 24}, {160, 50}} {
		for _, noAnimation := range []bool{false, true} {
			for _, details := range []bool{false, true} {
				t.Run(fmt.Sprintf("%dx%d/noAnimation=%v/details=%v", size[0], size[1], noAnimation, details), func(t *testing.T) {
					m := readyModel()
					m.width, m.height = size[0], size[1]
					m.options.NoAnimation, m.showDetails = noAnimation, details
					content := m.View().Content
					if m.View().BackgroundColor != lipgloss.Color(backgroundColor) {
						t.Fatal("terminal default background is not set")
					}
					if lipgloss.Width(content) != m.width || lipgloss.Height(content) != m.height {
						t.Fatalf("screen does not fill %dx%d", m.width, m.height)
					}
					screen := uv.NewScreenBuffer(m.width, m.height)
					uv.NewStyledString(content).Draw(screen, screen.Bounds())
					lines := screen.Lines
					for y, line := range lines {
						width := 0
						for _, cell := range line {
							width += cell.Width
							if y == m.height-1 || width == m.width {
								if cell.Style.Bg != nil {
									t.Fatalf("edge overrides terminal background at %d,%d", width-1, y)
								}
							}
						}
						if width != m.width {
							t.Fatalf("row %d: painted %d columns, want %d", y, width, m.width)
						}
					}
				})
			}
		}
	}
}

func TestScreenBackgroundSurvivesStyleResets(t *testing.T) {
	m := readyModel()
	m.width, m.height = 80, 28
	badge := lipgloss.NewStyle().Background(lipgloss.Color(colors[0])).Render("BADGE")
	content := badge + " " + accent.Render("styled text") + "\x1b[0m plain text"
	for frame := uint64(0); frame < 48; frame++ {
		m.sparkleFrame = frame
		view := m.screenView(m.sparkleBackground(content))
		screen := uv.NewScreenBuffer(m.width, m.height)
		uv.NewStyledString(view.Content).Draw(screen, screen.Bounds())
		lines := screen.Lines
		for y, line := range lines {
			for x, cell := range line {
				if y == 0 && x < len("BADGE") {
					if cell.Style.Bg == nil {
						t.Fatalf("frame %d: badge background reset at %d,%d", frame, x, y)
					}
					r, g, b, a := cell.Style.Bg.RGBA()
					wr, wg, wb, wa := lipgloss.Color(colors[0]).RGBA()
					if r != wr || g != wg || b != wb || a != wa {
						t.Fatalf("frame %d: wrong badge background at %d,%d", frame, x, y)
					}
				} else if cell.Style.Bg != nil {
					t.Fatalf("frame %d: cell overrides terminal background at %d,%d", frame, x, y)
				}
			}
		}
	}
}

func TestSparklesPreserveContent(t *testing.T) {
	m := readyModel()
	m.width, m.height = 80, 28
	content := lipgloss.NewStyle().Width(m.width).Height(m.height).
		Render("\n\n     " + accent.Render("日本語 · GH RECAP ✦") + "\n     " + mint.Render("━━━━ 42 commits"))
	original := strings.Split(ansi.Strip(content), "\n")
	first := m.sparkleBackground(content)
	changed := false
	for frame := uint64(0); frame < 48; frame++ {
		m.sparkleFrame = frame
		got := m.sparkleBackground(content)
		if lipgloss.Width(got) != m.width || lipgloss.Height(got) != m.height {
			t.Fatal("sparkles changed terminal dimensions")
		}
		changed = changed || got != first
		lines := strings.Split(ansi.Strip(got), "\n")
		for y, line := range original {
			for x := 0; x < m.width; x++ {
				cell := ansi.Cut(line, x, x+1)
				if strings.TrimSpace(cell) != "" && ansi.Cut(lines[y], x, x+1) != cell {
					t.Fatalf("sparkle overwrote content at %d,%d", x, y)
				}
			}
		}
	}
	if first == content || !changed {
		t.Fatal("expected visible, animated sparkles")
	}
}

func TestBackgroundWithoutAnimation(t *testing.T) {
	m := readyModel()
	m.options.NoAnimation = true
	if m.sparkleBackground("content") != "content" {
		t.Fatal("no-animation should disable sparkles")
	}
	m = update(m, tickMsg(m.lastTick.Add(time.Second)))
	if m.sparkleFrame != 0 {
		t.Fatal("no-animation should leave sparkle clock stopped")
	}
	view := m.View()
	if view.BackgroundColor != lipgloss.Color(backgroundColor) {
		t.Fatal("background color should remain with animation disabled")
	}
}
