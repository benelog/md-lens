package heading

import (
	"image"
	"image/color"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/benelog/md-lens/internal/highlight"
)

// sizes are the pixel font sizes per heading level 1..6 (index 0 unused).
var sizes = [7]int{0, 56, 44, 34, 28, 24, 21}

// FontImageHeading rasterizes heading text with an embedded bold sans-serif font into a transparent
// image, sized by heading level (h1 largest). The Go font is embedded in the binary, so it is always
// available without a display or system fonts; the constructor still probes it so a broken setup
// fails fast and the caller can fall back to a text heading.
type FontImageHeading struct {
	theme *highlight.Theme
	font  *opentype.Font
}

// NewFontImageHeading parses the embedded font and probes it, returning an error if rasterization
// is unavailable.
func NewFontImageHeading(theme *highlight.Theme) (*FontImageHeading, error) {
	f, err := opentype.Parse(gobold.TTF)
	if err != nil {
		return nil, err
	}
	face, err := newFace(f, 24)
	if err != nil {
		return nil, err
	}
	defer face.Close()
	_ = font.MeasureString(face, "probe")
	return &FontImageHeading{theme: theme, font: f}, nil
}

func newFace(f *opentype.Font, px int) (font.Face, error) {
	// DPI 72 makes one point equal one pixel, so px is the rendered font height.
	return opentype.NewFace(f, &opentype.FaceOptions{
		Size:    float64(px),
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

// Render rasterizes the heading text and returns a transparent image with the glyphs in the
// level's color.
func (h *FontImageHeading) Render(level int, text string) (image.Image, error) {
	label := text
	if label == "" {
		label = " "
	}
	px := sizes[clampLevel(level)]
	face, err := newFace(h.font, px)
	if err != nil {
		return nil, err
	}
	defer face.Close()

	metrics := face.Metrics()
	ascent := metrics.Ascent.Ceil()
	descent := metrics.Descent.Ceil()
	pad := max(4, px/6)
	advance := font.MeasureString(face, label).Ceil()
	width := advance + pad*2
	height := ascent + descent + pad*2

	img := image.NewRGBA(image.Rect(0, 0, max(1, width), max(1, height)))
	c := h.theme.HeadingColor(level)
	src := image.NewUniform(color.RGBA{R: uint8(c[0]), G: uint8(c[1]), B: uint8(c[2]), A: 255})
	d := font.Drawer{
		Dst:  img,
		Src:  src,
		Face: face,
		Dot:  fixed.P(pad, pad+ascent),
	}
	d.DrawString(label)
	return img, nil
}

func clampLevel(level int) int {
	return min(6, max(1, level))
}
