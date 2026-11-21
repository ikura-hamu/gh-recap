package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if options.video != "" {
		return recordVideo(ctx, options)
	}
	donePath := os.Getenv(videoDoneEnv)
	options.Recording = donePath != ""
	final, err := tea.NewProgram(tui.New(ctx, options.Options, &recap.GitHub{})).Run()
	if err != nil || donePath == "" {
		return err
	}
	if err := final.(tui.Model).RecordingResult(); err != nil {
		return err
	}
	return os.WriteFile(donePath, nil, 0600)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gh recap:", err)
		os.Exit(1)
	}
}
