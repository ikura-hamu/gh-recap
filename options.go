package main

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/ikura-hamu/gh-recap/internal/recap"
)

type commandOptions struct {
	recap.Options
	command string
	output  string
}

func parseOptions(args []string, out io.Writer) (commandOptions, error) {
	currentYear := time.Now().Year()
	options := commandOptions{Options: recap.Options{Year: currentYear}}
	flags := flag.NewFlagSet("gh recap", flag.ContinueOnError)
	flags.SetOutput(out)
	flags.BoolVar(&options.IncludePrivate, "include-private", false, "include accessible private repositories")
	flags.BoolVar(&options.Demo, "demo", false, "explore fictional activity without authentication")
	flags.BoolVar(&options.NoAnimation, "no-animation", false, "disable decorative animations (autoplay stays enabled)")
	flags.Func("output", "output file path for md or video", func(path string) error {
		if path == "" {
			return fmt.Errorf("output path cannot be empty")
		}
		options.output = path
		return nil
	})
	flags.Usage = func() {
		fmt.Fprintln(out, "Usage: gh recap [YEAR] [--include-private] [--demo] [--no-animation]")
		fmt.Fprintln(out, "       gh recap {md|video} [YEAR] [--output FILE] [--include-private] [--demo] [--no-animation]")
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
		if argument == "md" || argument == "video" {
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
	if options.command == "" {
		if options.output != "" {
			return options, fmt.Errorf("--output requires the md or video command")
		}
	} else if options.output == "" {
		extension := "md"
		if options.command == "video" {
			extension = "mp4"
		}
		options.output = fmt.Sprintf("gh-recap_%d.%s", options.Year, extension)
	}
	return options, nil
}
