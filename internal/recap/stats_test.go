package recap

import "testing"

func TestSummary(t *testing.T) {
	repository := Repo{NameWithOwner: "you/app", PrimaryLanguage: &Language{Name: "Go"}}
	events := []Event{
		{Date: "2024-02-28", Kind: KindCommits, Count: 3, Repo: repository},
		{Date: "2024-02-29", Kind: KindPullRequests, Count: 1, Repo: repository},
		{Date: "2024-03-01", Kind: KindReviews, Count: 2, Repo: repository},
		{Date: "2024-03-03", Kind: KindIssues, Count: 1, Repo: repository},
	}
	stats := Summarize(events)
	if stats.Total != 7 || stats.Streak != 3 || stats.ActiveDays != 4 || stats.Languages["Go"] != 1 || stats.Months[1] != 4 || stats.PeakDay != "2024-02-28" {
		t.Fatalf("unexpected summary: %+v", stats)
	}
}
