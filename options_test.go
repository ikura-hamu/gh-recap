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

func TestVideoArguments(t *testing.T) {
	for _, args := range [][]string{
		{"2025", "--video", "my recap.mp4", "--demo"},
		{"--video=my recap.mp4", "--demo", "2025"},
	} {
		options, err := parseOptions(args, io.Discard)
		if err != nil || options.video != "my recap.mp4" || options.Year != 2025 || !options.Demo {
			t.Fatalf("parse %v: %+v %v", args, options, err)
		}
	}
	for _, args := range [][]string{{"--video"}, {"--video="}, {"--video", "recap.gif"}} {
		if _, err := parseOptions(args, io.Discard); err == nil {
			t.Errorf("accepted invalid arguments %v", args)
		}
	}
}
