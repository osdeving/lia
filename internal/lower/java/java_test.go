package java

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
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
		if strings.Contains(err.Error(), "executable file not found") {
			t.Skip("javac not found in PATH, skipping compilation test")
		}
		t.Fatalf("javac failed: %v\n%s", err, string(output))
	}
}

func TestLowerProject_ProfilesEmitFrameworkMarkers(t *testing.T) {
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

	cases := []struct {
		name    string
		profile Profile
		markers []string
	}{
		{
			name:    "spring boot",
			profile: ProfileSpringBoot,
			markers: []string{"@SpringBootApplication", "@Configuration", "@Bean"},
		},
		{
			name:    "quarkus",
			profile: ProfileQuarkus,
			markers: []string{"@QuarkusMain", "@Produces", "@ApplicationScoped"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			project, err := LowerProjectWithOptions(linkedProgram, Options{Profile: tc.profile})
			if err != nil {
				t.Fatalf("lower failed: %v", err)
			}

			outDir := t.TempDir()
			if err := WriteProject(outDir, project); err != nil {
				t.Fatalf("write project failed: %v", err)
			}

			reportPath := filepath.Join(outDir, "LOWERING_REPORT.md")
			if _, err := os.Stat(reportPath); err != nil {
				t.Fatalf("expected lowering report: %v", err)
			}

			allText := readAllGeneratedText(t, outDir)
			for _, marker := range tc.markers {
				if !strings.Contains(allText, marker) {
					t.Fatalf("expected marker %q in generated project", marker)
				}
			}
		})
	}
}

func TestLowerProject_StrictFailsOnUnknownType(t *testing.T) {
	prog, err := parser.ParseString(`project Demo {
  module orders.domain as domain {
    type OrderId = String where nonEmpty;
  }
  module orders.port as port {
    port OrderRepository {
      fn Get(id: OrderId) -> (order: Order);
    }
  }
}`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	linkedProgram, _, _, err := linker.Link([]*ir.Program{prog}, nil)
	if err != nil {
		t.Fatalf("link failed: %v", err)
	}

	_, err = LowerProjectWithOptions(linkedProgram, Options{Profile: ProfilePlain, Strict: true})
	if err == nil || !strings.Contains(err.Error(), "requires all referenced types to be declared") {
		t.Fatalf("expected strict unknown type failure, got: %v", err)
	}
}

func TestLowerProject_StrictLowersSingleMethodAdapterWithoutStub(t *testing.T) {
	prog, err := parser.ParseString(`project Demo {
  module orders.domain as domain {
    type OrderId = String where nonEmpty;
  }
  module orders.port as port {
    port OrderRepository {
      fn Save(id: OrderId) -> (ok: Bool);
    }
  }
  module orders.app as usecase {
    usecase CreateOrder {
      input { id: OrderId };
      output { ok: Bool };
      effects [io];
      let saved = Save(id);
      return saved;
    }
    adapter OrdersDb implements orders.port::port:OrderRepository {
      input { id: OrderId };
      output { ok: Bool };
      effects [io];
      return true;
    }
    wiring OrdersWiring {
      bind orders.port::port:OrderRepository -> orders.app::adapter:OrdersDb;
    }
  }
}`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	linkedProgram, _, _, err := linker.Link([]*ir.Program{prog}, nil)
	if err != nil {
		t.Fatalf("link failed: %v", err)
	}

	project, err := LowerProjectWithOptions(linkedProgram, Options{Profile: ProfileSpringBoot, Strict: true})
	if err != nil {
		t.Fatalf("strict lower failed: %v", err)
	}

	outDir := t.TempDir()
	if err := WriteProject(outDir, project); err != nil {
		t.Fatalf("write project failed: %v", err)
	}
	allText := readAllGeneratedText(t, outDir)
	if strings.Contains(allText, "Deterministic adapter stub generated from LIA binding.") {
		t.Fatal("expected strict lower to avoid adapter stubs")
	}
	if strings.Contains(allText, "Placeholders conservadores") {
		t.Fatal("expected strict lower to avoid placeholders")
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

func readAllGeneratedText(t *testing.T, dir string) string {
	t.Helper()
	var contents []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr == nil {
			contents = append(contents, string(data))
		}
		return nil
	})
	return strings.Join(contents, "\n")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}
