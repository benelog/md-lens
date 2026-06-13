package render

import (
	"strings"

	"github.com/benelog/md-lens/internal/ansi"
)

// wrapText word-wraps styled text to a visible width. ANSI escape sequences inside words are kept
// intact and do not count toward the width; existing \n characters are honored as hard breaks.
func wrapText(styled string, width int) []string {
	if width < 1 {
		width = 1
	}
	var out []string
	for _, line := range strings.Split(styled, "\n") { // keep all hard lines (Java split(-1))
		wrapLine(line, width, &out)
	}
	return out
}

func wrapLine(line string, width int, out *[]string) {
	if ansi.Width(line) <= width {
		*out = append(*out, line)
		return
	}
	var cur strings.Builder
	curWidth := 0
	for _, word := range splitWords(line) {
		wordWidth := ansi.Width(word)
		switch {
		case curWidth == 0:
			cur.WriteString(word)
			curWidth = wordWidth
		case curWidth+1+wordWidth <= width:
			cur.WriteByte(' ')
			cur.WriteString(word)
			curWidth += 1 + wordWidth
		default:
			*out = append(*out, cur.String())
			cur.Reset()
			cur.WriteString(word)
			curWidth = wordWidth
		}
	}
	*out = append(*out, cur.String())
}

// splitWords splits on single spaces, dropping trailing empty fields to match Java's
// String.split(" ") (limit 0) semantics.
func splitWords(line string) []string {
	parts := strings.Split(line, " ")
	end := len(parts)
	for end > 0 && parts[end-1] == "" {
		end--
	}
	return parts[:end]
}
