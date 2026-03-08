package java

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/linker"
	"github.com/willams/lia/internal/parser"
)

func TestLowerProject_FullPipelineCompiles(t *testing.T) {
	repoRoot := repoRoot(t)
	input := filepath.Join(repoRoot, "examples", "full-pipeline", "project.lia")

	prog, err := parser.ParseFile(input)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	linkedProgram, _, _, err := linker.Link([]*ir.Program{prog}, nil)
	if err != nil {
		t.Fatalf("link failed: %v", err)
	}

	project, err := LowerProject(linkedProgram)
	if err != nil {
		t.Fatalf("lower failed: %v", err)
	}

	outDir := t.TempDir()
	if err := WriteProject(outDir, project); err != nil {
		t.Fatalf("write project failed: %v", err)
	}

	files := collectGeneratedJavaFiles(filepath.Join(outDir, "src", "main", "java"))
	if len(files) == 0 {
		t.Fatal("expected generated Java files")
	}

	buildDir := filepath.Join(outDir, ".build")
	args := append([]string{"-d", buildDir}, files...)
	cmd := exec.Command("javac", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("javac failed: %v\n%s", err, string(output))
	}
}

func collectGeneratedJavaFiles(dir string) []string {
	var files []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info != nil && !info.IsDir() && filepath.Ext(path) == ".java" {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}
