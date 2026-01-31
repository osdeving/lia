package python

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/willams/lia/internal/ir"
)

// Lower emits Python code for a minimal executable subset.
func Lower(p *ir.Program) ([]byte, error) {
	w := &pyWriter{}
	usecases := collectUsecases(p)
	if len(usecases) == 0 {
		w.line("# LIA lower (python) stub")
		w.line("# modules: " + fmt.Sprintf("%d", len(p.Modules)))
		return w.bytes(), nil
	}

	helpers := collectHelpers(usecases)
	if helpers["rand_int"] {
		w.line("import random")
		w.line("")
	}
	if helpers["read_line"] {
		w.line("def read_line(prompt):")
		w.indentIn()
		w.line("return input(prompt)")
		w.indentOut()
		w.line("")
	}
	if helpers["is_int"] {
		w.line("def is_int(value):")
		w.indentIn()
		w.line("return value.isdigit()")
		w.indentOut()
		w.line("")
	}
	if helpers["to_int"] {
		w.line("def to_int(value):")
		w.indentIn()
		w.line("return int(value)")
		w.indentOut()
		w.line("")
	}
	if helpers["rand_int"] {
		w.line("def rand_int(a, b):")
		w.indentIn()
		w.line("return random.randint(a, b)")
		w.indentOut()
		w.line("")
	}

	for _, uc := range usecases {
		params := []string{}
		for _, p := range uc.Inputs {
			params = append(params, p.Name)
		}
		w.line("def " + uc.Name + "(" + strings.Join(params, ", ") + "):")
		w.indentIn()
		emitBlock(w, uc.Body)
		w.indentOut()
		w.line("")
	}

	entry := selectEntry(usecases)
	if entry != "" {
		w.line("if __name__ == \"__main__\":")
		w.indentIn()
		w.line(entry + "()")
		w.indentOut()
	}

	return w.bytes(), nil
}

type pyWriter struct {
	buf    bytes.Buffer
	indent int
}

func (w *pyWriter) line(s string) {
	if s != "" {
		w.buf.WriteString(strings.Repeat(" ", w.indent))
		w.buf.WriteString(s)
	}
	w.buf.WriteString("\n")
}

func (w *pyWriter) indentIn() { w.indent += 4 }

func (w *pyWriter) indentOut() {
	w.indent -= 4
	if w.indent < 0 {
		w.indent = 0
	}
}

func (w *pyWriter) bytes() []byte { return w.buf.Bytes() }

func emitBlock(w *pyWriter, stmts []ir.Stmt) {
	if len(stmts) == 0 {
		w.line("pass")
		return
	}
	for _, s := range stmts {
		emitStmt(w, s)
	}
}

func emitStmt(w *pyWriter, s ir.Stmt) {
	switch s.Kind {
	case "let":
		w.line(s.Let.Name + " = " + emitExpr(s.Let.Value))
	case "assign":
		w.line(s.Assign.Name + " = " + emitExpr(s.Assign.Value))
	case "expr":
		w.line(emitExpr(*s.Expr))
	case "if":
		w.line("if " + emitExpr(s.If.Cond) + ":")
		w.indentIn()
		emitBlock(w, s.If.Then)
		w.indentOut()
		if len(s.If.Else) > 0 {
			w.line("else:")
			w.indentIn()
			emitBlock(w, s.If.Else)
			w.indentOut()
		}
	case "loop":
		w.line("while True:")
		w.indentIn()
		emitBlock(w, s.Loop.Body)
		w.indentOut()
	case "break":
		w.line("break")
	case "continue":
		w.line("continue")
	case "return":
		if s.Return.Value == nil {
			w.line("return")
			return
		}
		w.line("return " + emitExpr(*s.Return.Value))
	}
}

func emitExpr(e ir.Expr) string {
	switch e.Kind {
	case "lit":
		switch e.Lit.Kind {
		case "string":
			return strconv.Quote(e.Lit.Value)
		case "bool":
			if e.Lit.Value == "true" {
				return "True"
			}
			return "False"
		default:
			return e.Lit.Value
		}
	case "var":
		return e.Var.Name
	case "call":
		args := []string{}
		for _, a := range e.Call.Args {
			args = append(args, emitExpr(a))
		}
		name := mapCallName(e.Call.Name)
		return name + "(" + strings.Join(args, ", ") + ")"
	case "binary":
		return emitExpr(e.Binary.Left) + " " + mapOp(e.Binary.Op) + " " + emitExpr(e.Binary.Right)
	case "unary":
		if e.Unary.Op == "not" {
			return "not " + emitExpr(e.Unary.Expr)
		}
		return e.Unary.Op + emitExpr(e.Unary.Expr)
	default:
		return ""
	}
}

func mapCallName(name string) string {
	switch name {
	case "read_line", "is_int", "to_int", "rand_int", "print":
		return name
	default:
		return name
	}
}

func mapOp(op string) string {
	switch op {
	case "and", "or":
		return op
	default:
		return op
	}
}

func collectUsecases(p *ir.Program) []ir.UsecaseDecl {
	var list []ir.UsecaseDecl
	for _, m := range p.Modules {
		for _, uc := range m.Usecases {
			list = append(list, uc)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}

func collectHelpers(usecases []ir.UsecaseDecl) map[string]bool {
	used := map[string]bool{}
	for _, uc := range usecases {
		for _, s := range uc.Body {
			collectHelpersFromStmt(s, used)
		}
	}
	return used
}

func collectHelpersFromStmt(s ir.Stmt, used map[string]bool) {
	switch s.Kind {
	case "let":
		collectHelpersFromExpr(s.Let.Value, used)
	case "assign":
		collectHelpersFromExpr(s.Assign.Value, used)
	case "expr":
		collectHelpersFromExpr(*s.Expr, used)
	case "if":
		collectHelpersFromExpr(s.If.Cond, used)
		for _, t := range s.If.Then {
			collectHelpersFromStmt(t, used)
		}
		for _, e := range s.If.Else {
			collectHelpersFromStmt(e, used)
		}
	case "loop":
		for _, b := range s.Loop.Body {
			collectHelpersFromStmt(b, used)
		}
	case "return":
		if s.Return.Value != nil {
			collectHelpersFromExpr(*s.Return.Value, used)
		}
	}
}

func collectHelpersFromExpr(e ir.Expr, used map[string]bool) {
	switch e.Kind {
	case "call":
		name := e.Call.Name
		switch name {
		case "read_line", "is_int", "to_int", "rand_int":
			used[name] = true
		}
		for _, a := range e.Call.Args {
			collectHelpersFromExpr(a, used)
		}
	case "binary":
		collectHelpersFromExpr(e.Binary.Left, used)
		collectHelpersFromExpr(e.Binary.Right, used)
	case "unary":
		collectHelpersFromExpr(e.Unary.Expr, used)
	}
}

func selectEntry(usecases []ir.UsecaseDecl) string {
	for _, uc := range usecases {
		if uc.Name == "run" {
			return uc.Name
		}
	}
	if len(usecases) == 1 {
		return usecases[0].Name
	}
	return ""
}
