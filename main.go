package main

import (
	"context"
	"flag"
	"fmt"
	"os"

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err = tea.NewProgram(tui.New(ctx, options, &recap.GitHub{})).Run()
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gh recap:", err)
		os.Exit(1)
	}
}
