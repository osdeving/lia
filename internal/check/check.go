package check

import (
	"strings"

	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/policy"
)

// CheckProgram performs basic structural validation with optional packs.
func CheckProgram(p *ir.Program, packs []ir.Pack) []ir.Diagnostic {
	var diags []ir.Diagnostic
	if p == nil {
		return []ir.Diagnostic{{Severity: "error", Message: "nil program"}}
	}
	if strings.TrimSpace(p.Version) == "" {
		diags = append(diags, ir.Diagnostic{Severity: "error", Message: "program version is required"})
	}
	for i := range p.Modules {
		m := &p.Modules[i]
		if strings.TrimSpace(m.Name) == "" {
			diags = append(diags, ir.Diagnostic{Severity: "error", Message: "module name is required"})
		}
		if m.Gen != nil {
			if m.Gen.PromptRef != "" && m.Gen.PromptHash == "" {
				diags = append(diags, ir.Diagnostic{Severity: "error", Message: "gen.prompt_hash required when prompt_ref is set", Path: m.Name})
			}
		}
	}

	eng, pdiags := policy.NewEngine(packs)
	diags = append(diags, pdiags...)
	diags = append(diags, eng.ApplyLocal(p)...)
	return diags
}

// HasErrors returns true if any diagnostic is an error.
func HasErrors(diags []ir.Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == "error" {
			return true
		}
	}
	return false
}
