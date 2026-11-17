package recap

import (
	"sort"
	"time"
)

func Summarize(events []Event) Stats {
	stats := Stats{
		Days:      make(map[string]int),
		Kinds:     make(map[string]int),
		Repos:     make(map[string]int),
		Languages: make(map[string]int),
	}
	repositoryLanguages := make(map[string]string)
	for _, event := range events {
		date, err := time.Parse(time.DateOnly, event.Date)
		if err != nil {
			continue
		}
		stats.Total += event.Count
		stats.Days[event.Date] += event.Count
		stats.Months[int(date.Month())-1] += event.Count
		stats.Weekdays[date.Weekday()] += event.Count
		stats.Kinds[event.Kind] += event.Count
		stats.Repos[event.Repo.NameWithOwner] += event.Count
		if event.Repo.PrimaryLanguage != nil {
			repositoryLanguages[event.Repo.NameWithOwner] = event.Repo.PrimaryLanguage.Name
		}
	}
	// Languages count unique repositories, not individual contributions.
	for _, language := range repositoryLanguages {
		stats.Languages[language]++
	}
	stats.summarizeCalendar()
	return stats
}

func (s *Stats) summarizeCalendar() {
	dates := make([]string, 0, len(s.Days))
	for date, count := range s.Days {
		if count > 0 {
			dates = append(dates, date)
		}
	}
	sort.Strings(dates)
	s.ActiveDays = len(dates)

	streak := 0
	var previous time.Time
	for _, day := range dates {
		date, _ := time.Parse(time.DateOnly, day) // Summarize already validated the date.
		if date.Equal(previous.AddDate(0, 0, 1)) {
			streak++
		} else {
			streak = 1
		}
		s.Streak = max(s.Streak, streak)
		previous = date
		// Sorted dates make ties deterministic: the earliest peak wins.
		if s.Days[day] > s.PeakCount {
			s.PeakDay = day
			s.PeakCount = s.Days[day]
		}
	}
}

func Ranked(counts map[string]int) []Entry {
	entries := make([]Entry, 0, len(counts))
	for name, count := range counts {
		entries = append(entries, Entry{Name: name, Count: count})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Count == entries[j].Count {
			return entries[i].Name < entries[j].Name
		}
		return entries[i].Count > entries[j].Count
	})
	return entries
}
