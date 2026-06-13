package highlight

import "testing"

func tokenize(lang, code string) []Token {
	return NewRegistry().ForLanguage(lang).Tokenize(code)
}

func has(tokens []Token, typ TokenType, text string) bool {
	for _, t := range tokens {
		if t.Type == typ && t.Text == text {
			return true
		}
	}
	return false
}

func anyType(tokens []Token, typ TokenType) bool {
	for _, t := range tokens {
		if t.Type == typ {
			return true
		}
	}
	return false
}

func TestJavaKeywordNumberComment(t *testing.T) {
	tokens := tokenize("java", "int x = 5; // c")
	if !has(tokens, Keyword, "int") {
		t.Error("missing KEYWORD int")
	}
	if !has(tokens, Number, "5") {
		t.Error("missing NUMBER 5")
	}
	if !has(tokens, Comment, "// c") {
		t.Error("missing COMMENT // c")
	}
}

func TestJavaString(t *testing.T) {
	tokens := tokenize("java", `"hi"`)
	if !has(tokens, String, `"hi"`) {
		t.Error(`missing STRING "hi"`)
	}
}

func TestKeywordNotMatchedInsideIdentifier(t *testing.T) {
	tokens := tokenize("java", "internal")
	if anyType(tokens, Keyword) {
		t.Error("'int' must not be highlighted inside 'internal'")
	}
}

func TestPythonDefFunctionComment(t *testing.T) {
	tokens := tokenize("py", "def f(): # hi")
	if !has(tokens, Keyword, "def") {
		t.Error("missing KEYWORD def")
	}
	if !has(tokens, Function, "f") {
		t.Error("missing FUNCTION f")
	}
	if !has(tokens, Comment, "# hi") {
		t.Error("missing COMMENT # hi")
	}
}

func TestUnknownLanguageIsSinglePlainToken(t *testing.T) {
	tokens := tokenize("brainfuck", "abc")
	if len(tokens) != 1 {
		t.Fatalf("len = %d, want 1", len(tokens))
	}
	if tokens[0].Type != Plain {
		t.Errorf("type = %v, want Plain", tokens[0].Type)
	}
}
