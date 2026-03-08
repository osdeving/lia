package llmgen

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/parser"
	"github.com/willams/lia/internal/repro"
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
	asset, err := g.GenerateLIAModuleAsset(ctx, spec)
	if err != nil {
		return nil, nil, err
	}
	return asset.Module, asset.Meta, nil
}

// GenerateLIAModuleAsset generates a source module plus parsed metadata.
func (g *Generator) GenerateLIAModuleAsset(ctx context.Context, spec ModuleSpec) (*GeneratedModule, error) {
	prompt := g.buildPrompt(spec)
	promptRef, err := g.recordPrompt(prompt, spec)
	if err != nil {
		return nil, fmt.Errorf("record prompt: %w", err)
	}

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
		return nil, fmt.Errorf("llm generation failed: %w", err)
	}

	providerModel := resp.Model
	if providerModel == "" {
		providerModel = spec.Model
	}
	meta := &GenMetadata{
		PromptRef:  promptRef,
		PromptHash: repro.ComputePromptHash(prompt),
		ModelID:    providerModel,
		ModelParams: map[string]interface{}{
			"temperature": spec.Temperature,
		},
		Timestamp: resp.Timestamp,
	}

	module, err := g.parseModuleResponse(spec, resp.Content)
	if err != nil {
		return nil, err
	}
	module.Gen = buildIRGenMeta(meta, spec)

	return &GeneratedModule{
		Source: normalizedModuleSource(resp.Content),
		Module: module,
		Meta:   meta,
	}, nil
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

func (g *Generator) recordPrompt(prompt string, spec ModuleSpec) (string, error) {
	if g.TapeFile == "" {
		return fmt.Sprintf("p-%d", time.Now().UTC().UnixNano()), nil
	}
	return g.appendToTape(prompt, spec)
}

func (g *Generator) appendToTape(prompt string, spec ModuleSpec) (string, error) {
	tape := &repro.PromptTape{Version: "0.1"}
	if g.TapeFile != "" {
		loaded, err := repro.LoadTape(g.TapeFile)
		switch {
		case err == nil:
			tape = loaded
		case !errors.Is(err, os.ErrNotExist):
			return "", err
		}
	}

	hash := repro.ComputePromptHash(prompt)
	for _, entry := range tape.Prompts {
		if entry.Hash == hash && entry.Body == prompt {
			return entry.Ref, nil
		}
	}

	ref := nextPromptRef(tape)
	entry := repro.PromptEntry{
		Ref:  ref,
		Body: prompt,
		Hash: hash,
	}
	if spec.Context != "" {
		entry.ContextRefs = []string{spec.Context}
	}
	if spec.Temperature != 0 {
		entry.Params = append(entry.Params, ir.KV{
			Key:   "temperature",
			Value: strconv.FormatFloat(spec.Temperature, 'f', -1, 64),
		})
	}
	if err := repro.SaveTape(g.TapeFile, tapeWithEntry(tape, entry)); err != nil {
		return "", err
	}
	return ref, nil
}

func (g *Generator) parseModuleResponse(spec ModuleSpec, content string) (*ir.Module, error) {
	source := stripCodeFence(content)
	prog, err := parser.ParseString(source)
	if err != nil {
		return nil, fmt.Errorf("parse generated LIA: %w", err)
	}
	if len(prog.Modules) != 1 {
		return nil, fmt.Errorf("expected exactly 1 generated module, got %d", len(prog.Modules))
	}

	mod := prog.Modules[0]
	if spec.Name != "" && mod.Name != spec.Name {
		return nil, fmt.Errorf("generated module name mismatch: got %s want %s", mod.Name, spec.Name)
	}
	if spec.Role != "" && mod.Role != spec.Role {
		return nil, fmt.Errorf("generated module role mismatch: got %s want %s", mod.Role, spec.Role)
	}
	return &mod, nil
}

func stripCodeFence(content string) string {
	trim := strings.TrimSpace(content)
	if !strings.HasPrefix(trim, "```") {
		return trim
	}
	lines := strings.Split(trim, "\n")
	if len(lines) < 3 {
		return trim
	}
	if strings.HasPrefix(lines[0], "```") && strings.HasPrefix(lines[len(lines)-1], "```") {
		return strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
	}
	return trim
}

func normalizedModuleSource(content string) string {
	source := stripCodeFence(content)
	trim := strings.TrimSpace(source)
	if strings.HasPrefix(trim, "@gen") {
		if idx := strings.Index(trim, "module "); idx >= 0 {
			return strings.TrimSpace(trim[idx:])
		}
	}
	return trim
}

func buildIRGenMeta(meta *GenMetadata, spec ModuleSpec) *ir.GenMeta {
	if meta == nil {
		return nil
	}
	out := &ir.GenMeta{
		PromptRef:  meta.PromptRef,
		PromptHash: meta.PromptHash,
		ModelID:    meta.ModelID,
	}
	if !meta.Timestamp.IsZero() {
		out.Timestamp = meta.Timestamp.UTC().Format(time.RFC3339Nano)
	}
	if spec.Context != "" {
		out.ContextRefs = []string{spec.Context}
	}
	if spec.Temperature != 0 {
		out.ModelParams = append(out.ModelParams, ir.KV{
			Key:   "temperature",
			Value: strconv.FormatFloat(spec.Temperature, 'f', -1, 64),
		})
	}
	return out
}

func tapeWithEntry(tape *repro.PromptTape, entry repro.PromptEntry) *repro.PromptTape {
	if tape == nil {
		tape = &repro.PromptTape{Version: "0.1"}
	}
	if tape.Version == "" {
		tape.Version = "0.1"
	}
	tape.Prompts = append(tape.Prompts, entry)
	return tape
}

func nextPromptRef(tape *repro.PromptTape) string {
	maxID := 0
	for _, entry := range tape.Prompts {
		if !strings.HasPrefix(entry.Ref, "p-") {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(entry.Ref, "p-"))
		if err != nil {
			continue
		}
		if n > maxID {
			maxID = n
		}
	}
	return fmt.Sprintf("p-%03d", maxID+1)
}
