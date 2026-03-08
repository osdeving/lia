package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIGenProject_OpenAICompatible(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		prompt := req.Messages[0].Content
		content := generatedModuleFromPrompt(prompt)
		resp := map[string]any{
			"model": "mock-model",
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": content,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	specPath := filepath.Join(t.TempDir(), "project-spec.json")
	spec := `{
  "name": "AIDemo",
  "brief": "Order service bootstrap using LIA as an intermediate language.",
  "repro": "pinned",
  "packs": [
    { "name": "ArchBaseline", "version": "0.1.0" }
  ],
  "modules": [
    { "name": "orders.domain", "role": "domain", "context": "Define OrderId and Status." },
    { "name": "orders.port", "role": "port", "context": "Expose OrderRepository contract using OrderId." },
    { "name": "orders.app", "role": "usecase", "context": "CreateOrder usecase, adapter implementing OrderRepository, and wiring." }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	outDir := filepath.Join(t.TempDir(), "generated")
	packDir := filepath.Join(mustRepoRoot(t), "docs", "pt-br", "spec", "packs")
	out, err := executeCLI(
		"gen", "project",
		"--spec", specPath,
		"--provider", "openai-compatible",
		"--base-url", server.URL,
		"--model", "mock-model",
		"--out-dir", outDir,
		"--pack-dir", packDir,
	)
	if err != nil {
		t.Fatalf("gen project failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "replay ok: modules 3, generated 3, validated 3") {
		t.Fatalf("unexpected output: %s", out)
	}

	for _, path := range []string{
		filepath.Join(outDir, "project.lia"),
		filepath.Join(outDir, "project.liao"),
		filepath.Join(outDir, "project.lial"),
		filepath.Join(outDir, "project.lial.decision-log.json"),
		filepath.Join(outDir, "prompt-tape.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated artifact %s: %v", path, err)
		}
	}
}

func generatedModuleFromPrompt(prompt string) string {
	switch {
	case strings.Contains(prompt, "Module Name: orders.domain"):
		return `module orders.domain as domain {
  type OrderId = String where nonEmpty;
  enum Status { NEW, PAID };
}`
	case strings.Contains(prompt, "Module Name: orders.port"):
		return `module orders.port as port {
  port OrderRepository {
    fn Save(id: OrderId) -> (ok: Bool);
  }
  candidate orders.port::port:OrderRepository score 0.8;
}`
	default:
		return `module orders.app as usecase {
  usecase CreateOrder {
    input { id: OrderId };
    output { ok: Bool };
    effects [io];
    return true;
  }
  adapter OrdersDb implements orders.port::port:OrderRepository {
    effects [io];
    return true;
  }
  wiring OrdersWiring {
    bind orders.port::port:OrderRepository -> orders.app::adapter:OrdersDb;
  }
  prefer repo_preference: orders.port::port:OrderRepository weight 0.6;
}`
	}
}
