package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/willams/lia/internal/ir"
)

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

func astPosToIR(pos lexer.Position) *ir.Position {
	return &ir.Position{
		Filename: pos.Filename,
		Offset:   pos.Offset,
		Line:     pos.Line,
		Column:   pos.Column,
	}
}

func moduleToIR(m *Module) ir.Module {
	mod := ir.Module{Name: m.Name.String(), Pos: astPosToIR(m.Position)}
	if m.Role != "" {
		mod.Role = m.Role
	}
	if m.Gen != nil {
		mod.Gen = genToIR(m.Gen)
	}
	for _, item := range m.Items {
		switch {
		case item.Type != nil:
			mod.Types = append(mod.Types, ir.TypeDecl{Name: item.Type.Name, Pos: astPosToIR(item.Type.Position), Base: item.Type.Base.Value, Predicate: rawExprValue(item.Type.Predicate)})
		case item.Enum != nil:
			mod.Enums = append(mod.Enums, ir.EnumDecl{Name: item.Enum.Name, Pos: astPosToIR(item.Enum.Position), Values: item.Enum.Values})
		case item.Record != nil:
			mod.Records = append(mod.Records, recordToIR(item.Record))
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

func recordToIR(r *RecordDecl) ir.RecordDecl {
	rec := ir.RecordDecl{Name: r.Name, Pos: astPosToIR(r.Position)}
	for _, f := range r.Fields.Fields {
		rec.Fields = append(rec.Fields, ir.Field{Name: f.Name, Type: f.Type.Value})
	}
	return rec
}

func portToIR(p *PortDecl) ir.PortDecl {
	port := ir.PortDecl{Name: p.Name, Pos: astPosToIR(p.Position)}
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
	uc := ir.UsecaseDecl{Name: u.Name, Pos: astPosToIR(u.Position)}
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
	ad := ir.AdapterDecl{Name: a.Name, Pos: astPosToIR(a.Position)}
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
