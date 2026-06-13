package image

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/benelog/md-lens/internal/ansi"
	"github.com/benelog/md-lens/internal/term"
)

func TestTwoByTwoImageMapsTopToFgAndBottomToBg(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 0xFF, A: 0xFF}) // top-left red
	img.Set(1, 0, color.RGBA{G: 0xFF, A: 0xFF}) // top-right green
	img.Set(0, 1, color.RGBA{B: 0xFF, A: 0xFF}) // bottom-left blue
	img.Set(1, 1, color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF})

	a := ansi.New(term.Truecolor)
	var sb strings.Builder
	EmitHalfBlock(img, a, "", &sb)

	want := a.FgRgb(0xFF0000) + a.BgRgb(0x0000FF) + "▀" +
		a.FgRgb(0x00FF00) + a.BgRgb(0xFFFFFF) + "▀" +
		a.Reset() + "\n"
	if sb.String() != want {
		t.Errorf("got %q want %q", sb.String(), want)
	}
}
