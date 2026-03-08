package cli

import (
	"bytes"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLI_FullPipelineExample(t *testing.T) {
	repoRoot := mustRepoRoot(t)
	packDir := filepath.Join(repoRoot, "docs", "pt-br", "spec", "packs")
	input := filepath.Join(repoRoot, "examples", "full-pipeline", "project.lia")
	tape := filepath.Join(repoRoot, "examples", "full-pipeline", "prompt-tape.json")

	outDir := t.TempDir()
	liao := filepath.Join(outDir, "full-pipeline.liao")
	lial := filepath.Join(outDir, "full-pipeline.lial")
	logPath := filepath.Join(outDir, "full-pipeline.decision-log.json")

	out, err := executeCLI("parse", input, "-o", liao)
	if err != nil {
		t.Fatalf("parse failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "wrote "+liao) {
		t.Fatalf("unexpected parse output: %s", out)
	}

	out, err = executeCLI("check", liao, "--pack-dir", packDir)
	if err != nil {
		t.Fatalf("check failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "ok") {
		t.Fatalf("expected ok in check output: %s", out)
	}

	out, err = executeCLI("link", liao, "-o", lial, "--decision-log", logPath, "--pack-dir", packDir)
	if err != nil {
		t.Fatalf("link failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "wrote "+lial) {
		t.Fatalf("unexpected link output: %s", out)
	}

	out, err = executeCLI("explain", lial, "--decision-log", logPath)
	if err != nil {
		t.Fatalf("explain failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "decision log entries") {
		t.Fatalf("unexpected explain output: %s", out)
	}

	out, err = executeCLI("replay", lial, "--tape", tape)
	if err != nil {
		t.Fatalf("replay failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "modules 3, generated 3, validated 3") {
		t.Fatalf("unexpected replay output: %s", out)
	}
}

func executeCLI(args ...string) (string, error) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return out.String(), err
}

func mustRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
