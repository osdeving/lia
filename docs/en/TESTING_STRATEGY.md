# LIA Testing Strategy & LLM Integration

## Overview

This document outlines the comprehensive testing strategy for the LIA toolchain and the integration with local LLMs for generation testing.

## Current Test Coverage Analysis

### Before Implementation (Baseline)

- **Total Coverage**: 0%
- **Tested Packages**: 0/11
- **Critical Gaps**:
  - `parser`: No tests for file parsing logic
  - `linker`: No tests for symbol resolution
  - `codec`: No tests for serialization/hashing
  - `policy`: No tests for constraint enforcement
  - `repro`: No tests for reproducibility

### After Implementation (Target)

- **Parser**: ~85% coverage (basic parsing, edge cases, error handling)
- **Linker**: ~80% coverage (merging, decision log, hashing)
- **Codec**: ~90% coverage (serialization, canonical JSON, hashing)
- **LLMGen**: ~75% coverage (mock provider, integration tests)

## Test Architecture

### 1. Unit Tests

Located in each `internal/*` package as `*_test.go` files.

**Coverage Areas**:

- Parser: File parsing, pack references, error cases
- Linker: Program merging, decision logging, hash determinism
- Codec: Read/write operations, canonical serialization
- LLMGen: Generation requests, prompt building, metadata tracking

### 2. Integration Tests

Located in `tests/integration/` (to be created).

**Test Scenarios**:

- End-to-end: `.lia` → parse → check → link → `.lial`
- Policy enforcement across multiple modules
- Reproducibility: Same input + same config = same hash

### 3. LLM Generation Tests

Located in `tests/llm_integration/` (to be created).

**Test Types**:

- **Smoke Tests**: Verify LLM connectivity and basic generation
- **Reproducibility Tests**: Same prompt → compare hashes
- **Quality Tests**: Validate generated .lia syntax

## LLM Integration Architecture

### Supported Providers

#### 1. Ollama (Primary - Local)

- **Endpoint**: `http://localhost:11434`
- **Models**: qwen2.5-coder, codellama, deepseek-coder
- **Usage**: Main testing provider for offline development

#### 2. OpenAI-Compatible APIs

- **Endpoint**: Configurable (LMStudio, LocalAI, etc.)
- **Usage**: Alternative local providers

### Configuration

Environment variables for test configuration:

```bash
export LIA_LLM_PROVIDER=ollama           # or "openai-compatible"
export LIA_LLM_BASE_URL=http://localhost:11434
export LIA_LLM_MODEL=qwen2.5-coder:7b
export LIA_LLM_TEMPERATURE=0.1
```

## Test Execution

### Running All Tests

```bash
# Run all unit tests with coverage
go test ./internal/... -v -cover

# Generate coverage report
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Running LLM Integration Tests

```bash
# Requires Ollama running locally
ollama serve &

# Run LLM generation tests
go test ./internal/llmgen -v -tags=llm_integration

# Run with specific model
LIA_LLM_MODEL=deepseek-coder:6.7b go test ./internal/llmgen -v
```

### Docker-based Testing

```bash
# Run tests inside Docker container
docker build -t lia-test -f Dockerfile.test .
docker run --rm lia-test

# Run with Ollama access (network mode)
docker run --rm --network=host lia-test go test ./internal/llmgen -v
```

## Test Data & Fixtures

### Golden Files

Located in `tests/fixtures/golden/`:

- `basic_project.lia`: Minimal valid project
- `hex_example.lia`: Hexagonal architecture example
- `decision_log.json`: Expected linker output

### Prompt Tapes

Located in `tests/fixtures/tapes/`:

- `test_tape.json`: Sample prompt-tape for reproducibility tests

## Reproducibility Testing Protocol

### Phase 1: Deterministic Components

Test that non-LLM components produce identical outputs:

1. Parser: Same `.lia` → Same AST
2. Linker: Same inputs → Same decision log hash
3. Codec: Same IR → Same `.liao` hash

### Phase 2: LLM Replay

Test that LLM generations can be replayed:

1. Generate with prompt P1 → Output O1 (with metadata M1)
2. Store {P1, M1, O1} in tape
3. Replay: Use M1 to regenerate → Output O2
4. Compare: Parse(O1) ≈ Parse(O2) (semantic equivalence)

### Phase 3: Audit Trail

Verify that every `.liao` artifact contains:

- `@gen.prompt_ref` → valid entry in tape
- `@gen.prompt_hash` → matches tape entry
- `@gen.model_id` and `@gen.model_params`

## CI/CD Integration

### GitHub Actions Workflow

```yaml
name: Tests
on: [push, pull_request]
jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - run: go test ./internal/... -v -cover
  
  llm-integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Start Ollama
        run: |
          docker run -d -p 11434:11434 ollama/ollama
          docker exec ollama ollama pull qwen2.5-coder:1.5b
      - run: go test ./internal/llmgen -v -tags=llm_integration
```

## Quality Gates

### Pre-Commit

- Run `go test ./internal/...` (must pass)
- Check coverage: Parser, Linker, Codec > 75%

### Pre-Merge (PR)

- All unit tests pass
- LLM integration tests pass (if Ollama available)
- No decrease in coverage

### Pre-Release

- Full test suite including integration tests
- Reproducibility tests with 3 different models
- Decision log hash verification across platforms

## Future Enhancements

1. **Fuzz Testing**: Use `go-fuzz` for parser robustness
2. **Property-Based Testing**: Use `gopter` for linker invariants
3. **Benchmark Tests**: Track performance regression
4. **Multi-Model Testing**: Test across Ollama, LMStudio, and cloud LLMs
