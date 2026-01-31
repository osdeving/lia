package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/willams/lia/internal/ir"
)

// TestParseFile_BasicProject tests parsing a basic project with modules.
func TestParseFile_BasicProject(t *testing.T) {
	content := `project TestProj {
  repro strict;
  tape prompt_tape "./tape.json";

  use pack HexCore@1.0.0;

  module orders.core as domain {
  }
}
`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if prog.Version != "0.1" {
		t.Errorf("expected version 0.1, got %s", prog.Version)
	}

	if len(prog.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(prog.Projects))
	}

	proj := prog.Projects[0]
	if proj.Name != "TestProj" {
		t.Errorf("expected project name TestProj, got %s", proj.Name)
	}

	if proj.Repro != ir.ReproStrict {
		t.Errorf("expected repro strict, got %s", proj.Repro)
	}

	if proj.Tape != "./tape.json" {
		t.Errorf("expected tape ./tape.json, got %s", proj.Tape)
	}

	if len(proj.Uses) != 1 {
		t.Fatalf("expected 1 pack ref, got %d", len(proj.Uses))
	}

	packRef := proj.Uses[0]
	if packRef.Name != "HexCore" || packRef.Version != "1.0.0" {
		t.Errorf("expected HexCore@1.0.0, got %s@%s", packRef.Name, packRef.Version)
	}

	if len(prog.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(prog.Modules))
	}

	mod := prog.Modules[0]
	if mod.Name != "orders.core" {
		t.Errorf("expected module orders.core, got %s", mod.Name)
	}

	if mod.Role != "domain" {
		t.Errorf("expected role domain, got %s", mod.Role)
	}
}

// TestParseFile_Empty tests parsing an empty file.
func TestParseFile_Empty(t *testing.T) {
	tmpfile := createTempFile(t, "")
	defer os.Remove(tmpfile)

	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(prog.Projects) != 0 {
		t.Errorf("expected no projects, got %d", len(prog.Projects))
	}
}

// TestParseFile_NonExistent tests parsing a non-existent file.
func TestParseFile_NonExistent(t *testing.T) {
	_, err := ParseFile("/nonexistent/file.lia")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

// TestParsePackRef tests pack reference parsing.
func TestParsePackRef(t *testing.T) {
	tests := []struct {
		input   string
		name    string
		version string
	}{
		{"HexCore@1.0.0", "HexCore", "1.0.0"},
		{"MyPack", "MyPack", ""},
		{"foo@2.1.3-beta", "foo", "2.1.3-beta"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, version := parsePackRef(tt.input)
			if name != tt.name || version != tt.version {
				t.Errorf("parsePackRef(%q) = (%q, %q), want (%q, %q)",
					tt.input, name, version, tt.name, tt.version)
			}
		})
	}
}

// Helper: createTempFile creates a temporary .lia file for testing.
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpdir := t.TempDir()
	tmpfile := filepath.Join(tmpdir, "test.lia")
	if err := os.WriteFile(tmpfile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return tmpfile
}
