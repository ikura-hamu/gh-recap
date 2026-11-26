package tui

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/ikura-hamu/gh-recap/internal/recap"
)

func TestEncoreImageCapturesCurrentFrame(t *testing.T) {
	t.Chdir(t.TempDir())
	m := readyModel()
	m.page = pageSummary
	delete(m.loaded, recap.KindIssues)
	if _, cmd := m.Update(pressKey("s")); cmd != nil {
		t.Fatal("image export should wait for all ENCORE metrics")
	}
	m.loaded[recap.KindIssues] = true
	delete(m.loaded, recap.KindCoauthors)
	m.width, m.height = 80, fullDisplayHeight-1
	m.options.NoAnimation = true
	m.paused = true
	m.syncViewport()
	m.viewport.GotoBottom()
	if !strings.Contains(m.helpView(m.contentWidth()), "save PNG") {
		t.Fatal("ENCORE should advertise image export as soon as its metrics are ready")
	}
	frame := ansi.Strip(m.View().Content)
	capture := m.captureImageView()
	plain := ansi.Strip(capture.Content)
	assertBlockText(t, frame, display("ENCORE"), true)
	assertBlockText(t, plain, display("GITHUB"), true)
	assertBlockText(t, plain, display("RECAP"), true)
	assertBlockText(t, frame, display("2025"), true)
	assertBlockText(t, plain, display("2025"), true)
	if !strings.Contains(frame, "✦ GH RECAP") || strings.Contains(plain, "✦ GH RECAP") {
		t.Fatal("the badge should appear only in the TUI")
	}
	if !strings.Contains(plain, "Made with gh-recap") || !strings.Contains(plain, "PAUSED") {
		t.Fatal("capture should retain playback status and include the gh-recap credit")
	}
	hint := strings.Index(frame, "↑/↓ scroll · ")
	if hint == -1 || !strings.Contains(plain, strings.SplitN(frame[hint:], "\n", 2)[0]) {
		t.Fatal("capture should retain the TUI scroll position")
	}
	next, cmd := m.Update(pressKey("s"))
	if cmd == nil {
		t.Fatal("ENCORE save key did not start export")
	}
	result := update(next.(Model), cmd())
	filename := ImageFilename(m.options.Year)
	if !strings.HasPrefix(result.statusView(), "Saved "+filename) {
		t.Fatalf("image status = %q", result.statusView())
	}
	saved, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	interactive, err := png.Decode(bytes.NewReader(saved))
	if err != nil {
		t.Fatal(err)
	}
	wantBounds := image.Rect(0, 0, m.width*cellWidth, (m.height+compactDisplayHeight+1+2)*cellHeight)
	if interactive.Bounds() != wantBounds {
		t.Fatalf("capture bounds = %v, want %v", interactive.Bounds(), wantBounds)
	}
	expected, err := rasterizeEncore(capture)
	if err != nil {
		t.Fatal(err)
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, expected); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(saved, encoded.Bytes()) {
		t.Fatal("saved PNG differs from the branded capture")
	}
}

func TestHeadlessImageUsesWideLayout(t *testing.T) {
	m := readyModel()
	m.options.NoAnimation = true
	plain := ansi.Strip(m.encoreImageView().Content)
	assertBlockText(t, plain, display("GITHUB"), false)
	assertBlockText(t, plain, display("RECAP"), false)
	assertBlockText(t, plain, display("2025"), false)
	if strings.Contains(plain, "✦ GH RECAP") {
		t.Fatal("headless image should omit the TUI badge")
	}
	m.page, m.height = pageSummary, minimumHeight
	short := ansi.Strip(m.captureImageView().Content)
	assertBlockText(t, short, display("2025"), false)
	if !strings.Contains(short, "@you") {
		t.Fatal("the longer heading should fit in a capture from a short, wide terminal")
	}
	options := recap.Options{Year: 2025, Demo: true}
	path := filepath.Join(t.TempDir(), "custom recap.png")
	if _, err := ExportImage(t.Context(), options, nil, path); err != nil {
		t.Fatal(err)
	}
	headless, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	config, err := png.DecodeConfig(bytes.NewReader(headless))
	if err != nil || config.Width != 1200 || config.Height != 640 {
		t.Fatalf("expected a 1200 × 640 PNG, got %+v, %v", config, err)
	}
}

func assertBlockText(t *testing.T, frame, block string, compact bool) {
	t.Helper()
	rows := strings.Split(block, "\n")
	if compact {
		rows = compactRows(rows)
	}
	if !strings.Contains(frame, strings.TrimRight(rows[0], " ")) ||
		!strings.Contains(frame, strings.TrimRight(rows[len(rows)-1], " ")) {
		t.Fatalf("frame is missing block text %q", block)
	}
}
