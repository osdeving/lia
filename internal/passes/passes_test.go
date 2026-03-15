package passes

import (
	"strings"
	"testing"

	"github.com/willams/lia/internal/ir"
)

func TestRunner_Pipeline(t *testing.T) {
	runner := NewRunner()
	runner.Add(NewTransformDeduplicatePass())
	runner.Add(NewSynthFillHolesPass())
	runner.Add(NewVerifyHolesPass())

	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "test.core",
				Types: []ir.TypeDecl{
					{Name: "Id", Base: "String"},
					{Name: "Id", Base: "String"}, // Duplicate
				},
				Enums: []ir.EnumDecl{
					{Name: "Status", Values: []string{"A", "B"}},
					{Name: "Status", Values: []string{"A", "B"}}, // Duplicate
				},
				Holes: []ir.HoleDecl{
					{Name: "paymentGateway", Contract: "Payment contract"},
					{Name: "unresolvedHole", Contract: "Missing contract"},
				},
				Candidates: []ir.CandidateDecl{
					{Symbol: "paymentGateway", Score: 1.0},
				},
			},
		},
	}

	ctx := &Context{}
	res, err := runner.Run(ctx, prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(prog.Modules[0].Types) != 1 {
		t.Errorf("expected 1 type after dedup, got %d", len(prog.Modules[0].Types))
	}
	if len(prog.Modules[0].Enums) != 1 {
		t.Errorf("expected 1 enum after dedup, got %d", len(prog.Modules[0].Enums))
	}

	if len(prog.Modules[0].Holes) != 1 {
		t.Errorf("expected 1 unresolved hole remaining, got %d", len(prog.Modules[0].Holes))
	}

	foundResolvedInfo := false
	foundUnresolvedError := false
	foundDuplicateTypeWarning := false
	foundDuplicateEnumWarning := false

	for _, d := range res.Diags {
		if d.Severity == "info" && strings.Contains(d.Message, "resolved hole paymentGateway") {
			foundResolvedInfo = true
		}
		if d.Severity == "error" && strings.Contains(d.Message, "unresolved hole: unresolvedHole") {
			foundUnresolvedError = true
		}
		if d.Severity == "warning" && strings.Contains(d.Message, "removed duplicate type declaration: Id") {
			foundDuplicateTypeWarning = true
		}
		if d.Severity == "warning" && strings.Contains(d.Message, "removed duplicate enum declaration: Status") {
			foundDuplicateEnumWarning = true
		}
	}

	if !foundResolvedInfo {
		t.Error("expected info diagnostic for resolved hole")
	}
	if !foundUnresolvedError {
		t.Error("expected error diagnostic for unresolved hole")
	}
	if !foundDuplicateTypeWarning {
		t.Error("expected warning diagnostic for duplicate type")
	}
	if !foundDuplicateEnumWarning {
		t.Error("expected warning diagnostic for duplicate enum")
	}
}

func TestRunner_NilProgram(t *testing.T) {
	runner := NewRunner()
	runner.Add(NewVerifyHolesPass())
	res, _ := runner.Run(&Context{}, nil)
	if len(res.Diags) != 0 {
		t.Errorf("expected 0 diags for nil program, got %d", len(res.Diags))
	}
}
