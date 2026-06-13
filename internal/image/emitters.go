package image

import (
	"encoding/base64"
	"fmt"
	"image"
	"io"
	"strconv"

	"github.com/benelog/md-lens/internal/ansi"
)

const (
	apc = "\x1b_"
	st  = "\x1b\\"
	// KittyChunk is the kitty base64 payload chunk size (a multiple of 4, as the protocol requires).
	KittyChunk = 4096
)

// EmitKitty writes a PNG using the kitty graphics protocol: transmit-and-display as chunked base64
// over APC sequences. Cursor movement is suppressed with C=1; the caller positions the cursor.
//
// Frame: ESC _ G <control>; <base64-chunk> ESC \. Control keys only on the first chunk.
func EmitKitty(pngData []byte, cols, rows int, w io.Writer) {
	b64 := base64.StdEncoding.EncodeToString(pngData)
	n := len(b64)
	control := fmt.Sprintf("f=100,a=T,t=d,c=%d,r=%d,C=1,", cols, rows)

	if n <= KittyChunk {
		ws(w, apc+"G"+control+"m=0;"+b64+st)
		return
	}

	pos := 0
	first := true
	for pos < n {
		end := min(pos+KittyChunk, n)
		last := end == n
		ws(w, apc+"G")
		if first {
			ws(w, control)
		}
		m := "1"
		if last {
			m = "0"
		}
		ws(w, "m="+m+";")
		ws(w, b64[pos:end])
		ws(w, st)
		pos = end
		first = false
	}
}

// EmitIterm2 writes a PNG using the iTerm2 inline image protocol:
// ESC ] 1337 ; File=<args> : <base64> BEL. Width/height are in character cells.
func EmitIterm2(pngData []byte, cols, rows int, w io.Writer) {
	b64 := base64.StdEncoding.EncodeToString(pngData)
	ws(w, "\x1b]1337;File=inline=1;width="+strconv.Itoa(cols)+
		";height="+strconv.Itoa(rows)+
		";preserveAspectRatio=1;size="+strconv.Itoa(len(pngData))+
		":"+b64+"\x07")
}

const upperHalfBlock = "▀"

// EmitHalfBlock renders an image with the upper-half-block char ▀: each cell shows two vertical
// pixels — the top pixel as the foreground color, the bottom pixel as the background color.
// The image must already be sized to cols wide by rows*2 tall.
func EmitHalfBlock(img image.Image, a *ansi.Ansi, indent string, w io.Writer) {
	b := img.Bounds()
	width := b.Dx()
	height := b.Dy()
	for y := 0; y < height; y += 2 {
		ws(w, indent)
		for x := range width {
			top := rgbAt(img, b.Min.X+x, b.Min.Y+y)
			bottom := top
			if y+1 < height {
				bottom = rgbAt(img, b.Min.X+x, b.Min.Y+y+1)
			}
			ws(w, a.FgRgb(top)+a.BgRgb(bottom)+upperHalfBlock)
		}
		ws(w, a.Reset()+"\n")
	}
}

// rgbAt returns the pixel at (x, y) packed as 0xRRGGBB (alpha dropped).
func rgbAt(img image.Image, x, y int) int {
	r, g, b, _ := img.At(x, y).RGBA() // 16-bit per channel
	return int(r>>8)<<16 | int(g>>8)<<8 | int(b>>8)
}

// ws writes a string, ignoring the error (a wrapping errWriter short-circuits on a broken pipe).
func ws(w io.Writer, s string) {
	_, _ = io.WriteString(w, s)
}
