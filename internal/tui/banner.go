package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

var glyphs = map[rune][5]string{
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
	rows := make([]string, 5)
	for _, letter := range text {
		glyph := glyphs[letter]
		for i := range rows {
			rows[i] += glyph[i] + " "
		}
	}
	for i := range rows {
		rows[i] = strings.TrimRight(rows[i], " ")
	}
	return strings.Join(rows, "\n")
}

func (m Model) banner(width int) string {
	name := m.displayName()
	if m.showDetails {
		return ink(m.color()).Bold(true).Render("THE LINER NOTES")
	}
	large := display(name)
	if width < lipgloss.Width(large) || m.height < 24 {
		return ink(m.color()).Bold(true).Render(name)
	}
	rows := strings.Split(large, "\n")
	if !m.wide() && m.height < 30 {
		rows = compactRows(rows)
	}
	for i, row := range rows {
		color := m.color()
		if i >= len(rows)/2+1 {
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
