package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/willams/lia/internal/ir"
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
	Gen    *GenBlock     `@@?`
	Token1 string        `"module"`
	Name   *QName        `@@`
	Role   string        `( "as" @Ident )?`
	Items  []*ModuleItem `"{" @@* "}"`
}

type ModuleItem struct {
	Type       *TypeDecl       `  @@`
	Enum       *EnumDecl       `| @@`
	Port       *PortDecl       `| @@`
	Usecase    *UsecaseDecl    `| @@`
	Adapter    *AdapterDecl    `| @@`
	Wiring     *WiringDecl     `| @@`
	Constraint *ConstraintDecl `| @@`
	Prefer     *PreferDecl     `| @@`
	Hole       *HoleDecl       `| @@`
	Candidate  *CandidateDecl  `| @@`
}

type TypeDecl struct {
	Token1    string   `"type"`
	Name      string   `@Ident`
	Token2    string   `"="`
	Base      TypeRef  `@@`
	Predicate *RawExpr `( "where" @@ )?`
	Token3    string   `";"`
}

type EnumDecl struct {
	Token1 string   `"enum"`
	Name   string   `@Ident`
	Values []string `"{" @Ident ("," @Ident)* ","? "}"` // allow trailing comma
	Token2 string   `";"?`
}

type PortDecl struct {
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
	Token1 string         `"usecase"`
	Name   string         `@Ident`
	Items  []*UsecaseItem `"{" @@* "}"`
}

type AdapterDecl struct {
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
	for {
		peek := lex.Peek()
		if peek.EOF() {
			break
		}
		if isTokenType(*peek, "Ident") {
			switch peek.Value {
			case "as", "usecase", "adapter", "port", "wiring", "use", "project", "module", "pack", "repro", "tape", "policy", "constraint", "prefer", "hole", "candidate", "input", "output", "effects", "implements", "type", "enum", "where":
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

func (p *Program) ToIR() (*ir.Program, error) {
	prog := &ir.Program{Version: "0.1"}
	for _, item := range p.Items {
		switch {
		case item.Project != nil:
			proj, mods := projectToIR(item.Project)
			prog.Projects = append(prog.Projects, proj)
			prog.Modules = append(prog.Modules, mods...)
		case item.Pack != nil:
			prog.Packs = append(prog.Packs, packToIR(item.Pack))
		case item.Module != nil:
			prog.Modules = append(prog.Modules, moduleToIR(item.Module))
		}
	}
	return prog, nil
}

func projectToIR(p *Project) (ir.Project, []ir.Module) {
	proj := ir.Project{Name: p.Name.String()}
	var modules []ir.Module
	for _, item := range p.Items {
		if item.Repro != nil {
			proj.Repro = ir.ReproProfile(item.Repro.Value)
		}
		if item.Tape != nil {
			proj.Tape = item.Tape.Path.String()
		}
		if item.Use != nil {
			proj.Uses = append(proj.Uses, packRefToIR(item.Use.Pack))
		}
		if item.Policy != nil {
			proj.Policies = append(proj.Policies, ir.PolicyDecl{Name: item.Policy.Name, Body: item.Policy.Expr.Value})
		}
		if item.Constraint != nil {
			proj.Constraints = append(proj.Constraints, ir.ConstraintDecl{Name: item.Constraint.Name, Expr: item.Constraint.Expr.Value})
		}
		if item.Module != nil {
			modules = append(modules, moduleToIR(item.Module))
		}
	}
	return proj, modules
}

func packToIR(p *Pack) ir.Pack {
	pack := ir.Pack{Name: p.Name}
	if p.Version != nil {
		pack.Version = p.Version.String()
	}
	for _, item := range p.Items {
		if item.Module != nil {
			pack.Modules = append(pack.Modules, moduleToIR(item.Module))
		}
		if item.Policy != nil {
			pack.Policies = append(pack.Policies, ir.PolicyDecl{Name: item.Policy.Name, Body: item.Policy.Expr.Value})
		}
		if item.Constraint != nil {
			pack.Constraints = append(pack.Constraints, ir.ConstraintDecl{Name: item.Constraint.Name, Expr: item.Constraint.Expr.Value})
		}
	}
	return pack
}

func moduleToIR(m *Module) ir.Module {
	mod := ir.Module{Name: m.Name.String()}
	if m.Role != "" {
		mod.Role = m.Role
	}
	if m.Gen != nil {
		mod.Gen = genToIR(m.Gen)
	}
	for _, item := range m.Items {
		switch {
		case item.Type != nil:
			mod.Types = append(mod.Types, ir.TypeDecl{Name: item.Type.Name, Base: item.Type.Base.Value, Predicate: rawExprValue(item.Type.Predicate)})
		case item.Enum != nil:
			mod.Enums = append(mod.Enums, ir.EnumDecl{Name: item.Enum.Name, Values: item.Enum.Values})
		case item.Port != nil:
			mod.Ports = append(mod.Ports, portToIR(item.Port))
		case item.Usecase != nil:
			mod.Usecases = append(mod.Usecases, usecaseToIR(item.Usecase))
		case item.Adapter != nil:
			mod.Adapters = append(mod.Adapters, adapterToIR(item.Adapter))
		case item.Wiring != nil:
			mod.Wirings = append(mod.Wirings, wiringToIR(item.Wiring))
		case item.Constraint != nil:
			mod.Constraints = append(mod.Constraints, ir.ConstraintDecl{Name: item.Constraint.Name, Expr: item.Constraint.Expr.Value})
		case item.Prefer != nil:
			pref := ir.PreferDecl{Name: item.Prefer.Name, Expr: item.Prefer.Expr}
			if item.Prefer.HasWeight {
				pref.Weight = item.Prefer.Weight
			}
			mod.Preferences = append(mod.Preferences, pref)
		case item.Hole != nil:
			mod.Holes = append(mod.Holes, ir.HoleDecl{Name: item.Hole.Name, Contract: item.Hole.Contract.Value})
		case item.Candidate != nil:
			mod.Candidates = append(mod.Candidates, candidateToIR(item.Candidate))
		}
	}
	return mod
}

func genToIR(g *GenBlock) *ir.GenMeta {
	meta := &ir.GenMeta{}
	for _, entry := range g.Entries {
		switch entry.Key {
		case "prompt_ref":
			if entry.Value != nil {
				if val, ok := entry.Value.literal(); ok {
					meta.PromptRef = val
				}
			}
		case "prompt_hash":
			if entry.Value != nil {
				if val, ok := entry.Value.literal(); ok {
					meta.PromptHash = val
				}
			}
		case "model_id":
			if entry.Value != nil {
				if val, ok := entry.Value.literal(); ok {
					meta.ModelID = val
				}
			}
		case "generator_pass":
			if entry.Value != nil {
				if val, ok := entry.Value.literal(); ok {
					meta.GeneratorPass = val
				}
			}
		case "timestamp":
			if entry.Value != nil {
				if val, ok := entry.Value.literal(); ok {
					meta.Timestamp = val
				}
			}
		case "context_refs":
			if entry.Value != nil {
				if vals, ok := entry.Value.list(); ok {
					meta.ContextRefs = append(meta.ContextRefs, vals...)
				}
			}
		case "tools_trace_refs":
			if entry.Value != nil {
				if vals, ok := entry.Value.list(); ok {
					meta.ToolsTraceRefs = append(meta.ToolsTraceRefs, vals...)
				}
			}
		case "model_params":
			if entry.Value != nil {
				if kvs, ok := entry.Value.kvList(); ok {
					meta.ModelParams = append(meta.ModelParams, kvs...)
					continue
				}
				if val, ok := entry.Value.literal(); ok {
					meta.ModelParams = append(meta.ModelParams, ir.KV{Key: "raw", Value: val})
				}
			}
		}
	}
	return meta
}

func (v *GenValue) literal() (string, bool) {
	if v == nil || v.Literal == nil {
		return "", false
	}
	return v.Literal.String(), true
}

func (v *GenValue) list() ([]string, bool) {
	if v == nil || v.List == nil {
		return nil, false
	}
	return v.List.Values, true
}

func (v *GenValue) kvList() ([]ir.KV, bool) {
	if v == nil || v.Map == nil {
		return nil, false
	}
	kvs := make([]ir.KV, 0, len(v.Map.Entries))
	for _, entry := range v.Map.Entries {
		kvs = append(kvs, ir.KV{Key: entry.Key, Value: entry.Value.String()})
	}
	return kvs, true
}

func packRefToIR(ref *PackRef) ir.PackRef {
	if ref == nil {
		return ir.PackRef{}
	}
	out := ir.PackRef{Name: ref.Name}
	if ref.Version != nil {
		out.Version = ref.Version.String()
	}
	return out
}

func portToIR(p *PortDecl) ir.PortDecl {
	port := ir.PortDecl{Name: p.Name}
	for _, m := range p.Methods {
		port.Methods = append(port.Methods, methodToIR(m))
	}
	return port
}

func methodToIR(m *MethodDecl) ir.FuncDecl {
	decl := ir.FuncDecl{Name: m.Name}
	for _, f := range m.Params.Fields {
		decl.Params = append(decl.Params, fieldToIR(f))
	}
	if m.Returns != nil {
		for _, f := range m.Returns.Fields {
			decl.Returns = append(decl.Returns, fieldToIR(f))
		}
	}
	return decl
}

func fieldToIR(f *Field) ir.Field {
	return ir.Field{Name: f.Name, Type: f.Type.Value}
}

func usecaseToIR(u *UsecaseDecl) ir.UsecaseDecl {
	uc := ir.UsecaseDecl{Name: u.Name}
	for _, item := range u.Items {
		switch {
		case item.Input != nil:
			for _, f := range item.Input.Fields.Fields {
				uc.Inputs = append(uc.Inputs, fieldToIR(f))
			}
		case item.Output != nil:
			for _, f := range item.Output.Fields.Fields {
				uc.Outputs = append(uc.Outputs, fieldToIR(f))
			}
		case item.Effects != nil:
			uc.Effects = append(uc.Effects, item.Effects.List.Values...)
		case item.Stmt != nil:
			stmt := stmtToIR(item.Stmt)
			uc.Body = append(uc.Body, stmt)
		}
	}
	return uc
}

func adapterToIR(a *AdapterDecl) ir.AdapterDecl {
	ad := ir.AdapterDecl{Name: a.Name}
	if a.Implements != nil {
		ad.Implements = a.Implements.String()
	}
	for _, item := range a.Items {
		switch {
		case item.Input != nil:
			for _, f := range item.Input.Fields.Fields {
				ad.Inputs = append(ad.Inputs, fieldToIR(f))
			}
		case item.Output != nil:
			for _, f := range item.Output.Fields.Fields {
				ad.Outputs = append(ad.Outputs, fieldToIR(f))
			}
		case item.Effects != nil:
			ad.Effects = append(ad.Effects, item.Effects.List.Values...)
		case item.Stmt != nil:
			ad.Body = append(ad.Body, stmtToIR(item.Stmt))
		}
	}
	return ad
}

func wiringToIR(w *WiringDecl) ir.WiringDecl {
	decl := ir.WiringDecl{Name: w.Name}
	for _, b := range w.Binds {
		decl.Binds = append(decl.Binds, fmt.Sprintf("%s -> %s", b.Left.String(), b.Right.String()))
	}
	return decl
}

func candidateToIR(c *CandidateDecl) ir.CandidateDecl {
	decl := ir.CandidateDecl{Symbol: c.Symbol.String()}
	if c.Score != nil {
		if val, err := strconv.ParseFloat(c.Score.Value, 64); err == nil {
			decl.Score = val
		}
	}
	if c.Block != nil {
		for _, item := range c.Block.Items {
			if item.Score != nil {
				if val, err := strconv.ParseFloat(item.Score.Value, 64); err == nil {
					decl.Score = val
				}
			}
			if item.Constraint != nil {
				decl.Constraints = append(decl.Constraints, ir.ConstraintDecl{
					Name: item.Constraint.Name,
					Expr: item.Constraint.Expr.Value,
				})
			}
		}
	}
	return decl
}

func stmtToIR(s *Stmt) ir.Stmt {
	switch {
	case s.Let != nil:
		return ir.Stmt{Kind: "let", Let: &ir.LetStmt{Name: s.Let.Name, Value: exprToIR(s.Let.Value)}}
	case s.Return != nil:
		var val *ir.Expr
		if s.Return.Value != nil {
			v := exprToIR(s.Return.Value)
			val = &v
		}
		return ir.Stmt{Kind: "return", Return: &ir.ReturnStmt{Value: val}}
	case s.If != nil:
		thenStmts := stmtBlockToIR(s.If.Then)
		var elseStmts []ir.Stmt
		if s.If.Else != nil {
			elseStmts = stmtBlockToIR(s.If.Else)
		}
		return ir.Stmt{Kind: "if", If: &ir.IfStmt{Cond: exprToIR(s.If.Cond), Then: thenStmts, Else: elseStmts}}
	case s.While != nil:
		return ir.Stmt{Kind: "while", While: &ir.WhileStmt{Cond: exprToIR(s.While.Cond), Body: stmtBlockToIR(s.While.Body)}}
	case s.For != nil:
		return ir.Stmt{Kind: "for", For: &ir.ForStmt{Var: s.For.Var, Iter: exprToIR(s.For.Iter), Body: stmtBlockToIR(s.For.Body)}}
	case s.Break != nil:
		return ir.Stmt{Kind: "break"}
	case s.Continue != nil:
		return ir.Stmt{Kind: "continue"}
	case s.Assign != nil:
		return ir.Stmt{Kind: "assign", Assign: &ir.AssignStmt{Name: s.Assign.Name, Value: exprToIR(s.Assign.Value)}}
	case s.Expr != nil:
		v := exprToIR(s.Expr.Expr)
		return ir.Stmt{Kind: "expr", ExprStmt: &v}
	default:
		return ir.Stmt{}
	}
}

func stmtBlockToIR(b *StmtBlock) []ir.Stmt {
	if b == nil {
		return nil
	}
	out := make([]ir.Stmt, 0, len(b.Stmts))
	for _, st := range b.Stmts {
		out = append(out, stmtToIR(st))
	}
	return out
}

func exprToIR(e *Expr) ir.Expr {
	if e == nil || e.Or == nil {
		return ir.Expr{}
	}
	return orToIR(e.Or)
}

func orToIR(o *OrExpr) ir.Expr {
	left := andToIR(o.Left)
	for _, r := range o.Rights {
		right := andToIR(r)
		left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: "||", Left: left, Right: right}}
	}
	return left
}

func andToIR(a *AndExpr) ir.Expr {
	left := equalityToIR(a.Left)
	for _, r := range a.Rights {
		right := equalityToIR(r)
		left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: "&&", Left: left, Right: right}}
	}
	return left
}

func equalityToIR(e *EqualityExpr) ir.Expr {
	left := comparisonToIR(e.Left)
	for _, op := range e.Ops {
		right := comparisonToIR(op.Right)
		left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: op.Op, Left: left, Right: right}}
	}
	return left
}

func comparisonToIR(c *ComparisonExpr) ir.Expr {
	left := termToIR(c.Left)
	for _, op := range c.Ops {
		right := termToIR(op.Right)
		left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: op.Op, Left: left, Right: right}}
	}
	return left
}

func termToIR(t *TermExpr) ir.Expr {
	left := factorToIR(t.Left)
	for _, op := range t.Ops {
		right := factorToIR(op.Right)
		left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: op.Op, Left: left, Right: right}}
	}
	return left
}

func factorToIR(f *FactorExpr) ir.Expr {
	left := unaryToIR(f.Left)
	for _, op := range f.Ops {
		right := unaryToIR(op.Right)
		left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: op.Op, Left: left, Right: right}}
	}
	return left
}

func unaryToIR(u *UnaryExpr) ir.Expr {
	if u.Unary != nil {
		return ir.Expr{Kind: "unary", Unary: &ir.UnaryExpr{Op: u.Unary.Op, Expr: unaryToIR(u.Unary.Expr)}}
	}
	return postfixToIR(u.Postfix)
}

func postfixToIR(p *PostfixExpr) ir.Expr {
	expr := primaryToIR(p.Primary)
	for _, sfx := range p.Suffixes {
		switch {
		case sfx.Call != nil:
			var args []ir.Expr
			if sfx.Call.Args != nil {
				for _, arg := range sfx.Call.Args.Exprs {
					args = append(args, exprToIR(arg))
				}
			}
			expr = ir.Expr{Kind: "call", Call: &ir.CallExpr{Callee: expr, Args: args}}
		case sfx.Member != nil:
			expr = ir.Expr{Kind: "member", Member: &ir.MemberExpr{Object: expr, Field: sfx.Member.Field}}
		case sfx.Index != nil:
			expr = ir.Expr{Kind: "index", Index: &ir.IndexExpr{Object: expr, Index: exprToIR(sfx.Index.Index)}}
		}
	}
	return expr
}

func primaryToIR(p *PrimaryExpr) ir.Expr {
	switch {
	case p.Number != nil:
		return ir.Expr{Kind: "literal", Literal: &ir.Literal{Kind: "number", Value: *p.Number}}
	case p.String != nil:
		return ir.Expr{Kind: "literal", Literal: &ir.Literal{Kind: "string", Value: literalValue(*p.String)}}
	case p.Ident != nil:
		if *p.Ident == "true" || *p.Ident == "false" {
			return ir.Expr{Kind: "literal", Literal: &ir.Literal{Kind: "bool", Value: *p.Ident}}
		}
		return ir.Expr{Kind: "ident", Ident: *p.Ident}
	case p.List != nil:
		var elems []ir.Expr
		for _, e := range p.List.Exprs {
			elems = append(elems, exprToIR(e))
		}
		return ir.Expr{Kind: "list", List: &ir.ListExpr{Elements: elems}}
	case p.Group != nil:
		return exprToIR(p.Group)
	default:
		return ir.Expr{}
	}
}

func (r *RawExpr) Parse(lex *lexer.PeekingLexer) error {
	var parts []string
	for {
		peek := lex.Peek()
		if peek.EOF() {
			return fmt.Errorf("unexpected EOF in expression")
		}
		if peek.Value == ";" {
			break
		}
		tok := lex.Next()
		parts = append(parts, tok.Value)
	}
	r.Value = strings.TrimSpace(strings.Join(parts, " "))
	return nil
}

func rawExprValue(expr *RawExpr) string {
	if expr == nil {
		return ""
	}
	return expr.Value
}
