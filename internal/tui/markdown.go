package tui

import (
	"context"
	"fmt"
	"html"
	"net/url"
	"slices"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/ikura-hamu/gh-recap/internal/recap"
)

// CollectMarkdown runs the same collection as the recap without starting a terminal UI.
func CollectMarkdown(ctx context.Context, options recap.Options, api recap.API) (report string, warnings []string, err error) {
	m, err := collect(ctx, options, api, true)
	if err != nil {
		return "", m.CollectionWarnings(), err
	}
	return m.markdownDetails(), m.CollectionWarnings(), nil
}

func (m Model) markdownDetails() string {
	stats := recap.Summarize(m.events)
	var body strings.Builder
	fmt.Fprintf(&body, "# GH Recap %d · @%s\n\n%s\n", m.options.Year, markdownText(m.login), metricNotes)
	fmt.Fprintf(&body, "\n## %s\n\n| Activity | Count |\n| --- | ---: |\n", activityHeading)
	for _, kind := range recap.Kinds {
		fmt.Fprintf(&body, "| %s | %s |\n", kind, m.number(kind, stats.Kinds[kind]))
	}

	fmt.Fprintf(&body, "\n## %s\n\n", reposHeading)
	repoURLs := make(map[string]string, len(stats.Repos))
	for _, event := range m.events {
		if event.Repo.URL != "" {
			repoURLs[event.Repo.NameWithOwner] = event.Repo.URL
		}
	}
	markdownTable(&body, "Repository", "Contributions", recap.Ranked(stats.Repos), emptyRanking, func(name string) string {
		return markdownRepoLink(name, repoURLs[name])
	})

	fmt.Fprintf(&body, "\n## %s\n\n", languagesHeading)
	markdownTable(&body, "Language", "Repositories", recap.Ranked(stats.Languages), emptyRanking, markdownText)

	fmt.Fprintf(&body, "\n## %s\n\n", coauthorsHeading)
	markdownTable(&body, "Co-author", "Shared commits", recap.Ranked(m.coauthors), noCoauthors, markdownText)
	if m.failed[recap.KindCoauthors] {
		fmt.Fprint(&body, "\nIncomplete results.\n")
	}
	fmt.Fprintf(&body, "\n%s\n", strings.TrimSpace(coauthorsNotes))

	fmt.Fprintf(&body, "\n## %s\n\n", dailyHeading)
	markdownTable(&body, "Date", "Contributions", dailyActivity(stats, m.options.Year), emptyRanking, markdownText)

	if len(m.warnings) > 0 {
		fmt.Fprintf(&body, "\n## %s\n\n", warningsHeading)
		warnings := slices.Clone(m.warnings)
		slices.Sort(warnings)
		for _, warning := range warnings {
			fmt.Fprintf(&body, "- %s\n", markdownText(warning))
		}
	}
	return body.String()
}

func markdownTable(body *strings.Builder, label, count string, entries []recap.Entry, empty string, formatName func(string) string) {
	if len(entries) == 0 {
		fmt.Fprintln(body, empty)
		return
	}
	fmt.Fprintf(body, "| %s | %s |\n| --- | ---: |\n", label, count)
	for _, entry := range entries {
		fmt.Fprintf(body, "| %s | %d |\n", formatName(entry.Name), entry.Count)
	}
}

var markdownLinkEscapes = strings.NewReplacer("(", "%28", ")", "%29", "|", "%7C")

func markdownRepoLink(name, target string) string {
	link, err := url.Parse(target)
	if err != nil || link.Host == "" || (link.Scheme != "https" && link.Scheme != "http") {
		link = &url.URL{Scheme: "https", Host: "github.com", Path: "/" + name}
	}
	return fmt.Sprintf("[%s](%s)", markdownText(name), markdownLinkEscapes.Replace(link.String()))
}

var markdownEscapes = strings.NewReplacer(
	"\\", "\\\\", "`", "\\`", "*", "\\*", "_", "\\_", "[", "\\[", "]", "\\]", "|", "\\|", "~", "\\~",
)

func markdownText(value string) string {
	return markdownEscapes.Replace(html.EscapeString(strings.Join(strings.Fields(ansi.Strip(value)), " ")))
}
