package recap

import "time"

// Demo generates repeatable activity without GitHub access. The arithmetic
// spreads contributions across dates and repositories without a random seed.
func Demo(options Options, kind string) Result {
	result := Result{Kind: kind}
	if kind == KindCoauthors {
		result.Coauthors = map[string]int{"@aki": 42, "@haru": 27, "@sora": 18, "@ren": 9, "pair@example.com": 3}
		return result
	}
	end := time.Date(options.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)
	if now := time.Now(); now.Before(end) {
		end = now
	}
	repositories := []Repo{
		{NameWithOwner: "you/gh-recap", PrimaryLanguage: &Language{Name: "Go"}},
		{NameWithOwner: "charmbracelet/bubbletea", PrimaryLanguage: &Language{Name: "Go"}},
		{NameWithOwner: "you/dotfiles", PrimaryLanguage: &Language{Name: "Shell"}},
		{NameWithOwner: "you/weekend-lab", PrimaryLanguage: &Language{Name: "TypeScript"}},
		{NameWithOwner: "team/platform", IsPrivate: true, PrimaryLanguage: &Language{Name: "Rust"}},
	}
	kindIndex := 0
	for index, candidate := range Kinds {
		if candidate == kind {
			kindIndex = index
		}
	}

	date := time.Date(options.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	for day := 0; date.Before(end); day, date = day+1, date.AddDate(0, 0, 1) {
		inactiveDay := (day*7+kindIndex*11)%13 > 7
		skipRepositoryCreation := kind == KindRepositories && day%47 != 0
		if inactiveDay || skipRepositoryCreation {
			continue
		}

		repository := repositories[(day+kindIndex)%len(repositories)]
		if !repository.included(options.IncludePrivate) {
			continue
		}
		count := 1
		if kind == KindCommits {
			count = 1 + (day*17)%19
		}
		result.Events = append(result.Events, Event{
			Date: date.Format(time.DateOnly), Kind: kind, Count: count, Repo: repository,
		})
	}
	if kind == KindPullRequests && len(result.Events) > 0 {
		event := result.Events[len(result.Events)/2]
		result.PopularPR = &PopularPR{
			Repository:   event.Repo.NameWithOwner,
			Number:       42,
			Title:        "Add a scrollable activity timeline",
			Date:         event.Date,
			Description:  "Adds a timeline so terminal users can browse a full year of activity by day.",
			CommentCount: 8,
			Comments: []PRComment{
				{Author: "aki", Body: "The timeline now fits on my small terminal."},
				{Author: "haru", Body: "Could we show the selected date while scrolling?"},
				{Author: "you", Body: "Added the selected date above the timeline."},
			},
			People: []string{"you", "aki", "haru", "ren"},
		}
	}
	return result
}
