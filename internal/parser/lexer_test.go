package parser

import (
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
)

func TestLex_Tokens(t *testing.T) {
	src := `// line comment
/* block comment */
@gen { prompt_ref:"p\"1" }
a::b -> c <= d >= e == f != g && h || i
+ - * / % ! = ; , . : { } ( ) [ ]
123 4.56 "str\\n"
`
	tokens, err := lexTokens("test.lia", src)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	mustHaveValues := []string{"::", "->", "<=", ">=", "==", "!=", "&&", "||", "+", "-", "*", "/", "%", "!", "=", ";", ",", ".", ":", "{", "}", "(", ")", "[", "]", "@"}
	for _, val := range mustHaveValues {
		if !hasTokenValue(tokens, val) {
			t.Fatalf("expected token value %q", val)
		}
	}
	if !hasTokenType(tokens, "Ident") {
		t.Fatalf("expected Ident token")
	}
	if !hasTokenType(tokens, "Number") {
		t.Fatalf("expected Number token")
	}
	if !hasTokenType(tokens, "String") {
		t.Fatalf("expected String token")
	}
}

func hasTokenValue(tokens []lexer.Token, value string) bool {
	for _, tok := range tokens {
		if tok.Value == value {
			return true
		}
	}
	return false
}

func hasTokenType(tokens []lexer.Token, name string) bool {
	for _, tok := range tokens {
		if isTokenType(tok, name) {
			return true
		}
	}
	return false
}
