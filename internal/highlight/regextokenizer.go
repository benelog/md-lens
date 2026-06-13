package highlight

import (
	"strings"

	"github.com/dlclark/regexp2"
)

// regexTokenizer is an ordered-rule tokenizer engine. Rules are registered as (TokenType, regex);
// at each position the first rule that matches wins, otherwise one PLAIN char is consumed.
// Comment/string rules must be registered before keyword/operator rules.
//
// Each pattern is anchored with \G and matched with FindRunesMatchStartingAt, which makes regexp2
// attempt the match only at the current position while still seeing the whole input — so \b word
// boundaries and lookahead behave like Java's transparent, non-anchoring matcher bounds.
type regexTokenizer struct {
	rules []rule
}

type rule struct {
	typ TokenType
	re  *regexp2.Regexp
}

func (t *regexTokenizer) Tokenize(code string) []Token {
	runes := []rune(code)
	n := len(runes)
	out := make([]Token, 0, 16)
	plainStart := -1

	flushPlain := func(end int) {
		if plainStart >= 0 {
			out = append(out, Token{Plain, string(runes[plainStart:end])})
			plainStart = -1
		}
	}

	i := 0
	for i < n {
		matchedEnd := -1
		var matchedType TokenType
		for k := range t.rules {
			m, err := t.rules[k].re.FindRunesMatchStartingAt(runes, i)
			if err == nil && m != nil && m.Index == i && m.Length > 0 {
				matchedEnd = i + m.Length
				matchedType = t.rules[k].typ
				break
			}
		}
		if matchedEnd >= 0 {
			flushPlain(i)
			out = append(out, Token{matchedType, string(runes[i:matchedEnd])})
			i = matchedEnd
		} else {
			if plainStart < 0 {
				plainStart = i
			}
			i++
		}
	}
	flushPlain(n)
	return out
}

// rule registers a (type, regex) rule, anchored to the scan position with \G.
func (t *regexTokenizer) rule(typ TokenType, pattern string) {
	re := regexp2.MustCompile(`\G`+pattern, regexp2.None)
	t.rules = append(t.rules, rule{typ, re})
}

// --- common building blocks (call in order: comments → strings → ... → punctuation) ----------

func (t *regexTokenizer) cLineComment() { t.rule(Comment, `//[^\n]*`) }
func (t *regexTokenizer) hashComment()  { t.rule(Comment, `#[^\n]*`) }
func (t *regexTokenizer) dashComment()  { t.rule(Comment, `--[^\n]*`) }
func (t *regexTokenizer) blockComment() { t.rule(Comment, `/\*[\s\S]*?\*/`) }

func (t *regexTokenizer) doubleString()   { t.rule(String, `"(?:\\.|[^"\\\n])*"`) }
func (t *regexTokenizer) singleString()   { t.rule(String, `'(?:\\.|[^'\\\n])*'`) }
func (t *regexTokenizer) backtickString() { t.rule(String, "`(?:\\\\.|[^`\\\\])*`") }

func (t *regexTokenizer) tripleString() {
	t.rule(String, `"""[\s\S]*?"""`)
	t.rule(String, `'''[\s\S]*?'''`)
}

func (t *regexTokenizer) number() {
	t.rule(Number, `\b0[xX][0-9a-fA-F_]+\b`)
	t.rule(Number, `\b\d[\d_]*(?:\.\d+)?(?:[eE][+-]?\d+)?[fFdDlLuU]*\b`)
}

func (t *regexTokenizer) annotation() { t.rule(Annotation, `@[A-Za-z_][\w.]*`) }

func (t *regexTokenizer) keywords(kws ...string) {
	t.rule(Keyword, `\b(?:`+strings.Join(kws, "|")+`)\b`)
}

func (t *regexTokenizer) keywordsCI(kws ...string) {
	t.rule(Keyword, `(?i)\b(?:`+strings.Join(kws, "|")+`)\b`)
}

func (t *regexTokenizer) types(ts ...string) {
	t.rule(Type, `\b(?:`+strings.Join(ts, "|")+`)\b`)
}

func (t *regexTokenizer) builtins(bs ...string) {
	t.rule(Builtin, `\b(?:`+strings.Join(bs, "|")+`)\b`)
}

func (t *regexTokenizer) capitalizedTypes() { t.rule(Type, `\b[A-Z][A-Za-z0-9_]*\b`) }
func (t *regexTokenizer) functionCall()     { t.rule(Function, `\b[A-Za-z_]\w*(?=\s*\()`) }
func (t *regexTokenizer) operators()        { t.rule(Operator, `[+\-*/%=<>!&|^~?:]+`) }
func (t *regexTokenizer) punctuation()      { t.rule(Punctuation, `[\[\]{}().,;]`) }
