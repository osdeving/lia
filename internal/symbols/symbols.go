package symbols

import (
	"strings"

	"github.com/willams/lia/internal/ir"
)

// SymbolQName builds a canonical QName for a module symbol.
func SymbolQName(moduleName, kind, name string) string {
	moduleName = strings.TrimSpace(moduleName)
	kind = strings.TrimSpace(kind)
	name = strings.TrimSpace(name)
	if moduleName == "" {
		return kind + ":" + name
	}
	return moduleName + "::" + kind + ":" + name
}

// QualifySymbol ensures a symbol reference includes a module prefix.
func QualifySymbol(moduleName, kind, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.Contains(raw, "::") {
		return raw
	}
	if strings.Contains(raw, ":") {
		return moduleName + "::" + raw
	}
	return SymbolQName(moduleName, kind, raw)
}

// DeriveProgramSymbols computes provides/requires for all modules.
func DeriveProgramSymbols(p *ir.Program) []ir.Diagnostic {
	var diags []ir.Diagnostic
	if p == nil {
		return []ir.Diagnostic{{Severity: "error", Message: "nil program"}}
	}
	for i := range p.Modules {
		m := &p.Modules[i]
		provides, requires, d := DeriveModuleSymbols(m)
		diags = append(diags, d...)
		m.Provides = provides
		m.Requires = requires
	}
	return diags
}

// DeriveModuleSymbols computes provides/requires based on module contents.
func DeriveModuleSymbols(m *ir.Module) ([]ir.SymbolRef, []ir.SymbolRef, []ir.Diagnostic) {
	var diags []ir.Diagnostic
	if m == nil {
		return nil, nil, []ir.Diagnostic{{Severity: "error", Message: "nil module"}}
	}

	provides := []ir.SymbolRef{}
	requires := []ir.SymbolRef{}

	seenProvides := map[string]bool{}
	addProvide := func(kind, name string) {
		q := SymbolQName(m.Name, kind, name)
		if q == "" {
			return
		}
		if seenProvides[q] {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "duplicate symbol in module: " + q,
				Path:     m.Name,
			})
			return
		}
		seenProvides[q] = true
		provides = append(provides, ir.SymbolRef{QName: q, Kind: kind})
	}

	seenRequires := map[string]bool{}
	addRequire := func(kind, raw string) {
		q := QualifySymbol(m.Name, kind, raw)
		if q == "" {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "empty required symbol",
				Path:     m.Name,
			})
			return
		}
		if seenRequires[q] {
			return
		}
		seenRequires[q] = true
		requires = append(requires, ir.SymbolRef{QName: q, Kind: kind})
	}

	for _, t := range m.Types {
		addProvide("type", t.Name)
	}
	for _, e := range m.Enums {
		addProvide("enum", e.Name)
	}
	for _, p := range m.Ports {
		addProvide("port", p.Name)
	}
	for _, uc := range m.Usecases {
		addProvide("usecase", uc.Name)
	}
	for _, ad := range m.Adapters {
		addProvide("adapter", ad.Name)
		if ad.Implements != "" {
			addRequire("port", ad.Implements)
		}
	}
	for _, w := range m.Wirings {
		addProvide("wiring", w.Name)
		for _, bind := range w.Binds {
			left, right, ok := parseBind(bind)
			if !ok {
				diags = append(diags, ir.Diagnostic{
					Severity: "error",
					Message:  "invalid wiring bind: " + bind,
					Path:     m.Name,
				})
				continue
			}
			addRequire("port", left)
			addRequire("adapter", right)
		}
	}

	return provides, requires, diags
}

func parseBind(bind string) (string, string, bool) {
	parts := strings.SplitN(bind, "->", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])
	if left == "" || right == "" {
		return "", "", false
	}
	return left, right, true
}
