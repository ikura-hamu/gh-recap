package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ikura-hamu/gh-recap/internal/recap"
)

type pageContent struct {
	Eyebrow string
	Title   string
	Metric  string
	Body    string
}

func (m Model) number(kind string, count int) string {
	if !m.loaded[kind] {
		return "Loading…"
	}
	if m.failed[kind] {
		return fmt.Sprintf("%d recorded · incomplete", count)
	}
	return fmt.Sprint(count)
}

func (m Model) content() pageContent {
	stats := recap.Summarize(m.events)
	chartWidth := max(5, m.dataWidth()-14)
	if m.showDetails {
		return m.detailsContent(stats, chartWidth)
	}

	switch m.page {
	case pageOverview:
		content := pageContent{
			Eyebrow: "YOUR YEAR, REPLAYED.",
			Title:   "recorded contributions",
			Metric:  compactCount(stats.Total),
			Body:    m.overviewBody(stats),
		}
		if !m.ready() {
			content.Metric = ""
			content.Title = "Collecting your year"
		}
		return content
	case pageCalendar:
		return pageContent{
			Eyebrow: "LITTLE DAYS. BIG YEAR.", Title: "Your activity map",
			Body: calendar(stats, m.options.Year, m.dataWidth()),
		}
	case pageMonths:
		return pageContent{
			Eyebrow: "YOU FOUND YOUR RHYTHM.", Title: "Month by month",
			Body: monthlyActivity(stats, chartWidth),
		}
	case pageStreak:
		return pageContent{
			Eyebrow: "ONE DAY BECAME A STREAK.", Title: "consecutive active days",
			Metric: compactCount(stats.Streak),
			Body:   streakActivity(stats, chartWidth),
		}
	case pageCommits:
		return m.withKindMetric(pageContent{
			Eyebrow: "BUILT, ONE COMMIT AT A TIME.",
			Title:   m.number(recap.KindCommits, stats.Kinds[recap.KindCommits]) + " commits",
			Body:    m.kindRanking(recap.KindCommits, chartWidth) + "\nGitHub contribution commits, not every branch or push.",
		}, recap.KindCommits, stats.Kinds[recap.KindCommits], "commits")
	case pageCoauthors:
		content := pageContent{
			Eyebrow: "GOOD WORK HAS GOOD COMPANY.", Title: "Co-authors",
			Body: m.coauthorRanking(chartWidth),
		}
		if m.loaded[recap.KindCoauthors] {
			content.Metric = compactCount(len(m.coauthors))
			content.Title = "co-authors"
			if m.failed[recap.KindCoauthors] {
				content.Title += " · incomplete"
			}
		}
		return content
	case pagePullRequests:
		return m.withKindMetric(pageContent{
			Eyebrow: "IDEAS BECAME PULL REQUESTS.",
			Title:   m.number(recap.KindPullRequests, stats.Kinds[recap.KindPullRequests]) + " pull requests",
			Body:    m.kindRanking(recap.KindPullRequests, chartWidth) + "\nPRs opened during this year, including first and popular PRs.",
		}, recap.KindPullRequests, stats.Kinds[recap.KindPullRequests], "pull requests")
	case pageReviews:
		return m.withKindMetric(pageContent{
			Eyebrow: "YOU MADE OTHER WORK BETTER.",
			Title:   m.number(recap.KindReviews, stats.Kinds[recap.KindReviews]) + " review contributions",
			Body:    m.kindRanking(recap.KindReviews, chartWidth) + "\nLatest review contribution per reviewed PR during this year.",
		}, recap.KindReviews, stats.Kinds[recap.KindReviews], "review contributions")
	case pageRepositories:
		return pageContent{
			Eyebrow: "YOUR CORNER OF GITHUB.", Title: "active repositories",
			Metric: compactCount(len(stats.Repos)),
			Body:   ranking(recap.Ranked(stats.Repos), chartWidth),
		}
	case pageLanguages:
		return pageContent{
			Eyebrow: "A YEAR WITH MANY DIALECTS.", Title: "Languages in your active repositories",
			Body: ranking(recap.Ranked(stats.Languages), chartWidth) + "\nCurrent primary language per repository. Counts repositories, not code.",
		}
	default:
		return pageContent{
			Eyebrow: "THAT WAS YOUR YEAR.", Title: fmt.Sprintf("%d · @%s", m.options.Year, m.login),
			Body: m.encoreBody(stats),
		}
	}
}

func (m Model) overviewBody(stats recap.Stats) string {
	if !m.ready() {
		return m.collectionView(m.dataWidth())
	}
	width := m.dataWidth()
	compact := m.bodyHeight() < 20
	total := statistic(stats.Total, "recorded contributions", width, colors[0], compact)
	half := (width - 4) / 2
	activity := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(half).Render(statistic(stats.ActiveDays, "active days", half, colors[1], compact)),
		"    ", statistic(len(stats.Repos), "repositories", half, colors[4], compact))
	return fmt.Sprintf("@%s, you showed up.\n\n%s\n\n%s\n\nYour work has a story. Let's rewind.", m.login, total, activity)
}

func (m Model) encoreBody(stats recap.Stats) string {
	metrics := [...]struct {
		kind, label string
		count       int
	}{
		{recap.KindCommits, "commits", stats.Kinds[recap.KindCommits]},
		{recap.KindPullRequests, "pull requests", stats.Kinds[recap.KindPullRequests]},
		{recap.KindReviews, "reviews", stats.Kinds[recap.KindReviews]},
		{recap.KindIssues, "issues", stats.Kinds[recap.KindIssues]},
		{recap.KindRepositories, "new repositories", stats.Kinds[recap.KindRepositories]},
		{"", "active days", stats.ActiveDays},
		{"", "longest streak", stats.Streak},
	}
	columns := 3
	if m.bodyHeight() < 16 {
		// Use two rows on short terminals so the gap and all metrics fit.
		columns = 4
	}
	gap := "  "
	if columns == 4 && m.dataWidth() < columns*len("longest streak")+len(gap)*(columns-1) {
		gap = " "
	}
	width := max(1, (m.dataWidth()-len(gap)*(columns-1))/columns)
	compact := m.bodyHeight() < 24
	numberHeight := 5
	if compact {
		numberHeight = 3
	}

	var rows, cards []string
	incomplete := false
	for i, metric := range metrics {
		label := metric.label
		if metric.kind != "" && m.failed[metric.kind] {
			label += " *"
			incomplete = true
		}
		color := colors[(i+1)%len(colors)]
		card := statistic(metric.count, label, width, color, compact)
		if metric.kind != "" && !m.loaded[metric.kind] {
			card = strings.Repeat("\n", numberHeight-1) + muted.Render("Loading…") + "\n" + muted.Render(ansi.Wrap(label, width, " "))
		}
		if len(cards) > 0 {
			cards = append(cards, gap)
		}
		cards = append(cards, lipgloss.NewStyle().Width(width).Render(card))
		if (i+1)%columns == 0 || i == len(metrics)-1 {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cards...))
			cards = cards[:0]
		}
	}

	hint := "Enter to explore all the details."
	if incomplete {
		hint += "  * Incomplete"
	}
	body := strings.Join(rows, "\n\n")
	if m.bodyHeight() >= lipgloss.Height(body)+4 {
		body += "\n\nThanks for showing up. See you in the next commit."
	}
	return body + "\n" + muted.Render(hint)
}

func (m Model) withKindMetric(content pageContent, kind string, count int, label string) pageContent {
	if m.loaded[kind] {
		content.Metric = compactCount(count)
		content.Title = label
		if m.failed[kind] {
			content.Title += " recorded · incomplete"
		}
	}
	return content
}

const metricNotes = `Counts follow GitHub contribution rules; they are not all Git events.
Private activity requires access through your gh authentication.
Languages describe current primary languages of active repositories,
not code you wrote this year. Reviews count the latest review per PR, not comments.
Repository rankings combine the five recorded activity types.
Calendar uses dates returned by GitHub; the selected year uses UTC bounds.`

func (m Model) detailsContent(stats recap.Stats, chartWidth int) pageContent {
	var body strings.Builder
	fmt.Fprintln(&body, metricNotes)
	fmt.Fprintln(&body, "\nACTIVITY")
	for _, kind := range recap.Kinds {
		fmt.Fprintf(&body, "%-14s %s\n", kind, m.number(kind, stats.Kinds[kind]))
	}
	fmt.Fprintln(&body, "\nACTIVE REPOSITORIES\n"+ranking(recap.Ranked(stats.Repos), chartWidth))
	fmt.Fprintln(&body, "\nLANGUAGES · repository count\n"+ranking(recap.Ranked(stats.Languages), chartWidth))
	fmt.Fprintln(&body, "\nCO-AUTHORS · shared commits\n"+m.coauthorRanking(chartWidth))
	fmt.Fprintln(&body, "\nDAILY ACTIVITY")
	start := time.Date(m.options.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	for date := start; date.Year() == m.options.Year; date = date.AddDate(0, 0, 1) {
		day := date.Format(time.DateOnly)
		if stats.Days[day] > 0 {
			fmt.Fprintf(&body, "%s  %d\n", day, stats.Days[day])
		}
	}
	if len(m.warnings) > 0 {
		fmt.Fprintln(&body, "\nCOLLECTION WARNINGS\n"+strings.Join(m.warnings, "\n"))
	}
	return pageContent{Eyebrow: "THE LINER NOTES", Title: "Explore your year", Body: body.String()}
}

func (m Model) coauthorRanking(width int) string {
	if !m.loaded[recap.KindCoauthors] {
		return "Loading co-authors…"
	}
	var body string
	if len(m.coauthors) == 0 {
		body = "No co-authored commits found in the scanned history."
	} else {
		body = ranking(recap.Ranked(m.coauthors), width)
	}
	if m.failed[recap.KindCoauthors] {
		body += "\nIncomplete results. Enter for collection warnings.\n"
	}
	return body + "\nShared commits per author, excluding you.\nDefault branches of contribution repositories; commit dates in the selected year.\nEach commit counts once per co-author, including unlinked email identities."
}

func (m Model) kindRanking(kind string, width int) string {
	repositories := make(map[string]int)
	for _, event := range m.events {
		if event.Kind == kind {
			repositories[event.Repo.NameWithOwner] += event.Count
		}
	}
	return ranking(recap.Ranked(repositories), width)
}
