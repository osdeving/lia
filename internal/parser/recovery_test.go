package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/willams/lia/internal/ir"
)

func TestParseWithRecovery_ValidInput(t *testing.T) {
	input := `module orders.app as usecase {
  type OrderId = String where nonEmpty;
  enum Status { NEW, PAID };
  usecase CreateOrder {
    input { id: OrderId };
    output { ok: Bool };
    effects [io];
    return true;
  }
}
`
	result := ParseWithRecovery(input)
	if result.HasErrors() {
		t.Fatalf("expected no errors, got %v", result.Diagnostics)
	}
	if result.Parsed != 1 {
		t.Errorf("expected 1 parsed block, got %d", result.Parsed)
	}
	if len(result.Program.Modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(result.Program.Modules))
	}
}

func TestParseWithRecovery_PartialFailure(t *testing.T) {
	input := `module good.one as domain {
  type Email = String;
}

module bad.two {
  type X = String where

module good.three as domain {
  enum Color { RED, GREEN };
}
`
	result := ParseWithRecovery(input)
	if !result.HasErrors() {
		t.Fatal("expected errors for bad.two")
	}
	if result.TotalBlocks != 3 {
		t.Errorf("expected 3 blocks, got %d", result.TotalBlocks)
	}
	if result.Parsed != 2 {
		t.Errorf("expected 2 parsed, got %d", result.Parsed)
	}
	if len(result.Diagnostics) != 1 {
		t.Errorf("expected 1 diagnostic, got %d", len(result.Diagnostics))
	}
	if len(result.Program.Modules) != 2 {
		t.Errorf("expected 2 modules in partial result, got %d", len(result.Program.Modules))
	}
}

func TestParseWithRecovery_AllInvalid(t *testing.T) {
	input := `module bad1 { invalid syntax
module bad2 { also invalid`
	result := ParseWithRecovery(input)
	if !result.HasErrors() {
		t.Fatal("expected errors")
	}
	if result.Parsed != 0 {
		t.Errorf("expected 0 parsed, got %d", result.Parsed)
	}
	if len(result.Diagnostics) != 2 {
		t.Errorf("expected 2 diagnostics, got %d", len(result.Diagnostics))
	}
}

func TestParseWithRecovery_EmptyInput(t *testing.T) {
	result := ParseWithRecovery("")
	if result.HasErrors() {
		t.Fatalf("expected no errors for empty input, got %v", result.Diagnostics)
	}
	if result.Program == nil {
		t.Fatal("expected non-nil program")
	}
}

func TestParseWithRecovery_MultipleValid(t *testing.T) {
	input := `module one.core as domain {
  type Id = String;
}

module two.core as domain {
  type Name = String;
}

pack TestPack@1.0 {
  policy p: allow_roles ["domain"];
}
`
	result := ParseWithRecovery(input)
	if result.HasErrors() {
		t.Fatalf("expected no errors, got %v", result.Diagnostics)
	}
	if result.Parsed != 3 {
		t.Errorf("expected 3 parsed, got %d", result.Parsed)
	}
	if len(result.Program.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(result.Program.Modules))
	}
	if len(result.Program.Packs) != 1 {
		t.Errorf("expected 1 pack, got %d", len(result.Program.Packs))
	}
}

func TestParseWithRecovery_WithProject(t *testing.T) {
	input := `project Demo {
  repro strict;

  module demo.core as domain {
    type Email = String;
  }
}
`
	result := ParseWithRecovery(input)
	if result.HasErrors() {
		t.Fatalf("expected no errors, got %v", result.Diagnostics)
	}
	if len(result.Program.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(result.Program.Projects))
	}
}

func TestParseWithRecovery_WithGenBlock(t *testing.T) {
	input := `@gen { prompt_ref:"p1", model_id:"m1" }
module gen.core as domain {
  type Id = String;
}
`
	result := ParseWithRecovery(input)
	if result.HasErrors() {
		t.Fatalf("expected no errors, got %v", result.Diagnostics)
	}
	if len(result.Program.Modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(result.Program.Modules))
	}
	if result.Program.Modules[0].Gen == nil {
		t.Error("expected gen metadata")
	}
}

func TestParseFileWithRecovery_FileNotFound(t *testing.T) {
	result := ParseFileWithRecovery("/nonexistent/file.lia")
	if !result.HasErrors() {
		t.Fatal("expected error for missing file")
	}
	if result.Program == nil {
		t.Fatal("expected non-nil program even on error")
	}
}

func TestParseFileWithRecovery_ValidFile(t *testing.T) {
	content := `module file.test as domain {
  type Id = String;
}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.lia")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	result := ParseFileWithRecovery(path)
	if result.HasErrors() {
		t.Fatalf("expected no errors, got %v", result.Diagnostics)
	}
	if len(result.Program.Modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(result.Program.Modules))
	}
}

func TestParseDiagnostic_String(t *testing.T) {
	d := ParseDiagnostic{Line: 5, Column: 10, Message: "syntax error"}
	s := d.String()
	if s != "line 5:10: syntax error" {
		t.Errorf("unexpected diagnostic string: %s", s)
	}
}

func TestRecoveryResult_HasErrors(t *testing.T) {
	r := &RecoveryResult{Program: &ir.Program{}}
	if r.HasErrors() {
		t.Error("expected no errors when diagnostics is empty")
	}
	r.Diagnostics = append(r.Diagnostics, ParseDiagnostic{Message: "err"})
	if !r.HasErrors() {
		t.Error("expected HasErrors to return true")
	}
}

func TestSplitTopLevelBlocks_NoBlocks(t *testing.T) {
	blocks := splitTopLevelBlocks("just some random text")
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks, got %d", len(blocks))
	}
}

func TestSplitTopLevelBlocks_Multiple(t *testing.T) {
	input := `module a as domain {}
module b as domain {}
pack C@1 {}`
	blocks := splitTopLevelBlocks(input)
	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(blocks))
	}
	if blocks[0].StartLine != 1 {
		t.Errorf("block 0 start line: expected 1, got %d", blocks[0].StartLine)
	}
	if blocks[1].StartLine != 2 {
		t.Errorf("block 1 start line: expected 2, got %d", blocks[1].StartLine)
	}
	if blocks[2].StartLine != 3 {
		t.Errorf("block 2 start line: expected 3, got %d", blocks[2].StartLine)
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"abc", 0},
		{"a\nb", 1},
		{"a\nb\nc", 2},
		{"\n\n\n", 3},
	}
	for _, tc := range tests {
		got := countLines(tc.input)
		if got != tc.want {
			t.Errorf("countLines(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestMergeProgram(t *testing.T) {
	dst := &ir.Program{Version: "0.1"}
	src := &ir.Program{
		Modules:  []ir.Module{{Name: "a"}},
		Projects: []ir.Project{{Name: "p"}},
		Packs:    []ir.Pack{{Name: "k"}},
	}
	mergeProgram(dst, src)
	if len(dst.Modules) != 1 || len(dst.Projects) != 1 || len(dst.Packs) != 1 {
		t.Errorf("merge failed: modules=%d projects=%d packs=%d", len(dst.Modules), len(dst.Projects), len(dst.Packs))
	}
}
