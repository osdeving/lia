package parser

import "testing"

func TestLex_Tokens(t *testing.T) {
	src := `// line comment
/* block comment */
@gen { prompt_ref:"p\"1" }
a::b -> c <= d >= e == f != g && h || i
+ - * / % ! = ; , . : { } ( ) [ ]
123 4.56 "str\\n"
`
	tokens, err := lex(src)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	want := map[tokenType]bool{
		tokenDoubleColon: true,
		tokenArrow:       true,
		tokenLTE:         true,
		tokenGTE:         true,
		tokenEqEq:        true,
		tokenNotEq:       true,
		tokenAndAnd:      true,
		tokenOrOr:        true,
		tokenPlus:        true,
		tokenMinus:       true,
		tokenStar:        true,
		tokenSlash:       true,
		tokenPercent:     true,
		tokenBang:        true,
		tokenAssign:      true,
		tokenSemicolon:   true,
		tokenComma:       true,
		tokenDot:         true,
		tokenColon:       true,
		tokenLBrace:      true,
		tokenRBrace:      true,
		tokenLParen:      true,
		tokenRParen:      true,
		tokenLBracket:    true,
		tokenRBracket:    true,
		tokenNumber:      true,
		tokenString:      true,
		tokenAt:          true,
	}
	seen := map[tokenType]bool{}
	for _, tok := range tokens {
		seen[tok.typ] = true
	}
	for typ := range want {
		if !seen[typ] {
			t.Fatalf("expected token type %v", typ)
		}
	}
}
