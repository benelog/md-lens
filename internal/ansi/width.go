package ansi

const escByte = rune(27)

// Width returns the visible width of a string: ANSI escape sequences count as zero, East-Asian
// wide characters count as two. Used so word-wrapping and gutters line up regardless of styling.
func Width(s string) int {
	r := []rune(s)
	n := len(r)
	width := 0
	i := 0
	for i < n {
		c := r[i]
		switch c {
		case escByte:
			i = skipEscape(r, i)
		case '\n', '\r':
			i++
		default:
			if isWide(c) {
				width += 2
			} else {
				width++
			}
			i++
		}
	}
	return width
}

// skipEscape returns the index just past the escape sequence starting at i (which is ESC).
func skipEscape(s []rune, i int) int {
	n := len(s)
	if i+1 >= n {
		return n
	}
	next := s[i+1]
	if next == '[' {
		// CSI: ESC [ ... final-byte(0x40-0x7e)
		j := i + 2
		for j < n && (s[j] < 0x40 || s[j] > 0x7e) {
			j++
		}
		return min(j+1, n)
	}
	if next == ']' {
		// OSC: ESC ] ... terminated by BEL or ST (ESC \)
		j := i + 2
		for j < n {
			ch := s[j]
			if ch == 0x07 {
				return j + 1
			}
			if ch == escByte && j+1 < n && s[j+1] == '\\' {
				return j + 2
			}
			j++
		}
		return n
	}
	// Other two-char escapes (e.g. ESC \, ESC =) — skip both.
	return i + 2
}

// isWide is a rough East-Asian wide / fullwidth detection — good enough for layout.
func isWide(cp rune) bool {
	return (cp >= 0x1100 && cp <= 0x115F) || // Hangul Jamo
		(cp >= 0x2E80 && cp <= 0x303E) || // CJK radicals, Kangxi
		(cp >= 0x3041 && cp <= 0x33FF) || // Hiragana, Katakana, CJK symbols
		(cp >= 0x3400 && cp <= 0x4DBF) || // CJK Ext A
		(cp >= 0x4E00 && cp <= 0x9FFF) || // CJK Unified
		(cp >= 0xA000 && cp <= 0xA4CF) || // Yi
		(cp >= 0xAC00 && cp <= 0xD7A3) || // Hangul syllables
		(cp >= 0xF900 && cp <= 0xFAFF) || // CJK compatibility
		(cp >= 0xFE30 && cp <= 0xFE4F) || // CJK compatibility forms
		(cp >= 0xFF00 && cp <= 0xFF60) || // Fullwidth forms
		(cp >= 0xFFE0 && cp <= 0xFFE6) ||
		(cp >= 0x1F300 && cp <= 0x1FAFF) || // emoji & pictographs
		(cp >= 0x20000 && cp <= 0x3FFFD) // CJK Ext B+
}
