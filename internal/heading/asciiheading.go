// Package heading renders a markdown heading as a large font-rasterized image on pixel-graphics
// terminals, or as differentiated styled text otherwise.
package heading

import (
	"io"
	"strings"

	"github.com/benelog/md-lens/internal/ansi"
	"github.com/benelog/md-lens/internal/highlight"
)

// RenderAscii is the text fallback for headings when no pixel graphics protocol is available.
// Each level is differentiated by a distinct color plus its own decoration:
//
//	h1  BOLD + heavy underline (═)
//	h2  BOLD + light underline (─)
//	h3  ◆ + BOLD
//	h4  ▸ + BOLD
//	h5  • + normal
//	h6  · + dim
func RenderAscii(level int, text string, a *ansi.Ansi, theme *highlight.Theme, cols int, w io.Writer) {
	color := theme.HeadingColor(level)
	underlineLen := min(max(1, cols), max(1, ansi.Width(text)))

	switch level {
	case 1:
		ws(w, a.Bold()+a.Fg3(color)+text+a.Reset()+"\n")
		ws(w, a.Fg3(color)+strings.Repeat("═", underlineLen)+a.Reset()+"\n")
	case 2:
		ws(w, a.Bold()+a.Fg3(color)+text+a.Reset()+"\n")
		ws(w, a.Fg3(color)+strings.Repeat("─", underlineLen)+a.Reset()+"\n")
	case 3:
		ws(w, a.Fg3(color)+"◆ "+a.Bold()+text+a.Reset()+"\n")
	case 4:
		ws(w, a.Fg3(color)+"▸ "+a.Bold()+text+a.Reset()+"\n")
	case 5:
		ws(w, a.Fg3(color)+"• "+text+a.Reset()+"\n")
	default:
		ws(w, a.Dim()+a.Fg3(color)+"· "+text+a.Reset()+"\n")
	}
}

func ws(w io.Writer, s string) {
	_, _ = io.WriteString(w, s)
}
