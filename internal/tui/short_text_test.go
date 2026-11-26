package tui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ikura-hamu/gh-recap/internal/recap"
)

func TestShortText(t *testing.T) {
	for _, tc := range []struct {
		stats recap.Stats
		want  string
	}{
		{recap.Stats{Total: 9, Repos: map[string]int{"you/z": 4, "you/a": 4, "you/m": 1}},
			"In 2025, I made 9 GitHub contributions. My top repo was you/a. #ghrecap"},
		{recap.Stats{Total: 1, Repos: map[string]int{"you/repo": 1}},
			"In 2025, I made 1 GitHub contribution. My top repo was you/repo. #ghrecap"},
		{recap.Stats{}, "In 2025, I made 0 GitHub contributions. #ghrecap"},
	} {
		if got := shortText(2025, tc.stats); got != tc.want {
			t.Errorf("shortText = %q, want %q", got, tc.want)
		}
	}

	longName := "you/" + strings.Repeat("界", 100)
	got := shortText(2025, recap.Stats{Total: 2, Repos: map[string]int{longName: 2}})
	if !utf8.ValidString(got) || utf8.RuneCountInString(got) != 140 || !strings.HasSuffix(got, "… #ghrecap") {
		t.Fatalf("expected a valid 140-character summary ending in the hashtag, got %q", got)
	}
}
