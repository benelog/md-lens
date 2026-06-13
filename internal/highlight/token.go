// Package highlight turns fenced code blocks into classified, colorizable tokens using a set of
// ordered-rule regex tokenizers (one per language).
package highlight

// TokenType is a category a tokenizer can assign to a slice of source code.
type TokenType int

const (
	Keyword TokenType = iota
	Type
	String
	Number
	Comment
	Function
	Annotation
	Operator
	Punctuation
	Builtin
	// Plain is anything not specially highlighted.
	Plain
)

// Token is a classified slice of source code. Text may span multiple lines.
type Token struct {
	Type TokenType
	Text string
}

// Tokenizer splits source code into classified tokens.
type Tokenizer interface {
	Tokenize(code string) []Token
}
