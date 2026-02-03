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
