package main

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ikura-hamu/gh-recap/internal/recap"
)

type commandOptions struct {
	recap.Options
	video string
}

func parseOptions(args []string, out io.Writer) (commandOptions, error) {
	options := commandOptions{Options: recap.Options{Year: time.Now().Year()}}
	flags := flag.NewFlagSet("gh recap", flag.ContinueOnError)
	flags.SetOutput(out)
	flags.BoolVar(&options.IncludePrivate, "include-private", false, "include accessible private repositories")
	flags.BoolVar(&options.Demo, "demo", false, "explore fictional activity without authentication")
	flags.BoolVar(&options.NoAnimation, "no-animation", false, "disable decorative animations (autoplay stays enabled)")
	flags.Func("video", "record the recap as an MP4 using VHS (FILE.mp4)", func(path string) error {
		if !strings.EqualFold(filepath.Ext(path), ".mp4") {
			return fmt.Errorf("video output must have an .mp4 extension")
		}
		options.video = path
		return nil
	})
	flags.Usage = func() {
		fmt.Fprintln(out, "Usage: gh recap [YEAR] [--include-private] [--demo] [--no-animation] [--video FILE.mp4]")
		flags.PrintDefaults()
	}
	flagArgs := []string{}
	year := ""
	// flag stops at the first positional argument; reorder flags to also accept YEAR first.
	for i := 0; i < len(args); i++ {
		argument := args[i]
		if argument == "--video" || argument == "-video" {
			if i+1 == len(args) {
				return options, fmt.Errorf("%s requires an MP4 output path", argument)
			}
			flagArgs = append(flagArgs, argument, args[i+1])
			i++
			continue
		}
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
