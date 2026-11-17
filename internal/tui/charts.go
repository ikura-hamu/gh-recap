package tui

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/lipgloss/v2"
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
		return "No recorded activity in this scope."
	}
	maximum := entries[0].Count
	var body strings.Builder
	for index, entry := range entries {
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
		peak = "No recorded activity yet"
	}
	var body strings.Builder
	fmt.Fprintf(&body, "%s consecutive active days\n\nPeak day: %s · %d contributions\n\n",
		mint.Render(fmt.Sprint(stats.Streak)), peak, stats.PeakCount)
	maximum := slices.Max(stats.Weekdays[:])
	for weekday, count := range stats.Weekdays {
		label := time.Weekday(weekday).String()[:3]
		bar := coloredBar(count, maximum, width, colors[weekday%len(colors)])
		fmt.Fprintf(&body, "%s %s %d\n", label, bar, count)
	}
	return body.String()
}

func calendar(stats recap.Stats, year, width int) string {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Now()
	cellWidth := 1
	if width >= 110 {
		cellWidth = 2
	}
	var body strings.Builder
	fmt.Fprint(&body, "     YOUR CONTRIBUTION CONSTELLATION\n\n")
	// Columns are weeks, aligned to Sunday. 54 columns also cover leap years
	// that start on Saturday; dates outside the selected year stay blank.
	for weekday := 0; weekday < 7; weekday++ {
		body.WriteString(time.Weekday(weekday).String()[:3] + "  ")
		for week := 0; week < 54; week++ {
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
	fmt.Fprintf(&body, "\n%d active days. Each one counts.\nSelected-scope contribution records; not the profile graph.", stats.ActiveDays)
	return body.String()
}

func calendarCell(count, width int) string {
	symbol := "·"
	style := muted
	switch {
	case count > 15:
		symbol = "█"
		style = ink("#FF78C6")
	case count > 5:
		symbol = "■"
		style = ink("#C4A1FF")
	case count > 0:
		symbol = "▪"
		style = ink("#67E8F9")
	}
	return style.Render(symbol + strings.Repeat(" ", width-1))
}
