package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	videoDoneEnv = "GH_RECAP_VIDEO_DONE"
	videoName    = "recap.mp4"
)

func recordVideo(ctx context.Context, options commandOptions) error {
	vhs, err := exec.LookPath("vhs")
	if err != nil {
		return fmt.Errorf("vhs is unavailable on PATH; install VHS (for example, brew install vhs) and its dependencies: https://github.com/charmbracelet/vhs#installation")
	}
	vhs, err = filepath.Abs(vhs)
	if err != nil {
		return err
	}
	output, err := filepath.Abs(options.video)
	if err != nil {
		return err
	}
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	// Keep VHS's output and completion marker private until playback succeeds.
	dir, err := os.MkdirTemp(filepath.Dir(output), ".gh-recap-video-")
	if err != nil {
		return fmt.Errorf("create video workspace: %w", err)
	}
	defer os.RemoveAll(dir)

	done := filepath.Join(dir, "done")
	log := filepath.Join(dir, "recap.log")
	cmd := exec.CommandContext(ctx, vhs, "-")
	cmd.Dir = dir
	cmd.Env = append(cmd.Environ(),
		"GH_RECAP_VIDEO_BIN="+binary,
		"GH_RECAP_VIDEO_DIR="+cwd,
		videoDoneEnv+"="+done,
		"GH_RECAP_VIDEO_LOG="+log,
	)
	cmd.Stdin = strings.NewReader(videoTape(options))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Fprintln(os.Stderr, "Recording your recap with VHS…")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("VHS recording failed: %w", err)
	}
	if _, err := os.Stat(done); err != nil {
		if message, err := os.ReadFile(log); err == nil && len(message) > 0 {
			return fmt.Errorf("recording did not finish: %s", strings.TrimSpace(string(message)))
		}
		return fmt.Errorf("recording did not reach the end of playback")
	}

	recording := filepath.Join(dir, videoName)
	info, err := os.Stat(recording)
	if err != nil {
		return fmt.Errorf("VHS did not produce an MP4: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return fmt.Errorf("VHS produced an invalid MP4")
	}
	// The workspace is on the same filesystem as the destination, so a
	// successful recording can replace the old file in one operation.
	if err := os.Rename(recording, output); err != nil {
		return fmt.Errorf("save video: %w", err)
	}
	fmt.Fprintf(os.Stdout, "Saved %s\n", output)
	return nil
}

func videoTape(options commandOptions) string {
	// Pass paths through the environment so they need no VHS or shell escaping.
	command := fmt.Sprintf(`cd -- "$GH_RECAP_VIDEO_DIR" && "$GH_RECAP_VIDEO_BIN" %d`, options.Year)
	if options.IncludePrivate {
		command += " --include-private"
	}
	if options.Demo {
		command += " --demo"
	}
	if options.NoAnimation {
		command += " --no-animation"
	}
	// Keep the shell's echoed command from matching the exit marker.
	command += ` 2>"$GH_RECAP_VIDEO_LOG"; echo GH_RECAP_CAPTURE_""EXITED`
	return fmt.Sprintf(`Output %s
Set Shell bash
Set Width 1280
Set Height 720
Set FontSize 16
Set Padding 16
Set Framerate 24
Set TypingSpeed 1ms
Set CursorBlink false

Hide
Type '%s'
Enter
Wait+Screen@300s /PLAYING · 1×|CONNECTION FAILED|Terminal too small|GH_RECAP_CAPTURE_EXITED/
Show
Wait+Screen@240s /FINISHED · 1×|CONNECTION FAILED|Terminal too small|GH_RECAP_CAPTURE_EXITED/
Sleep 6s
Hide
Type "q"
Wait+Screen@10s /GH_RECAP_CAPTURE_EXITED/
`, videoName, command)
}
