package recap

import (
	"context"
	"strings"
	"testing"
)

func TestPopularPRUsesMostCommentedInScope(t *testing.T) {
	for _, includePrivate := range []bool{false, true} {
		wantID := "public-most"
		if includePrivate {
			wantID = "private"
		}
		t.Run(wantID, func(t *testing.T) {
			detailCalls := 0
			client := testAPI(t, func(request graphQLRequest) (string, error) {
				query := compactQuery(request.Query)
				if strings.Contains(query, "pullRequestContributions(") {
					if request.Variables["after"] == nil {
						return `{"data":{"viewer":{"contributionsCollection":{"pullRequestContributions":{
							"nodes":[
								{"occurredAt":"2025-04-03","pullRequest":{"id":"private","repository":{"nameWithOwner":"you/private","isPrivate":true},"comments":{"totalCount":40}}},
								{"occurredAt":"2025-04-02","pullRequest":{"id":"public-less","repository":{"nameWithOwner":"you/public"},"comments":{"totalCount":3}}}
							],"pageInfo":{"hasNextPage":true,"endCursor":"next"}
						}}}}}`, nil
					}
					return `{"data":{"viewer":{"contributionsCollection":{"pullRequestContributions":{
						"nodes":[{"occurredAt":"2025-04-01","pullRequest":{"id":"public-most","repository":{"nameWithOwner":"you/public"},"comments":{"totalCount":9}}}],
						"pageInfo":{"hasNextPage":false}
					}}}}}`, nil
				}
				detailCalls++
				if request.Variables["id"] != wantID {
					t.Fatalf("queried %v, want %s", request.Variables["id"], wantID)
				}
				for _, field := range []string{"$id:ID!", "...onPullRequest", "comments(first:3)", "participants(first:20)"} {
					if !strings.Contains(query, field) {
						t.Errorf("missing %s in %s", field, query)
					}
				}
				return `{"data":{"node":{
					"repository":{"nameWithOwner":"you/recap"},
					"number":42,"title":"Activity timeline","createdAt":"2025-04-01T01:00:00Z","bodyText":"Adds a timeline.",
					"author":{"login":"you"},
					"comments":{"totalCount":9,"nodes":[{"author":{"login":"aki"},"bodyText":"Looks good."},{"author":null,"bodyText":"Thanks."}]},
					"participants":{"nodes":[{"login":"you"},{"login":"aki"},{"login":"ren"}],"pageInfo":{"hasNextPage":true}}
				}}}`, nil
			})
			result := Fetch(context.Background(), client, Options{Year: 2025, IncludePrivate: includePrivate}, KindPullRequests)
			pr := result.PopularPR
			if result.Err != nil || result.PopularPRErr != nil || pr == nil || detailCalls != 1 {
				t.Fatalf("unexpected result: %+v", result)
			}
			if pr.Repository != "you/recap" || pr.Title != "Activity timeline" || pr.Date != "2025-04-01" || pr.Description != "Adds a timeline." ||
				pr.CommentCount != 9 || len(pr.Comments) != 2 || pr.Comments[1].Author != "" ||
				strings.Join(pr.People, ",") != "you,aki,ren" || !pr.MorePeople {
				t.Fatalf("unexpected PR: %+v", pr)
			}
		})
	}
}
