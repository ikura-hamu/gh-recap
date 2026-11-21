package recap

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	graphql "github.com/cli/shurcooL-graphql"
)

type GitTimestamp struct{ time.Time }

type commitActor struct {
	Name  string
	Email string
	User  *struct{ Login string }
}

type authorConnection = contributionConnection[commitActor]

type sharedCommit struct {
	ID            string
	OID           string
	CommittedDate string
	Authors       authorConnection `graphql:"authors(first: 100)"`
}

type coauthorHistoryQuery struct {
	Repository *struct {
		IsPrivate        bool
		Visibility       string
		DefaultBranchRef *struct {
			Target struct {
				Commit struct {
					History contributionConnection[sharedCommit] `graphql:"history(first: 100, after: $after, since: $since, until: $until)"`
				} `graphql:"... on Commit"`
			}
		}
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type coauthorAuthorsQuery struct {
	Node struct {
		Commit struct {
			Authors authorConnection `graphql:"authors(first: 100, after: $after)"`
		} `graphql:"... on Commit"`
	} `graphql:"node(id: $id)"`
}

// FetchCoauthors scans default-branch histories without an author filter: a
// viewer can be a secondary author, and must still be included. OIDs deduplicate
// commits shared by forks. This supplementary result never contributes Events.
func FetchCoauthors(ctx context.Context, client API, options Options, login string, repositories []Repo) Result {
	result := Result{Kind: KindCoauthors, Coauthors: make(map[string]int)}
	start, end := yearWindow(options.Year, time.Now())
	if end.Before(start) {
		return result
	}
	seenRepos, seenCommits := make(map[string]bool), make(map[string]bool)
	for _, repository := range repositories {
		key := strings.ToLower(repository.NameWithOwner)
		if !repository.included(options.IncludePrivate) || seenRepos[key] {
			continue
		}
		seenRepos[key] = true
		if err := scanCoauthorRepository(ctx, client, options, login, repository, start, end, seenCommits, result.Coauthors); err != nil {
			result.Err = errors.Join(result.Err, fmt.Errorf("%s: %w", repository.NameWithOwner, err))
		}
		if ctx.Err() != nil {
			break
		}
	}
	return result
}

func scanCoauthorRepository(ctx context.Context, client API, options Options, login string, repository Repo, start, end time.Time, seen map[string]bool, counts map[string]int) error {
	owner, name, ok := strings.Cut(repository.NameWithOwner, "/")
	if !ok || owner == "" || name == "" {
		return fmt.Errorf("invalid repository name")
	}
	var cursor *graphql.String
	cursors := make(map[string]bool)
	for {
		var query coauthorHistoryQuery
		variables := map[string]any{
			"owner": graphql.String(owner), "name": graphql.String(name),
			"since": GitTimestamp{start}, "until": GitTimestamp{end}, "after": cursor,
		}
		if err := client.QueryWithContext(ctx, "RecapCoauthors", &query, variables); err != nil {
			return err
		}
		if query.Repository == nil {
			return fmt.Errorf("repository is unavailable")
		}
		// Recheck visibility: it may have changed since contribution collection.
		visibility := Repo{IsPrivate: query.Repository.IsPrivate, Visibility: query.Repository.Visibility}
		if !visibility.included(options.IncludePrivate) || query.Repository.DefaultBranchRef == nil {
			return nil
		}
		history := query.Repository.DefaultBranchRef.Target.Commit.History
		for _, commit := range history.Nodes {
			date, err := time.Parse(time.RFC3339, commit.CommittedDate)
			if err != nil || date.Before(start) || date.After(end) || commit.OID == "" || seen[commit.OID] {
				continue
			}
			authors, err := allCommitAuthors(ctx, client, commit)
			if err != nil {
				return err
			}
			seen[commit.OID] = true
			addCoauthors(counts, login, authors)
		}
		if !history.PageInfo.HasNextPage {
			return nil
		}
		next := history.PageInfo.EndCursor
		if next == "" || cursors[next] {
			return fmt.Errorf("commit pagination did not advance")
		}
		cursors[next] = true
		value := graphql.String(next)
		cursor = &value
	}
}

func allCommitAuthors(ctx context.Context, client API, commit sharedCommit) ([]commitActor, error) {
	authors := commit.Authors.Nodes
	page := commit.Authors.PageInfo
	cursors := make(map[string]bool)
	for page.HasNextPage {
		if page.EndCursor == "" || cursors[page.EndCursor] {
			return nil, fmt.Errorf("author pagination did not advance")
		}
		cursors[page.EndCursor] = true
		var query coauthorAuthorsQuery
		variables := map[string]any{"id": graphql.ID(commit.ID), "after": graphql.String(page.EndCursor)}
		if err := client.QueryWithContext(ctx, "RecapCommitAuthors", &query, variables); err != nil {
			return nil, err
		}
		authors = append(authors, query.Node.Commit.Authors.Nodes...)
		page = query.Node.Commit.Authors.PageInfo
	}
	return authors, nil
}

func addCoauthors(counts map[string]int, login string, authors []commitActor) {
	foundViewer := false
	for _, author := range authors {
		if author.User != nil && strings.EqualFold(author.User.Login, login) {
			foundViewer = true
			break
		}
	}
	if !foundViewer {
		return
	}
	unique := make(map[string]bool)
	for _, author := range authors {
		var key string
		if author.User != nil && author.User.Login != "" {
			if strings.EqualFold(author.User.Login, login) {
				continue
			}
			key = "@" + strings.ToLower(author.User.Login)
		} else if author.Email != "" {
			key = strings.ToLower(author.Email)
		} else if author.Name != "" {
			key = author.Name + " (unlinked)"
		} else {
			continue
		}
		if !unique[key] {
			counts[key]++
			unique[key] = true
		}
	}
}
