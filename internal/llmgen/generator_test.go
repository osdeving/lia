package llmgen

import (
	"context"
	"testing"
	"time"
)

// MockProvider is a mock LLM provider for testing.
type MockProvider struct {
	Response string
	Err      error
}

func (m *MockProvider) Name() string {
	return "mock"
}

func (m *MockProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return &GenerateResponse{
		Content:   m.Response,
		Model:     req.Model,
		Timestamp: time.Now(),
	}, nil
}

// TestGenerator_GenerateLIAModule tests basic module generation.
func TestGenerator_GenerateLIAModule(t *testing.T) {
	mock := &MockProvider{
		Response: "module test.module as domain {}",
	}

	gen := NewGenerator(mock, "")

	spec := ModuleSpec{
		Name:        "test.module",
		Role:        "domain",
		Model:       "test-model",
		Temperature: 0.1,
		Context:     "Test context",
	}

	module, meta, err := gen.GenerateLIAModule(context.Background(), spec)
	if err != nil {
		t.Fatalf("GenerateLIAModule failed: %v", err)
	}

	if module.Name != "test.module" {
		t.Errorf("expected module name test.module, got %s", module.Name)
	}

	if module.Role != "domain" {
		t.Errorf("expected role domain, got %s", module.Role)
	}

	if meta == nil {
		t.Fatal("expected non-nil metadata")
	}

	if meta.PromptRef == "" {
		t.Error("expected prompt_ref to be set")
	}

	if meta.PromptHash == "" {
		t.Error("expected prompt_hash to be set")
	}

	if meta.ModelID != "test-model" {
		t.Errorf("expected model_id test-model, got %s", meta.ModelID)
	}
}

// TestGenerator_PromptBuilding tests prompt construction.
func TestGenerator_PromptBuilding(t *testing.T) {
	gen := &Generator{}

	spec := ModuleSpec{
		Name:    "orders.repository",
		Role:    "adapter",
		Context: "Database access for orders",
	}

	prompt := gen.buildPrompt(spec)

	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}

	// Verify prompt contains key information
	if !contains(prompt, "orders.repository") {
		t.Error("prompt should contain module name")
	}

	if !contains(prompt, "adapter") {
		t.Error("prompt should contain role")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
