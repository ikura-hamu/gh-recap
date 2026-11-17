package recap

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPaginationAndPrivacy(t *testing.T) {
	tests := []struct {
		name           string
		includePrivate bool
		wantEvents     int
	}{
		{name: "public", wantEvents: 1},
		{name: "private", includePrivate: true, wantEvents: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			client := testAPI(t, func(request graphQLRequest) (string, error) {
				query := compactQuery(request.Query)
				selections := []string{
					"issue{repository{", "$from:DateTime!", "$to:DateTime!",
					"$after:String", "excludeFirst:false", "excludePopular:false",
				}
				for _, selection := range selections {
					if !strings.Contains(query, selection) {
						t.Fatalf("missing %q in query: %s", selection, request.Query)
					}
				}
				calls++
				switch calls {
				case 1:
					if request.Variables["after"] != nil {
						t.Fatal("first page must have a null cursor")
					}
					return fixture(t, "issues_public"), nil
				case 2:
					if request.Variables["after"] != "next" {
						t.Fatal("lost pagination cursor")
					}
					return fixture(t, "issues_private"), nil
				default:
					t.Fatal("unexpected extra page request")
					return "", nil
				}
			})
			result := Fetch(context.Background(), client, Options{Year: 2025, IncludePrivate: test.includePrivate}, KindIssues)
			if result.Err != nil || len(result.Events) != test.wantEvents || calls != 2 {
				t.Fatalf("unexpected result: %+v", result)
			}
		})
	}
}

func TestPartialFailure(t *testing.T) {
	calls := 0
	client := testAPI(t, func(request graphQLRequest) (string, error) {
		calls++
		if calls > 1 {
			return "", errors.New("rate limit")
		}
		return fixture(t, "reviews_first"), nil
	})
	result := Fetch(context.Background(), client, Options{Year: 2025}, KindReviews)
	if result.Err == nil || len(result.Events) != 1 {
		t.Fatalf("partial results lost: %+v", result)
	}
}

func TestCommitsUseMonthlyWindows(t *testing.T) {
	calls := 0
	client := testAPI(t, func(request graphQLRequest) (string, error) {
		calls++
		from, err := time.Parse(time.RFC3339, request.Variables["from"].(string))
		if err != nil {
			t.Fatal(err)
		}
		to, err := time.Parse(time.RFC3339, request.Variables["to"].(string))
		if err != nil {
			t.Fatal(err)
		}
		if int(from.Month()) != calls || to.Month() != from.Month() || !strings.Contains(request.Query, "commitCount") {
			t.Fatal("invalid monthly window")
		}
		return strings.ReplaceAll(fixture(t, "commits"), "{{date}}", from.Format(time.RFC3339)), nil
	})
	result := Fetch(context.Background(), client, Options{Year: 2024}, KindCommits)
	if result.Err != nil || calls != 12 || Summarize(result.Events).Total != 48 {
		t.Fatalf("unexpected commits: %+v", result)
	}
}

func TestOtherContributionQueries(t *testing.T) {
	tests := []struct{ kind, field, selection string }{
		{KindPullRequests, "pullRequestContributions", "pullRequest{repository{"},
		{KindReviews, "pullRequestReviewContributions", "repository{"},
		{KindRepositories, "repositoryContributions", "repository{"},
	}
	for _, test := range tests {
		t.Run(test.kind, func(t *testing.T) {
			client := testAPI(t, func(request graphQLRequest) (string, error) {
				query := compactQuery(request.Query)
				if !strings.Contains(query, test.field+"(") || !strings.Contains(query, test.selection) {
					t.Fatalf("wrong query: %s", request.Query)
				}
				return `{"data":{"viewer":{"contributionsCollection":{"` + test.field + `":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}}}`, nil
			})
			result := Fetch(context.Background(), client, Options{Year: 2025}, test.kind)
			if result.Err != nil || len(result.Events) != 0 {
				t.Fatalf("unexpected result: %+v", result)
			}
		})
	}
}
