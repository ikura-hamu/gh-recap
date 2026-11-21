package recap

import "time"

// DateTime's Go type name tells the GraphQL client to use GitHub's DateTime
// scalar for variables; embedding time.Time supplies its JSON encoding.
type DateTime struct{ time.Time }

type viewerQuery struct {
	Viewer struct{ Login string }
}

// Every activity query uses the same viewer and year window. Collection contains
// only the fields needed for that activity, so each source can load independently.
type contributionsQuery[Collection any] struct {
	Viewer struct {
		ContributionsCollection Collection `graphql:"contributionsCollection(from: $from, to: $to)"`
	}
}

type pageInfo struct {
	HasNextPage bool
	EndCursor   string
}

type contributionConnection[Node any] struct {
	Nodes    []Node
	PageInfo pageInfo
}

type commitContribution struct {
	OccurredAt  string
	CommitCount int
}

type commitRepository struct {
	Repository    Repo
	Contributions contributionConnection[commitContribution] `graphql:"contributions(first: 100)"`
}

type commitCollection struct {
	CommitContributionsByRepository []commitRepository `graphql:"commitContributionsByRepository(maxRepositories: 100)"`
}

type pullRequestContribution struct {
	OccurredAt  string
	PullRequest pullRequestCandidate
}

type pullRequestCandidate struct {
	Repository Repo
	ID         string
	Comments   struct{ TotalCount int }
}

type pullRequestActor struct{ Login string }

type popularPRQuery struct {
	Node struct {
		PullRequest struct {
			Repository Repo
			Number     int
			Title      string
			CreatedAt  string
			URL        string
			BodyText   string
			Author     *pullRequestActor
			Comments   struct {
				TotalCount int
				Nodes      []struct {
					Author   *pullRequestActor
					BodyText string
				}
			} `graphql:"comments(first: 3)"`
			Participants struct {
				Nodes    []pullRequestActor
				PageInfo struct{ HasNextPage bool }
			} `graphql:"participants(first: 20)"`
		} `graphql:"... on PullRequest"`
	} `graphql:"node(id: $id)"`
}

type pullRequestCollection struct {
	PullRequestContributions contributionConnection[pullRequestContribution] `graphql:"pullRequestContributions(first: 100, after: $after, excludeFirst: false, excludePopular: false)"`
}

type issueContribution struct {
	OccurredAt string
	Issue      struct{ Repository Repo }
}

type issueCollection struct {
	IssueContributions contributionConnection[issueContribution] `graphql:"issueContributions(first: 100, after: $after, excludeFirst: false, excludePopular: false)"`
}

type repositoryContribution struct {
	OccurredAt string
	Repository Repo
}

type reviewCollection struct {
	PullRequestReviewContributions contributionConnection[repositoryContribution] `graphql:"pullRequestReviewContributions(first: 100, after: $after)"`
}

type repositoryCollection struct {
	RepositoryContributions contributionConnection[repositoryContribution] `graphql:"repositoryContributions(first: 100, after: $after, excludeFirst: false)"`
}
