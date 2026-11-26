package tui

import (
	"context"
	_ "embed"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/ikura-hamu/gh-recap/internal/recap"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

const (
	imageColumns, imageRows = 120, 32
	cellWidth, cellHeight   = 10, 20
)

//go:embed fonts/JetBrainsMono-Regular.ttf
var imageFontRegular []byte

//go:embed fonts/JetBrainsMono-Bold.ttf
var imageFontBold []byte

// ImageFontLicense holds the license and copyright notice for the embedded fonts.
//
//go:embed fonts/OFL.txt
var ImageFontLicense string

// ImageFilename returns the default filename for an ENCORE card.
func ImageFilename(year int) string { return fmt.Sprintf("gh-recap_%d.png", year) }

// ExportImage collects the activity shown on ENCORE and saves it as a PNG.
func ExportImage(ctx context.Context, options recap.Options, api recap.API, path string) ([]string, error) {
	m, err := collect(ctx, options, api, false)
	if err == nil {
		err = saveEncoreImage(ctx, path, m.encoreImageView())
	}
	return m.CollectionWarnings(), err
}

func (m Model) encoreImageView() tea.View {
	m.imageExport = true
	m.page, m.width, m.height = pageSummary, imageColumns, imageRows
	m.elapsed, m.titleFrame, m.sparkleFrame = pageDuration, 4, 0
	m.showDetails = false
	m.transition = pageTransition{}

	layout := m.layoutContent(m.encoreContent(recap.Summarize(m.events), ""))
	body := m.arrangeBody(layout.Heading, layout.Body)
	return m.renderScreen(body, "\n"+m.imageCredit())
}

func (m Model) imageCredit() string {
	credit := ink(m.color()).Bold(true).Render("Made with gh-recap")
	project := muted.Render("github.com/ikura-hamu/gh-recap")
	gap := m.contentWidth() - lipgloss.Width(credit) - lipgloss.Width(project)
	return credit + strings.Repeat(" ", gap) + project
}

func (m Model) captureImageView() tea.View {
	m.imageExport = true
	m.viewportCache = nil
	m.transition = pageTransition{}
	view := m.View()
	credit := lipgloss.NewStyle().Width(m.width).Padding(0, horizontalPadding).Render(m.imageCredit())
	view.Content += "\n" + credit + "\n" + strings.Repeat(" ", m.width)
	return view
}

type imageSavedMsg struct {
	name string
	err  error
}

func (m Model) canSaveImage() bool {
	return m.page == pageSummary && !m.showDetails && !m.options.Recording && m.login != "" && m.identityError == nil && m.ready()
}

func (m *Model) saveImage() tea.Cmd {
	if m.savingImage {
		return nil
	}
	ctx, view := m.ctx, m.captureImageView()
	m.savingImage = true
	m.imageStatus = "Saving PNG…"
	name := ImageFilename(m.options.Year)
	return func() tea.Msg {
		return imageSavedMsg{name, saveEncoreImage(ctx, name, view)}
	}
}

func saveEncoreImage(ctx context.Context, path string, view tea.View) error {
	img, err := rasterizeEncore(view)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".gh-recap-png-*")
	if err != nil {
		return fmt.Errorf("create PNG output: %w", err)
	}
	defer os.Remove(temp.Name())
	writeErr := png.Encode(temp, img)
	closeErr := temp.Close()
	if writeErr != nil {
		return fmt.Errorf("encode PNG: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close PNG output: %w", closeErr)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.Rename(temp.Name(), path); err != nil {
		return fmt.Errorf("save PNG: %w", err)
	}
	return nil
}

func rasterizeEncore(view tea.View) (*image.RGBA, error) {
	regular, err := terminalFace(imageFontRegular)
	if err != nil {
		return nil, fmt.Errorf("load regular PNG font: %w", err)
	}
	defer regular.Close()
	bold, err := terminalFace(imageFontBold)
	if err != nil {
		return nil, fmt.Errorf("load bold PNG font: %w", err)
	}
	defer bold.Close()

	columns, rows := lipgloss.Size(view.Content)
	screen := uv.NewScreenBuffer(columns, rows)
	uv.NewStyledString(view.Content).Draw(screen, screen.Bounds())
	img := image.NewRGBA(image.Rect(0, 0, columns*cellWidth, rows*cellHeight))
	draw.Draw(img, img.Bounds(), image.NewUniform(view.BackgroundColor), image.Point{}, draw.Src)
	star := terminalStar()
	for y := range rows {
		for x := range columns {
			cell := screen.CellAt(x, y)
			if cell == nil || cell.Width == 0 {
				continue
			}
			left, top := x*cellWidth, y*cellHeight
			rect := image.Rect(left, top, left+cell.Width*cellWidth, top+cellHeight)
			if cell.Style.Bg != nil {
				draw.Draw(img, rect, image.NewUniform(cell.Style.Bg), image.Point{}, draw.Src)
			}
			if cell.Content == " " || cell.Content == "" {
				continue
			}
			fg := cell.Style.Fg
			if fg == nil {
				fg = view.ForegroundColor
			}
			brush := image.NewUniform(fg)
			if cell.Content == "✦" {
				draw.DrawMask(img, rect, brush, image.Point{}, star, image.Point{}, draw.Over)
				continue
			}
			if shape, ok := terminalShape(cell.Content, rect); ok {
				draw.Draw(img, shape, brush, image.Point{}, draw.Src)
				continue
			}
			face := regular
			if cell.Style.Attrs&uv.AttrBold != 0 {
				face = bold
			}
			drawer := font.Drawer{Dst: img, Src: brush, Face: face, Dot: fixed.P(left, top+16)}
			drawer.DrawString(cell.Content)
		}
	}
	return img, nil
}

func terminalStar() *image.Alpha {
	mask := image.NewAlpha(image.Rect(0, 0, cellWidth, cellHeight))
	raster := vector.NewRasterizer(cellWidth, cellHeight)
	x, y := float32(cellWidth)/2, float32(cellHeight)/2
	raster.MoveTo(x, y-6)
	raster.LineTo(x+1.5, y-2)
	raster.LineTo(x+4, y)
	raster.LineTo(x+1.5, y+2)
	raster.LineTo(x, y+6)
	raster.LineTo(x-1.5, y+2)
	raster.LineTo(x-4, y)
	raster.LineTo(x-1.5, y-2)
	raster.ClosePath()
	raster.Draw(mask, mask.Bounds(), image.Opaque, image.Point{})
	return mask
}

func terminalFace(ttf []byte) (font.Face, error) {
	typeface, err := opentype.Parse(ttf)
	if err != nil {
		return nil, err
	}
	// JetBrains Mono's advance is 0.6 em. Match it to the terminal cell width.
	return opentype.NewFace(typeface, &opentype.FaceOptions{Size: cellWidth / 0.6, DPI: 72})
}

// Draw block and line glyphs at cell boundaries so adjacent cells join.
func terminalShape(glyph string, cell image.Rectangle) (image.Rectangle, bool) {
	x, y := cell.Min.X, cell.Min.Y
	switch glyph {
	case "█":
		return cell, true
	case "▀":
		cell.Max.Y = y + cellHeight/2
	case "▄":
		cell.Min.Y = y + cellHeight/2
	case "│":
		cell.Min.X, cell.Max.X = x+cellWidth/2, x+cellWidth/2+1
	case "━":
		cell.Min.Y, cell.Max.Y = y+cellHeight/2-1, y+cellHeight/2+2
	case "─":
		cell.Min.Y, cell.Max.Y = y+cellHeight/2, y+cellHeight/2+1
	default:
		return cell, false
	}
	return cell, true
}
