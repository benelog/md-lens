package render

import (
	"io"
	"strings"

	"github.com/benelog/md-lens/internal/ansi"
	"github.com/benelog/md-lens/internal/highlight"
	"github.com/benelog/md-lens/internal/term"
)

// Context is the mutable rendering state shared across the visitor: the output sink,
// terminal/style helpers, the wrap width, and the stack of line prefixes (blockquote gutters,
// list indents).
//
// A "pending marker" lets a list item show its bullet on the first printed line while continuation
// lines fall back to spaces.
type Context struct {
	out        io.Writer
	caps       term.Capabilities
	ansi       *ansi.Ansi
	theme      *highlight.Theme
	width      int
	hyperlinks bool

	prefixes      []string
	pendingMarker string
}

// NewContext creates a render context writing to out.
func NewContext(out io.Writer, caps term.Capabilities, a *ansi.Ansi, theme *highlight.Theme, width int) *Context {
	return &Context{
		out:        out,
		caps:       caps,
		ansi:       a,
		theme:      theme,
		width:      width,
		hyperlinks: caps.IsTTY,
	}
}

func (c *Context) pushPrefix(prefix string) {
	c.prefixes = append(c.prefixes, prefix)
}

func (c *Context) popPrefix() {
	c.prefixes = c.prefixes[:len(c.prefixes)-1]
}

// setPendingMarker makes the next printed line replace the innermost prefix with marker.
func (c *Context) setPendingMarker(marker string) {
	c.pendingMarker = marker
}

// indentWidth is the total visible width of the current prefixes — used to shrink the wrap width.
func (c *Context) indentWidth() int {
	w := 0
	for _, p := range c.prefixes {
		w += ansi.Width(p)
	}
	return w
}

func (c *Context) contentWidth() int {
	return max(1, c.width-c.indentWidth())
}

func (c *Context) linePrefix() string {
	if len(c.prefixes) == 0 {
		m := c.pendingMarker
		c.pendingMarker = ""
		return m
	}
	var sb strings.Builder
	last := len(c.prefixes) - 1
	for i, p := range c.prefixes {
		if i == last && c.pendingMarker != "" {
			sb.WriteString(c.pendingMarker)
		} else {
			sb.WriteString(p)
		}
	}
	c.pendingMarker = ""
	return sb.String()
}

// line prints one line with the current prefix and a trailing newline.
func (c *Context) line(content string) {
	ws(c.out, c.linePrefix()+content+"\n")
}

// blank prints a blank separator line (no prefix, to avoid trailing whitespace).
func (c *Context) blank() {
	ws(c.out, "\n")
}

func ws(w io.Writer, s string) {
	_, _ = io.WriteString(w, s)
}
