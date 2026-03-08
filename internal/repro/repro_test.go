package repro

import (
	"testing"

	"github.com/willams/lia/internal/ir"
)

func TestValidateProgramAgainstTape_OK(t *testing.T) {
	body := "module demo.core as domain { }"
	hash := ComputePromptHash(body)
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "demo.core",
				Gen: &ir.GenMeta{
					PromptRef:  "p-001",
					PromptHash: hash,
					ModelID:    "mock",
					ContextRefs: []string{
						"spec:lia-v0.1",
					},
					ModelParams: []ir.KV{
						{Key: "temperature", Value: "0.1"},
					},
				},
			},
		},
	}
	tape := &PromptTape{
		Version: "0.1",
		Prompts: []PromptEntry{
			{
				Ref:  "p-001",
				Body: body,
				Hash: hash,
				ContextRefs: []string{
					"spec:lia-v0.1",
				},
				Params: []ir.KV{
					{Key: "temperature", Value: "0.1"},
				},
			},
		},
	}

	diags, summary := ValidateProgramAgainstTape(prog, tape)
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if summary.GeneratedModules != 1 || summary.ValidatedModules != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestValidateProgramAgainstTape_MissingPromptRef(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "demo.core",
				Gen: &ir.GenMeta{
					PromptRef:  "p-404",
					PromptHash: "abc",
					ModelID:    "mock",
				},
			},
		},
	}
	tape := &PromptTape{Version: "0.1"}

	diags, _ := ValidateProgramAgainstTape(prog, tape)
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics")
	}
	if diags[0].Message != "prompt ref not found in tape: p-404" {
		t.Fatalf("unexpected diagnostic: %+v", diags[0])
	}
}

func TestValidateProgramAgainstTape_DuplicateRefs(t *testing.T) {
	hash := ComputePromptHash("module demo.core as domain { }")
	tape := &PromptTape{
		Version: "0.1",
		Prompts: []PromptEntry{
			{Ref: "p-001", Body: "a", Hash: hash},
			{Ref: "p-001", Body: "b", Hash: hash},
		},
	}

	diags, _ := ValidateProgramAgainstTape(&ir.Program{Version: "0.1"}, tape)
	if len(diags) == 0 {
		t.Fatalf("expected duplicate ref diagnostic")
	}
	if diags[0].Message != "duplicate prompt ref in tape: p-001" {
		t.Fatalf("unexpected diagnostic: %+v", diags[0])
	}
}
