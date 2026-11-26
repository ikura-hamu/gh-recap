// Package recap collects and summarizes the authenticated user's GitHub activity.
package recap

const (
	KindCommits      = "commits"
	KindPullRequests = "prs"
	KindReviews      = "reviews"
	KindIssues       = "issues"
	KindRepositories = "repositories"
	KindCoauthors    = "coauthors"
)

// Kinds defines the collection order used by the UI and demo data.
var Kinds = []string{KindCommits, KindPullRequests, KindReviews, KindIssues, KindRepositories}

// Sources includes supplementary data which must not inflate contribution totals.
var Sources = append(append([]string{}, Kinds...), KindCoauthors)

type Options struct {
	Year           int
	IncludePrivate bool
	Demo           bool
	NoAnimation    bool
	DPGK           bool
	Recording      bool
}

type Language struct{ Name string }

type Repo struct {
	NameWithOwner   string
	URL             string
	IsPrivate       bool
	Visibility      string
	PrimaryLanguage *Language
}

func (r Repo) included(includePrivate bool) bool {
	if includePrivate {
		return true
	}
	// Internal repositories are also excluded from public-only results.
	return !r.IsPrivate && (r.Visibility == "" || r.Visibility == "PUBLIC")
}

// Event is a contribution record. Commit records can represent several commits
// on the same day; the other activity types have a count of one.
type Event struct {
	Date  string
	Kind  string
	Count int
	Repo  Repo
}

// Result retains collected events even when a later request fails.
type Result struct {
	Kind         string
	Events       []Event
	Warnings     []string
	Err          error
	Coauthors    map[string]int
	PopularPR    *PopularPR
	PopularPRErr error
}

type PopularPR struct {
	Repository   string
	Number       int
	Title        string
	Date         string
	URL          string
	Description  string
	CommentCount int
	Comments     []PRComment
	People       []string
	MorePeople   bool
}

type PRComment struct {
	Author string
	Body   string
}

type Stats struct {
	Total      int
	ActiveDays int
	Streak     int
	Months     [12]int
	Weekdays   [7]int
	Days       map[string]int
	Kinds      map[string]int
	Repos      map[string]int
	Languages  map[string]int
	PeakDay    string
	PeakCount  int
}

type Entry struct {
	Name  string
	Count int
}
