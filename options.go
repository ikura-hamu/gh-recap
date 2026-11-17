package main

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/ikura-hamu/gh-recap/internal/recap"
)

func parseOptions(args []string, out io.Writer) (recap.Options, error) {
	options := recap.Options{Year: time.Now().Year()}
	flags := flag.NewFlagSet("gh recap", flag.ContinueOnError)
	flags.SetOutput(out)
	flags.BoolVar(&options.IncludePrivate, "include-private", false, "include accessible private repositories")
	flags.BoolVar(&options.Demo, "demo", false, "explore fictional activity without authentication")
	flags.BoolVar(&options.NoAnimation, "no-animation", false, "disable decorative animations (autoplay stays enabled)")
	flags.Usage = func() {
		fmt.Fprintln(out, "Usage: gh recap [YEAR] [--include-private] [--demo] [--no-animation]")
		flags.PrintDefaults()
	}
	flagArgs := []string{}
	year := ""
	// flag stops at the first positional argument; reorder flags to also accept YEAR first.
	for _, argument := range args {
		if len(argument) > 0 && argument[0] != '-' {
			if year != "" {
				return options, fmt.Errorf("specify only one year")
			}
			year = argument
		} else {
			flagArgs = append(flagArgs, argument)
		}
	}
	if err := flags.Parse(flagArgs); err != nil {
		return options, err
	}
	if year != "" {
		parsedYear, err := strconv.Atoi(year)
		if err != nil || parsedYear < 2008 || parsedYear > time.Now().Year() {
			return options, fmt.Errorf("year must be between 2008 and %d", time.Now().Year())
		}
		options.Year = parsedYear
	}
	return options, nil
}
