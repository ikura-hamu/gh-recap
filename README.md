# gh-recap

Your GitHub year, replayed in your terminal. A Go / Bubble Tea v2 GitHub CLI extension with eleven colorful highlights, autoplay, and a browsable breakdown. Large terminal titles, multicolor charts, and a two-column layout on wider terminals make room for the story and the numbers.

## Run

Requires Go 1.25.8+ and [GitHub CLI](https://cli.github.com/).

The layout switches to stacked sections in narrower terminals. The minimum size is calculated from the activity grid and the compact heading and scroll area. Windows below that size show the required dimensions and suspend playback until resized. Quitting remains available.

```sh
gh auth login
go build -o gh-recap .
gh extension install .
gh recap
```

```sh
gh recap 2025                       # another calendar year
gh recap --include-private          # include private repositories you can access
gh recap 2025 --include-private      # flags also work before the year
gh recap --demo                     # fictional activity, no login required
gh recap --demo --no-animation      # no decorative animation; autoplay still works
```

During development, run `go run . --demo`.

The default year is the current local year. API date boundaries use UTC, January 1 through December 31 (or now for the current year). Future years and years before 2008 are rejected.

## Controls

| Key | Action |
| --- | --- |
| `←` / `→`, `h` / `l` | Previous / next highlight |
| `j` / `k` | Halve / double autoplay speed (0.5× to 4×) |
| `Space` | Pause / resume autoplay and background animation |
| `Enter` | Open / close details |
| `Esc` | Return from details |
| `↑` / `↓` | Scroll. In details, `k` / `j` also scroll |
| `Home` / `End` | Scroll to top / bottom |
| `PgUp` / `PgDn`, `Ctrl+B` / `Ctrl+F` | Scroll one page |
| `q`, `Ctrl+C` | Quit |

Autoplay uses eight seconds per page, twelve for the calendar, repository ranking, and languages. Manual navigation resets the timer. Details suspend autoplay without changing your pause preference. Playback waits for the current page's data and finishes on the last page.

Highlights slide horizontally when changing pages: the next page enters from the right, and the previous page enters from the left. The header and controls stay in place. Use `--no-animation` for instant page changes.

## What's in your recap

Overview, contribution calendar, monthly activity, streaks and weekdays, commits, co-authors, pull requests, reviews, active repositories, repository languages, and a closing summary. Details include issues, new repositories, all rankings, daily records, and collection warnings.

Five activity sources load concurrently. Once commit contributions load, a sixth source scans co-authors without blocking other highlights. Results appear as each source completes. Failures preserve collected records and are marked incomplete. Nothing is cached or exported.

Public-only mode filters every record by its repository's visibility; profile totals that could contain private activity are never substituted. `--include-private` includes only activity accessible to your existing `gh` authentication. GitHub permissions, organization SSO, and contribution eligibility can limit the result.

Metrics follow [GitHub's contribution API](https://docs.github.com/en/graphql/reference/users#contributionscollection), rather than counting all pushes or all branches:

- Commits are fetched in monthly windows so more than 100 active days do not truncate a year. GitHub returns at most 100 commit repositories per window; reaching this threshold displays an incompleteness warning.
- PRs and issues count records opened in the selected year; first and popular contributions are included. Connections are paginated.
- Review contributions represent the latest review per reviewed PR, not every review submission or comment.
- The calendar and streaks are rebuilt from the five selected-scope activity sources. They can differ from the GitHub profile graph. Unavailable data is identified in the collection status and details.
- Repository rankings combine commits, PRs, review contributions, issues, and repository creation.
- Languages count the **current primary language of each active repository**, not the language or volume of code you personally wrote that year.
- Co-authors are ranked by shared commits with you, whether you are the primary author or a secondary author. The scan covers default branches of repositories found in your commit contributions, using commit dates within the selected UTC year. Each commit counts once per other author; duplicate commit IDs across forks are counted once. Linked accounts are grouped by login; unlinked authors by email, or name when email is unavailable. Co-author counts do not add to contribution totals. Other branches, including `gh-pages`, are outside this scan. History scanning has a two-minute timeout and marks partial results incomplete.

## Verify

```sh
go test ./...
go vet ./...
```

The TUI uses Charm Bubbles for its viewport, progress bars, loading spinner, key bindings, and width-aware help.

All repository tests run without GitHub authentication or API access. An in-memory HTTP transport exercises go-gh’s generated GraphQL queries and response decoding, covering privacy filtering, pagination, partial failures, and monthly commit windows; other tests cover leap-year streaks, CLI parsing, and playback controls. Releases use the existing GitHub extension precompile workflow.

## Code layout

Start with `main.go` and `options.go` for startup and argument parsing.

- `internal/recap/types.go` defines the activity records shared by collection and presentation.
- `internal/recap/queries.go` defines GraphQL selections as Go structs and tags. `github.go` supplies the authenticated go-gh client; `fetch.go` handles date windows, pagination, and visibility filtering.
- `internal/recap/stats.go` summarizes records; `demo.go` generates offline sample data. JSON responses used by tests live in `testdata/`.
- `internal/tui/model.go` routes Bubble Tea messages and collects results. `controls.go` owns keyboard actions and help; `playback.go` defines page order, data prerequisites, and timing.
- `internal/tui/pages.go` contains page text and metrics. `view.go` arranges the header, content viewport, and footer. `charts.go`, `design.go`, and `banner.go` handle charts, shared styling, and block lettering.

To add a highlight, add its page definition in `playback.go` and its content in `pages.go`. To change a GitHub selection, edit the corresponding struct in `queries.go` and its collection code in `fetch.go`.
