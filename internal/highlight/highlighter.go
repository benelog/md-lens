package highlight

import (
	"strings"

	"github.com/benelog/md-lens/internal/ansi"
)

// Highlighter turns a code block into styled tokens. The renderer drives layout (indenting,
// gutters) by walking Tokenize and styling each piece with StylePiece; Highlight is a convenience
// that returns the whole styled string in one go (used by tests).
type Highlighter struct {
	registry *Registry
	theme    *Theme
	ansi     *ansi.Ansi
}

// NewHighlighter returns a highlighter using the given theme and ansi emitter.
func NewHighlighter(theme *Theme, a *ansi.Ansi) *Highlighter {
	return &Highlighter{registry: NewRegistry(), theme: theme, ansi: a}
}

// Tokenize splits code into classified tokens for the given language.
func (h *Highlighter) Tokenize(code, lang string) []Token {
	return h.registry.ForLanguage(lang).Tokenize(code)
}

// StylePiece wraps a single piece of token text in the theme's color (and italic for comments).
func (h *Highlighter) StylePiece(typ TokenType, piece string) string {
	if !h.ansi.Enabled() || typ == Plain || piece == "" {
		return piece
	}
	var sb strings.Builder
	sb.WriteString(h.ansi.Fg3(h.theme.Color(typ)))
	if h.theme.Italic(typ) {
		sb.WriteString(h.ansi.Italic())
	}
	sb.WriteString(piece)
	sb.WriteString(h.ansi.Reset())
	return sb.String()
}

// Highlight returns the whole code block as one styled string.
func (h *Highlighter) Highlight(code, lang string) string {
	var sb strings.Builder
	for _, t := range h.Tokenize(code, lang) {
		sb.WriteString(h.StylePiece(t.Type, t.Text))
	}
	return sb.String()
}
