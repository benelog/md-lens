package highlight

// Theme is a truecolor color scheme. It holds both syntax-token colors and a few document-level
// colors (headings, links, rules, ...) so the renderer has one place to ask. Colors are
// {r, g, b}; the ansi helper downgrades them to 256/16-color when needed.
type Theme struct {
	syntax         map[TokenType][3]int
	headingPalette [6][3]int

	// Document-level colors.
	Heading   [3]int
	Heading2  [3]int
	Link      [3]int
	CodeFg    [3]int
	CodeBg    [3]int
	QuoteBar  [3]int
	QuoteText [3]int
	Rule      [3]int
	Marker    [3]int
	TaskDone  [3]int
	Dim       [3]int
}

// Default returns the default theme.
func Default() *Theme {
	return &Theme{
		Heading:  [3]int{97, 175, 239}, // blue
		Heading2: [3]int{130, 170, 255},
		Link:     [3]int{86, 182, 194}, // cyan
		// Distinct color per heading level (1..6) so each level reads differently.
		headingPalette: [6][3]int{
			{97, 175, 239},  // h1 blue
			{78, 201, 176},  // h2 teal
			{152, 195, 121}, // h3 green
			{229, 192, 123}, // h4 yellow
			{224, 108, 117}, // h5 red
			{198, 120, 221}, // h6 magenta
		},
		CodeFg:    [3]int{206, 145, 120}, // soft orange
		CodeBg:    [3]int{45, 45, 45},    // dim grey
		QuoteBar:  [3]int{97, 175, 239},
		QuoteText: [3]int{160, 160, 160},
		Rule:      [3]int{90, 90, 90},
		Marker:    [3]int{120, 200, 120},
		TaskDone:  [3]int{120, 200, 120},
		Dim:       [3]int{110, 110, 110},
		syntax: map[TokenType][3]int{
			Keyword:     {197, 134, 192},
			Type:        {78, 201, 176},
			String:      {206, 145, 120},
			Number:      {181, 206, 168},
			Comment:     {106, 153, 85},
			Function:    {220, 220, 170},
			Annotation:  {215, 186, 125},
			Operator:    {212, 212, 212},
			Punctuation: {200, 200, 200},
			Builtin:     {86, 156, 214},
			Plain:       {212, 212, 212},
		},
	}
}

// HeadingColor returns a distinct color for the heading level (1..6, clamped).
func (t *Theme) HeadingColor(level int) [3]int {
	return t.headingPalette[min(6, max(1, level))-1]
}

// Color returns the token color, falling back to PLAIN.
func (t *Theme) Color(typ TokenType) [3]int {
	if c, ok := t.syntax[typ]; ok {
		return c
	}
	return t.syntax[Plain]
}

// Italic reports whether the token type should be rendered italic.
func (t *Theme) Italic(typ TokenType) bool {
	return typ == Comment
}
