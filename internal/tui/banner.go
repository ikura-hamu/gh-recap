package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	displayHeight        = 5
	compactDisplayHeight = (displayHeight + 1) / 2
)

var glyphs = map[rune][displayHeight]string{
	'0': {"███", "█ █", "█ █", "█ █", "███"},
	'1': {" █ ", "██ ", " █ ", " █ ", "███"},
	'2': {"███", "  █", "███", "█  ", "███"},
	'3': {"███", "  █", "███", "  █", "███"},
	'4': {"█ █", "█ █", "███", "  █", "  █"},
	'5': {"███", "█  ", "███", "  █", "███"},
	'6': {"███", "█  ", "███", "█ █", "███"},
	'7': {"███", "  █", "  █", "  █", "  █"},
	'8': {"███", "█ █", "███", "█ █", "███"},
	'9': {"███", "█ █", "███", "  █", "███"},
	'.': {" ", " ", " ", " ", "█"},
	'+': {"   ", " █ ", "███", " █ ", "   "},
	'A': {" █ ", "█ █", "███", "█ █", "█ █"},
	'B': {"██ ", "█ █", "██ ", "█ █", "██ "},
	'C': {" ██", "█  ", "█  ", "█  ", " ██"},
	'D': {"██ ", "█ █", "█ █", "█ █", "██ "},
	'E': {"███", "█  ", "██ ", "█  ", "███"},
	'F': {"███", "█  ", "██ ", "█  ", "█  "},
	'G': {" ██", "█  ", "█ █", "█ █", " ██"},
	'H': {"█ █", "█ █", "███", "█ █", "█ █"},
	'I': {"███", " █ ", " █ ", " █ ", "███"},
	'J': {"  █", "  █", "  █", "█ █", " █ "},
	'K': {"█ █", "█ █", "██ ", "█ █", "█ █"},
	'L': {"█  ", "█  ", "█  ", "█  ", "███"},
	'M': {"█ █", "███", "███", "█ █", "█ █"},
	'N': {"█ █", "███", "███", "███", "█ █"},
	'O': {" █ ", "█ █", "█ █", "█ █", " █ "},
	'P': {"██ ", "█ █", "██ ", "█  ", "█  "},
	'Q': {" █ ", "█ █", "█ █", "███", " ██"},
	'R': {"██ ", "█ █", "██ ", "█ █", "█ █"},
	'S': {" ██", "█  ", " █ ", "  █", "██ "},
	'T': {"███", " █ ", " █ ", " █ ", " █ "},
	'U': {"█ █", "█ █", "█ █", "█ █", "███"},
	'V': {"█ █", "█ █", "█ █", "█ █", " █ "},
	'W': {"█ █", "█ █", "███", "███", "█ █"},
	'X': {"█ █", "█ █", " █ ", "█ █", "█ █"},
	'Y': {"█ █", "█ █", " █ ", " █ ", " █ "},
	'Z': {"███", "  █", " █ ", "█  ", "███"},
	' ': {"   ", "   ", "   ", "   ", "   "},
}

// A 3×5 display face stays readable without requiring terminal font support.
func display(text string) string {
	return displayWithSpacing(text, " ")
}

func displayWithSpacing(text, spacing string) string {
	rows := make([]string, displayHeight)
	for index, letter := range text {
		glyph := glyphs[letter]
		for i := range rows {
			if index > 0 {
				rows[i] += spacing
			}
			rows[i] += glyph[i]
		}
	}
	for i := range rows {
		rows[i] = strings.TrimRight(rows[i], " ")
	}
	return strings.Join(rows, "\n")
}

func (m Model) banner(width int) string {
	if m.showDetails {
		return ink(m.color()).Bold(true).Render(linerNotes)
	}
	if m.imageExport {
		return m.bannerText("GITHUB", width) + "\n\n" + m.bannerText("RECAP", width)
	}
	return m.bannerText(m.displayName(), width)
}

func (m Model) compactBanner() bool { return !m.wide() && m.height < fullDisplayHeight }

// In a stacked image, make room for the second word without shrinking the viewport.
func (m Model) imageBannerExtraRows() int {
	if !m.imageExport || m.wide() {
		return 0
	}
	if m.compactBanner() {
		return compactDisplayHeight + 1
	}
	return displayHeight + 1
}

func (m Model) bannerText(name string, width int) string {
	large := display(name)
	if width < lipgloss.Width(large) || m.height < minimumHeight {
		return ink(m.color()).Bold(true).Render(name)
	}
	rows := strings.Split(large, "\n")
	if m.compactBanner() {
		rows = compactRows(rows)
	}
	phase := m.page
	if m.dpgkEffects() && m.titleFlashing() {
		phase++
	}
	for i, row := range rows {
		color := m.color()
		if m.dpgkEffects() {
			color = dpgkColor(phase + 2*i)
		} else if i >= len(rows)/2+1 {
			color = colors[(m.page+1)%len(colors)]
		}
		rows[i] = ink(color).Render(row)
	}
	return strings.Join(rows, "\n")
}

// Pack two rows of pixels into each terminal row using half-block characters.
func compactRows(rows []string) []string {
	compact := []string{}
	for y := 0; y < len(rows); y += 2 {
		top := []rune(rows[y])
		bottom := []rune{}
		if y+1 < len(rows) {
			bottom = []rune(rows[y+1])
		}
		var line strings.Builder
		for x := 0; x < max(len(top), len(bottom)); x++ {
			topFilled := x < len(top) && top[x] != ' '
			bottomFilled := x < len(bottom) && bottom[x] != ' '
			switch {
			case topFilled && bottomFilled:
				line.WriteRune('█')
			case topFilled:
				line.WriteRune('▀')
			case bottomFilled:
				line.WriteRune('▄')
			default:
				line.WriteRune(' ')
			}
		}
		compact = append(compact, line.String())
	}
	return compact
}
