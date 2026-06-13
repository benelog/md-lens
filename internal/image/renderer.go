package image

import (
	"image"
	"io"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/benelog/md-lens/internal/ansi"
	"github.com/benelog/md-lens/internal/term"
)

const csi = "\x1b["

// Renderer decides how to draw an image: it fits the image to the terminal, then dispatches to the
// kitty / iTerm2 / half-block emitter.
//
// For the pixel protocols the cursor is reserved-and-repositioned so following text lands just
// below the image (the classic overwrite bug). Half-blocks are real text rows, so they need no trick.
type Renderer struct {
	caps term.Capabilities
	ansi *ansi.Ansi
}

// NewRenderer returns an image renderer for the given capabilities.
func NewRenderer(caps term.Capabilities, a *ansi.Ansi) *Renderer {
	return &Renderer{caps: caps, ansi: a}
}

// Render renders an image file, falling back to alt text if it can't be shown. Never fails.
func (r *Renderer) Render(imagePath, alt string, indentCols int, w io.Writer) {
	if !r.caps.Graphics.CanRenderImages() {
		r.fallback(alt, imagePath, indentCols, w)
		return
	}
	img, err := Load(imagePath)
	if err != nil {
		r.fallback(alt, imagePath, indentCols, w)
		return
	}
	if err := r.RenderImage(img, indentCols, w); err != nil {
		r.fallback(alt, imagePath, indentCols, w)
	}
}

// RenderImage renders an already-decoded image.
func (r *Renderer) RenderImage(img image.Image, indentCols int, w io.Writer) error {
	dispCols, dispRows := r.fit(img.Bounds().Dx(), img.Bounds().Dy(), indentCols)
	switch r.caps.Graphics {
	case term.Kitty:
		return r.emitKitty(img, indentCols, dispCols, dispRows, w)
	case term.Iterm2:
		return r.emitIterm(img, indentCols, dispCols, dispRows, w)
	case term.HalfBlock:
		resized := ResizeExact(img, dispCols, dispRows*2)
		EmitHalfBlock(resized, r.ansi, strings.Repeat(" ", indentCols), w)
		return nil
	default:
		// NONE / SIXEL: nothing to draw here.
		return nil
	}
}

// fit computes display size in cells: native size, capped to available width and screen height.
func (r *Renderer) fit(imgW, imgH, indentCols int) (int, int) {
	availCols := max(1, r.caps.Cols-indentCols)
	cellW := r.caps.CellPxW
	cellH := r.caps.CellPxH
	aspect := float64(imgH) / float64(imgW) // height / width

	nativeCols := max(1, int(math.Ceil(float64(imgW)/float64(cellW))))
	dispCols := min(availCols, nativeCols)
	dispRows := max(1, int(math.Round(float64(dispCols)*(float64(cellW)/float64(cellH))*aspect)))

	maxRows := max(1, r.caps.Rows-2)
	if dispRows > maxRows {
		dispRows = maxRows
		dispCols = max(1, int(math.Round(float64(dispRows)*(float64(cellH)/float64(cellW))/aspect)))
		dispCols = min(dispCols, availCols)
	}
	return dispCols, dispRows
}

func (r *Renderer) emitKitty(img image.Image, indentCols, dispCols, dispRows int, w io.Writer) error {
	pngData, err := ToPNG(img)
	if err != nil {
		return err
	}
	for range dispRows {
		ws(w, "\n") // reserve the rows (scrolls if at bottom)
	}
	ws(w, csi+strconv.Itoa(dispRows)+"A") // back up to the top row
	if indentCols > 0 {
		ws(w, csi+strconv.Itoa(indentCols)+"C") // move to indent column
	}
	EmitKitty(pngData, dispCols, dispRows, w)  // C=1 → cursor stays put
	ws(w, csi+strconv.Itoa(dispRows)+"B"+"\r") // land below image
	return nil
}

func (r *Renderer) emitIterm(img image.Image, indentCols, dispCols, dispRows int, w io.Writer) error {
	pngData, err := ToPNG(img)
	if err != nil {
		return err
	}
	if indentCols > 0 {
		ws(w, strings.Repeat(" ", indentCols))
	}
	EmitIterm2(pngData, dispCols, dispRows, w)
	ws(w, "\n")
	return nil
}

func (r *Renderer) fallback(alt, path string, indentCols int, w io.Writer) {
	label := alt
	if strings.TrimSpace(label) == "" {
		label = filepath.Base(path)
	}
	ws(w, strings.Repeat(" ", indentCols)+r.ansi.Dim()+"[image: "+label+"]"+r.ansi.Reset()+"\n")
}
