package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Program struct {
	Items []*TopLevel `@@*`
}

type TopLevel struct {
	Project *Project `  @@`
	Pack    *Pack    `| @@`
	Module  *Module  `| @@`
}

type Project struct {
	Token1 string         `"project"`
	Name   *QName         `@@`
	Items  []*ProjectItem `"{" @@* "}"`
}

type ProjectItem struct {
	Repro      *ReproDecl      `  @@`
	Tape       *TapeDecl       `| @@`
	Use        *UseDecl        `| @@`
	Policy     *PolicyDecl     `| @@`
	Constraint *ConstraintDecl `| @@`
	Module     *Module         `| @@`
}

type Pack struct {
	Token1  string      `"pack"`
	Name    string      `@Ident`
	Version *Version    `( "@" @@ )?`
	Items   []*PackItem `"{" @@* "}"`
}

type PackItem struct {
	Module     *Module         `  @@`
	Policy     *PolicyDecl     `| @@`
	Constraint *ConstraintDecl `| @@`
}

type Version struct {
	Value string `(@Ident | @Number | @String)`
}

type ReproDecl struct {
	Token1 string `"repro"`
	Value  string `@Ident`
	Token2 string `";"`
}

type TapeDecl struct {
	Token1 string   `"tape"`
	Path   TapePath `@@`
	Token2 string   `";"`
}

type TapePath struct {
	Value string
}

type UseDecl struct {
	Token1 string   `"use" "pack"`
	Pack   *PackRef `@@`
	Token2 string   `";"`
}

type PackRef struct {
	Name    string   `@Ident`
	Version *Version `( "@" @@ )?`
}

type PolicyDecl struct {
	Token1 string  `"policy"`
	Name   string  `@Ident`
	Token2 string  `":"`
	Expr   RawExpr `@@`
	Token3 string  `";"`
}

type ConstraintDecl struct {
	Token1 string  `"constraint"`
	Name   string  `@Ident`
	Token2 string  `":"`
	Expr   RawExpr `@@`
	Token3 string  `";"`
}

type PreferDecl struct {
	Name      string
	Expr      string
	Weight    float64
	HasWeight bool
}

type HoleDecl struct {
	Token1   string  `"hole"`
	Name     string  `@Ident`
	Token2   string  `":"`
	Contract RawExpr `@@`
	Token3   string  `";"`
}

type GenBlock struct {
	Token1  string   `@"@" "gen" "{"`
	Entries []*GenKV `(@@ ( ("," | ";") @@ )*)?`
	Token2  string   `"}"`
}

type GenKV struct {
	Key    string    `@Ident`
	Token1 string    `":"`
	Value  *GenValue `@@`
}

type GenValue struct {
	Map     *KVList      `  @@`
	List    *LiteralList `| @@`
	Literal *Literal     `| @@`
}

type KVList struct {
	Entries []*KVPair `"{" ( @@ ( ("," | ";") @@ )* ("," | ";")? )? "}"`
}

type KVPair struct {
	Key    string  `@Ident`
	Token1 string  `":"`
	Value  Literal `@@`
}

type Literal struct {
	Value string `(@String | @Number | @Ident)`
}

type Module struct {
	lexer.Position
	Gen    *GenBlock     `@@?`
	Token1 string        `"module"`
	Name   *QName        `@@`
	Role   string        `( "as" @Ident )?`
	Items  []*ModuleItem `"{" @@* "}"`
}

type ModuleItem struct {
	Type       *TypeDecl       `  @@`
	Enum       *EnumDecl       `| @@`
	Record     *RecordDecl     `| @@`
	Port       *PortDecl       `| @@`
	Usecase    *UsecaseDecl    `| @@`
	Adapter    *AdapterDecl    `| @@`
	Wiring     *WiringDecl     `| @@`
	Constraint *ConstraintDecl `| @@`
	Prefer     *PreferDecl     `| @@`
	Hole       *HoleDecl       `| @@`
	Candidate  *CandidateDecl  `| @@`
}

// RecordDecl represents a multi-field structured type in the AST.
type RecordDecl struct {
	lexer.Position
	Token1 string    `"record"`
	Name   string    `@Ident`
	Fields FieldList `"{" @@ "}"`
}

type TypeDecl struct {
	lexer.Position
	Token1    string   `"type"`
	Name      string   `@Ident`
	Token2    string   `"="`
	Base      TypeRef  `@@`
	Predicate *RawExpr `( "where" @@ )?`
	Token3    string   `";"`
}

type EnumDecl struct {
	lexer.Position
	Token1 string   `"enum"`
	Name   string   `@Ident`
	Values []string `"{" @Ident ("," @Ident)* ","? "}"` // allow trailing comma
	Token2 string   `";"?`
}

type PortDecl struct {
	lexer.Position
	Token1  string        `"port"`
	Name    string        `@Ident`
	Methods []*MethodDecl `"{" @@* "}"`
}

type MethodDecl struct {
	Token1  string     `@("fn" | "method")`
	Name    string     `@Ident`
	Params  FieldList  `"(" @@ ")"`
	Returns *FieldList `( "->" "(" @@ ")" )?`
	Token2  string     `";"`
}

type FieldList struct {
	Fields []*Field `( @@ ( ("," | ";") @@ )* ("," | ";")? )?`
}

type Field struct {
	Name   string  `@Ident`
	Token1 string  `":"`
	Type   TypeRef `@@`
}

type UsecaseDecl struct {
	lexer.Position
	Token1 string         `"usecase"`
	Name   string         `@Ident`
	Items  []*UsecaseItem `"{" @@* "}"`
}

type AdapterDecl struct {
	lexer.Position
	Token1     string         `"adapter"`
	Name       string         `@Ident`
	Implements *QName         `( "implements" @@ )?`
	Items      []*UsecaseItem `"{" @@* "}"`
}

type UsecaseItem struct {
	Input   *InputBlock  `  @@`
	Output  *OutputBlock `| @@`
	Effects *EffectsDecl `| @@`
	Stmt    *Stmt        `| @@`
}

type InputBlock struct {
	Token1 string    `"input"`
	Fields FieldList `"{" @@ "}"`
	Token2 string    `";"?`
}

type OutputBlock struct {
	Token1 string    `"output"`
	Fields FieldList `"{" @@ "}"`
	Token2 string    `";"?`
}

type EffectsDecl struct {
	Token1 string     `"effects"`
	List   EffectList `@@`
	Token2 string     `";"?`
}

type WiringDecl struct {
	Token1 string      `"wiring"`
	Name   string      `@Ident`
	Binds  []*BindDecl `"{" @@* "}"`
}

type BindDecl struct {
	Token1 string `"bind"`
	Left   *QName `@@`
	Token2 string `"->"`
	Right  *QName `@@`
	Token3 string `";"`
}

type CandidateDecl struct {
	Token1 string          `"candidate"`
	Symbol *QName          `@@`
	Score  *ScoreDecl      `( @@ )?`
	Block  *CandidateBlock `( @@ )?`
	Token2 string          `";"?`
}

type ScoreDecl struct {
	Token1 string `"score"`
	Value  string `(@Number | @Ident)`
}

type CandidateBlock struct {
	Items []*CandidateItem `"{" @@* "}"`
}

type CandidateItem struct {
	Score      *ScoreStmt      `  @@`
	Constraint *ConstraintDecl `| @@`
}

type ScoreStmt struct {
	Token1 string `"score"`
	Value  string `(@Number | @Ident)`
	Token2 string `";"`
}

type Stmt struct {
	Let      *LetStmt      `  @@`
	Return   *ReturnStmt   `| @@`
	If       *IfStmt       `| @@`
	While    *WhileStmt    `| @@`
	For      *ForStmt      `| @@`
	Break    *BreakStmt    `| @@`
	Continue *ContinueStmt `| @@`
	Assign   *AssignStmt   `| @@`
	Expr     *ExprStmt     `| @@`
}

type LetStmt struct {
	Token1 string `"let"`
	Name   string `@Ident`
	Token2 string `"="`
	Value  *Expr  `@@`
	Token3 string `";"`
}

type AssignStmt struct {
	Name   string `@Ident`
	Token1 string `"="`
	Value  *Expr  `@@`
	Token2 string `";"`
}

type ReturnStmt struct {
	Token1 string `"return"`
	Value  *Expr  `@@?`
	Token2 string `";"`
}

type BreakStmt struct {
	Keyword string `"break"`
	Token1  string `";"`
}

type ContinueStmt struct {
	Keyword string `"continue"`
	Token1  string `";"`
}

type IfStmt struct {
	Token1 string     `"if"`
	Cond   *Expr      `@@`
	Then   *StmtBlock `@@`
	Else   *StmtBlock `( "else" @@ )?`
}

type WhileStmt struct {
	Token1 string     `"while"`
	Cond   *Expr      `@@`
	Body   *StmtBlock `@@`
}

type ForStmt struct {
	Token1 string     `"for"`
	Var    string     `@Ident`
	Token2 string     `"in"`
	Iter   *Expr      `@@`
	Body   *StmtBlock `@@`
}

type StmtBlock struct {
	Stmts []*Stmt `"{" @@* "}"`
}

type ExprStmt struct {
	Expr   *Expr  `@@`
	Token1 string `";"`
}

type Expr struct {
	Or *OrExpr `@@`
}

type OrExpr struct {
	Left   *AndExpr   `@@`
	Rights []*AndExpr `("||" @@)*`
}

type AndExpr struct {
	Left   *EqualityExpr   `@@`
	Rights []*EqualityExpr `("&&" @@)*`
}

type EqualityExpr struct {
	Left *ComparisonExpr `@@`
	Ops  []*EqualityOp   `@@*`
}

type EqualityOp struct {
	Op    string          `@("==" | "!=")`
	Right *ComparisonExpr `@@`
}

type ComparisonExpr struct {
	Left *TermExpr    `@@`
	Ops  []*CompareOp `@@*`
}

type CompareOp struct {
	Op    string    `@("<=" | ">=" | "<" | ">")`
	Right *TermExpr `@@`
}

type TermExpr struct {
	Left *FactorExpr `@@`
	Ops  []*TermOp   `@@*`
}

type TermOp struct {
	Op    string      `@("+" | "-")`
	Right *FactorExpr `@@`
}

type FactorExpr struct {
	Left *UnaryExpr  `@@`
	Ops  []*FactorOp `@@*`
}

type FactorOp struct {
	Op    string     `@("*" | "/" | "%")`
	Right *UnaryExpr `@@`
}

type UnaryExpr struct {
	Unary   *UnaryOp     `  @@`
	Postfix *PostfixExpr `| @@`
}

type UnaryOp struct {
	Op   string     `@("!" | "-")`
	Expr *UnaryExpr `@@`
}

type PostfixExpr struct {
	Primary  *PrimaryExpr `@@`
	Suffixes []*Suffix    `@@*`
}

type Suffix struct {
	Call   *CallSuffix   `  @@`
	Member *MemberSuffix `| @@`
	Index  *IndexSuffix  `| @@`
}

type CallSuffix struct {
	Args *ExprList `"(" @@? ")"`
}

type MemberSuffix struct {
	Token1 string `"."`
	Field  string `@Ident`
}

type IndexSuffix struct {
	Token1 string `"["`
	Index  *Expr  `@@`
	Token2 string `"]"`
}

type ExprList struct {
	Exprs []*Expr `@@ ("," @@)*`
}

type PrimaryExpr struct {
	Number *string   `  @Number`
	String *string   `| @String`
	Ident  *string   `| @Ident`
	List   *ExprList `| "[" @@? "]"`
	Group  *Expr     `| "(" @@ ")"`
}

type QName struct {
	Head string       `@(Ident | Keyword)`
	Tail []*QNameTail `@@*`
}

type QNameTail struct {
	Sep  string `@("::" | "." | ":")`
	Part string `@(Ident | Keyword)`
}

type TypeRef struct {
	Value string
}

type RawExpr struct {
	Value string
}

type EffectList struct {
	Values []string
}

type LiteralList struct {
	Values []string
}

func (q *QName) String() string {
	if q == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(q.Head)
	for _, t := range q.Tail {
		b.WriteString(t.Sep)
		b.WriteString(t.Part)
	}
	return b.String()
}

func (v Version) String() string {
	return literalValue(v.Value)
}

func (l Literal) String() string {
	return literalValue(l.Value)
}

func (t TapePath) String() string {
	return t.Value
}

func (p *PreferDecl) Parse(lex *lexer.PeekingLexer) error {
	tok := lex.Peek()
	if tok.Value != "prefer" {
		return participle.NextMatch
	}
	lex.Next()
	nameTok := lex.Next()
	if !isTokenType(*nameTok, "Ident") {
		return fmt.Errorf("expected prefer name")
	}
	colon := lex.Next()
	if colon.Value != ":" {
		return fmt.Errorf("expected ':' after prefer name")
	}
	var exprTokens []string
	for {
		peek := lex.Peek()
		if peek.EOF() {
			return fmt.Errorf("unexpected EOF in prefer")
		}
		if peek.Value == ";" {
			lex.Next()
			break
		}
		if peek.Value == "weight" {
			lex.Next()
			valTok := lex.Next()
			if !isTokenType(*valTok, "Number") && !isTokenType(*valTok, "Ident") {
				return fmt.Errorf("expected weight number")
			}
			weight, err := strconv.ParseFloat(valTok.Value, 64)
			if err != nil {
				return fmt.Errorf("invalid weight")
			}
			p.Weight = weight
			p.HasWeight = true
			continue
		}
		tok := lex.Next()
		exprTokens = append(exprTokens, tok.Value)
	}
	p.Name = nameTok.Value
	p.Expr = strings.TrimSpace(strings.Join(exprTokens, " "))
	return nil
}

func (b *InputBlock) Parse(lex *lexer.PeekingLexer) error {
	tok := lex.Peek()
	if tok.Value != "input" {
		return participle.NextMatch
	}
	lex.Next()
	if lex.Peek().Value != "{" {
		return fmt.Errorf("expected '{' after input")
	}
	lex.Next()
	fields, err := parseFieldList(lex, "}")
	if err != nil {
		return err
	}
	if lex.Peek().Value != "}" {
		return fmt.Errorf("expected '}' to close input block")
	}
	lex.Next()
	if lex.Peek().Value == ";" {
		lex.Next()
	}
	b.Fields.Fields = fields
	return nil
}

func (b *OutputBlock) Parse(lex *lexer.PeekingLexer) error {
	tok := lex.Peek()
	if tok.Value != "output" {
		return participle.NextMatch
	}
	lex.Next()
	if lex.Peek().Value != "{" {
		return fmt.Errorf("expected '{' after output")
	}
	lex.Next()
	fields, err := parseFieldList(lex, "}")
	if err != nil {
		return err
	}
	if lex.Peek().Value != "}" {
		return fmt.Errorf("expected '}' to close output block")
	}
	lex.Next()
	if lex.Peek().Value == ";" {
		lex.Next()
	}
	b.Fields.Fields = fields
	return nil
}

func (e *EffectsDecl) Parse(lex *lexer.PeekingLexer) error {
	tok := lex.Peek()
	if tok.Value != "effects" {
		return participle.NextMatch
	}
	lex.Next()
	if err := e.List.Parse(lex); err != nil {
		if err == participle.NextMatch {
			return fmt.Errorf("expected effects list")
		}
		return err
	}
	if lex.Peek().Value == ";" {
		lex.Next()
	}
	return nil
}

func (t *TapePath) Parse(lex *lexer.PeekingLexer) error {
	peek := lex.Peek()
	if isTokenType(*peek, "String") {
		tok := lex.Next()
		t.Value = literalValue(tok.Value)
		return nil
	}
	if isTokenType(*peek, "Ident") {
		first := lex.Next()
		next := lex.Peek()
		if isTokenType(*next, "String") {
			tok := lex.Next()
			t.Value = literalValue(tok.Value)
			return nil
		}
		t.Value = first.Value
		return nil
	}
	return fmt.Errorf("expected tape path")
}

func (t *TypeRef) Parse(lex *lexer.PeekingLexer) error {
	var parts []string
	depthAngle := 0
	depthBracket := 0
	terminators := LIATypeTerminators()
	for {
		peek := lex.Peek()
		if peek.EOF() {
			break
		}
		if isTokenType(*peek, "Ident") || isTokenType(*peek, "Keyword") {
			if terminators[peek.Value] {
				if depthAngle == 0 && depthBracket == 0 {
					goto done
				}
			}
		}
		if depthAngle == 0 && depthBracket == 0 {
			if peek.Value == "," || peek.Value == ";" || peek.Value == "}" || peek.Value == ")" {
				break
			}
		}
		if peek.Value == "<" {
			depthAngle++
		} else if peek.Value == ">" && depthAngle > 0 {
			depthAngle--
		} else if peek.Value == "[" {
			depthBracket++
		} else if peek.Value == "]" && depthBracket > 0 {
			depthBracket--
		}
		tok := lex.Next()
		parts = append(parts, tok.Value)
	}

done:
	if len(parts) == 0 {
		return fmt.Errorf("expected type reference")
	}
	t.Value = strings.Join(parts, "")
	return nil
}

func (l *EffectList) Parse(lex *lexer.PeekingLexer) error {
	peek := lex.Peek()
	if peek.Value != "[" {
		return participle.NextMatch
	}
	lex.Next()
	for {
		peek := lex.Peek()
		if peek.EOF() {
			return fmt.Errorf("unexpected EOF in effects")
		}
		if peek.Value == "]" {
			lex.Next()
			return nil
		}
		if peek.Value == "," {
			lex.Next()
			continue
		}
		if !isTokenType(*peek, "Ident") && !isTokenType(*peek, "String") {
			return fmt.Errorf("expected effect name")
		}
		tok := lex.Next()
		val := tok.Value
		if isTokenType(*tok, "String") {
			val = literalValue(tok.Value)
		}
		l.Values = append(l.Values, val)
	}
}

func (l *LiteralList) Parse(lex *lexer.PeekingLexer) error {
	peek := lex.Peek()
	if peek.Value != "[" {
		return participle.NextMatch
	}
	lex.Next()
	for {
		peek := lex.Peek()
		if peek.EOF() {
			return fmt.Errorf("unexpected EOF in list")
		}
		if peek.Value == "]" {
			lex.Next()
			return nil
		}
		if peek.Value == "," {
			lex.Next()
			continue
		}
		if !isTokenType(*peek, "Ident") && !isTokenType(*peek, "String") && !isTokenType(*peek, "Number") {
			return fmt.Errorf("expected literal")
		}
		tok := lex.Next()
		l.Values = append(l.Values, literalValue(tok.Value))
	}
}

func literalValue(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "\"") {
		unq, err := strconv.Unquote(raw)
		if err == nil {
			return unq
		}
	}
	return raw
}

func parseFieldList(lex *lexer.PeekingLexer, endToken string) ([]*Field, error) {
	var fields []*Field
	for {
		peek := lex.Peek()
		if peek.EOF() {
			return nil, fmt.Errorf("unexpected EOF in field list")
		}
		if peek.Value == endToken {
			return fields, nil
		}
		if peek.Value == "," || peek.Value == ";" {
			lex.Next()
			continue
		}
		if !isTokenType(*peek, "Ident") {
			return nil, fmt.Errorf("expected field name")
		}
		nameTok := lex.Next()
		if lex.Peek().Value != ":" {
			return nil, fmt.Errorf("expected ':' after field name")
		}
		lex.Next()
		var typ TypeRef
		if err := typ.Parse(lex); err != nil {
			return nil, err
		}
		fields = append(fields, &Field{Name: nameTok.Value, Type: typ})
	}
}
