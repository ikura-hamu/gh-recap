package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ikura-hamu/gh-recap/internal/recap"
)

func TestRecordVideo(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a shell executable")
	}
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	options := commandOptions{
		Options: recap.Options{Year: 2025, Demo: true, IncludePrivate: true, NoAnimation: true, DPGK: true},
		output:  filepath.Join(dir, "test video.mp4"),
	}
	ctx := context.Background()
	if err := recordVideo(ctx, options); err == nil || !strings.Contains(err.Error(), "brew install vhs") {
		t.Fatalf("expected VHS installation instructions, got %v", err)
	}

	fakeVHS := `#!/bin/sh
[ "$1" = "-" ] || exit 1
tape=
while IFS= read -r line; do tape="$tape
$line"; done
case "$tape" in
  *"Output recap.mp4"*"Set Width 1280"*"Set Height 720"*"2025 --include-private --demo --no-animation --dpgk"*"FINISHED"*"Sleep 6s"*) ;;
  *) exit 2 ;;
esac
printf 'test recording' > recap.mp4
if [ "$GH_RECAP_TEST_FINISH" = 1 ]; then
  : > "$GH_RECAP_VIDEO_DONE"
else
  printf 'playback did not reach the final page' > "$GH_RECAP_VIDEO_LOG"
fi
`
	if err := os.WriteFile(filepath.Join(dir, "vhs"), []byte(fakeVHS), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(options.output, []byte("previous recording"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GH_RECAP_TEST_FINISH", "0")
	if err := recordVideo(ctx, options); err == nil || !strings.Contains(err.Error(), "playback did not reach the final page") {
		t.Fatalf("expected incomplete playback error, got %v", err)
	}
	if output, err := os.ReadFile(options.output); err != nil || string(output) != "previous recording" {
		t.Fatalf("previous video should survive an incomplete recording: %q, %v", output, err)
	}
	t.Setenv("GH_RECAP_TEST_FINISH", "1")
	if err := recordVideo(ctx, options); err != nil {
		t.Fatal(err)
	}
	if output, err := os.ReadFile(options.output); err != nil || string(output) != "test recording" {
		t.Fatalf("unexpected video output: %q, %v", output, err)
	}
}
