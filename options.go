package main

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/ikura-hamu/gh-recap/internal/recap"
	"github.com/ikura-hamu/gh-recap/internal/tui"
)

type commandOptions struct {
	recap.Options
	command     string
	output      string
	fontLicense bool
}

func parseOptions(args []string, out io.Writer) (commandOptions, error) {
	currentYear := time.Now().Year()
	options := commandOptions{Options: recap.Options{Year: currentYear}}
	flags := flag.NewFlagSet("gh recap", flag.ContinueOnError)
	flags.SetOutput(out)
	flags.BoolVar(&options.IncludePrivate, "include-private", false, "include accessible private repositories")
	flags.BoolVar(&options.Demo, "demo", false, "explore fictional activity without authentication")
	flags.BoolVar(&options.NoAnimation, "no-animation", false, "disable decorative animations (autoplay stays enabled)")
	flags.BoolVar(&options.DPGK, "dpgk", false, "enable flashy Dopagaki effects (ignored with --no-animation)")
	flags.BoolVar(&options.fontLicense, "font-license", false, "print the bundled JetBrains Mono license")
	flags.Func("output", "export file path (defaults to gh-recap_YEAR with the format's extension)", func(path string) error {
		if path == "" {
			return fmt.Errorf("output path cannot be empty")
		}
		options.output = path
		return nil
	})
	flags.Usage = func() {
		fmt.Fprintln(out, `Usage: gh recap [YEAR] [OPTIONS]
       gh recap {md|image|video|short-text} [YEAR] [OPTIONS]

YEAR defaults to the current year. Without a subcommand, gh recap opens the
terminal recap. Press s on ENCORE to save a PNG.

Subcommands:
  md          Export activity details as Markdown (.md)
  image       Export the recap summary as a PNG (.png)
  video       Record the full recap as an MP4 (.mp4). Requires Bash, vhs, ttyd, and ffmpeg.
  short-text  Print a recap of up to 140 characters for social media (e.g. X).
              Also tries to copy it to the clipboard.

Options:`)
		flags.PrintDefaults()
	}
	flagArgs := []string{}
	var year *string
	positionalOnly := false
	// flag stops at the first positional argument; reorder flags to also accept YEAR first.
	for i := 0; i < len(args); i++ {
		argument := args[i]
		if !positionalOnly && argument == "--" {
			positionalOnly = true
			continue
		}
		if !positionalOnly && (argument == "--output" || argument == "-output") {
			if i+1 == len(args) {
				return options, fmt.Errorf("%s requires a file path", argument)
			}
			flagArgs = append(flagArgs, argument, args[i+1])
			i++
			continue
		}
		switch argument {
		case "md", "image", "video", "short-text":
			if options.command != "" {
				return options, fmt.Errorf("specify only one export command")
			}
			options.command = argument
			continue
		}
		if !positionalOnly && len(argument) > 1 && argument[0] == '-' {
			flagArgs = append(flagArgs, argument)
			continue
		}
		if year != nil {
			return options, fmt.Errorf("specify only one year")
		}
		year = &argument
	}
	if err := flags.Parse(flagArgs); err != nil {
		return options, err
	}
	if year != nil {
		parsedYear, err := strconv.Atoi(*year)
		if err != nil || parsedYear < 2008 || parsedYear > currentYear {
			return options, fmt.Errorf("year must be between 2008 and %d", currentYear)
		}
		options.Year = parsedYear
	}
	if options.output != "" && (options.command == "" || options.command == "short-text") {
		return options, fmt.Errorf("--output requires the md, image, or video command")
	}
	if options.output == "" {
		switch options.command {
		case "md":
			options.output = fmt.Sprintf("gh-recap_%d.md", options.Year)
		case "image":
			options.output = tui.ImageFilename(options.Year)
		case "video":
			options.output = fmt.Sprintf("gh-recap_%d.mp4", options.Year)
		}
	}
	return options, nil
}
