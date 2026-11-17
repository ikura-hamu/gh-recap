package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Keep stars fixed in space, with staggered brightness cycles. Only empty
// patches are decorated so text, charts, and their immediate spacing stay clear.
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
	palette := [...]string{"#443A57", "#726084", "#B69AC7", "#E9CFFF"}
	glyphs := [...]string{"·", "·", "·", "+"}
	for y := 1; y < len(lines)-1; y++ {
		var row strings.Builder
		start := 0
		for x := 1; x < m.width-1; x++ {
			seed := sparkleSeed(x, y)
			if seed%79 != 0 {
				continue
			}
			phase := (m.sparkleFrame + uint64(seed>>8)) % 48
			if phase >= 28 {
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
			brightness := int(min(phase, 27-phase) / 4)
			row.WriteString(ansi.Cut(lines[y], start, x))
			row.WriteString(ink(palette[brightness]).Render(glyphs[brightness]))
			start = x + 1
		}
		if start > 0 {
			row.WriteString(ansi.Cut(lines[y], start, m.width))
			lines[y] = row.String()
		}
	}
	return strings.Join(lines, "\n")
}

func sparkleSeed(x, y int) uint32 {
	seed := uint32(x)*374761393 + uint32(y)*668265263
	seed = (seed ^ (seed >> 13)) * 1274126177
	return seed ^ (seed >> 16)
}
