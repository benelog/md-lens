// Package ansi builds SGR (Select Graphic Rendition) escape sequences, downgrading color to
// whatever the terminal supports. When the color depth is term.None every method returns an
// empty string, so the same rendering code produces clean text when piped.
package ansi

import (
	"math"
	"strconv"

	"github.com/benelog/md-lens/internal/term"
)

const csi = "\x1b["

// Ansi emits escape sequences appropriate to a color depth.
type Ansi struct {
	depth term.ColorDepth
}

// New returns an Ansi emitter for the given color depth.
func New(depth term.ColorDepth) *Ansi {
	return &Ansi{depth: depth}
}

// Enabled reports whether any styling will be emitted.
func (a *Ansi) Enabled() bool {
	return a.depth != term.None
}

func (a *Ansi) Reset() string {
	if a.Enabled() {
		return csi + "0m"
	}
	return ""
}

func (a *Ansi) Bold() string      { return a.sgr(1) }
func (a *Ansi) Dim() string       { return a.sgr(2) }
func (a *Ansi) Italic() string    { return a.sgr(3) }
func (a *Ansi) Underline() string { return a.sgr(4) }
func (a *Ansi) Strike() string    { return a.sgr(9) }

// Style-off codes so nested inline styling can be closed without resetting siblings.
func (a *Ansi) BoldOff() string      { return a.sgr(22) }
func (a *Ansi) ItalicOff() string    { return a.sgr(23) }
func (a *Ansi) UnderlineOff() string { return a.sgr(24) }
func (a *Ansi) StrikeOff() string    { return a.sgr(29) }
func (a *Ansi) FgDefault() string    { return a.sgr(39) }
func (a *Ansi) BgDefault() string    { return a.sgr(49) }

func (a *Ansi) sgr(code int) string {
	if a.Enabled() {
		return csi + strconv.Itoa(code) + "m"
	}
	return ""
}

func (a *Ansi) Fg(r, g, b int) string { return a.color(true, r, g, b) }
func (a *Ansi) Bg(r, g, b int) string { return a.color(false, r, g, b) }

// Fg3 sets the foreground from an {r, g, b} triple.
func (a *Ansi) Fg3(c [3]int) string { return a.Fg(c[0], c[1], c[2]) }

// Bg3 sets the background from an {r, g, b} triple.
func (a *Ansi) Bg3(c [3]int) string { return a.Bg(c[0], c[1], c[2]) }

// FgRgb sets the foreground from a packed 0xRRGGBB (alpha ignored).
func (a *Ansi) FgRgb(rgb int) string {
	return a.Fg((rgb>>16)&0xff, (rgb>>8)&0xff, rgb&0xff)
}

// BgRgb sets the background from a packed 0xRRGGBB (alpha ignored).
func (a *Ansi) BgRgb(rgb int) string {
	return a.Bg((rgb>>16)&0xff, (rgb>>8)&0xff, rgb&0xff)
}

func (a *Ansi) color(fg bool, r, g, b int) string {
	switch a.depth {
	case term.None:
		return ""
	case term.Truecolor:
		lead := 38
		if !fg {
			lead = 48
		}
		return csi + strconv.Itoa(lead) + ";2;" + strconv.Itoa(r) + ";" + strconv.Itoa(g) + ";" + strconv.Itoa(b) + "m"
	case term.Ansi256:
		lead := 38
		if !fg {
			lead = 48
		}
		return csi + strconv.Itoa(lead) + ";5;" + strconv.Itoa(To256(r, g, b)) + "m"
	case term.Ansi16:
		return csi + strconv.Itoa(To16(fg, r, g, b)) + "m"
	}
	return ""
}

// To256 maps an RGB triple to the closest xterm-256 palette index.
func To256(r, g, b int) int {
	if r == g && g == b {
		if r < 8 {
			return 16
		}
		if r > 248 {
			return 231
		}
		return int(math.Round(float64(r-8)/247.0*24)) + 232
	}
	ri := int(math.Round(float64(r) / 255.0 * 5))
	gi := int(math.Round(float64(g) / 255.0 * 5))
	bi := int(math.Round(float64(b) / 255.0 * 5))
	return 16 + 36*ri + 6*gi + bi
}

var palette16 = [16][3]int{
	{0, 0, 0}, {128, 0, 0}, {0, 128, 0}, {128, 128, 0},
	{0, 0, 128}, {128, 0, 128}, {0, 128, 128}, {192, 192, 192},
	{128, 128, 128}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0},
	{0, 0, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
}

// To16 returns the SGR parameter for the nearest 16-color match
// (fg 30-37/90-97, bg 40-47/100-107).
func To16(fg bool, r, g, b int) int {
	best := 0
	var bestDist int64 = math.MaxInt64
	for i, c := range palette16 {
		dr := int64(r - c[0])
		dg := int64(g - c[1])
		db := int64(b - c[2])
		dist := dr*dr + dg*dg + db*db
		if dist < bestDist {
			bestDist = dist
			best = i
		}
	}
	base := 30
	if !fg {
		base = 40
	}
	if best < 8 {
		return base + best
	}
	return base + 60 + (best - 8)
}
