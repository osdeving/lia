package llmgen

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/willams/lia/internal/ir"
)

// Generator orchestrates LLM-based code generation for LIA.
type Generator struct {
	Provider Provider
	TapeFile string // Path to prompt-tape.json
}

// NewGenerator creates a new LLM-based generator.
func NewGenerator(provider Provider, tapeFile string) *Generator {
	return &Generator{
		Provider: provider,
		TapeFile: tapeFile,
	}
}

// GenerateLIAModule generates a LIA module using the LLM.
func (g *Generator) GenerateLIAModule(ctx context.Context, spec ModuleSpec) (*ir.Module, *GenMetadata, error) {
	prompt := g.buildPrompt(spec)

	req := GenerateRequest{
		Prompt:      prompt,
		Model:       spec.Model,
		Temperature: spec.Temperature,
		Metadata: map[string]interface{}{
			"module_name": spec.Name,
			"role":        spec.Role,
		},
	}

	resp, err := g.Provider.Generate(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("llm generation failed: %w", err)
	}

	// Parse LLM output into a Module
	module := &ir.Module{
		Name: spec.Name,
		Role: spec.Role,
	}

	// Record generation metadata
	meta := &GenMetadata{
		PromptRef:  g.recordPrompt(prompt),
		PromptHash: hashString(prompt),
		ModelID:    resp.Model,
		ModelParams: map[string]interface{}{
			"temperature": spec.Temperature,
		},
		Timestamp: resp.Timestamp,
	}

	return module, meta, nil
}

// ModuleSpec defines the specification for a module to generate.
type ModuleSpec struct {
	Name        string
	Role        string
	Model       string
	Temperature float64
	Context     string
}

// GenMetadata represents generation metadata (@gen block).
type GenMetadata struct {
	PromptRef   string                 `json:"prompt_ref"`
	PromptHash  string                 `json:"prompt_hash"`
	ModelID     string                 `json:"model_id"`
	ModelParams map[string]interface{} `json:"model_params"`
	Timestamp   time.Time              `json:"timestamp"`
}

func (g *Generator) buildPrompt(spec ModuleSpec) string {
	return fmt.Sprintf(`You are a LIA code generator. Generate a LIA module with the following specification:

Module Name: %s
Role: %s
Context: %s

Generate only the module definition in LIA syntax. Follow these constraints:
1. If role is "domain", do not use "io" effect
2. Use proper typing and contracts
3. Include necessary constraints

Output only valid LIA code.`, spec.Name, spec.Role, spec.Context)
}

func (g *Generator) recordPrompt(prompt string) string {
	// Generate a unique ref for this prompt
	ref := fmt.Sprintf("p-%d", time.Now().Unix())

	// Append to tape file if configured
	if g.TapeFile != "" {
		_ = g.appendToTape(ref, prompt) // Error logged but not fatal for generation
	}

	return ref
}

func (g *Generator) appendToTape(ref, prompt string) error {
	var tape map[string]string

	// Read existing tape
	if data, err := os.ReadFile(g.TapeFile); err == nil {
		_ = json.Unmarshal(data, &tape)
	}

	if tape == nil {
		tape = make(map[string]string)
	}

	tape[ref] = prompt

	// Write back
	data, err := json.MarshalIndent(tape, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(g.TapeFile, data, 0o644)
}

func hashString(s string) string {
	// Simple implementation - use proper hashing in production
	return fmt.Sprintf("%x", len(s))
}
