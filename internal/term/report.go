package term

import (
	"fmt"
	"strings"
)

// FormatReport returns a human-readable summary of what mdl detected about the terminal (for --caps).
func FormatReport(caps Capabilities, headingImagesWanted bool) string {
	var color string
	switch caps.Color {
	case Truecolor:
		color = "truecolor (24-bit)"
	case Ansi256:
		color = "256 colors"
	case Ansi16:
		color = "16 colors"
	default:
		color = "none"
	}

	var graphics string
	switch caps.Graphics {
	case Kitty:
		graphics = "kitty graphics protocol"
	case Iterm2:
		graphics = "iTerm2 inline images"
	case Sixel:
		graphics = "sixel"
	case HalfBlock:
		graphics = "half-block (▀ truecolor fallback)"
	default:
		graphics = "none (images shown as alt text)"
	}

	fontHeadings := headingImagesWanted && caps.Graphics.IsPixelProtocol()
	headings := "styled text (font images need kitty/iTerm2)"
	if fontHeadings {
		headings = "large font images"
	}

	tty := "no (piped/redirected)"
	if caps.IsTTY {
		tty = "yes"
	}

	var sb strings.Builder
	sb.WriteString("mdl — detected terminal capabilities\n")
	sb.WriteString(row("interactive tty", tty))
	sb.WriteString(row("size", fmt.Sprintf("%d x %d cells", caps.Cols, caps.Rows)))
	sb.WriteString(row("cell size", fmt.Sprintf("~%d x %d px (estimated)", caps.CellPxW, caps.CellPxH)))
	sb.WriteString(row("color depth", color))
	sb.WriteString(row("images", graphics))
	sb.WriteString(row("headings", headings))
	return sb.String()
}

func row(key, value string) string {
	return "  " + pad(key+":", 18) + value + "\n"
}

func pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
