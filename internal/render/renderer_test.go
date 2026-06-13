package render

import (
	"os"
	"strings"
	"testing"

	"github.com/benelog/md-lens/internal/cli"
	"github.com/benelog/md-lens/internal/term"
)

func renderSamplePlain(t *testing.T) string {
	t.Helper()
	md, err := os.ReadFile("testdata/sample.md")
	if err != nil {
		t.Fatalf("read sample.md: %v", err)
	}
	caps := term.Capabilities{IsTTY: false, Cols: 80, Rows: 24, CellPxW: 10, CellPxH: 20,
		Color: term.None, Graphics: term.GraphicsNone}
	opts := cli.Args{Plain: true, ForceGraphics: term.GraphicsAuto}
	var sb strings.Builder
	NewRenderer(caps, opts, "testdata").Render(string(md), &sb)
	return sb.String()
}

func TestPlainOutputHasNoEscapeSequences(t *testing.T) {
	out := renderSamplePlain(t)
	if strings.Contains(out, "\x1b") {
		t.Error("NONE color depth must produce no ANSI escapes")
	}
}

func TestRendersAllMajorBlockKinds(t *testing.T) {
	out := renderSamplePlain(t)
	wants := []struct{ text, desc string }{
		{"mdl Sample Document", "heading text"},
		{"First bullet", "bullet list item"},
		{"•", "bullet glyph"},
		{"☑", "checked task list item"},
		{"public class Hello", "code block content"},
		{"Feature", "table header cell"},
		{"│", "table column separator"},
		{"https://example.com", "link destination"},
		{"[image:", "image alt-text fallback"},
		{"─", "thematic break / divider"},
	}
	for _, w := range wants {
		if !strings.Contains(out, w.text) {
			t.Errorf("missing %s (%q)", w.desc, w.text)
		}
	}
}
