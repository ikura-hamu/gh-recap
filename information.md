# Data sources and metrics

gh-recap uses the [GitHub GraphQL contribution API](https://docs.github.com/en/graphql/reference/users#contributionscollection). The year is measured in UTC. The default scope is public repositories. Use `--include-private` to include private activity your GitHub CLI account can access. Partial results are marked incomplete.

- **Activity** includes contribution commits, opened pull requests and issues, review contributions, and new repositories. Reviews count once per reviewed pull request. The activity calendar and streaks use these records and may differ from your GitHub profile.
- **Repository rankings** combine the five activity types. Language rankings count each active repository's current primary language.
- **Co-authors** are ranked by shared commits dated in the selected year, using the default branches of repositories found in your commit contributions.
