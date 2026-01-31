# AGENT.md — Guidelines for AI Agents Working on LIA

> This document defines the rules and best practices for AI agents (LLMs) contributing to the LIA project.

---

## 1. CONTEXT AND MISSION

**LIA (Language Intermediate Assisted)** is a model-first toolchain for AI software generation. Unlike conventional languages, LIA was designed to be:

- **Model-generated**: Syntax and semantics optimized for LLMs.
- **Verifiable**: Determinism and reproducibility by construction.
- **Linkable**: Modules generated in parallel can be merged via linker.

**Your mission as an agent**: Contribute code, tests, and documentation while maintaining the project's reproducibility and quality principles.

---

## 2. MANDATORY PRINCIPLES

### 2.1 Reproducibility First

- **Every code generation** must include a `@gen` block with metadata:
  - `prompt_ref`: Reference to the prompt used.
  - `prompt_hash`: SHA-256 hash of the prompt.
  - `model_id`: Model identification (e.g., `gpt-4o-mini`, `qwen2.5-coder:7b`).
  - `model_params`: Parameters such as `temperature`, `top_p`, `seed`.
  
- **Example of @gen block in LIA**:

  ```lia
  @gen {
    prompt_ref: "p-001",
    prompt_hash: "a3f2b1c...",
    model_id: "qwen2.5-coder:7b",
    model_params: "temperature=0.1"
  }
  module orders.core as domain {
    // ...
  }
  ```

### 2.2 Docs First

- **Before changing code**, update the specification in `docs/en/spec/` and `docs/pt-br/spec/`.
- **Before adding a feature**, document it in the implementation plan.
- **Use ADRs** (Architecture Decision Records) for significant decisions.

### 2.3 Tests are Non-Negotiable

- **All Go code must have unit tests** (`*_test.go`).
- **Minimum Coverage**: 75% for new packages.
- **Linker and Parser**: Must maintain >85% coverage.

---

## 3. CONTRIBUTION WORKFLOW

> **Note**: For full details on Roadmap, RFCs, and Release, refer to [docs/en/PROCESS.md](docs/en/PROCESS.md).

### 3.1 Branch Naming Convention

**Format**: `<type>/<scope>/<short-description>`

#### Branch Types

- `feat/<scope>/<description>`: New features
  - Ex: `feat/parser/add-hole-support`, `feat/linker/multi-candidate-selection`
- `fix/<scope>/<description>`: Bug fixes
  - Ex: `fix/codec/canonical-json-sorting`, `fix/linker/hash-determinism`
- `spec/<scope>/<description>`: Changes to the LIA specification
  - Ex: `spec/grammar/effect-system`, `spec/policies/budget-constraints`
- `docs/<scope>/<description>`: Documentation improvements
  - Ex: `docs/pt-br/guide/getting-started`, `docs/en/guide/cli`
- `test/<scope>/<description>`: Adding/improving tests
  - Ex: `test/parser/edge-cases`, `test/integration/e2e-workflow`
- `refactor/<scope>/<description>`: Refactors without behavioral changes
  - Ex: `refactor/codec/extract-sorting`, `refactor/cli/command-structure`
- `perf/<scope>/<description>`: Performance optimizations
  - Ex: `perf/linker/graph-resolution`, `perf/parser/streaming-mode`

#### Valid Scopes

- `parser`, `linker`, `codec`, `llmgen`, `policy`, `repro`, `cli`, `check`, `lower`
- `spec`, `docs`, `examples`, `packs`

#### Rules

- Use kebab-case (hyphen-separated words)
- Maximum 50 characters total
- Be descriptive but concise
- If related to an issue, use: `fix/linker/issue-123-symbol-collision`

### 3.2 Conventional Commits

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat(parser): add support for hole declarations`
- `fix(linker): resolve duplicate symbols correctly`
- `docs(guide): add examples for pack usage`
- `test(codec): increase coverage to 85%`

### 3.3 Pull Request Checklist

Before creating a PR, **verify**:

- [ ] Specification in `docs/en/spec/` and `docs/pt-br/spec/` is updated (if applicable)
- [ ] Unit tests added/updated
- [ ] `go test ./internal/... -cover` passes without errors
- [ ] Parsed/Linker changes updated in `examples/`
- [ ] AI-generated artifacts include `@gen` block or metadata in commit

---

## 4. GO CODE STRUCTURE

### 4.1 Package Organization

```
internal/
  parser/     → Parsing .lia to IR
  linker/     → Symbol resolution and linking
  codec/      → Canonical serialization and hashing
  llmgen/     → Integration with LLMs (Ollama, etc.)
  policy/     → Constraints enforcement
  repro/      → Reproducibility metadata
```

### 4.2 Code Standards

- **Follow [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)**
- **Typed Errors**: Use `%w` for wrapping
- **Minimal Exports**: Keep interfaces small
- **Canonicalization**: Use `codec.SortAll()` before serializing

### 4.3 Tests

```go
// Naming: Test<FunctionName>_<Scenario>
func TestParseFile_BasicProject(t *testing.T) {
    // Arrange
    content := `project TestProj { ... }`
    tmpfile := createTempFile(t, content)
    
    // Act
    prog, err := ParseFile(tmpfile)
    
    // Assert
    if err != nil {
        t.Fatalf("ParseFile failed: %v", err)
    }
    // ... other assertions
}
```

---

## 5. LIA CODE GENERATION

### 5.1 Using LLMGen

If you are **using LIA to generate LIA** (meta-generation):

```go
provider := llmgen.NewOllamaProvider("http://localhost:11434")
gen := llmgen.NewGenerator(provider, "./prompt-tape.json")

module, meta, err := gen.GenerateLIAModule(ctx, llmgen.ModuleSpec{
    Name:        "user.repository",
    Role:        "adapter",
    Model:       "qwen2.5-coder:7b",
    Temperature: 0.1,
    Context:     "Database access layer for User entity",
})
```

### 5.2 Structured Prompts

When creating prompts to generate LIA:

- **Be explicit about roles**: `"Create a domain module (no IO effects)"`
- **Include constraints**: `"Must not depend on adapters"`
- **Provide context**: Used packs, architecture styles

### 5.3 Post-Generation Validation

After generating LIA code:

1. **Parse**: `lia parse output.lia -o output.liao`
2. **Check**: `lia check output.liao`
3. **Link**: `lia link output.liao -o linked.lial`
4. **Verify hash**: The `.lial` hash must be deterministic

---

## 6. CONSTRAINTS AND BEST PRACTICES

### 6.1 Domain-Driven Design (via Packs)

LIA does not enforce Clean/Hexagonal, but **packs** can:

- `HexCore@1.0.0`: Enforces hexagonal rules via policies
- `LayeredArch@1.0.0`: Enforces traditional layers

**As an agent, respect project policies**:

```bash
lia check mycode.liao --pack-dir ./docs/pt-br/spec/packs  # or ./docs/en/spec/packs
```

### 6.2 Avoid Over-Engineering

- **Budgets**: Some projects have layer/module limits.
- **Simplicity**: If the project is marked "simple", do not generate complex architecture.

### 6.3 Canonical Naming

- **Modules**: `snake_case` or `dotted.hierarchy` (e.g., `orders.core`)
- **Types**: `PascalCase` (e.g., `OrderId`, `PaymentStatus`)
- **Holes**: `hole_<description>` (e.g., `hole_user_repository`)

---

## 7. INTEGRATION WITH OLLAMA

If you are an agent **running locally** and need to test generation:

### 7.1 Setup

```bash
# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Pull code model
ollama pull qwen2.5-coder:7b
```

### 7.2 Configuration

Create `.env.local`:

```bash
LIA_LLM_PROVIDER=ollama
LIA_LLM_BASE_URL=http://localhost:11434
LIA_LLM_MODEL=qwen2.5-coder:7b
LIA_LLM_TEMPERATURE=0.1
```

### 7.3 Tests

```bash
# Run tests with mock (fast, no LLM)
go test ./internal/llmgen -v

# Run tests with real LLM (requires Ollama running)
go test ./internal/llmgen -v -tags=llm_integration
```

---

## 8. PRE-COMMIT CHECKLIST (AGENTS)

Before submitting any change, **verify**:

- [ ] Code compiles: `go build ./...`
- [ ] Tests pass: `go test ./internal/... -cover`
- [ ] Coverage did not decrease (use `go tool cover`)
- [ ] Documentation updated (if applicable)
- [ ] `@gen` block present in AI-generated artifacts
- [ ] Commit message follows Conventional Commits
- [ ] No sensitive info (API keys, secrets) in code

---

## 9. ADDITIONAL RESOURCES

- **LIA Specification**: `docs/pt-br/spec/lia-v0.1.md` (ou `docs/en/spec/lia-v0.1.md`)
- **Manifesto**: `docs/pt-br/LIA_MANIFESTO.md` (ou `docs/en/LIA_MANIFESTO.md`)
- **Testing Guide**: `docs/pt-br/TESTING_GUIDE.md` (ou `docs/en/TESTING_GUIDE.md`)
- **Testing Strategy**: `docs/pt-br/TESTING_STRATEGY.md` (ou `docs/en/TESTING_STRATEGY.md`)
- **Contribution**: `docs/pt-br/CONTRIBUTING.md` (ou `docs/en/CONTRIBUTING.md`)

---

## 10. MODEL-FIRST PHILOSOPHY

**Remember**: LIA is not a language for humans to write by hand. It is designed for:

1. **Models to generate** (you!)
2. **Machines to validate** (linker, checker)
3. **Humans to audit** (via decision logs and prompt tapes)

Your contribution as an agent is to **generate verifiable and reproducible artifacts**, not "pretty" code for human reading.

---

**Last Update**: 2026-01-31
**LIA Version**: 0.1 (pre-alpha)
