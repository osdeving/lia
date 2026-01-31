package parser

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/willams/lia/internal/ir"
)

// ParseFile parses a .lia source file into a minimal IR program.
func ParseFile(path string) (*ir.Program, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	lines, err := readLines(f)
	if err != nil {
		return nil, err
	}

	p := &ir.Program{Version: "0.1"}
	var currentProject *ir.Project
	var currentModule *ir.Module
	var scopeStack []string

	for i := 0; i < len(lines); i++ {
		raw := lines[i]
		line := cleanLine(raw)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "project ") {
			name := strings.Fields(strings.TrimPrefix(line, "project "))
			if len(name) > 0 {
				p.Projects = append(p.Projects, ir.Project{Name: name[0]})
				currentProject = &p.Projects[len(p.Projects)-1]
				scopeStack = append(scopeStack, "project")
			}
			continue
		}

		if strings.HasPrefix(line, "module ") {
			modLine := strings.TrimPrefix(line, "module ")
			modName := strings.Fields(modLine)
			if len(modName) == 0 {
				continue
			}
			role := ""
			if strings.Contains(line, " as ") {
				parts := strings.Split(line, " as ")
				if len(parts) == 2 {
					role = strings.Fields(parts[1])[0]
				}
			}
			p.Modules = append(p.Modules, ir.Module{Name: modName[0], Role: role})
			currentModule = &p.Modules[len(p.Modules)-1]
			scopeStack = append(scopeStack, "module")
			continue
		}

		if line == "}" {
			if len(scopeStack) == 0 {
				continue
			}
			top := scopeStack[len(scopeStack)-1]
			scopeStack = scopeStack[:len(scopeStack)-1]
			if top == "module" {
				currentModule = nil
			}
			if top == "project" {
				currentProject = nil
			}
			continue
		}

		if strings.HasPrefix(line, "repro ") && currentProject != nil {
			value := strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, "repro ")), ";")
			currentProject.Repro = ir.ReproProfile(value)
			continue
		}

		if strings.HasPrefix(line, "tape ") && currentProject != nil {
			// e.g. tape prompt_tape "./prompt-tape.json";
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				path := strings.TrimSuffix(parts[len(parts)-1], ";")
				path = strings.Trim(path, "\"")
				currentProject.Tape = path
			}
			continue
		}

		if strings.HasPrefix(line, "use pack ") && currentProject != nil {
			ref := strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, "use pack ")), ";")
			name, version := parsePackRef(ref)
			currentProject.Uses = append(currentProject.Uses, ir.PackRef{Name: name, Version: version})
			continue
		}

		if strings.HasPrefix(line, "usecase ") && currentModule != nil {
			name, params, effects, err := parseUsecaseHeader(line)
			if err != nil {
				return nil, err
			}
			blockLines, end, err := collectBlock(lines, i)
			if err != nil {
				return nil, err
			}
			stmts, err := parseStmts(blockLines)
			if err != nil {
				return nil, err
			}
			uc := ir.UsecaseDecl{
				Name:    name,
				Inputs:  params,
				Effects: effects,
				Body:    stmts,
			}
			currentModule.Usecases = append(currentModule.Usecases, uc)
			i = end
			continue
		}
	}

	return p, nil
}

func readLines(f *os.File) ([]string, error) {
	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return normalizeLines(lines), nil
}

func parsePackRef(ref string) (string, string) {
	ref = strings.TrimSpace(ref)
	if strings.Contains(ref, "@") {
		parts := strings.SplitN(ref, "@", 2)
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return ref, ""
}

func parseUsecaseHeader(line string) (string, []ir.Field, []string, error) {
	trim := strings.TrimSpace(strings.TrimPrefix(line, "usecase "))
	if !strings.Contains(trim, "(") || !strings.Contains(trim, ")") {
		return "", nil, nil, fmt.Errorf("invalid usecase header: %s", line)
	}
	name := strings.TrimSpace(trim[:strings.Index(trim, "(")])
	paramsPart := trim[strings.Index(trim, "(")+1 : strings.Index(trim, ")")]
	paramNames := splitCSV(paramsPart)
	var params []ir.Field
	for _, p := range paramNames {
		params = append(params, ir.Field{Name: p, Type: "Any"})
	}

	rest := strings.TrimSpace(trim[strings.Index(trim, ")")+1:])
	effects := []string{}
	if strings.HasPrefix(rest, "effects ") {
		rest = strings.TrimPrefix(rest, "effects ")
		if strings.Contains(rest, "{") {
			rest = strings.TrimSpace(rest[:strings.Index(rest, "{")])
		}
		effects = splitCSV(rest)
	}

	if !strings.Contains(line, "{") {
		return "", nil, nil, fmt.Errorf("usecase must open block on same line: %s", line)
	}

	return name, params, effects, nil
}

func collectBlock(lines []string, start int) ([]string, int, error) {
	depth := 0
	started := false
	var body []string

	for i := start; i < len(lines); i++ {
		line := lines[i]
		open := strings.Count(line, "{")
		close := strings.Count(line, "}")

		if !started {
			if open == 0 {
				continue
			}
			started = true
			depth += open - close
			if depth == 0 {
				return body, i, nil
			}
			continue
		}

		depth += open - close
		if depth < 0 {
			return nil, i, fmt.Errorf("unbalanced braces")
		}
		if depth == 0 {
			return body, i, nil
		}
		body = append(body, line)
	}

	return nil, len(lines) - 1, fmt.Errorf("unterminated block")
}

func parseStmts(lines []string) ([]ir.Stmt, error) {
	var stmts []ir.Stmt
	for i := 0; i < len(lines); i++ {
		line := cleanLine(lines[i])
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "if ") {
			condText := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "if "), "{"))
			cond, err := parseExpr(condText)
			if err != nil {
				return nil, err
			}
			block, end, err := collectBlock(lines, i)
			if err != nil {
				return nil, err
			}
			thenStmts, err := parseStmts(block)
			if err != nil {
				return nil, err
			}
			ifStmt := ir.IfStmt{Cond: *cond, Then: thenStmts}
			i = end

			// optional else
			j := i + 1
			for j < len(lines) && cleanLine(lines[j]) == "" {
				j++
			}
			if j < len(lines) && strings.HasPrefix(cleanLine(lines[j]), "else") {
				elseLine := cleanLine(lines[j])
				if !strings.Contains(elseLine, "{") {
					return nil, fmt.Errorf("else must open block on same line")
				}
				elseBlock, elseEnd, err := collectBlock(lines, j)
				if err != nil {
					return nil, err
				}
				elseStmts, err := parseStmts(elseBlock)
				if err != nil {
					return nil, err
				}
				ifStmt.Else = elseStmts
				i = elseEnd
			}

			stmts = append(stmts, ir.Stmt{Kind: "if", If: &ifStmt})
			continue
		}

		if strings.HasPrefix(line, "loop") {
			if !strings.Contains(line, "{") {
				return nil, fmt.Errorf("loop must open block on same line")
			}
			block, end, err := collectBlock(lines, i)
			if err != nil {
				return nil, err
			}
			body, err := parseStmts(block)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ir.Stmt{Kind: "loop", Loop: &ir.LoopStmt{Body: body}})
			i = end
			continue
		}

		if strings.HasPrefix(line, "break") {
			stmts = append(stmts, ir.Stmt{Kind: "break", Break: &ir.BreakStmt{}})
			continue
		}

		if strings.HasPrefix(line, "continue") {
			stmts = append(stmts, ir.Stmt{Kind: "continue", Continue: &ir.ContinueStmt{}})
			continue
		}

		if strings.HasPrefix(line, "return") {
			valueText := strings.TrimSpace(strings.TrimPrefix(line, "return"))
			valueText = strings.TrimSuffix(valueText, ";")
			if valueText == "" {
				stmts = append(stmts, ir.Stmt{Kind: "return", Return: &ir.ReturnStmt{}})
				continue
			}
			expr, err := parseExpr(valueText)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ir.Stmt{Kind: "return", Return: &ir.ReturnStmt{Value: expr}})
			continue
		}

		if strings.HasPrefix(line, "let ") {
			stmt, err := parseLet(line)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, stmt)
			continue
		}

		if isAssignment(line) {
			stmt, err := parseAssign(line)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, stmt)
			continue
		}

		exprLine := strings.TrimSuffix(line, ";")
		expr, err := parseExpr(exprLine)
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, ir.Stmt{Kind: "expr", Expr: expr})
	}

	return stmts, nil
}

func parseLet(line string) (ir.Stmt, error) {
	trim := strings.TrimSpace(strings.TrimPrefix(line, "let "))
	trim = strings.TrimSuffix(trim, ";")
	parts := strings.SplitN(trim, "=", 2)
	if len(parts) != 2 {
		return ir.Stmt{}, fmt.Errorf("invalid let: %s", line)
	}
	name := strings.TrimSpace(parts[0])
	valueText := strings.TrimSpace(parts[1])
	expr, err := parseExpr(valueText)
	if err != nil {
		return ir.Stmt{}, err
	}
	return ir.Stmt{Kind: "let", Let: &ir.LetStmt{Name: name, Value: *expr}}, nil
}

func parseAssign(line string) (ir.Stmt, error) {
	trim := strings.TrimSuffix(line, ";")
	parts := strings.SplitN(trim, "=", 2)
	if len(parts) != 2 {
		return ir.Stmt{}, fmt.Errorf("invalid assignment: %s", line)
	}
	name := strings.TrimSpace(parts[0])
	valueText := strings.TrimSpace(parts[1])
	expr, err := parseExpr(valueText)
	if err != nil {
		return ir.Stmt{}, err
	}
	return ir.Stmt{Kind: "assign", Assign: &ir.AssignStmt{Name: name, Value: *expr}}, nil
}

func isAssignment(line string) bool {
	inQuote := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if ch == '"' {
			inQuote = !inQuote
		}
		if inQuote {
			continue
		}
		if ch == '=' {
			prev := byte(0)
			next := byte(0)
			if i > 0 {
				prev = line[i-1]
			}
			if i+1 < len(line) {
				next = line[i+1]
			}
			if prev == '=' || prev == '!' || prev == '<' || prev == '>' || next == '=' {
				continue
			}
			return true
		}
	}
	return false
}

func cleanLine(line string) string {
	trim := strings.TrimSpace(line)
	if trim == "" {
		return ""
	}
	trim = stripInlineComment(trim)
	return strings.TrimSpace(trim)
}

func normalizeLines(lines []string) []string {
	var out []string
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "} else") {
			out = append(out, "}")
			out = append(out, strings.TrimSpace(strings.TrimPrefix(trim, "}")))
			continue
		}
		out = append(out, line)
	}
	return out
}

func stripInlineComment(line string) string {
	var out strings.Builder
	inQuote := false
	for i := 0; i < len(line); i++ {
		if line[i] == '"' {
			inQuote = !inQuote
		}
		if !inQuote && i+1 < len(line) && line[i] == '/' && line[i+1] == '/' {
			break
		}
		out.WriteByte(line[i])
	}
	return out.String()
}

func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		trim := strings.TrimSpace(strings.Trim(part, "\""))
		if trim != "" {
			out = append(out, trim)
		}
	}
	return out
}

// --- Expression parser (minimal, precedence-based) ---

type tokenKind int

const (
	tokEOF tokenKind = iota
	tokIdent
	tokNumber
	tokString
	tokSymbol
)

type token struct {
	kind  tokenKind
	value string
}

type lexer struct {
	input []rune
	pos   int
}

func newLexer(s string) *lexer {
	return &lexer{input: []rune(s)}
}

func (l *lexer) nextToken() token {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
	if l.pos >= len(l.input) {
		return token{kind: tokEOF}
	}
	ch := l.input[l.pos]
	if isIdentStart(ch) {
		start := l.pos
		l.pos++
		for l.pos < len(l.input) && isIdentPart(l.input[l.pos]) {
			l.pos++
		}
		return token{kind: tokIdent, value: string(l.input[start:l.pos])}
	}
	if unicode.IsDigit(ch) {
		start := l.pos
		l.pos++
		for l.pos < len(l.input) && unicode.IsDigit(l.input[l.pos]) {
			l.pos++
		}
		return token{kind: tokNumber, value: string(l.input[start:l.pos])}
	}
	if ch == '"' {
		start := l.pos
		l.pos++
		for l.pos < len(l.input) {
			if l.input[l.pos] == '\\' {
				l.pos += 2
				continue
			}
			if l.input[l.pos] == '"' {
				l.pos++
				break
			}
			l.pos++
		}
		return token{kind: tokString, value: string(l.input[start:l.pos])}
	}

	symbols := []string{"==", "!=", "<=", ">=", "(", ")", "{", "}", "+", "-", "*", "/", "<", ">", ","}
	for _, sym := range symbols {
		if strings.HasPrefix(string(l.input[l.pos:]), sym) {
			l.pos += len(sym)
			return token{kind: tokSymbol, value: sym}
		}
	}

	l.pos++
	return token{kind: tokSymbol, value: string(ch)}
}

func isIdentStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isIdentPart(r rune) bool {
	return isIdentStart(r) || unicode.IsDigit(r)
}

type parser struct {
	lex    *lexer
	curr   token
	peeked bool
}

func newParser(s string) *parser {
	p := &parser{lex: newLexer(s)}
	p.curr = p.lex.nextToken()
	return p
}

func (p *parser) next() token {
	p.curr = p.lex.nextToken()
	return p.curr
}

func parseExpr(s string) (*ir.Expr, error) {
	p := newParser(s)
	expr, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *parser) parseOr() (*ir.Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.curr.kind == tokIdent && p.curr.value == "or" {
		op := p.curr.value
		p.next()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: op, Left: *left, Right: *right}}
	}
	return left, nil
}

func (p *parser) parseAnd() (*ir.Expr, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for p.curr.kind == tokIdent && p.curr.value == "and" {
		op := p.curr.value
		p.next()
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = &ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: op, Left: *left, Right: *right}}
	}
	return left, nil
}

func (p *parser) parseEquality() (*ir.Expr, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	for p.curr.kind == tokSymbol && (p.curr.value == "==" || p.curr.value == "!=") {
		op := p.curr.value
		p.next()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: op, Left: *left, Right: *right}}
	}
	return left, nil
}

func (p *parser) parseComparison() (*ir.Expr, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for p.curr.kind == tokSymbol && (p.curr.value == "<" || p.curr.value == ">" || p.curr.value == "<=" || p.curr.value == ">=") {
		op := p.curr.value
		p.next()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = &ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: op, Left: *left, Right: *right}}
	}
	return left, nil
}

func (p *parser) parseTerm() (*ir.Expr, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for p.curr.kind == tokSymbol && (p.curr.value == "+" || p.curr.value == "-") {
		op := p.curr.value
		p.next()
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		left = &ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: op, Left: *left, Right: *right}}
	}
	return left, nil
}

func (p *parser) parseFactor() (*ir.Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.curr.kind == tokSymbol && (p.curr.value == "*" || p.curr.value == "/") {
		op := p.curr.value
		p.next()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: op, Left: *left, Right: *right}}
	}
	return left, nil
}

func (p *parser) parseUnary() (*ir.Expr, error) {
	if p.curr.kind == tokIdent && p.curr.value == "not" {
		op := p.curr.value
		p.next()
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ir.Expr{Kind: "unary", Unary: &ir.UnaryExpr{Op: op, Expr: *expr}}, nil
	}
	if p.curr.kind == tokSymbol && p.curr.value == "-" {
		op := p.curr.value
		p.next()
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ir.Expr{Kind: "unary", Unary: &ir.UnaryExpr{Op: op, Expr: *expr}}, nil
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() (*ir.Expr, error) {
	if p.curr.kind == tokNumber {
		val := p.curr.value
		p.next()
		return &ir.Expr{Kind: "lit", Lit: &ir.LiteralExpr{Kind: "int", Value: val}}, nil
	}
	if p.curr.kind == tokString {
		val := p.curr.value
		p.next()
		s, err := strconv.Unquote(val)
		if err != nil {
			return nil, err
		}
		return &ir.Expr{Kind: "lit", Lit: &ir.LiteralExpr{Kind: "string", Value: s}}, nil
	}
	if p.curr.kind == tokIdent {
		name := p.curr.value
		p.next()
		if p.curr.kind == tokSymbol && p.curr.value == "(" {
			p.next()
			args, err := p.parseArgs()
			if err != nil {
				return nil, err
			}
			return &ir.Expr{Kind: "call", Call: &ir.CallExpr{Name: name, Args: args}}, nil
		}
		if name == "true" || name == "false" {
			return &ir.Expr{Kind: "lit", Lit: &ir.LiteralExpr{Kind: "bool", Value: name}}, nil
		}
		return &ir.Expr{Kind: "var", Var: &ir.VarExpr{Name: name}}, nil
	}
	if p.curr.kind == tokSymbol && p.curr.value == "(" {
		p.next()
		expr, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.curr.kind != tokSymbol || p.curr.value != ")" {
			return nil, fmt.Errorf("expected )")
		}
		p.next()
		return expr, nil
	}
	return nil, fmt.Errorf("unexpected token: %s", p.curr.value)
}

func (p *parser) parseArgs() ([]ir.Expr, error) {
	var args []ir.Expr
	if p.curr.kind == tokSymbol && p.curr.value == ")" {
		p.next()
		return args, nil
	}
	for {
		expr, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		args = append(args, *expr)
		if p.curr.kind == tokSymbol && p.curr.value == "," {
			p.next()
			continue
		}
		if p.curr.kind == tokSymbol && p.curr.value == ")" {
			p.next()
			break
		}
		return nil, fmt.Errorf("expected , or )")
	}
	return args, nil
}
