package recap

import (
	"context"
	"fmt"
	"strings"

	graphql "github.com/cli/shurcooL-graphql"
)

func fetchPopularPR(ctx context.Context, client API, options Options, id string) (*PopularPR, error) {
	var query popularPRQuery
	if err := client.QueryWithContext(ctx, "RecapPopularPR", &query, map[string]any{"id": graphql.ID(id)}); err != nil {
		return nil, err
	}
	node := query.Node.PullRequest
	if node.Repository.NameWithOwner == "" || !node.Repository.included(options.IncludePrivate) {
		return nil, fmt.Errorf("pull request is no longer available in this scope")
	}
	date, _, _ := strings.Cut(node.CreatedAt, "T")
	pr := &PopularPR{
		Repository:   node.Repository.NameWithOwner,
		Number:       node.Number,
		Title:        node.Title,
		Date:         date,
		URL:          node.URL,
		Description:  node.BodyText,
		CommentCount: node.Comments.TotalCount,
		MorePeople:   node.Participants.PageInfo.HasNextPage,
	}
	seen := make(map[string]bool)
	addPerson := func(actor *pullRequestActor) {
		if actor != nil && actor.Login != "" && !seen[actor.Login] {
			seen[actor.Login] = true
			pr.People = append(pr.People, actor.Login)
		}
	}
	addPerson(node.Author)
	for _, comment := range node.Comments.Nodes {
		var author string
		if comment.Author != nil {
			author = comment.Author.Login
		}
		pr.Comments = append(pr.Comments, PRComment{Author: author, Body: comment.BodyText})
		addPerson(comment.Author)
	}
	for i := range node.Participants.Nodes {
		addPerson(&node.Participants.Nodes[i])
	}
	return pr, nil
}
