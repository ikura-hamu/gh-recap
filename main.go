package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/ikura-hamu/gh-recap/internal/recap"
	"github.com/ikura-hamu/gh-recap/internal/tui"
)

func run() error {
	options, err := parseOptions(os.Args[1:], os.Stderr)
	if err == flag.ErrHelp {
		return nil
	}
	if err != nil {
		return err
	}
	if options.fontLicense {
		_, err := fmt.Fprint(os.Stdout, tui.ImageFontLicense)
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	switch options.command {
	case "video":
		return recordVideo(ctx, options)
	case "md":
		return exportMarkdown(ctx, options, &recap.GitHub{})
	case "image":
		output, err := filepath.Abs(options.output)
		if err != nil {
			return err
		}
		warnings, err := tui.ExportImage(ctx, options.Options, &recap.GitHub{}, output)
		printCollectionWarnings(warnings)
		if err == nil {
			fmt.Fprintf(os.Stdout, "Saved %s\n", output)
		}
		return err
	}
	donePath := os.Getenv(videoDoneEnv)
	options.Recording = donePath != ""
	final, err := tea.NewProgram(tui.New(ctx, options.Options, &recap.GitHub{})).Run()
	model, ok := final.(tui.Model)
	if ok {
		printCollectionWarnings(model.CollectionWarnings())
	}
	if err != nil || donePath == "" {
		return err
	}
	if err := model.RecordingResult(); err != nil {
		return err
	}
	return os.WriteFile(donePath, nil, 0600)
}

func printCollectionWarnings(warnings []string) {
	for _, warning := range warnings {
		fmt.Fprintln(os.Stderr, "gh recap: warning:", warning)
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gh recap:", err)
		os.Exit(1)
	}
}
