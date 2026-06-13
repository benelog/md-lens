package highlight

import "strings"

// Registry maps a fenced-code-block language name (and common aliases) to a Tokenizer.
// Unknown or empty languages fall back to a plain tokenizer (single PLAIN token).
type Registry struct {
	byName map[string]Tokenizer
}

// NewRegistry builds a registry with all supported languages registered.
func NewRegistry() *Registry {
	r := &Registry{byName: make(map[string]Tokenizer)}
	r.register(newJavaTokenizer(), "java")
	r.register(newPythonTokenizer(), "python", "py")
	r.register(newJavaScriptTokenizer(), "javascript", "js", "jsx", "node")
	r.register(newTypeScriptTokenizer(), "typescript", "ts", "tsx")
	r.register(newRustTokenizer(), "rust", "rs")
	r.register(newGoTokenizer(), "go", "golang")
	r.register(newCTokenizer(), "c", "h")
	r.register(newCppTokenizer(), "cpp", "c++", "cc", "cxx", "hpp")
	r.register(newJSONTokenizer(), "json", "json5")
	r.register(newYAMLTokenizer(), "yaml", "yml")
	r.register(newBashTokenizer(), "bash", "sh", "shell", "zsh")
	r.register(newSQLTokenizer(), "sql")
	r.register(newXMLTokenizer(), "xml", "html", "htm", "svg")
	return r
}

func (r *Registry) register(t Tokenizer, names ...string) {
	for _, n := range names {
		r.byName[n] = t
	}
}

// ForLanguage returns the tokenizer for the fenced info string, normalizing aliases.
func (r *Registry) ForLanguage(info string) Tokenizer {
	key := strings.ToLower(strings.TrimSpace(info))
	if sp := strings.IndexByte(key, ' '); sp >= 0 {
		key = key[:sp]
	}
	if t, ok := r.byName[key]; ok {
		return t
	}
	return plainTokenizer{}
}

// plainTokenizer returns the whole input as one PLAIN token.
type plainTokenizer struct{}

func (plainTokenizer) Tokenize(code string) []Token {
	return []Token{{Plain, code}}
}

// -------------------------------------------------------------------------
// Languages. Rules are registered in priority order: comments → strings →
// annotations → numbers → keywords → types/builtins → functions → operators.
// -------------------------------------------------------------------------

func newJavaTokenizer() Tokenizer {
	t := &regexTokenizer{}
	t.cLineComment()
	t.blockComment()
	t.doubleString()
	t.singleString()
	t.annotation()
	t.number()
	t.keywords("abstract", "assert", "boolean", "break", "byte", "case", "catch", "char",
		"class", "const", "continue", "default", "do", "double", "else", "enum",
		"extends", "final", "finally", "float", "for", "goto", "if", "implements",
		"import", "instanceof", "int", "interface", "long", "native", "new", "package",
		"private", "protected", "public", "return", "short", "static", "strictfp",
		"super", "switch", "synchronized", "this", "throw", "throws", "transient",
		"try", "void", "volatile", "while", "var", "record", "sealed", "permits",
		"yield", "true", "false", "null")
	t.capitalizedTypes()
	t.functionCall()
	t.operators()
	t.punctuation()
	return t
}

func newPythonTokenizer() Tokenizer {
	t := &regexTokenizer{}
	t.hashComment()
	t.tripleString()
	t.doubleString()
	t.singleString()
	t.rule(Annotation, `@[A-Za-z_][\w.]*`)
	t.number()
	t.keywords("and", "as", "assert", "async", "await", "break", "class", "continue", "def",
		"del", "elif", "else", "except", "False", "finally", "for", "from", "global",
		"if", "import", "in", "is", "lambda", "None", "nonlocal", "not", "or", "pass",
		"raise", "return", "True", "try", "while", "with", "yield", "match", "case")
	t.builtins("print", "len", "range", "int", "str", "float", "list", "dict", "set", "tuple",
		"bool", "bytes", "object", "type", "self", "super", "open", "enumerate", "zip",
		"map", "filter", "sorted", "sum", "min", "max", "abs", "Exception")
	t.functionCall()
	t.operators()
	t.punctuation()
	return t
}

func buildJSRules(t *regexTokenizer) {
	t.cLineComment()
	t.blockComment()
	t.doubleString()
	t.singleString()
	t.backtickString()
	t.number()
	t.keywords("abstract", "async", "await", "break", "case", "catch", "class", "const",
		"continue", "debugger", "default", "delete", "do", "else", "enum", "export",
		"extends", "false", "finally", "for", "function", "if", "implements", "import",
		"in", "instanceof", "let", "new", "null", "of", "return", "static", "super",
		"switch", "this", "throw", "true", "try", "typeof", "undefined", "var", "void",
		"while", "with", "yield", "get", "set")
	t.builtins("console", "window", "document", "Math", "JSON", "Promise", "Array", "Object",
		"String", "Number", "Boolean", "Symbol", "Map", "Set", "NaN", "Infinity")
	t.capitalizedTypes()
	t.functionCall()
	t.operators()
	t.punctuation()
}

func newJavaScriptTokenizer() Tokenizer {
	t := &regexTokenizer{}
	buildJSRules(t)
	return t
}

func newTypeScriptTokenizer() Tokenizer {
	t := &regexTokenizer{}
	buildJSRules(t)
	// Extra TS keywords/types layered after the JS ones (first-match wins by order, but JS
	// keywords already cover most; add the TS-only words).
	t.keywords("interface", "type", "namespace", "declare", "readonly", "abstract", "as",
		"is", "keyof", "infer", "satisfies", "override", "public", "private",
		"protected", "enum")
	t.types("number", "string", "boolean", "any", "unknown", "never", "void", "object",
		"symbol", "bigint")
	return t
}

func newRustTokenizer() Tokenizer {
	t := &regexTokenizer{}
	t.cLineComment()
	t.blockComment()
	t.doubleString()
	t.singleString()
	t.rule(Annotation, `#!?\[[^\]]*\]`)
	t.rule(Function, `\b[A-Za-z_]\w*!`)
	t.number()
	t.keywords("as", "break", "const", "continue", "crate", "dyn", "else", "enum", "extern",
		"false", "fn", "for", "if", "impl", "in", "let", "loop", "match", "mod", "move",
		"mut", "pub", "ref", "return", "self", "Self", "static", "struct", "super",
		"trait", "true", "type", "unsafe", "use", "where", "while", "async", "await")
	t.types("i8", "i16", "i32", "i64", "i128", "isize", "u8", "u16", "u32", "u64", "u128",
		"usize", "f32", "f64", "bool", "char", "str", "String", "Vec", "Option",
		"Result", "Box", "Rc", "Arc")
	t.capitalizedTypes()
	t.functionCall()
	t.operators()
	t.punctuation()
	return t
}

func newGoTokenizer() Tokenizer {
	t := &regexTokenizer{}
	t.cLineComment()
	t.blockComment()
	t.doubleString()
	t.backtickString()
	t.singleString()
	t.number()
	t.keywords("break", "case", "chan", "const", "continue", "default", "defer", "else",
		"fallthrough", "for", "func", "go", "goto", "if", "import", "interface", "map",
		"package", "range", "return", "select", "struct", "switch", "type", "var")
	t.types("bool", "string", "int", "int8", "int16", "int32", "int64", "uint", "uint8",
		"uint16", "uint32", "uint64", "uintptr", "byte", "rune", "float32", "float64",
		"complex64", "complex128", "error")
	t.builtins("append", "cap", "close", "complex", "copy", "delete", "imag", "len", "make",
		"new", "panic", "print", "println", "real", "recover", "true", "false", "nil", "iota")
	t.functionCall()
	t.operators()
	t.punctuation()
	return t
}

func cKeywordList() []string {
	return []string{"auto", "break", "case", "char", "const", "continue", "default", "do",
		"double", "else", "enum", "extern", "float", "for", "goto", "if", "inline",
		"int", "long", "register", "restrict", "return", "short", "signed", "sizeof",
		"static", "struct", "switch", "typedef", "union", "unsigned", "void",
		"volatile", "while"}
}

func buildCTokenizer(extra func(*regexTokenizer)) Tokenizer {
	t := &regexTokenizer{}
	t.cLineComment()
	t.blockComment()
	t.doubleString()
	t.singleString()
	t.rule(Annotation, `#[ \t]*[A-Za-z_]+`)
	t.number()
	t.keywords(cKeywordList()...)
	if extra != nil {
		extra(t)
	}
	t.capitalizedTypes()
	t.functionCall()
	t.operators()
	t.punctuation()
	return t
}

func newCTokenizer() Tokenizer {
	return buildCTokenizer(nil)
}

func newCppTokenizer() Tokenizer {
	return buildCTokenizer(func(t *regexTokenizer) {
		t.keywords("class", "namespace", "template", "typename", "public", "private", "protected",
			"virtual", "override", "final", "new", "delete", "this", "nullptr", "using",
			"friend", "operator", "constexpr", "noexcept", "explicit", "mutable", "try",
			"catch", "throw", "and", "or", "not", "auto", "decltype", "static_cast",
			"dynamic_cast", "const_cast", "reinterpret_cast", "bool", "true", "false")
	})
}

func newJSONTokenizer() Tokenizer {
	t := &regexTokenizer{}
	t.doubleString()
	t.number()
	t.keywords("true", "false", "null")
	t.punctuation()
	return t
}

func newYAMLTokenizer() Tokenizer {
	t := &regexTokenizer{}
	t.hashComment()
	t.doubleString()
	t.singleString()
	t.rule(Keyword, `[A-Za-z_][\w-]*(?=\s*:)`)
	t.number()
	t.keywords("true", "false", "null", "yes", "no", "on", "off")
	t.rule(Operator, `[-:|>]`)
	return t
}

func newBashTokenizer() Tokenizer {
	t := &regexTokenizer{}
	t.hashComment()
	t.doubleString()
	t.singleString()
	t.rule(Type, `\$\{[^}]*\}`)
	t.rule(Type, `\$[A-Za-z_]\w*`)
	t.number()
	t.keywords("if", "then", "else", "elif", "fi", "case", "esac", "for", "while", "do",
		"done", "in", "function", "select", "until", "time", "return", "break", "continue")
	t.builtins("echo", "cd", "ls", "export", "source", "local", "read", "set", "unset",
		"alias", "exit", "printf", "test", "eval", "exec", "trap", "shift")
	t.functionCall()
	t.operators()
	t.punctuation()
	return t
}

func newSQLTokenizer() Tokenizer {
	t := &regexTokenizer{}
	t.dashComment()
	t.blockComment()
	t.singleString()
	t.doubleString()
	t.number()
	t.keywordsCI("select", "from", "where", "insert", "into", "values", "update", "set",
		"delete", "create", "table", "drop", "alter", "add", "column", "join", "inner",
		"left", "right", "outer", "full", "on", "group", "by", "order", "having",
		"limit", "offset", "as", "and", "or", "not", "null", "is", "in", "like",
		"between", "distinct", "union", "all", "primary", "key", "foreign", "references",
		"index", "view", "default", "case", "when", "then", "else", "end", "asc", "desc",
		"count", "sum", "avg", "min", "max")
	t.operators()
	t.punctuation()
	return t
}

func newXMLTokenizer() Tokenizer {
	t := &regexTokenizer{}
	t.rule(Comment, `<!--[\s\S]*?-->`)
	t.doubleString()
	t.singleString()
	t.rule(Keyword, `</?[A-Za-z][\w:-]*`)
	t.rule(Keyword, `/?>`)
	t.rule(Type, `[A-Za-z_][\w:-]*(?=\s*=)`)
	t.rule(Builtin, `&[A-Za-z#0-9]+;`)
	return t
}
