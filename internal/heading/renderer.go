package heading

import (
	"io"
	"strings"

	"github.com/benelog/md-lens/internal/ansi"
	"github.com/benelog/md-lens/internal/highlight"
	"github.com/benelog/md-lens/internal/image"
	"github.com/benelog/md-lens/internal/term"
)

// Renderer renders a heading as a large font-rasterized image on pixel-graphics terminals, or as
// styled text otherwise. If font rendering is unavailable (or disabled), it degrades to RenderAscii.
type Renderer struct {
	caps        term.Capabilities
	ansi        *ansi.Ansi
	theme       *highlight.Theme
	images      *image.Renderer
	fontHeading *FontImageHeading
}

// NewRenderer builds a heading renderer, preparing font-image rendering when the terminal supports
// a pixel protocol and it is not disabled.
func NewRenderer(caps term.Capabilities, a *ansi.Ansi, theme *highlight.Theme,
	images *image.Renderer, noHeadingImages bool) *Renderer {
	var fh *FontImageHeading
	if caps.Graphics.IsPixelProtocol() && !noHeadingImages {
		if f, err := NewFontImageHeading(theme); err == nil {
			fh = f
		}
	}
	return &Renderer{caps: caps, ansi: a, theme: theme, images: images, fontHeading: fh}
}

// Render writes the heading as a font image when possible, otherwise as styled text.
func (r *Renderer) Render(level int, text string, w io.Writer) {
	if r.fontHeading != nil && strings.TrimSpace(text) != "" {
		if img, err := r.fontHeading.Render(level, text); err == nil {
			if err := r.images.RenderImage(img, 0, w); err == nil {
				return
			}
		}
	}
	RenderAscii(level, text, r.ansi, r.theme, r.caps.Cols, w)
}
