package tui

import "charm.land/lipgloss/v2"

const (
	backgroundColor   = "#191425"
	foregroundColor   = "#EEE7F5"
	horizontalPadding = 4
	verticalPadding   = 1
	sidebarWidth      = 43
	columnGap         = 3 // divider and space to its right
	headingGap        = "\n\n"
	metricPadding     = "\n"
	scrollHintRows    = 1
	// One header, one blank row, and three footer rows.
	reservedRows = 2*verticalPadding + 2 + 3
	// The activity grid must fit on one line, even in the stacked layout.
	minimumWidth = calendarLabelWidth + calendarWeeks + 2*horizontalPadding
	wideWidth    = minimumWidth + sidebarWidth + columnGap
	// Keep enough scrollable rows to show an entire compact number at once.
	stackedBodyExtraRows = 2 + 2*len(metricPadding) + len(headingGap) + compactDisplayHeight + scrollHintRows
	minimumHeight        = reservedRows + 2*compactDisplayHeight + stackedBodyExtraRows
	fullDisplayHeight    = reservedRows + 2*displayHeight + stackedBodyExtraRows
)

var colors = []string{
	"#FF78C6", "#67E8F9", "#C4A1FF", "#FFD580", "#7CF7B0",
	"#FF987D", "#89B4FF", "#F5A9E1", "#E2F58B", "#FF78C6",
}

var (
	accent = ink("#FF75B5").Bold(true)
	mint   = ink("#73F5C4")
	muted  = ink("#9B93B5")
)

func ink(color string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color))
}

func (m Model) color() string { return colors[m.page%len(colors)] }

func (m Model) wide() bool {
	return m.width >= wideWidth && !m.showDetails
}

func (m Model) tooSmall() bool { return m.width < minimumWidth || m.height < minimumHeight }

func (m Model) contentWidth() int { return max(1, m.width-2*horizontalPadding) }

func (m Model) bodyHeight() int { return max(1, m.height-reservedRows) }

func (m Model) dataWidth() int {
	width := m.contentWidth()
	if m.wide() {
		return width - sidebarWidth - columnGap
	}
	return width
}

func (m Model) displayName() string { return pageDefinitions[m.page].Banner }
