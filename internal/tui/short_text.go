package tui

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/ikura-hamu/gh-recap/internal/recap"
)

// CollectShortText returns a shareable summary of the selected year's activity.
func CollectShortText(ctx context.Context, options recap.Options, api recap.API) (string, []string, error) {
	m, err := collect(ctx, options, api, false)
	if err != nil {
		return "", m.CollectionWarnings(), err
	}
	return shortText(options.Year, recap.Summarize(m.events)), m.CollectionWarnings(), nil
}

func shortText(year int, stats recap.Stats) string {
	const limit = 140
	contributions := "contributions"
	if stats.Total == 1 {
		contributions = "contribution"
	}
	text := fmt.Sprintf("In %d, I made %d GitHub %s.", year, stats.Total, contributions)
	ranking := recap.Ranked(stats.Repos)
	if len(ranking) == 0 {
		return text + " #ghrecap"
	}

	prefix, suffix := text+" My top repo was ", ". #ghrecap"
	name := ranking[0].Name
	budget := limit - utf8.RuneCountInString(prefix+suffix)
	if utf8.RuneCountInString(name) > budget {
		suffix = "… #ghrecap"
		budget = limit - utf8.RuneCountInString(prefix+suffix)
		name = string([]rune(name)[:budget])
	}
	return prefix + name + suffix
}
