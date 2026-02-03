package check

import (
	"strings"

	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/policy"
	"github.com/willams/lia/internal/symbols"
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

	diags = append(diags, symbols.DeriveProgramSymbols(p)...)
	diags = append(diags, validateEffects(p)...)

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

func validateEffects(p *ir.Program) []ir.Diagnostic {
	var diags []ir.Diagnostic
	allowed := map[string]bool{
		"pure": true,
		"io":   true,
		"tx":   true,
		"emit": true,
	}

	for i := range p.Modules {
		m := &p.Modules[i]
		for _, uc := range m.Usecases {
			diags = append(diags, validateEffectList(m.Name, uc.Effects, allowed)...)
		}
		for _, ad := range m.Adapters {
			diags = append(diags, validateEffectList(m.Name, ad.Effects, allowed)...)
		}
	}

	return diags
}

func validateEffectList(path string, effects []string, allowed map[string]bool) []ir.Diagnostic {
	var diags []ir.Diagnostic
	if len(effects) == 0 {
		return diags
	}
	seenPure := false
	for _, ef := range effects {
		if !allowed[ef] {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "unknown effect: " + ef,
				Path:     path,
			})
			continue
		}
		if ef == "pure" {
			seenPure = true
		}
	}
	if seenPure && len(effects) > 1 {
		diags = append(diags, ir.Diagnostic{
			Severity: "error",
			Message:  "effect 'pure' cannot be combined with other effects",
			Path:     path,
		})
	}
	return diags
}
