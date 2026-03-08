package parser

import (
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
)

var (
	liaLexer = lexer.MustStateful(lexer.Rules{
		"Root": {
			{Name: "Comment", Pattern: `//[^\n]*|/\*[^*]*\*+(?:[^/*][^*]*\*+)*/`},
			{Name: "Whitespace", Pattern: `\s+`},
			{Name: "DoubleColon", Pattern: `::`},
			{Name: "Arrow", Pattern: `->`},
			{Name: "EqEq", Pattern: `==`},
			{Name: "NotEq", Pattern: `!=`},
			{Name: "LTE", Pattern: `<=`},
			{Name: "GTE", Pattern: `>=`},
			{Name: "AndAnd", Pattern: `&&`},
			{Name: "OrOr", Pattern: `\|\|`},
			{Name: "Dot", Pattern: `\.`},
			{Name: "Colon", Pattern: `:`},
			{Name: "At", Pattern: `@`},
			{Name: "Number", Pattern: `\d+(?:\.\d+)*`},
			{Name: "String", Pattern: `"(?:\\.|[^"\\])*"`},
			{Name: "Keyword", Pattern: `\b(?:input|output|effects)\b`},
			{Name: "Ident", Pattern: `[A-Za-z_][A-Za-z0-9_]*`},
			{Name: "Punct", Pattern: `[{}()\[\],;=+\-*/%<>!]`},
		},
	})
	liaSymbols = liaLexer.Symbols()
)

// DebugLexer exposes the lexer for debugging within this module.
func DebugLexer() *lexer.StatefulDefinition {
	return liaLexer
}

func isTokenType(tok lexer.Token, name string) bool {
	if tok.Type == lexer.EOF {
		return name == "EOF"
	}
	typ, ok := liaSymbols[name]
	if !ok {
		return false
	}
	return tok.Type == typ
}

func lexTokens(filename, input string) ([]lexer.Token, error) {
	lex, err := liaLexer.Lex(filename, strings.NewReader(input))
	if err != nil {
		return nil, err
	}
	return lexer.ConsumeAll(lex)
}

// DebugLexTokens exposes lexing for diagnostics within this module.
func DebugLexTokens(filename, input string) ([]lexer.Token, error) {
	return lexTokens(filename, input)
}
