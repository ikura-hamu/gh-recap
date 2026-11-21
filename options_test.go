package main

import (
	"io"
	"testing"
	"time"
)

func TestArguments(t *testing.T) {
	for _, args := range [][]string{{"2025", "--demo", "--include-private"}, {"--demo", "2025", "--include-private"}} {
		o, err := parseOptions(args, io.Discard)
		if err != nil || o.Year != 2025 || !o.Demo || !o.IncludePrivate {
			t.Fatalf("parse %v: %+v %v", args, o, err)
		}
	}
	o, err := parseOptions(nil, io.Discard)
	if err != nil || o.Year != time.Now().Year() {
		t.Fatal("default must be current year")
	}
	for _, args := range [][]string{{"2025", "2024"}, {"nope"}, {"2007"}, {"9999"}, {"--unknown"}} {
		if _, err := parseOptions(args, io.Discard); err == nil {
			t.Fatalf("accepted invalid arguments %v", args)
		}
	}
}

func TestExportArguments(t *testing.T) {
	for _, test := range []struct {
		args    []string
		command string
		output  string
	}{
		{[]string{"md", "2025", "--demo"}, "md", "gh-recap_2025.md"},
		{[]string{"video", "2025", "--demo"}, "video", "gh-recap_2025.mp4"},
		{[]string{"video", "2025", "--output", "my recap.mp4", "--demo"}, "video", "my recap.mp4"},
		{[]string{"md", "--output=my recap.md", "--demo", "2025"}, "md", "my recap.md"},
	} {
		options, err := parseOptions(test.args, io.Discard)
		if err != nil || options.command != test.command || options.output != test.output || options.Year != 2025 || !options.Demo {
			t.Fatalf("parse %v: %+v %v", test.args, options, err)
		}
	}
	for _, args := range [][]string{{"video", "--output"}, {"md", "--output="}, {"--output", "recap.md"}, {"md", "video"}} {
		if _, err := parseOptions(args, io.Discard); err == nil {
			t.Errorf("accepted invalid arguments %v", args)
		}
	}
}
