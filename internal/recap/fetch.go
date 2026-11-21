package recap

import (
	"context"
	"fmt"
	"time"

	graphql "github.com/cli/shurcooL-graphql"
)

func Fetch(ctx context.Context, client API, options Options, kind string) Result {
	start, end := yearWindow(options.Year, time.Now())
	if end.Before(start) {
		return Result{Kind: kind}
	}

	if kind == KindCommits {
		return fetchCommits(ctx, client, options, start, end)
	}
	return fetchContributions(ctx, client, options, kind, start, end)
}

func yearWindow(year int, now time.Time) (start, end time.Time) {
	start = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end = start.AddDate(1, 0, 0).Add(-time.Second)
	if now.Before(end) {
		end = now.UTC()
	}
	return start, end
}

func dateVariables(start, end time.Time) map[string]any {
	return map[string]any{"from": DateTime{start}, "to": DateTime{end}}
}

func fetchCommits(ctx context.Context, client API, options Options, start, end time.Time) Result {
	result := Result{Kind: KindCommits}
	// Monthly windows have at most 31 daily records per repository, avoiding
	// the connection's 100-node limit without cutting off a year's active days.
	for month := start; !month.After(end); month = month.AddDate(0, 1, 0) {
		until := month.AddDate(0, 1, 0).Add(-time.Second)
		if until.After(end) {
			until = end
		}
		var query contributionsQuery[commitCollection]
		if err := client.QueryWithContext(ctx, "RecapCommits", &query, dateVariables(month, until)); err != nil {
			result.Err = err
			return result
		}
		repositories := query.Viewer.ContributionsCollection.CommitContributionsByRepository
		if len(repositories) == 100 {
			result.Warnings = append(result.Warnings, month.Format("January")+": reached 100 commit repositories; totals may be incomplete.")
		}
		for _, repository := range repositories {
			for _, node := range repository.Contributions.Nodes {
				result.addContribution(options, node.OccurredAt, node.CommitCount, repository.Repository)
			}
			if repository.Contributions.PageInfo.HasNextPage {
				result.Warnings = append(result.Warnings, "Unexpected daily commit truncation in "+month.Format("January"))
			}
		}
	}
	return result
}

func fetchContributions(ctx context.Context, client API, options Options, kind string, start, end time.Time) Result {
	result := Result{Kind: kind}
	var bestPR pullRequestCandidate
	variables := dateVariables(start, end)
	// A typed nil generates a nullable String variable for the first page.
	var cursor *graphql.String
	seenCursors := make(map[string]bool)
	for {
		variables["after"] = cursor
		page, err := fetchContributionPage(ctx, client, options, &result, variables, &bestPR)
		if err != nil {
			result.Err = err
			break
		}
		if !page.HasNextPage {
			break
		}
		if page.EndCursor == "" || seenCursors[page.EndCursor] {
			result.Err = fmt.Errorf("pagination did not advance")
			break
		}
		seenCursors[page.EndCursor] = true
		next := graphql.String(page.EndCursor)
		cursor = &next
	}
	if bestPR.ID != "" {
		result.PopularPR, result.PopularPRErr = fetchPopularPR(ctx, client, options, bestPR.ID)
		if result.PopularPRErr != nil {
			result.Warnings = append(result.Warnings, "Popular PR: "+result.PopularPRErr.Error())
		}
	}
	return result
}

func fetchContributionPage(ctx context.Context, client API, options Options, result *Result, variables map[string]any, bestPR *pullRequestCandidate) (pageInfo, error) {
	switch result.Kind {
	case KindPullRequests:
		var query contributionsQuery[pullRequestCollection]
		if err := client.QueryWithContext(ctx, "RecapPullRequests", &query, variables); err != nil {
			return pageInfo{}, err
		}
		connection := query.Viewer.ContributionsCollection.PullRequestContributions
		for _, node := range connection.Nodes {
			pr := node.PullRequest
			if result.addContribution(options, node.OccurredAt, 1, pr.Repository) &&
				pr.ID != "" && (bestPR.ID == "" || pr.Comments.TotalCount > bestPR.Comments.TotalCount) {
				*bestPR = pr
			}
		}
		return connection.PageInfo, nil
	case KindIssues:
		var query contributionsQuery[issueCollection]
		if err := client.QueryWithContext(ctx, "RecapIssues", &query, variables); err != nil {
			return pageInfo{}, err
		}
		connection := query.Viewer.ContributionsCollection.IssueContributions
		for _, node := range connection.Nodes {
			result.addContribution(options, node.OccurredAt, 1, node.Issue.Repository)
		}
		return connection.PageInfo, nil
	case KindReviews:
		var query contributionsQuery[reviewCollection]
		if err := client.QueryWithContext(ctx, "RecapReviews", &query, variables); err != nil {
			return pageInfo{}, err
		}
		connection := query.Viewer.ContributionsCollection.PullRequestReviewContributions
		for _, node := range connection.Nodes {
			result.addContribution(options, node.OccurredAt, 1, node.Repository)
		}
		return connection.PageInfo, nil
	case KindRepositories:
		var query contributionsQuery[repositoryCollection]
		if err := client.QueryWithContext(ctx, "RecapRepositories", &query, variables); err != nil {
			return pageInfo{}, err
		}
		connection := query.Viewer.ContributionsCollection.RepositoryContributions
		for _, node := range connection.Nodes {
			result.addContribution(options, node.OccurredAt, 1, node.Repository)
		}
		return connection.PageInfo, nil
	default:
		return pageInfo{}, fmt.Errorf("unknown activity: %s", result.Kind)
	}
}

func (r *Result) addContribution(options Options, occurredAt string, count int, repository Repo) bool {
	if !repository.included(options.IncludePrivate) {
		return false
	}
	if len(occurredAt) < len(time.DateOnly) || repository.NameWithOwner == "" {
		return false
	}
	r.Events = append(r.Events, Event{
		Date:  occurredAt[:len(time.DateOnly)],
		Kind:  r.Kind,
		Count: count,
		Repo:  repository,
	})
	return true
}
