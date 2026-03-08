package repro

import (
	"slices"
	"strings"

	"github.com/willams/lia/internal/ir"
)

// ValidationSummary reports how much of the program was covered by tape checks.
type ValidationSummary struct {
	TapeEntries      int
	Modules          int
	GeneratedModules int
	ValidatedModules int
}

// ValidateProgramAgainstTape checks whether @gen metadata is backed by prompt tape entries.
func ValidateProgramAgainstTape(p *ir.Program, tape *PromptTape) ([]ir.Diagnostic, ValidationSummary) {
	summary := ValidationSummary{}
	if p == nil {
		return []ir.Diagnostic{{Severity: "error", Message: "nil program"}}, summary
	}
	if tape == nil {
		return []ir.Diagnostic{{Severity: "error", Message: "nil prompt tape"}}, summary
	}

	summary.Modules = len(p.Modules)
	summary.TapeEntries = len(tape.Prompts)

	entries, diags := indexTape(tape)
	for _, mod := range p.Modules {
		if mod.Gen == nil {
			continue
		}
		if strings.TrimSpace(mod.Gen.PromptRef) == "" && strings.TrimSpace(mod.Gen.PromptHash) == "" {
			continue
		}
		summary.GeneratedModules++

		if strings.TrimSpace(mod.Gen.PromptRef) == "" {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "gen.prompt_ref is required when generation metadata is present",
				Path:     mod.Name,
			})
			continue
		}

		entry, ok := entries[mod.Gen.PromptRef]
		if !ok {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "prompt ref not found in tape: " + mod.Gen.PromptRef,
				Path:     mod.Name,
			})
			continue
		}
		if strings.TrimSpace(mod.Gen.PromptHash) == "" {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "gen.prompt_hash is required for replay validation",
				Path:     mod.Name,
			})
			continue
		}
		if entry.Hash != mod.Gen.PromptHash {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "prompt hash mismatch for ref: " + mod.Gen.PromptRef,
				Path:     mod.Name,
			})
			continue
		}

		if len(mod.Gen.ContextRefs) > 0 && len(entry.ContextRefs) > 0 && !sameStrings(mod.Gen.ContextRefs, entry.ContextRefs) {
			diags = append(diags, ir.Diagnostic{
				Severity: "warning",
				Message:  "gen.context_refs differ from prompt tape",
				Path:     mod.Name,
			})
		}
		if len(mod.Gen.ModelParams) > 0 && len(entry.Params) > 0 && !sameKVs(mod.Gen.ModelParams, entry.Params) {
			diags = append(diags, ir.Diagnostic{
				Severity: "warning",
				Message:  "gen.model_params differ from prompt tape params",
				Path:     mod.Name,
			})
		}

		summary.ValidatedModules++
	}

	return diags, summary
}

func indexTape(tape *PromptTape) (map[string]PromptEntry, []ir.Diagnostic) {
	entries := make(map[string]PromptEntry, len(tape.Prompts))
	var diags []ir.Diagnostic
	for _, entry := range tape.Prompts {
		ref := strings.TrimSpace(entry.Ref)
		if ref == "" {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "prompt tape entry has empty ref",
			})
			continue
		}
		if _, exists := entries[ref]; exists {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "duplicate prompt ref in tape: " + ref,
			})
			continue
		}
		entries[ref] = entry
	}
	return entries, diags
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	a := append([]string(nil), left...)
	b := append([]string(nil), right...)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

func sameKVs(left, right []ir.KV) bool {
	if len(left) != len(right) {
		return false
	}
	a := append([]ir.KV(nil), left...)
	b := append([]ir.KV(nil), right...)
	slices.SortFunc(a, compareKV)
	slices.SortFunc(b, compareKV)
	return slices.Equal(a, b)
}

func compareKV(a, b ir.KV) int {
	if a.Key < b.Key {
		return -1
	}
	if a.Key > b.Key {
		return 1
	}
	if a.Value < b.Value {
		return -1
	}
	if a.Value > b.Value {
		return 1
	}
	return 0
}
