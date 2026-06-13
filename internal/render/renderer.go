// Package render parses markdown (with GFM tables/strikethrough/task-lists) and renders it to the
// terminal.
package render

import (
	"io"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"

	"github.com/benelog/md-lens/internal/ansi"
	"github.com/benelog/md-lens/internal/cli"
	"github.com/benelog/md-lens/internal/heading"
	"github.com/benelog/md-lens/internal/highlight"
	"github.com/benelog/md-lens/internal/image"
	"github.com/benelog/md-lens/internal/term"
)

// Renderer renders markdown to a terminal-styled stream.
type Renderer struct {
	caps    term.Capabilities
	opts    cli.Args
	baseDir string
	theme   *highlight.Theme
}

// NewRenderer creates a renderer for the given capabilities, options, and base directory (used to
// resolve relative image paths).
func NewRenderer(caps term.Capabilities, opts cli.Args, baseDir string) *Renderer {
	return &Renderer{caps: caps, opts: opts, baseDir: baseDir, theme: highlight.Default()}
}

// Render parses markdown and writes the rendered output. A downstream closed pipe (e.g. | head)
// stops output quietly.
func (r *Renderer) Render(markdown string, out io.Writer) {
	ew := &errWriter{w: out}
	a := ansi.New(r.caps.Color)

	source := []byte(markdown)
	md := goldmark.New(goldmark.WithExtensions(
		extension.Table,
		extension.Strikethrough,
		extension.TaskList,
	))
	doc := md.Parser().Parse(text.NewReader(source))

	ctx := NewContext(ew, r.caps, a, r.theme, r.caps.Cols)
	highlighter := highlight.NewHighlighter(r.theme, a)
	images := image.NewRenderer(r.caps, a)
	headings := heading.NewRenderer(r.caps, a, r.theme, images, r.opts.NoHeadingImages)

	v := &visitor{
		ctx:         ctx,
		ansi:        a,
		theme:       r.theme,
		highlighter: highlighter,
		images:      images,
		headings:    headings,
		baseDir:     r.baseDir,
		source:      source,
	}
	v.renderDocument(doc)
}

// errWriter short-circuits once the underlying writer fails (e.g. a broken pipe), so the rest of
// the render becomes a cheap no-op instead of repeatedly failing syscalls.
type errWriter struct {
	w   io.Writer
	err error
}

func (e *errWriter) Write(p []byte) (int, error) {
	if e.err != nil {
		return len(p), nil
	}
	n, err := e.w.Write(p)
	e.err = err
	return n, err
}
