package check

import (
	"testing"

	"github.com/willams/lia/internal/ir"
)

func TestCheckProgram_NilProgram(t *testing.T) {
	diags := CheckProgram(nil, nil)
	if len(diags) == 0 {
		t.Fatal("expected error for nil program")
	}
	if diags[0].Message != "nil program" {
		t.Errorf("unexpected message: %s", diags[0].Message)
	}
}

func TestCheckProgram_EmptyVersion(t *testing.T) {
	prog := &ir.Program{}
	diags := CheckProgram(prog, nil)
	found := false
	for _, d := range diags {
		if d.Message == "program version is required" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected version required diagnostic")
	}
}

func TestCheckProgram_EmptyModuleName(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{{Name: ""}},
	}
	diags := CheckProgram(prog, nil)
	found := false
	for _, d := range diags {
		if d.Message == "module name is required" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected module name required diagnostic")
	}
}

func TestCheckProgram_GenPromptHashWithoutRef(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "test.core",
				Gen:  &ir.GenMeta{PromptHash: "abc"},
			},
		},
	}
	diags := CheckProgram(prog, nil)
	found := false
	for _, d := range diags {
		if d.Message == "gen.prompt_ref required when prompt_hash is set" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected prompt_ref required diagnostic")
	}
}

func TestCheckProgram_UnknownEffect(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "test.app",
				Usecases: []ir.UsecaseDecl{
					{Name: "Test", Effects: []string{"unknown_effect"}},
				},
			},
		},
	}
	diags := CheckProgram(prog, nil)
	found := false
	for _, d := range diags {
		if d.Message == "unknown effect: unknown_effect" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected unknown effect diagnostic")
	}
}

func TestCheckProgram_InvalidReproProfile(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Projects: []ir.Project{
			{Name: "test", Repro: ir.ReproProfile("invalid_profile")},
		},
	}
	diags := CheckProgram(prog, nil)
	found := false
	for _, d := range diags {
		if d.Severity == "error" && d.Message == "invalid repro profile: invalid_profile" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected invalid repro profile diagnostic")
	}
}

func TestCheckProgram_PinnedReproRequiresTape(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Projects: []ir.Project{
			{Name: "test", Repro: ir.ReproPinned},
		},
	}
	diags := CheckProgram(prog, nil)
	found := false
	for _, d := range diags {
		if d.Message == "project tape is required when repro is strict or pinned" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected tape required diagnostic")
	}
}

func TestCheckProgram_ValidProgram(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Projects: []ir.Project{{Name: "test"}},
		Modules: []ir.Module{
			{
				Name: "test.core",
				Role: "domain",
				Types: []ir.TypeDecl{{Name: "Id", Base: "String"}},
				Enums: []ir.EnumDecl{{Name: "Status", Values: []string{"A"}}},
			},
		},
	}
	diags := CheckProgram(prog, nil)
	if HasErrors(diags) {
		t.Errorf("expected no errors for valid program, got: %v", diags)
	}
}

func TestCheckProgram_WithPolicies(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "test.core",
				Role: "infra",
			},
		},
	}
	packs := []ir.Pack{
		{
			Name: "TestPack",
			Policies: []ir.PolicyDecl{
				{Name: "roles", Body: `allow_roles ["domain", "usecase"]`},
			},
		},
	}
	diags := CheckProgram(prog, packs)
	found := false
	for _, d := range diags {
		if d.Severity == "error" && d.Message == "role not allowed by policy: infra" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected role not allowed diagnostic")
	}
}

func TestHasErrors(t *testing.T) {
	if HasErrors(nil) {
		t.Error("nil should have no errors")
	}
	if HasErrors([]ir.Diagnostic{{Severity: "warning"}}) {
		t.Error("warnings should not be errors")
	}
	if !HasErrors([]ir.Diagnostic{{Severity: "error"}}) {
		t.Error("error should be detected")
	}
}

func TestCheckProgram_GenPromptHashNoPromptRef(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "m1",
				Gen: &ir.GenMeta{
					PromptRef:  "ref",
					PromptHash: "hash",
					ModelID:    "model",
				},
			},
		},
	}
	diags := CheckProgram(prog, nil)
	for _, d := range diags {
		if d.Severity == "error" {
			t.Errorf("unexpected error with complete gen: %s", d.Message)
		}
	}
}

func TestCheckProgram_AdapterEffects(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "test.infra",
				Adapters: []ir.AdapterDecl{
					{Name: "PgAdapter", Effects: []string{"pure", "io"}},
				},
			},
		},
	}
	diags := CheckProgram(prog, nil)
	found := false
	for _, d := range diags {
		if d.Message == "effect 'pure' cannot be combined with other effects" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected pure+io diagnostic for adapter")
	}
}

func TestCheckProgram_GenericTypeNotReported(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "test.core",
				Ports: []ir.PortDecl{
					{
						Name: "Repo",
						Methods: []ir.FuncDecl{
							{
								Name:    "Get",
								Params:  []ir.Field{{Name: "id", Type: "list<String>"}},
								Returns: []ir.Field{{Name: "ok", Type: "Bool"}},
							},
						},
					},
				},
			},
		},
	}
	diags := CheckProgram(prog, nil)
	for _, d := range diags {
		if d.Severity == "error" && d.Message != "module name is required" {
			t.Errorf("unexpected error for generic type: %s", d.Message)
		}
	}
}

func TestCheckProgram_ArrayTypeNotReported(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "test.core",
				Types: []ir.TypeDecl{
					{Name: "Ids", Base: "[]String"},
				},
			},
		},
	}
	diags := CheckProgram(prog, nil)
	for _, d := range diags {
		if d.Message == "unknown type reference: String" {
			t.Error("[]String should not produce unknown type error")
		}
	}
}
