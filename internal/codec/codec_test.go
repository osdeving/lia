package codec

import (
	"os"
	"testing"

	"github.com/willams/lia/internal/ir"
)

// TestWriteProgramFile tests writing a program to a file.
func TestWriteProgramFile(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Projects: []ir.Project{
			{Name: "TestProject", Repro: ir.ReproStrict},
		},
		Modules: []ir.Module{
			{Name: "test.module", Role: "domain"},
		},
	}

	tmpfile := t.TempDir() + "/test.liao"
	if err := WriteProgramFile(tmpfile, prog); err != nil {
		t.Fatalf("WriteProgramFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpfile); os.IsNotExist(err) {
		t.Fatalf("output file not created: %v", err)
	}
}

// TestLoadProgramFile tests reading a program from a file.
func TestLoadProgramFile(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{Name: "test.module", Role: "domain"},
		},
	}

	tmpfile := t.TempDir() + "/test.liao"
	if err := WriteProgramFile(tmpfile, prog); err != nil {
		t.Fatalf("WriteProgramFile failed: %v", err)
	}

	loaded, err := LoadProgramFile(tmpfile)
	if err != nil {
		t.Fatalf("LoadProgramFile failed: %v", err)
	}

	if loaded.Version != prog.Version {
		t.Errorf("expected version %s, got %s", prog.Version, loaded.Version)
	}

	if len(loaded.Modules) != len(prog.Modules) {
		t.Errorf("expected %d modules, got %d", len(prog.Modules), len(loaded.Modules))
	}
}

func ReadProgramFile(tmpfile string) (any, any) {
	panic("unimplemented")
}

// TestHashBytes tests byte hashing.
func TestHashBytes(t *testing.T) {
	data := []byte("test data")
	hash1 := HashBytes(data)
	hash2 := HashBytes(data)

	if hash1 != hash2 {
		t.Errorf("hash not deterministic: %s != %s", hash1, hash2)
	}

	if hash1 == "" {
		t.Error("expected non-empty hash")
	}

	// Different data should produce different hash
	differentData := []byte("different data")
	hash3 := HashBytes(differentData)
	if hash1 == hash3 {
		t.Error("different data produced same hash")
	}
}
