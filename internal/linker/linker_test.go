package linker

import (
	"testing"

	"github.com/willams/lia/internal/ir"
)

// TestLink_EmptyInputs tests linker with no inputs.
func TestLink_EmptyInputs(t *testing.T) {
	_, _, _, err := Link(nil, nil)
	if err == nil {
		t.Fatal("expected error for empty inputs, got nil")
	}
}

// TestLink_SingleProgram tests linking a single program.
func TestLink_SingleProgram(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{Name: "test.module", Role: "domain"},
		},
	}

	linked, log, diags, err := Link([]*ir.Program{prog}, nil)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	if linked == nil {
		t.Fatal("expected non-nil linked program")
	}

	if len(linked.Modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(linked.Modules))
	}

	if log == nil {
		t.Fatal("expected non-nil decision log")
	}

	if len(log.Entries) != 1 {
		t.Errorf("expected 1 decision entry, got %d", len(log.Entries))
	}

	if log.Hash == "" {
		t.Error("expected decision log hash to be set")
	}

	// Diagnostics may be nil or empty, both are valid
	_ = diags
}

// TestLink_MultiplePrograms tests linking multiple programs.
func TestLink_MultiplePrograms(t *testing.T) {
	prog1 := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{Name: "mod1", Role: "domain"},
		},
	}

	prog2 := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{Name: "mod2", Role: "adapter"},
		},
	}

	linked, log, _, err := Link([]*ir.Program{prog1, prog2}, nil)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	if len(linked.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(linked.Modules))
	}

	if len(log.Entries) != 2 {
		t.Errorf("expected 2 decision entries, got %d", len(log.Entries))
	}
}

// TestWriteDecisionLog tests decision log serialization.
func TestWriteDecisionLog(t *testing.T) {
	log := &DecisionLog{
		Entries: []DecisionEntry{
			{Symbol: "test.symbol", Chosen: "test.impl", Reason: "test reason"},
		},
	}

	tmpfile := t.TempDir() + "/decision.json"
	if err := WriteDecisionLog(tmpfile, log); err != nil {
		t.Fatalf("WriteDecisionLog failed: %v", err)
	}
}

// TestWriteDecisionLog_Nil tests writing nil log.
func TestWriteDecisionLog_Nil(t *testing.T) {
	err := WriteDecisionLog("/tmp/test.json", nil)
	if err == nil {
		t.Fatal("expected error for nil log, got nil")
	}
}

// TestHashDecisionLog_Determinism tests hash reproducibility.
func TestHashDecisionLog_Determinism(t *testing.T) {
	log := &DecisionLog{
		Entries: []DecisionEntry{
			{Symbol: "a", Chosen: "impl_a"},
			{Symbol: "b", Chosen: "impl_b"},
		},
	}

	hash1 := hashDecisionLog(log)
	hash2 := hashDecisionLog(log)

	if hash1 != hash2 {
		t.Errorf("hash not deterministic: %s != %s", hash1, hash2)
	}

	if hash1 == "" {
		t.Error("expected non-empty hash")
	}
}
