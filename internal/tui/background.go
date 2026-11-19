package tui

import (
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Keep a full cell of clearance around content, including wide characters
// and chart bars, regardless of which animation is playing.
func (m Model) sparkleBackground(content string) string {
	if m.options.NoAnimation || m.width < 1 || m.height < 1 {
		return content
	}
	content = lipgloss.NewStyle().Width(m.width).Height(m.height).Render(content)
	lines := strings.Split(content, "\n")
	plain := make([]string, len(lines))
	for y, line := range lines {
		plain[y] = ansi.Strip(line)
	}
	effect := m.backgroundEffect()
	for y := 1; y < len(lines)-1; y++ {
		var row strings.Builder
		start := 0
		for x := 1; x < m.width-1; x++ {
			glyph, color := effect.cell(x, y)
			if glyph == "" {
				continue
			}
			clear := true
			for dy := -1; dy <= 1; dy++ {
				if ansi.Cut(plain[y+dy], x-1, x+2) != "   " {
					clear = false
					break
				}
			}
			if !clear {
				continue
			}
			row.WriteString(ansi.Cut(lines[y], start, x))
			row.WriteString(ink(color).Render(glyph))
			start = x + 1
		}
		if start > 0 {
			row.WriteString(ansi.Cut(lines[y], start, m.width))
			lines[y] = row.String()
		}
	}
	return strings.Join(lines, "\n")
}

type backgroundEffect struct {
	page, width, height int
	frame               int
	dim, bright         string
}

func (m Model) backgroundEffect() backgroundEffect {
	palette := [pageCount][2]string{
		pageOverview:     {"#584266", "#B77EB2"},
		pageCalendar:     {"#345562", "#69ADB7"},
		pageMonths:       {"#52436D", "#A087C8"},
		pageStreak:       {"#665239", "#CEAA65"},
		pageCommits:      {"#315647", "#6BA789"},
		pageCoauthors:    {"#66473F", "#C58C79"},
		pagePullRequests: {"#3F506F", "#7C9DD4"},
		pageReviews:      {"#614466", "#B989BE"},
		pageRepositories: {"#4F593D", "#A6B46C"},
		pageLanguages:    {"#574264", "#AD83C3"},
		pageSummary:      {"#61455D", "#D694BA"},
	}
	shades := palette[m.page]
	return backgroundEffect{m.page, m.width, m.height, int(m.sparkleFrame % 100000), shades[0], shades[1]}
}

func (e backgroundEffect) cell(x, y int) (string, string) {
	t := float64(e.frame) * 0.1
	switch e.page {
	case pageOverview: // Aurora ribbons bend and drift across the screen.
		wave := float64(e.height)*(0.52+0.26*math.Sin(float64(x)*0.045+t*0.55)) +
			2*math.Sin(float64(x)*0.09-t*0.4)
		distance := math.Abs(float64(y) - wave)
		if distance < 0.75 {
			return "▒", e.bright
		}
		if distance < 2.8 {
			return "░", e.dim
		}
	case pageCalendar: // A scan crosses a grid of activity cells.
		if x%5 == 0 && y%2 == 0 {
			scan := (e.frame*2 + x) % 45
			switch {
			case scan < 5:
				return "▪", e.bright
			case scan < 16:
				return "▪", e.dim
			default:
				return "·", e.dim
			}
		}
	case pageMonths: // Two sound waves travel in opposite directions.
		center := float64(e.height) * 0.5
		amplitude := center * 0.68
		if math.Abs(float64(y)-center-amplitude*math.Sin(float64(x)*0.14-t*2)) < 0.65 {
			return "•", e.bright
		}
		if math.Abs(float64(y)-center-amplitude*math.Sin(float64(x)*0.09+t*1.4)) < 0.65 {
			return "·", e.dim
		}
	case pageStreak: // Sparks race along diagonal lanes with fading tails.
		if (x+y*2)%11 == 0 {
			tail := (e.frame*3 + y*2 - x%e.width + e.width) % 34
			if tail < 4 {
				return "◆", e.bright
			}
			if tail < 16 {
				return "╱", e.dim
			}
		}
	case pageCommits: // Drops slide down staggered columns.
		if x%7 == 0 {
			column := x / 7
			period := e.height + 9
			head := (e.frame*(1+column%2) + column*17) % period
			switch head - y {
			case 0:
				return "▪", e.bright
			case 1, 2, 3:
				return "│", e.dim
			case 4, 5:
				return "·", e.dim
			}
		}
	case pageCoauthors: // Paired pulses spread out from opposite sides.
		left := math.Hypot(float64(x-e.width/4)*0.45, float64(y-e.height/2))
		right := math.Hypot(float64(x-3*e.width/4)*0.45, float64(y-e.height/2))
		a := math.Mod(left+14-math.Mod(t*3, 14), 14)
		b := math.Mod(right+14-math.Mod(t*3+6, 14), 14)
		if a < 1.2 && b < 1.2 {
			return "✦", e.bright
		}
		if a < 0.85 || b < 0.85 {
			return "·", e.dim
		}
	case pagePullRequests: // Chevrons move toward the center from both sides.
		distance := int(math.Abs(float64(x-e.width/2))) + int(math.Abs(float64(y-e.height/2)))*2
		phase := (distance + e.frame*2) % 24
		if phase < 3 {
			if x < e.width/2 {
				return "›", e.bright
			}
			return "‹", e.bright
		}
		if phase < 6 {
			return "·", e.dim
		}
	case pageReviews: // A horizontal beam sweeps down and back up.
		travel := e.height - 1
		beam := travel - int(math.Abs(float64(e.frame%(travel*2)-travel)))
		if x%3 == 0 && y == beam {
			return "━", e.bright
		}
		if x%4 == 0 && (y == beam-1 || y == beam+1) {
			return "·", e.dim
		}
	case pageRepositories: // Towers rise and fall like a skyline.
		if x%8 < 3 {
			column := x / 8
			height := float64(e.height) * (0.28 + 0.23*math.Sin(float64(column)*1.8+t))
			top := e.height - 1 - int(height)
			if y == top {
				return "▄", e.bright
			}
			if y > top {
				return "░", e.dim
			}
		}
	case pageLanguages: // Each row scrolls a different sequence of symbols.
		if y%3 == 1 {
			offset := x + (e.frame*(1+y%2)+y*13)%e.width
			symbols := [...]string{"<", ">", "{", "}", "+", "*"}
			if offset%7 == 0 {
				return symbols[(offset/7+y)%len(symbols)], e.bright
			}
			if offset%7 == 2 {
				return "·", e.dim
			}
		}
	case pageSummary: // Rotating rays pulse outward like fireworks.
		dx := float64(x-e.width/2) * 0.45
		dy := float64(y - e.height/2)
		radius := math.Hypot(dx, dy)
		ray := math.Cos(math.Atan2(dy, dx)*10 - t)
		phase := math.Mod(radius+16-math.Mod(t*5, 16), 16)
		if ray > 0.84 && phase < 2 {
			return "✦", e.bright
		}
		if ray > 0.65 && phase < 6 {
			return "·", e.dim
		}
	}
	return "", ""
}
