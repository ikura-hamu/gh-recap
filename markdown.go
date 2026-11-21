package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ikura-hamu/gh-recap/internal/recap"
	"github.com/ikura-hamu/gh-recap/internal/tui"
)

func exportMarkdown(ctx context.Context, options commandOptions, api recap.API) error {
	report, warnings, err := tui.CollectMarkdown(ctx, options.Options, api)
	printCollectionWarnings(warnings)
	if err != nil {
		return err
	}
	output, err := filepath.Abs(options.output)
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(output), ".gh-recap-md-*")
	if err != nil {
		return fmt.Errorf("create Markdown output: %w", err)
	}
	defer os.Remove(temp.Name())
	_, writeErr := temp.WriteString(report)
	closeErr := temp.Close()
	if writeErr != nil {
		return fmt.Errorf("write Markdown output: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close Markdown output: %w", closeErr)
	}
	if err := os.Rename(temp.Name(), output); err != nil {
		return fmt.Errorf("save Markdown output: %w", err)
	}
	fmt.Fprintf(os.Stdout, "Saved %s\n", output)
	return nil
}
