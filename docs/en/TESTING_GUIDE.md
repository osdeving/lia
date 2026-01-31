# LIA Test Execution Guide

## Quick Start

### Run All Unit Tests

```bash
go test ./internal/... -v -cover
```

### Generate Coverage Report

```bash
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Run Specific Package Tests

```bash
# Parser tests
go test ./internal/parser -v

# Linker tests
go test ./internal/linker -v

# Codec tests
go test ./internal/codec -v

# LLM Generator tests (mock only, no real LLM needed)
go test ./internal/llmgen -v
```

## LLM Integration Testing

### Prerequisites

1. Install Ollama: <https://ollama.ai>
2. Pull a code model:

```bash
# Recommended models (from lightest to heaviest):
ollama pull phi3:mini          # 3.8GB - Fastest, ideal for CI/CD
ollama pull qwen2.5-coder:1.5b # 1.5GB - Very light, good for quick tests
ollama pull qwen2.5-coder:7b   # 4.7GB - Balanced quality/speed
ollama pull deepseek-coder:6.7b # 3.8GB - Excellent for code
```

### Run LLM Integration Tests

```bash
# Set environment variables
export LIA_LLM_PROVIDER=ollama
export LIA_LLM_BASE_URL=http://localhost:11434
export LIA_LLM_MODEL=qwen2.5-coder:7b

# Run with integration tag (when implemented)
go test ./internal/llmgen -v -tags=llm_integration
```

### Docker Testing

```bash
# Build test image
docker build -t lia-test -f Dockerfile .

# Run tests in container
docker run --rm lia-test go test ./internal/... -v -cover

# Run with Ollama access (host network mode)
docker run --rm --network=host lia-test go test ./internal/llmgen -v
```

## Current Coverage Status

| Package | Coverage | Status |
|---------|----------|--------|
| parser | 96.4% | ✅ Excellent |
| linker | 78.4% | ✅ Good |
| codec | 43.9% | ⚠️ Needs improvement |
| llmgen | 16.5% | ⚠️ Basic coverage |
| check | 0.0% | ❌ No tests |
| cli | 0.0% | ❌ No tests |
| policy | 0.0% | ❌ No tests |
| repro | 0.0% | ❌ No tests |

## Next Steps

1. Add tests for `policy` package (constraint enforcement)
2. Add tests for `repro` package (reproducibility metadata)
3. Add integration tests for end-to-end workflows
4. Implement real LLM integration tests with Ollama
