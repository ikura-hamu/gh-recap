package tui

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ikura-hamu/gh-recap/internal/recap"
)

func coloredBar(count, maximum, width int, color string) string {
	if maximum <= 0 {
		return ""
	}
	return progressBar(float64(count)/float64(maximum), width, color)
}

func progressBar(fraction float64, width int, color string) string {
	bar := progress.New(
		progress.WithWidth(width),
		progress.WithoutPercentage(),
		progress.WithColors(lipgloss.Color(color)),
		progress.WithFillCharacters('━', '─'),
	)
	bar.EmptyColor = lipgloss.Color("#443A57")
	return bar.ViewAs(fraction)
}

func ranking(entries []recap.Entry, width int) string {
	if len(entries) == 0 {
		return emptyRanking
	}
	maximum := entries[0].Count
	var body strings.Builder
	for index, entry := range entries {
		if index < 3 {
			color := []string{"#FFD580", "#CAD8F0", "#F5AD85"}[index]
			badge := lipgloss.NewStyle().Background(lipgloss.Color(color)).
				Foreground(lipgloss.Color(backgroundColor)).Bold(true).
				PaddingRight(1).Render(fmt.Sprintf("#%d", index+1))
			number := ink(color).Bold(true).Render(strings.Join(compactRows(strings.Split(display(compactCount(entry.Count)), "\n")), "\n"))
			labelWidth := max(8, width-lipgloss.Width(badge)-lipgloss.Width(number)-4)
			label := ink(color).Bold(true).Render(ansi.Wrap(entry.Name, labelWidth, " /"))
			label += "\n" + coloredBar(entry.Count, maximum, labelWidth, color)
			fmt.Fprintln(&body, lipgloss.JoinHorizontal(lipgloss.Top, badge, " ", number, "   ", label))
			body.WriteByte('\n')
			continue
		}
		color := colors[index%len(colors)]
		count := ink(color).Bold(true).Render(fmt.Sprint(entry.Count))
		bar := coloredBar(entry.Count, maximum, width, color)
		fmt.Fprintf(&body, "%02d  %s  %s\n    %s\n", index+1, entry.Name, count, bar)
	}
	return body.String()
}

func monthlyActivity(stats recap.Stats, width int) string {
	maximum := slices.Max(stats.Months[:])
	var body strings.Builder
	for month, count := range stats.Months {
		label := time.Month(month + 1).String()[:3]
		bar := coloredBar(count, maximum, width, colors[month%len(colors)])
		fmt.Fprintf(&body, "%s  %s %d\n", label, bar, count)
	}
	return body.String()
}

func streakActivity(stats recap.Stats, width int) string {
	peak := stats.PeakDay
	if peak == "" {
		peak = noPeak
	}
	var body strings.Builder
	fmt.Fprintf(&body, streakBody,
		mint.Render(fmt.Sprint(stats.Streak)), peak, stats.PeakCount)
	maximum := slices.Max(stats.Weekdays[:])
	for weekday, count := range stats.Weekdays {
		label := time.Weekday(weekday).String()[:3]
		bar := coloredBar(count, maximum, width, colors[weekday%len(colors)])
		fmt.Fprintf(&body, "%s %s %d\n", label, bar, count)
	}
	return body.String()
}

const (
	calendarWeeks      = 54 // covers leap years that start on Saturday
	calendarLabelGap   = "  "
	calendarLabelWidth = len("Sun" + calendarLabelGap)
)

func calendar(stats recap.Stats, year, width int) string {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Now()
	cellWidth := 1
	if width >= calendarLabelWidth+2*calendarWeeks {
		cellWidth = 2
	}
	var body strings.Builder
	body.WriteString("     " + calendarHeading + "\n\n")
	// Columns are weeks, aligned to Sunday. 54 columns also cover leap years
	// that start on Saturday; dates outside the selected year stay blank.
	for weekday := 0; weekday < 7; weekday++ {
		body.WriteString(time.Weekday(weekday).String()[:3] + calendarLabelGap)
		for week := 0; week < calendarWeeks; week++ {
			date := start.AddDate(0, 0, week*7+weekday-int(start.Weekday()))
			if date.Year() != year || date.After(now) {
				body.WriteString(strings.Repeat(" ", cellWidth))
				continue
			}
			count := stats.Days[date.Format(time.DateOnly)]
			body.WriteString(calendarCell(count, cellWidth))
		}
		body.WriteByte('\n')
	}
	fmt.Fprintf(&body, calendarNotes, stats.ActiveDays)
	return body.String()
}

func calendarCell(count, width int) string {
	color := "#40384D"
	switch {
	case count > 15:
		color = "#23834A"
	case count > 5:
		color = "#4CB873"
	case count > 0:
		color = "#9BE9AA"
	}
	return ink(color).Render("■" + strings.Repeat(" ", width-1))
}
