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

func TestCLIDemoCompare_OpenAICompatible(t *testing.T) {
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
		content := generatedComparisonResponse(prompt)
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

	repoRoot := mustRepoRoot(t)
	specPath := filepath.Join(repoRoot, "examples", "demo-compare", "project-spec.json")
	packDir := filepath.Join(repoRoot, "docs", "pt-br", "spec", "packs")
	referenceDir := filepath.Join(repoRoot, "examples", "java-reference", "orders-service")
	outDir := filepath.Join(t.TempDir(), "compare")

	out, err := executeCLI(
		"demo", "compare",
		"--spec", specPath,
		"--provider", "openai-compatible",
		"--base-url", server.URL,
		"--model", "mock-model",
		"--compare-profile", "plain",
		"--lia-java-profiles", "plain",
		"--out-dir", outDir,
		"--reference-dir", referenceDir,
		"--pack-dir", packDir,
	)
	if err != nil {
		t.Fatalf("demo compare failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "comparison report:") {
		t.Fatalf("unexpected output: %s", out)
	}

	for _, path := range []string{
		filepath.Join(outDir, "COMPARISON.md"),
		filepath.Join(outDir, "direct-java", "pom.xml"),
		filepath.Join(outDir, "lia-artifacts", "project.lia"),
		filepath.Join(outDir, "lia-artifacts", "project.lial.decision-log.json"),
		filepath.Join(outDir, "lia-java-plain", "pom.xml"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected artifact %s: %v", path, err)
		}
	}
}

func generatedComparisonResponse(prompt string) string {
	if strings.Contains(prompt, "Generate a plain Java 21 project directly") {
		return `=== FILE: pom.xml ===
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>demo.compare</groupId>
  <artifactId>orders-comparison-demo</artifactId>
  <version>1.0.0-SNAPSHOT</version>
</project>
=== FILE: src/main/java/demo/compare/orderscomparisondemo/domain/OrderId.java ===
package demo.compare.orderscomparisondemo.domain;

public record OrderId(String value) {
}
=== FILE: src/main/java/demo/compare/orderscomparisondemo/domain/OrderStatus.java ===
package demo.compare.orderscomparisondemo.domain;

public enum OrderStatus {
  NEW,
  PAID
}
=== FILE: src/main/java/demo/compare/orderscomparisondemo/port/OrderRepository.java ===
package demo.compare.orderscomparisondemo.port;

import demo.compare.orderscomparisondemo.domain.OrderId;

public interface OrderRepository {
  boolean save(OrderId id);
}
=== FILE: src/main/java/demo/compare/orderscomparisondemo/app/CreateOrderUseCase.java ===
package demo.compare.orderscomparisondemo.app;

import demo.compare.orderscomparisondemo.domain.OrderId;
import demo.compare.orderscomparisondemo.port.OrderRepository;

public final class CreateOrderUseCase {
  private final OrderRepository orderRepository;

  public CreateOrderUseCase(OrderRepository orderRepository) {
    this.orderRepository = orderRepository;
  }

  public boolean execute(OrderId id) {
    return orderRepository.save(id);
  }
}
=== FILE: src/main/java/demo/compare/orderscomparisondemo/bootstrap/OrdersWiring.java ===
package demo.compare.orderscomparisondemo.bootstrap;

import demo.compare.orderscomparisondemo.app.CreateOrderUseCase;
import demo.compare.orderscomparisondemo.domain.OrderId;
import demo.compare.orderscomparisondemo.port.OrderRepository;

public final class OrdersWiring {
  private OrdersWiring() {
  }

  public static CreateOrderUseCase createUseCase() {
    OrderRepository repo = new OrderRepository() {
      @Override
      public boolean save(OrderId id) {
        return true;
      }
    };
    return new CreateOrderUseCase(repo);
  }
}
`
	}
	return generatedModuleFromPrompt(prompt)
}
