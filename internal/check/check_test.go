package check

import (
	"testing"

	"github.com/willams/lia/internal/ir"
)

func TestCheckProgram_EffectValidation(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "orders.app",
				Usecases: []ir.UsecaseDecl{{
					Name:    "Create",
					Effects: []string{"pure", "io"},
				}},
			},
		},
	}
	// No packs needed for this test.
	diags := CheckProgram(prog, nil)
	found := false
	for _, d := range diags {
		if d.Severity == "error" && d.Message == "effect 'pure' cannot be combined with other effects" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected pure+io diagnostic")
	}
}

func TestCheckProgram_ReproRequiresTape(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Projects: []ir.Project{
			{Name: "demo", Repro: ir.ReproStrict},
		},
	}

	diags := CheckProgram(prog, nil)
	found := false
	for _, d := range diags {
		if d.Severity == "error" && d.Message == "project tape is required when repro is strict or pinned" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected repro/tape diagnostic")
	}
}

func TestCheckProgram_GenModelIDRequired(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "demo.core",
				Gen: &ir.GenMeta{
					PromptRef:  "p-001",
					PromptHash: "abc",
				},
			},
		},
	}

	diags := CheckProgram(prog, nil)
	found := false
	for _, d := range diags {
		if d.Severity == "error" && d.Message == "gen.model_id required when prompt_ref is set" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected gen.model_id diagnostic")
	}
}
