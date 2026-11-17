package recap

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func actorJSON(login, email string) map[string]any {
	actor := map[string]any{"name": login, "email": email, "user": nil}
	if login != "" {
		actor["user"] = map[string]any{"login": login}
	}
	return actor
}

func commitJSON(oid, date string, authors ...map[string]any) map[string]any {
	return map[string]any{
		"id": "id-" + oid, "oid": oid, "committedDate": date,
		"authors": map[string]any{"nodes": authors, "pageInfo": map[string]any{"hasNextPage": false}},
	}
}

func historyJSON(t *testing.T, visibility, next string, commits ...map[string]any) string {
	t.Helper()
	response := map[string]any{"data": map[string]any{"repository": map[string]any{
		"isPrivate": visibility == "PRIVATE", "visibility": visibility,
		"defaultBranchRef": map[string]any{"target": map[string]any{
			"history": map[string]any{"nodes": commits, "pageInfo": map[string]any{"hasNextPage": next != "", "endCursor": next}},
		}},
	}}}
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestCoauthorsPaginationAndDeduplication(t *testing.T) {
	self, alice := actorJSON("You", "you@example.com"), actorJSON("Alice", "alice@example.com")
	bob, unlinked := actorJSON("Bob", "bob@example.com"), actorJSON("", "pair@example.com")
	shared := commitJSON("shared", "2025-01-01T00:00:00Z", self, alice, alice, unlinked)
	secondary := commitJSON("secondary", "2025-12-31T23:59:59Z", alice)
	secondary["authors"].(map[string]any)["pageInfo"] = map[string]any{"hasNextPage": true, "endCursor": "authors-next"}
	calls := 0
	client := testAPI(t, func(request graphQLRequest) (string, error) {
		calls++
		query := compactQuery(request.Query)
		if strings.Contains(query, "RecapCommitAuthors") {
			if request.Variables["id"] != "id-secondary" || request.Variables["after"] != "authors-next" || !strings.Contains(query, "$id:ID!") {
				t.Fatalf("invalid author pagination: %+v", request)
			}
			data, _ := json.Marshal(map[string]any{"data": map[string]any{"node": map[string]any{"authors": map[string]any{
				"nodes": []any{self, bob}, "pageInfo": map[string]any{"hasNextPage": false},
			}}}})
			return string(data), nil
		}
		for _, selection := range []string{"$since:GitTimestamp!", "$until:GitTimestamp!", "...onCommit", "authors(first:100)", "committedDate", "oid"} {
			if !strings.Contains(query, selection) {
				t.Fatalf("missing %q: %s", selection, query)
			}
		}
		if request.Variables["since"] != "2025-01-01T00:00:00Z" || request.Variables["until"] != "2025-12-31T23:59:59Z" {
			t.Fatal("incorrect year boundaries")
		}
		if request.Variables["name"] == "fork" {
			return historyJSON(t, "PUBLIC", "", shared), nil
		}
		if request.Variables["after"] == nil {
			return historyJSON(t, "PUBLIC", "next", shared,
				commitJSON("unrelated", "2025-06-01T00:00:00Z", alice, bob),
				commitJSON("old", "2024-12-31T23:59:59Z", self, bob),
				commitJSON("future", "2026-01-01T00:00:00Z", self, bob)), nil
		}
		if request.Variables["after"] != "next" {
			t.Fatal("incorrect history cursor")
		}
		return historyJSON(t, "PUBLIC", "", secondary, shared, commitJSON("solo", "2025-02-01T00:00:00Z", self)), nil
	})
	result := FetchCoauthors(context.Background(), client, Options{Year: 2025}, "you", []Repo{
		{NameWithOwner: "you/app"}, {NameWithOwner: "YOU/APP"}, {NameWithOwner: "you/fork"},
		{NameWithOwner: "you/private", IsPrivate: true}, {NameWithOwner: "you/internal", Visibility: "INTERNAL"},
	})
	want := map[string]int{"@alice": 2, "@bob": 1, "pair@example.com": 1}
	if result.Err != nil || !reflect.DeepEqual(result.Coauthors, want) || calls != 4 {
		t.Fatalf("unexpected result (%d calls): %+v", calls, result)
	}
	if len(result.Events) != 0 || Summarize(result.Events).Total != 0 {
		t.Fatal("coauthors must not inflate contribution totals")
	}
}

func TestCoauthorsPrivacyAndPartialFailure(t *testing.T) {
	for _, includePrivate := range []bool{false, true} {
		calls := 0
		client := testAPI(t, func(request graphQLRequest) (string, error) {
			calls++
			if request.Variables["name"] == "broken" {
				return "", errors.New("rate limit")
			}
			visibility := "PUBLIC"
			if request.Variables["name"] == "private" || request.Variables["name"] == "changed" {
				visibility = "PRIVATE"
			}
			return historyJSON(t, visibility, "", commitJSON(request.Variables["name"].(string), "2025-05-01T00:00:00Z",
				actorJSON("you", "you@example.com"), actorJSON("alice", "alice@example.com"))), nil
		})
		result := FetchCoauthors(context.Background(), client, Options{Year: 2025, IncludePrivate: includePrivate}, "you", []Repo{
			{NameWithOwner: "you/app"}, {NameWithOwner: "you/broken"},
			{NameWithOwner: "you/private", IsPrivate: true}, {NameWithOwner: "you/changed"},
		})
		wantCount, wantCalls := 1, 3
		if includePrivate {
			wantCount, wantCalls = 3, 4
		}
		if result.Err == nil || result.Coauthors["@alice"] != wantCount || calls != wantCalls {
			t.Fatalf("private=%v: calls=%d result=%+v", includePrivate, calls, result)
		}
	}
}

func TestCoauthorsBadPagination(t *testing.T) {
	calls := 0
	client := testAPI(t, func(request graphQLRequest) (string, error) {
		calls++
		return historyJSON(t, "PUBLIC", "stuck", commitJSON("one", "2025-02-01T00:00:00Z",
			actorJSON("you", "you@example.com"), actorJSON("alice", "alice@example.com"))), nil
	})
	result := FetchCoauthors(context.Background(), client, Options{Year: 2025}, "you", []Repo{{NameWithOwner: "you/app"}})
	if result.Err == nil || calls != 2 || result.Coauthors["@alice"] != 1 {
		t.Fatalf("bad cursor should stop with partial results: %+v", result)
	}
}

func TestCoauthorsEmptyRepository(t *testing.T) {
	client := testAPI(t, func(request graphQLRequest) (string, error) {
		return `{"data":{"repository":{"isPrivate":false,"visibility":"PUBLIC","defaultBranchRef":null}}}`, nil
	})
	result := FetchCoauthors(context.Background(), client, Options{Year: 2025}, "you", []Repo{{NameWithOwner: "you/app"}})
	if result.Err != nil || len(result.Coauthors) != 0 {
		t.Fatalf("empty repository should be empty: %+v", result)
	}
}
