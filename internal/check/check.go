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
	diags = append(diags, validateRepro(p)...)
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
	diags = append(diags, validateTypeRefs(p)...)

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

func validateRepro(p *ir.Program) []ir.Diagnostic {
	var diags []ir.Diagnostic
	allowedProfiles := map[ir.ReproProfile]bool{
		"":                 true,
		ir.ReproStrict:     true,
		ir.ReproPinned:     true,
		ir.ReproBestEffort: true,
	}

	for _, proj := range p.Projects {
		if !allowedProfiles[proj.Repro] {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "invalid repro profile: " + string(proj.Repro),
				Path:     proj.Name,
			})
		}
		if (proj.Repro == ir.ReproStrict || proj.Repro == ir.ReproPinned) && strings.TrimSpace(proj.Tape) == "" {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "project tape is required when repro is strict or pinned",
				Path:     proj.Name,
			})
		}
	}

	for _, mod := range p.Modules {
		if mod.Gen == nil {
			continue
		}
		if strings.TrimSpace(mod.Gen.PromptHash) != "" && strings.TrimSpace(mod.Gen.PromptRef) == "" {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "gen.prompt_ref required when prompt_hash is set",
				Path:     mod.Name,
			})
		}
		if strings.TrimSpace(mod.Gen.PromptRef) != "" && strings.TrimSpace(mod.Gen.ModelID) == "" {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "gen.model_id required when prompt_ref is set",
				Path:     mod.Name,
			})
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

func validateTypeRefs(p *ir.Program) []ir.Diagnostic {
	known := map[string]bool{
		"String":  true,
		"Bool":    true,
		"bool":    true,
		"Boolean": true,
		"Int":     true,
		"int":     true,
		"Integer": true,
		"Long":    true,
		"long":    true,
		"Float":   true,
		"float":   true,
		"Double":  true,
		"double":  true,
	}
	for _, mod := range p.Modules {
		for _, decl := range mod.Types {
			known[strings.TrimSpace(decl.Name)] = true
		}
		for _, decl := range mod.Enums {
			known[strings.TrimSpace(decl.Name)] = true
		}
	}

	var diags []ir.Diagnostic
	seen := map[string]bool{}
	addDiag := func(path, raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		key := path + "::" + raw
		if seen[key] {
			return
		}
		seen[key] = true
		diags = append(diags, ir.Diagnostic{
			Severity: "error",
			Message:  "unknown type reference: " + raw,
			Path:     path,
		})
	}

	var validate func(path, raw string)
	validate = func(path, raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		for strings.HasPrefix(raw, "[]") {
			raw = strings.TrimSpace(strings.TrimPrefix(raw, "[]"))
		}
		if raw == "" || known[raw] {
			return
		}
		if strings.Contains(raw, "<") || strings.Contains(raw, ">") {
			return
		}
		addDiag(path, raw)
	}

	for _, mod := range p.Modules {
		path := mod.Name
		for _, decl := range mod.Types {
			validate(path, decl.Base)
		}
		for _, decl := range mod.Ports {
			for _, method := range decl.Methods {
				for _, field := range method.Params {
					validate(path, field.Type)
				}
				for _, field := range method.Returns {
					validate(path, field.Type)
				}
			}
		}
		for _, decl := range mod.Usecases {
			for _, field := range decl.Inputs {
				validate(path, field.Type)
			}
			for _, field := range decl.Outputs {
				validate(path, field.Type)
			}
		}
		for _, decl := range mod.Adapters {
			for _, field := range decl.Inputs {
				validate(path, field.Type)
			}
			for _, field := range decl.Outputs {
				validate(path, field.Type)
			}
		}
	}
	return diags
}
