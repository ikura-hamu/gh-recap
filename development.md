# Development

Use the Go version in [go.mod](go.mod) or newer.

```sh
go run . --demo
go test ./...
go vet ./...
```

The demo and tests run without GitHub authentication. To try your local build as a GitHub CLI extension, run `go build -o gh-recap .` and `gh extension install .`. See `gh recap --help` for commands and flags.

- [CLI options](options.go)
- [GitHub data collection and metrics](internal/recap/)
- [Terminal UI and PNG export](internal/tui/)
- [Bundled JetBrains Mono font license](internal/tui/fonts/OFL.txt)
