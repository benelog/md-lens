package render

// bulletGlyph returns the bullet glyph for a list nesting depth.
func bulletGlyph(depth int) string {
	switch ((depth % 3) + 3) % 3 {
	case 0:
		return "•"
	case 1:
		return "◦"
	default:
		return "▪"
	}
}
