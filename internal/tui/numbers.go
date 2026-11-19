package tui

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// The decimal point is punctuation: 10.0K has four digits/unit letters.
// Truncate the fractional part so rounding cannot overstate activity or grow
// the display past its four-character budget at a unit boundary.
func compactCount(count int) string {
	if count < 10_000 {
		return strconv.Itoa(count)
	}
	unit, suffix := 1_000, "K"
	if count >= 1_000_000 {
		unit, suffix = 1_000_000, "M"
	}
	whole := count / unit
	if whole > 999 {
		return "999M+"
	}
	decimals := 3 - len(strconv.Itoa(whole))
	if decimals == 0 {
		return strconv.Itoa(whole) + suffix
	}
	scale := 10
	if decimals == 2 {
		scale = 100
	}
	fraction := (count % unit) / (unit / scale)
	return fmt.Sprintf("%d.%0*d%s", whole, decimals, fraction, suffix)
}

func (m Model) metricView(text string, width int) string {
	large := display(text)
	if width < lipgloss.Width(large) || m.height < minimumHeight {
		return ink(m.color()).Bold(true).Render(text)
	}
	if !m.wide() && m.height < fullDisplayHeight {
		large = strings.Join(compactRows(strings.Split(large, "\n")), "\n")
	}
	return ink(m.color()).Render(large)
}

func statistic(count int, label string, width int, color string, compact bool) string {
	text := compactCount(count)
	large := display(text)
	if lipgloss.Width(large) > width {
		large = displayWithSpacing(text, "")
	}
	if compact {
		large = strings.Join(compactRows(strings.Split(large, "\n")), "\n")
	}
	if lipgloss.Width(large) > width {
		large = text
	}
	if text != strconv.Itoa(count) {
		fullLabel := fmt.Sprintf("%d %s", count, label)
		if lipgloss.Width(fullLabel) <= width {
			label = fullLabel
		}
	}
	return ink(color).Bold(true).Render(large) + "\n" + muted.Render(ansi.Wrap(label, max(1, width), " "))
}
