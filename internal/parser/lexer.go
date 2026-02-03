package parser

import (
	"fmt"
	"unicode"
)

type tokenType int

const (
	tokenEOF tokenType = iota
	tokenIdent
	tokenNumber
	tokenString
	tokenLBrace
	tokenRBrace
	tokenLParen
	tokenRParen
	tokenLBracket
	tokenRBracket
	tokenComma
	tokenSemicolon
	tokenColon
	tokenDot
	tokenAssign
	tokenArrow
	tokenDoubleColon
	tokenPlus
	tokenMinus
	tokenStar
	tokenSlash
	tokenPercent
	tokenEqEq
	tokenNotEq
	tokenLT
	tokenLTE
	tokenGT
	tokenGTE
	tokenAndAnd
	tokenOrOr
	tokenBang
	tokenAt
)

type token struct {
	typ    tokenType
	lexeme string
	line   int
	col    int
}

func lex(input string) ([]token, error) {
	var tokens []token
	line := 1
	col := 1
	for i := 0; i < len(input); {
		ch := input[i]

		if ch == ' ' || ch == '\t' || ch == '\r' {
			i++
			col++
			continue
		}
		if ch == '\n' {
			i++
			line++
			col = 1
			continue
		}

		if ch == '/' && i+1 < len(input) && input[i+1] == '/' {
			i += 2
			col += 2
			for i < len(input) && input[i] != '\n' {
				i++
				col++
			}
			continue
		}

		if ch == '/' && i+1 < len(input) && input[i+1] == '*' {
			i += 2
			col += 2
			for i < len(input)-1 {
				if input[i] == '\n' {
					line++
					col = 1
					i++
					continue
				}
				if input[i] == '*' && input[i+1] == '/' {
					i += 2
					col += 2
					break
				}
				i++
				col++
			}
			continue
		}

		startCol := col
		if isIdentStart(ch) {
			start := i
			for i < len(input) && isIdentPart(input[i]) {
				i++
				col++
			}
			lexeme := input[start:i]
			tokens = append(tokens, token{typ: tokenIdent, lexeme: lexeme, line: line, col: startCol})
			continue
		}

		if isDigit(ch) {
			start := i
			for i < len(input) && (isDigit(input[i]) || input[i] == '.') {
				i++
				col++
			}
			lexeme := input[start:i]
			tokens = append(tokens, token{typ: tokenNumber, lexeme: lexeme, line: line, col: startCol})
			continue
		}

		if ch == '"' {
			start := i
			i++
			col++
			escaped := false
			for i < len(input) {
				if input[i] == '\n' {
					return nil, fmt.Errorf("unterminated string at %d:%d", line, startCol)
				}
				if !escaped && input[i] == '"' {
					i++
					col++
					lexeme := input[start:i]
					tokens = append(tokens, token{typ: tokenString, lexeme: lexeme, line: line, col: startCol})
					break
				}
				if input[i] == '\\' && !escaped {
					escaped = true
				} else {
					escaped = false
				}
				i++
				col++
			}
			if i >= len(input) {
				return nil, fmt.Errorf("unterminated string at %d:%d", line, startCol)
			}
			continue
		}

		switch {
		case ch == ':' && i+1 < len(input) && input[i+1] == ':':
			tokens = append(tokens, token{typ: tokenDoubleColon, lexeme: "::", line: line, col: startCol})
			i += 2
			col += 2
			continue
		case ch == '-' && i+1 < len(input) && input[i+1] == '>':
			tokens = append(tokens, token{typ: tokenArrow, lexeme: "->", line: line, col: startCol})
			i += 2
			col += 2
			continue
		case ch == '=' && i+1 < len(input) && input[i+1] == '=':
			tokens = append(tokens, token{typ: tokenEqEq, lexeme: "==", line: line, col: startCol})
			i += 2
			col += 2
			continue
		case ch == '!' && i+1 < len(input) && input[i+1] == '=':
			tokens = append(tokens, token{typ: tokenNotEq, lexeme: "!=", line: line, col: startCol})
			i += 2
			col += 2
			continue
		case ch == '<' && i+1 < len(input) && input[i+1] == '=':
			tokens = append(tokens, token{typ: tokenLTE, lexeme: "<=", line: line, col: startCol})
			i += 2
			col += 2
			continue
		case ch == '>' && i+1 < len(input) && input[i+1] == '=':
			tokens = append(tokens, token{typ: tokenGTE, lexeme: ">=", line: line, col: startCol})
			i += 2
			col += 2
			continue
		case ch == '&' && i+1 < len(input) && input[i+1] == '&':
			tokens = append(tokens, token{typ: tokenAndAnd, lexeme: "&&", line: line, col: startCol})
			i += 2
			col += 2
			continue
		case ch == '|' && i+1 < len(input) && input[i+1] == '|':
			tokens = append(tokens, token{typ: tokenOrOr, lexeme: "||", line: line, col: startCol})
			i += 2
			col += 2
			continue
		}

		switch ch {
		case '{':
			tokens = append(tokens, token{typ: tokenLBrace, lexeme: "{", line: line, col: startCol})
		case '}':
			tokens = append(tokens, token{typ: tokenRBrace, lexeme: "}", line: line, col: startCol})
		case '(':
			tokens = append(tokens, token{typ: tokenLParen, lexeme: "(", line: line, col: startCol})
		case ')':
			tokens = append(tokens, token{typ: tokenRParen, lexeme: ")", line: line, col: startCol})
		case '[':
			tokens = append(tokens, token{typ: tokenLBracket, lexeme: "[", line: line, col: startCol})
		case ']':
			tokens = append(tokens, token{typ: tokenRBracket, lexeme: "]", line: line, col: startCol})
		case ',':
			tokens = append(tokens, token{typ: tokenComma, lexeme: ",", line: line, col: startCol})
		case ';':
			tokens = append(tokens, token{typ: tokenSemicolon, lexeme: ";", line: line, col: startCol})
		case ':':
			tokens = append(tokens, token{typ: tokenColon, lexeme: ":", line: line, col: startCol})
		case '.':
			tokens = append(tokens, token{typ: tokenDot, lexeme: ".", line: line, col: startCol})
		case '=':
			tokens = append(tokens, token{typ: tokenAssign, lexeme: "=", line: line, col: startCol})
		case '+':
			tokens = append(tokens, token{typ: tokenPlus, lexeme: "+", line: line, col: startCol})
		case '-':
			tokens = append(tokens, token{typ: tokenMinus, lexeme: "-", line: line, col: startCol})
		case '*':
			tokens = append(tokens, token{typ: tokenStar, lexeme: "*", line: line, col: startCol})
		case '/':
			tokens = append(tokens, token{typ: tokenSlash, lexeme: "/", line: line, col: startCol})
		case '%':
			tokens = append(tokens, token{typ: tokenPercent, lexeme: "%", line: line, col: startCol})
		case '<':
			tokens = append(tokens, token{typ: tokenLT, lexeme: "<", line: line, col: startCol})
		case '>':
			tokens = append(tokens, token{typ: tokenGT, lexeme: ">", line: line, col: startCol})
		case '!':
			tokens = append(tokens, token{typ: tokenBang, lexeme: "!", line: line, col: startCol})
		case '@':
			tokens = append(tokens, token{typ: tokenAt, lexeme: "@", line: line, col: startCol})
		default:
			return nil, fmt.Errorf("unexpected character %q at %d:%d", ch, line, col)
		}
		i++
		col++
	}

	tokens = append(tokens, token{typ: tokenEOF, lexeme: "", line: line, col: col})
	return tokens, nil
}

func isIdentStart(ch byte) bool {
	return ch == '_' || unicode.IsLetter(rune(ch))
}

func isIdentPart(ch byte) bool {
	return isIdentStart(ch) || isDigit(ch)
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
