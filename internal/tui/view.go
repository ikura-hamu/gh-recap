package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ikura-hamu/gh-recap/internal/recap"
)

type contentLayout struct {
	Heading        string
	Body           string
	ViewportHeight int
}

func (m Model) scrollLayout() contentLayout {
	content := m.content()
	headingWidth := m.contentWidth()
	if m.wide() {
		headingWidth = sidebarWidth - 3
	}

	titleStyle := ink(m.color()).Bold(true)
	if m.titleFlashing() {
		titleStyle = titleStyle.Foreground(lipgloss.Color("#FFF1FA"))
	}
	parts := []string{
		muted.Render(ansi.Wrap(content.Eyebrow, headingWidth, " ")),
		m.banner(headingWidth),
	}
	if content.Metric != "" {
		parts = append(parts, metricPadding+m.metricView(content.Metric, headingWidth)+metricPadding)
	}
	parts = append(parts, titleStyle.Render(ansi.Wrap(content.Title, headingWidth, " ")))
	heading := strings.Join(parts, "\n")
	height := max(1, m.bodyHeight()-lipgloss.Height(heading)-len(headingGap)-scrollHintRows)
	if m.wide() {
		height = max(1, m.bodyHeight()-scrollHintRows)
	}
	return contentLayout{
		Heading:        heading,
		Body:           ansi.Wrap(content.Body, m.dataWidth(), " "),
		ViewportHeight: height,
	}
}

type viewportKey struct {
	page, width, height int
	details, flash      bool
	loadingFrame        uint64
	calendarDate        string
}

type cachedViewport struct {
	key     viewportKey
	heading string
}

func (m Model) titleFlashing() bool {
	return !m.options.NoAnimation && !m.paused && m.transitionFrame < 4
}

func (m Model) viewportKey() viewportKey {
	key := viewportKey{page: m.page, width: m.width, height: m.height, details: m.showDetails}
	if m.page == pageCalendar && !m.showDetails {
		// The calendar starts showing a new cell at midnight UTC.
		now := time.Now().UTC()
		if m.options.Year == now.Year() {
			key.calendarDate = now.Format(time.DateOnly)
		}
	}
	if !m.options.NoAnimation {
		key.flash = m.titleFlashing()
		if m.page == pageOverview && !m.showDetails && !m.ready() {
			key.loadingFrame = m.sparkleFrame % collectionOrbitSize
		}
	}
	return key
}

func (m *Model) syncViewport() {
	if m.tooSmall() || m.login == "" || m.identityError != nil {
		return
	}
	key := m.viewportKey()
	if m.viewportCache != nil && m.viewportCache.key == key {
		return
	}
	heading, view := m.buildViewport()
	m.viewport = view
	m.viewportCache = &cachedViewport{key: key, heading: heading}
}

func (m Model) scrollViewport() (string, viewport.Model) {
	if m.viewportCache != nil && m.viewportCache.key == m.viewportKey() {
		return m.viewportCache.heading, m.viewport
	}
	return m.buildViewport()
}

func (m Model) buildViewport() (string, viewport.Model) {
	layout := m.scrollLayout()
	view := m.viewport
	view.SetWidth(m.dataWidth())
	view.SetHeight(layout.ViewportHeight)
	view.SetContent(layout.Body)
	// Resizing can reduce the valid range. Clamp the stored position as well.
	view.SetYOffset(view.YOffset())
	return layout.Heading, view
}

func (m Model) View() tea.View {
	if m.tooSmall() {
		return m.screenView(m.sizeWarning())
	}
	width := m.contentWidth()
	body := lipgloss.NewStyle().Height(m.bodyHeight()).MaxHeight(m.bodyHeight()).Render(m.transitionBodyView())
	content := m.headerView(width) + "\n\n" + body + "\n" +
		m.playbackBar(width) + "\n" + ink(m.color()).Render(ansi.Truncate(m.statusView(), width, "…")) + "\n" + m.helpView(width)
	content = lipgloss.NewStyle().Width(width).Render(content)
	content = lipgloss.NewStyle().
		MaxWidth(max(1, m.width)).
		MaxHeight(max(1, m.height)).
		Padding(verticalPadding, horizontalPadding).
		Render(content)
	// The slide is already moving the entire body. Resume the background when
	// it settles so the high-frequency slide ticks only draw one animation.
	if !m.transition.active() {
		content = m.sparkleBackground(content)
	}
	return m.screenView(content)
}

func (m Model) sizeWarning() string {
	width, height := max(1, m.width), max(1, m.height)
	message := fmt.Sprintf("Terminal too small to display GH RECAP.\n\nMinimum: %d columns × %d rows\nCurrent: %d columns × %d rows\n\nResize your terminal to continue.\nq / Ctrl+C to quit", minimumWidth, minimumHeight, m.width, m.height)
	message = lipgloss.NewStyle().MaxWidth(width).MaxHeight(height).
		Render(ansi.Wrap(message, width, " "))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, message)
}

func (m Model) screenView(content string) tea.View {
	width, height := max(1, m.width), max(1, m.height)
	content = lipgloss.NewStyle().
		Width(width).Height(height).
		MaxWidth(width).MaxHeight(height).
		Foreground(lipgloss.Color(foregroundColor)).
		Render(content)
	// Nested styles reset all ANSI attributes, including the screen colors.
	// Keep the background at the terminal default, set by View.BackgroundColor,
	// so cells and terminal padding use the same color and opacity. Restore
	// only the foreground after resets; badges can still set a background.
	baseColors := ansi.Style{}.
		ForegroundColor(lipgloss.Color(foregroundColor)).String()
	content = strings.NewReplacer(
		ansi.ResetStyle, ansi.ResetStyle+baseColors,
		"\x1b[0m", "\x1b[0m"+baseColors,
	).Replace(content)
	view := tea.NewView(content)
	view.AltScreen = true
	view.BackgroundColor = lipgloss.Color(backgroundColor)
	view.ForegroundColor = lipgloss.Color(foregroundColor)
	return view
}

func (m Model) headerView(width int) string {
	scope := "PUBLIC ONLY"
	if m.options.IncludePrivate {
		scope = "PUBLIC + PRIVATE"
	}
	if m.options.Demo {
		scope = "DEMO · " + scope
	}
	badge := lipgloss.NewStyle().
		Background(lipgloss.Color(m.color())).
		Foreground(lipgloss.Color("#191425")).
		Bold(true).
		Padding(0, 1).
		Render("✦ GH RECAP")
	label := muted.Render(fmt.Sprintf("%d / %s", m.options.Year, scope))
	gap := max(1, width-lipgloss.Width(badge)-lipgloss.Width(label))
	return badge + strings.Repeat(" ", gap) + label
}

func (m Model) bodyView() string {
	if m.identityError != nil {
		return accent.Render("Couldn't load your GitHub identity.") + "\n\n" +
			m.identityError.Error() + "\n\nRun gh auth login, then try again.\nOr explore with: gh recap --demo"
	}
	if m.login == "" {
		return m.banner(m.contentWidth()) + "\n\n" + m.collectionView(m.contentWidth())
	}

	heading, view := m.scrollViewport()
	body := view.View()
	if view.TotalLineCount() > view.Height() {
		first := view.YOffset() + 1
		last := min(view.TotalLineCount(), view.YOffset()+view.Height())
		body += muted.Render(fmt.Sprintf("\n↑/↓ scroll · %d–%d / %d", first, last, view.TotalLineCount()))
	}
	if !m.wide() {
		return heading + headingGap + body
	}
	divider := strings.TrimSuffix(strings.Repeat("│\n", m.bodyHeight()), "\n")
	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(sidebarWidth).Render(heading),
		ink(m.color()).Render(divider),
		strings.Repeat(" ", columnGap-1),
		lipgloss.NewStyle().Width(m.dataWidth()).Render(body),
	)
}

const collectionOrbitSize = 8

func (m Model) collectionView(width int) string {
	// A large orbit remains visible even before the first source returns.
	positions := [collectionOrbitSize][2]int{{2, 0}, {4, 0}, {4, 1}, {4, 2}, {2, 2}, {0, 2}, {0, 1}, {0, 0}}
	frame := int(m.sparkleFrame % uint64(len(positions)))
	var rows [3]string
	for y := range rows {
		for x := 0; x < 5; x++ {
			cell := " "
			for i, position := range positions {
				if position == [2]int{x, y} {
					cell = muted.Render("·")
					if i == frame || m.options.NoAnimation {
						cell = accent.Render("●")
					}
				}
			}
			rows[y] += cell
		}
	}
	message := "Collecting your GitHub activity…"
	if m.login == "" {
		message = "Connecting to GitHub…"
	}
	label := accent.Render("COLLECTING") + "\n" + message +
		fmt.Sprintf("\n%d / %d sources collected", len(m.loaded), len(recap.Sources))
	panel := lipgloss.JoinHorizontal(lipgloss.Center, strings.Join(rows[:], "\n"), "   ", label)
	bar := progressBar(float64(len(m.loaded))/float64(len(recap.Sources)), min(40, max(1, width-8)), m.color())
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(panel + "\n\n" + bar)
}

func (m Model) statusView() string {
	state := "PLAYING"
	switch {
	case !m.ready():
		state = "COLLECTING"
	case m.playbackFinished:
		state = "FINISHED"
	case m.showDetails:
		state = "DETAILS"
	}
	if m.paused {
		if state == "PLAYING" {
			state = "PAUSED"
		} else {
			state = "PAUSED · " + state
		}
	}
	if len(m.failed) > 0 || len(m.warnings) > 0 {
		state += " · INCOMPLETE"
	}
	if m.identityError != nil {
		state = "CONNECTION FAILED"
	}

	indicator := "●"
	if !m.options.NoAnimation && len(m.loaded) < len(recap.Sources) && m.identityError == nil {
		indicator = m.loading.View()
	}
	status := fmt.Sprintf("%s %s · %g× · %d/%d sources · %02d/%d", indicator, state, m.speed, len(m.loaded), len(recap.Sources), m.page+1, pageCount)
	if len(m.warnings) > 0 && m.identityError == nil {
		status += " · Enter for warnings"
	}
	return status
}

func (m Model) playbackBar(width int) string {
	segmentWidth := max(1, (width-pageCount)/pageCount)
	var bar strings.Builder
	for page := 0; page < pageCount; page++ {
		if page == m.page {
			fraction := float64(m.elapsed) / float64(m.duration())
			bar.WriteString(progressBar(fraction, segmentWidth, m.color()) + " ")
			continue
		}
		color := "#443A57"
		if page < m.page {
			color = colors[page]
		}
		bar.WriteString(ink(color).Render(strings.Repeat("━", segmentWidth) + " "))
	}
	return bar.String()
}
