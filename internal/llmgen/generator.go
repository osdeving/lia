package llmgen

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
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

const maxGenerationAttempts = 3

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

	resp, module, meta, err := g.generateAndRepair(ctx, spec, req, promptRef, prompt)
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
	return fmt.Sprintf(`You are generating code in LIA v0.1.

Return exactly one LIA module and nothing else.
Do not use markdown fences.
Do not explain the result.
Do not invent syntax outside the supported subset.

Supported top-level module shape:
module <qname> as <role> {
  // allowed declarations depend on the role
}

Supported declarations:
- type Name = Base where predicate;
- enum Name { A, B, C };
- port Name { fn Method(arg: Type) -> (out: Type); }
- usecase Name { input { x: Type }; output { y: Type }; effects [pure|io|tx|emit]; let x = expr; if expr { ... } else { ... } while expr { ... } for item in [1,2] { ... } return expr; }
- adapter Name implements module::port:PortName { input { ... }; output { ... }; effects [io]; return true; }
- wiring Name { bind module::port:PortName -> module::adapter:AdapterName; }
- prefer pref_name: module::port:PortName weight 0.6;
- candidate module::port:PortName score 0.8;
- hole MissingThing: need.some.contract;
- constraint name: raw_expr;

Supported expressions:
- literals: numbers, strings, true, false, lists
- operators: || && == != < <= > >= + - * / %% !
- calls: Save(id);
- member/index access: obj.field, list[0]

Hard rules:
- domain modules should use only type/enum/usecase with effects [pure] when needed
- port modules should declare ports and optional candidates
- usecase modules may declare usecase, adapter, wiring, prefer, hole, constraint
- never emit keywords such as contract, record, class, interface, struct, impl, package, import, match
- every field and parameter must use Name: Type syntax
- every non-built-in type you reference must be declared somewhere in the project; never invent undeclared types
- every statement must end with ';' when required by the grammar
- if role is domain, do not use io

Module Name: %s
Role: %s
Context:
%s

Role examples:
domain:
module orders.domain as domain {
  type OrderId = String where nonEmpty;
  enum Status { NEW, PAID };
}

port:
module orders.port as port {
  port OrderRepository {
    fn Save(id: OrderId) -> (ok: Bool);
  }
}

usecase:
module orders.app as usecase {
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
}

Return only the final LIA module.`, spec.Name, spec.Role, spec.Context)
}

func (g *Generator) generateAndRepair(ctx context.Context, spec ModuleSpec, req GenerateRequest, promptRef, prompt string) (*GenerateResponse, *ir.Module, *GenMetadata, error) {
	currentReq := req
	var lastResp *GenerateResponse
	var lastErr error

	for attempt := 1; attempt <= maxGenerationAttempts; attempt++ {
		resp, err := g.Provider.Generate(ctx, currentReq)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("llm generation failed: %w", err)
		}
		lastResp = resp

		module, parseErr := g.parseModuleResponse(spec, resp.Content)
		if parseErr == nil {
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
			return resp, module, meta, nil
		}

		lastErr = parseErr
		if attempt == maxGenerationAttempts {
			break
		}
		currentReq.Prompt = buildRepairPrompt(prompt, spec, resp.Content, parseErr.Error())
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("unknown generation failure")
	}
	if lastResp != nil && strings.TrimSpace(lastResp.Content) != "" {
		return nil, nil, nil, fmt.Errorf("%w; last response: %s", lastErr, compactText(lastResp.Content, 600))
	}
	return nil, nil, nil, lastErr
}

func buildRepairPrompt(originalPrompt string, spec ModuleSpec, previousOutput, parseErr string) string {
	return fmt.Sprintf(`The previous answer was not valid LIA and failed to parse.

Original task:
%s

Required module name: %s
Required role: %s

Parser error:
%s

Previous invalid output:
%s

Now repair the output.
Return exactly one valid LIA module.
Do not use markdown fences.
Do not include prose.
Do not invent unsupported syntax.
The output must start with "module %s as %s {" or "@gen {" followed by that module.`, originalPrompt, spec.Name, spec.Role, parseErr, previousOutput, spec.Name, spec.Role)
}

func compactText(input string, limit int) string {
	trim := strings.TrimSpace(input)
	if len(trim) <= limit {
		return trim
	}
	return trim[:limit] + "..."
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

	if block := extractFencedModuleBlock(trim); block != "" {
		return block
	}
	if block := extractModuleBlock(trim); block != "" {
		return block
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

func extractFencedModuleBlock(content string) string {
	lines := strings.Split(content, "\n")
	inFence := false
	var block []string

	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			if inFence {
				candidate := strings.TrimSpace(strings.Join(block, "\n"))
				if containsModule(candidate) {
					return candidate
				}
				block = nil
				inFence = false
				continue
			}
			inFence = true
			block = nil
			continue
		}
		if inFence {
			block = append(block, line)
		}
	}
	return ""
}

func extractModuleBlock(content string) string {
	re := regexp.MustCompile(`(?s)(@gen\s*\{.*?\}\s*)?module\s+[A-Za-z_][A-Za-z0-9_:.]*.*`)
	match := re.FindString(content)
	if match == "" {
		return ""
	}
	return strings.TrimSpace(match)
}

func containsModule(content string) bool {
	return strings.Contains(content, "\nmodule ") ||
		strings.HasPrefix(strings.TrimSpace(content), "module ") ||
		strings.Contains(content, "\n@gen") ||
		strings.HasPrefix(strings.TrimSpace(content), "@gen")
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
