// Package term figures out what the output terminal can do — color depth, image protocol,
// and size — and reports it for the --caps flag.
package term

// ColorDepth is how many colors the output stream can show.
type ColorDepth int

const (
	// None means no ANSI styling at all (piped output, --no-color, dumb terminal).
	None ColorDepth = iota
	// Ansi16 means 16 ANSI colors.
	Ansi16
	// Ansi256 means 256 indexed colors.
	Ansi256
	// Truecolor means 24-bit truecolor.
	Truecolor
)

// GraphicsMode is how inline images (and font-rendered headings) get drawn.
type GraphicsMode int

const (
	// GraphicsAuto is a sentinel meaning "no forced override" (auto-detect). It is never the
	// result of detection, only an input to it.
	GraphicsAuto GraphicsMode = iota
	// GraphicsNone means no image support; images degrade to their alt text.
	GraphicsNone
	// Kitty is the kitty graphics protocol (chunked base64 PNG over APC).
	Kitty
	// Iterm2 is the iTerm2 inline image protocol (OSC 1337).
	Iterm2
	// Sixel is reserved — not implemented yet.
	Sixel
	// HalfBlock is the Unicode upper-half-block (▀) with truecolor fg/bg per two pixels.
	HalfBlock
)

// IsPixelProtocol is true for protocols that consume a real PNG (so font-image headings make sense).
func (g GraphicsMode) IsPixelProtocol() bool {
	return g == Kitty || g == Iterm2
}

// CanRenderImages is true when any kind of image can be shown (pixel protocol or half-block).
func (g GraphicsMode) CanRenderImages() bool {
	return g != GraphicsNone && g != GraphicsAuto
}

// Capabilities is an immutable snapshot of what the output terminal can do, produced by Detect.
type Capabilities struct {
	IsTTY    bool       // whether stdout is an interactive terminal
	Cols     int        // terminal width in character cells
	Rows     int        // terminal height in character cells
	CellPxW  int        // estimated pixel width of one cell
	CellPxH  int        // estimated pixel height of one cell
	Color    ColorDepth // color depth of the output
	Graphics GraphicsMode
}

// ColorEnabled reports whether the output supports any color.
func (c Capabilities) ColorEnabled() bool {
	return c.Color != None
}
