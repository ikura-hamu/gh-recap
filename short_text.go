package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/charmbracelet/x/ansi"
	"github.com/ikura-hamu/gh-recap/internal/recap"
	"github.com/ikura-hamu/gh-recap/internal/tui"
)

func exportShortText(ctx context.Context, options commandOptions, api recap.API) error {
	text, warnings, err := tui.CollectShortText(ctx, options.Options, api)
	printCollectionWarnings(warnings)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(os.Stdout, text); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "gh recap:", copyShortText(ctx, text))
	return nil
}

func copyShortText(ctx context.Context, text string) string {
	const failed = "warning: couldn't copy the summary to the clipboard"
	commandCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	commands := [][]string{{"wl-copy"}, {"xclip", "-selection", "clipboard", "-in"}, {"xsel", "--clipboard", "--input"}}
	switch runtime.GOOS {
	case "darwin":
		commands = [][]string{{"pbcopy"}}
	case "windows":
		commands = [][]string{{"clip.exe"}}
	case "linux":
		if os.Getenv("WSL_INTEROP") != "" || os.Getenv("WSL_DISTRO_NAME") != "" {
			commands = append(commands, []string{"clip.exe"})
		}
	}
	for _, args := range commands {
		input := text
		if args[0] == "clip.exe" {
			// A UTF-16LE byte order mark lets clip.exe recognize Unicode input.
			codeUnits := utf16.Encode([]rune("\ufeff" + text))
			encoded := make([]byte, 2*len(codeUnits))
			for i, codeUnit := range codeUnits {
				binary.LittleEndian.PutUint16(encoded[2*i:], codeUnit)
			}
			input = string(encoded)
		}
		cmd := exec.CommandContext(commandCtx, args[0], args[1:]...)
		cmd.Stdin = strings.NewReader(input)
		if cmd.Run() == nil {
			return "Copied the summary to the clipboard"
		}
	}

	if ctx.Err() != nil || runtime.GOOS == "windows" || os.Getenv("TERM") == "" || os.Getenv("TERM") == "dumb" {
		return failed
	}
	terminal, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return failed
	}
	defer terminal.Close()
	if _, err := io.WriteString(terminal, ansi.SetSystemClipboard(text)); err != nil {
		return failed
	}
	return "Sent a clipboard copy request to the terminal"
}
