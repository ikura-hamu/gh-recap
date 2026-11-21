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
		return loading
	}
	if m.failed[kind] {
		return fmt.Sprintf(partialCount, count)
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
			Eyebrow: overviewKicker,
			Title:   contributions,
			Metric:  compactCount(stats.Total),
			Body:    m.overviewBody(stats),
		}
		if !m.ready() {
			content.Metric = ""
			content.Title = collectingYear
		}
		return content
	case pageCalendar:
		return pageContent{
			Eyebrow: calendarKicker, Title: calendarTitle,
			Body: calendar(stats, m.options.Year, m.dataWidth()),
		}
	case pageMonths:
		return pageContent{
			Eyebrow: monthsKicker, Title: monthsTitle,
			Body: monthlyActivity(stats, chartWidth),
		}
	case pageStreak:
		return pageContent{
			Eyebrow: streakKicker, Title: streakTitle,
			Metric: compactCount(stats.Streak),
			Body:   streakActivity(stats, chartWidth),
		}
	case pageCommits:
		return m.withKindMetric(pageContent{
			Eyebrow: commitsKicker,
			Title:   m.number(recap.KindCommits, stats.Kinds[recap.KindCommits]) + " " + commits,
			Body:    m.kindRanking(recap.KindCommits, chartWidth) + commitsNotes,
		}, recap.KindCommits, stats.Kinds[recap.KindCommits], commits)
	case pageCoauthors:
		content := pageContent{
			Eyebrow: coauthorsKicker, Title: coauthorsTitle,
			Body: m.coauthorRanking(chartWidth),
		}
		if m.loaded[recap.KindCoauthors] {
			content.Metric = compactCount(len(m.coauthors))
			content.Title = coauthors
			if m.failed[recap.KindCoauthors] {
				content.Title += incomplete
			}
		}
		return content
	case pagePullRequests:
		return m.withKindMetric(pageContent{
			Eyebrow: prsKicker,
			Title:   m.number(recap.KindPullRequests, stats.Kinds[recap.KindPullRequests]) + " " + prs,
			Body:    m.kindRanking(recap.KindPullRequests, chartWidth) + prsNotes,
		}, recap.KindPullRequests, stats.Kinds[recap.KindPullRequests], prs)
	case pageReviews:
		return m.withKindMetric(pageContent{
			Eyebrow: reviewsKicker,
			Title:   m.number(recap.KindReviews, stats.Kinds[recap.KindReviews]) + " " + reviewContributions,
			Body:    m.kindRanking(recap.KindReviews, chartWidth) + reviewsNotes,
		}, recap.KindReviews, stats.Kinds[recap.KindReviews], reviewContributions)
	case pageRepositories:
		return pageContent{
			Eyebrow: reposKicker, Title: reposTitle,
			Metric: compactCount(len(stats.Repos)),
			Body:   ranking(recap.Ranked(stats.Repos), chartWidth),
		}
	case pageLanguages:
		return pageContent{
			Eyebrow: languagesKicker, Title: languagesTitle,
			Body: ranking(recap.Ranked(stats.Languages), chartWidth) + languagesNotes,
		}
	default:
		return pageContent{
			Eyebrow: summaryKicker, Title: fmt.Sprintf("%d · @%s", m.options.Year, m.login),
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
	total := statistic(stats.Total, contributions, width, colors[0], compact)
	half := (width - 4) / 2
	activity := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(half).Render(statistic(stats.ActiveDays, activeDays, half, colors[1], compact)),
		"    ", statistic(len(stats.Repos), repositories, half, colors[4], compact))
	return fmt.Sprintf(overviewBody, m.login, total, activity)
}

func (m Model) encoreBody(stats recap.Stats) string {
	metrics := [...]struct {
		kind, label string
		count       int
	}{
		{recap.KindCommits, commits, stats.Kinds[recap.KindCommits]},
		{recap.KindPullRequests, prs, stats.Kinds[recap.KindPullRequests]},
		{recap.KindReviews, reviews, stats.Kinds[recap.KindReviews]},
		{recap.KindIssues, issues, stats.Kinds[recap.KindIssues]},
		{recap.KindRepositories, newRepos, stats.Kinds[recap.KindRepositories]},
		{"", activeDays, stats.ActiveDays},
		{"", longestStreak, stats.Streak},
	}
	columns := 3
	if m.bodyHeight() < 16 {
		// Use two rows on short terminals so the gap and all metrics fit.
		columns = 4
	}
	gap := "  "
	if columns == 4 && m.dataWidth() < columns*lipgloss.Width(longestStreak)+len(gap)*(columns-1) {
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
			card = strings.Repeat("\n", numberHeight-1) + muted.Render(loading) + "\n" + muted.Render(ansi.Wrap(label, width, " "))
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

	hint := explore
	if incomplete {
		hint += incompleteMark
	}
	body := strings.Join(rows, "\n\n")
	if m.bodyHeight() >= lipgloss.Height(body)+4 {
		body += farewell
	}
	return body + "\n" + muted.Render(hint)
}

func (m Model) withKindMetric(content pageContent, kind string, count int, label string) pageContent {
	if m.loaded[kind] {
		content.Metric = compactCount(count)
		content.Title = label
		if m.failed[kind] {
			content.Title += recordedIncomplete
		}
	}
	return content
}

func (m Model) detailsContent(stats recap.Stats, chartWidth int) pageContent {
	var body strings.Builder
	heading := ink(m.color()).Bold(true)
	fmt.Fprintln(&body, metricNotes)
	fmt.Fprintln(&body, "\n"+heading.Render(activityHeading))
	for _, kind := range recap.Kinds {
		fmt.Fprintf(&body, "%-14s %s\n", kind, m.number(kind, stats.Kinds[kind]))
	}
	fmt.Fprintln(&body, "\n"+heading.Render(reposHeading)+"\n"+ranking(recap.Ranked(stats.Repos), chartWidth))
	fmt.Fprintln(&body, "\n"+heading.Render(languagesHeading)+languageCount+ranking(recap.Ranked(stats.Languages), chartWidth))
	fmt.Fprintln(&body, "\n"+heading.Render(coauthorsHeading)+sharedCommits+m.coauthorRanking(chartWidth))
	fmt.Fprintln(&body, "\n"+heading.Render(dailyHeading))
	start := time.Date(m.options.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	for date := start; date.Year() == m.options.Year; date = date.AddDate(0, 0, 1) {
		day := date.Format(time.DateOnly)
		if stats.Days[day] > 0 {
			fmt.Fprintf(&body, "%s  %d\n", day, stats.Days[day])
		}
	}
	if len(m.warnings) > 0 {
		fmt.Fprintln(&body, "\n"+heading.Render(warningsHeading)+"\n"+strings.Join(m.warnings, "\n"))
	}
	return pageContent{Title: detailsTitle, Body: body.String()}
}

func (m Model) coauthorRanking(width int) string {
	if !m.loaded[recap.KindCoauthors] {
		return loadingCoauthors
	}
	var body string
	if len(m.coauthors) == 0 {
		body = noCoauthors
	} else {
		body = ranking(recap.Ranked(m.coauthors), width)
	}
	if m.failed[recap.KindCoauthors] {
		body += coauthorsIncomplete
	}
	return body + coauthorsNotes
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
