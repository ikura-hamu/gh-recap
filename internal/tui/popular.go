package tui

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"
	"github.com/ikura-hamu/gh-recap/internal/recap"
)

func (m Model) popularPRContent() pageContent {
	content := pageContent{Title: popularPRTitle}
	switch {
	case !m.loaded[recap.KindPullRequests]:
		content.Body = loading
	case m.popularPR == nil && (m.popularPRFailed || m.failed[recap.KindPullRequests]):
		content.Body = popularPRUnavailable
	case m.popularPR == nil:
		content.Body = popularPRMissing
	default:
		content.Body = m.popularPRBody(m.popularPR)
	}
	return content
}

func (m Model) popularPRBody(pr *recap.PopularPR) string {
	heading := ink(m.color()).Bold(true)
	var body strings.Builder
	fmt.Fprintf(&body, "%s #%d · %s\n%s\n",
		prExcerpt(pr.Repository, 150), pr.Number, prExcerpt(pr.Date, 20),
		heading.Render(prExcerpt(pr.Title, 220)))
	if pr.URL != "" {
		fmt.Fprintln(&body, prExcerpt(pr.URL, 300))
	}
	if m.failed[recap.KindPullRequests] {
		fmt.Fprintln(&body, muted.Render(strings.TrimSpace(statusIncomplete)))
	}

	description := prExcerpt(pr.Description, 600)
	if description == "" {
		description = prNoDescription
	}
	fmt.Fprintf(&body, "\n%s\n%s\n", heading.Render(prDescription), description)

	fmt.Fprintf(&body, "\n%s\n", heading.Render(fmt.Sprintf("%s · %d", prComments, pr.CommentCount)))
	if len(pr.Comments) == 0 {
		fmt.Fprintln(&body, prNoComments)
	}
	for _, comment := range pr.Comments {
		author := prUnknownAuthor
		if comment.Author != "" {
			author = "@" + prExcerpt(comment.Author, 60)
		}
		fmt.Fprintf(&body, "%s  %s\n", mint.Render(author), prExcerpt(comment.Body, 180))
	}

	fmt.Fprintf(&body, "\n%s\n", heading.Render(prPeople))
	people := make([]string, 0, len(pr.People))
	for _, login := range pr.People {
		people = append(people, "@"+prExcerpt(login, 60))
	}
	if len(people) == 0 {
		fmt.Fprintln(&body, prNoPeople)
	} else {
		fmt.Fprint(&body, strings.Join(people, ", "))
		if pr.MorePeople {
			body.WriteString(", …")
		}
		body.WriteByte('\n')
	}
	fmt.Fprintf(&body, "\n%s", muted.Render(popularPRNotes))
	return body.String()
}

func prExcerpt(value string, limit int) string {
	// GitHub titles and bodies are external input to the terminal.
	value = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, ansi.Strip(value))
	runes := []rune(strings.Join(strings.Fields(value), " "))
	if len(runes) <= limit {
		return string(runes)
	}
	return strings.TrimSpace(string(runes[:limit])) + "…"
}
