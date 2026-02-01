---
name: golang-best-practices
description: Best practices for writing idiomatic, performant, and maintainable Go code in the LIA project
version: 1.0.0
tags: [golang, best-practices, architecture, performance]
---

# Golang Best Practices for LIA

<persona>
Act as a Senior Go Developer with expertise in compiler design, DSP systems, and reproducible builds. Your code prioritizes determinism, performance, and maintainability.
</persona>

<project_context>
  Project Name: LIA (Linguagem Intermediária Assistida)
  Objective: A model-first toolchain for generating software via LLMs with reproducibility guarantees.
  Environment: Cross-platform (Linux, macOS, Windows), Go 1.23+
  Target Performance: Sub-second parsing and linking for medium projects (<1000 modules)
</project_context>

<architectural_constraints>

- Pattern: Hexagonal Architecture in Python orchestrator; Go toolchain follows modular design
- Principles: SOLID, deterministic output, canonical representations
- Performance: Minimize allocations in hot paths (parser, linker, codec)
- Reproducibility: All outputs must be deterministic and hashable
</architectural_constraints>

---

## 1. Code Organization

### Package Structure

```
internal/
  parser/      → Parsing .lia to IR (deterministic)
  linker/      → Symbol resolution and linking (reproducible)
  codec/       → Canonical serialization and hashing
  llmgen/      → LLM integration (Ollama, OpenAI-compatible)
  policy/      → Constraint enforcement engine
  repro/       → Reproducibility metadata (@gen blocks)
  check/       → Validation passes
  cli/         → Command-line interface (Cobra)
```

<output_requirements>

  1. Keep `internal/` packages focused and single-purpose
  2. Minimize cross-package dependencies (DAG structure)
  3. Export only necessary types/functions from packages
  4. Use `pkg/api/` for stable public APIs (if needed in future)
</output_requirements>

---

## 2. Naming Conventions

### Variables

- **Short names in small scopes**: `i`, `j`, `k` for loops; `p` for parser/program
- **Descriptive names in large scopes**: `decisionLog`, `promptTape`
- **No stuttering**: `codec.CanonicalJSON`, not `codec.CodecCanonicalJSON`

### Functions

- **Start with verb**: `ParseFile`, `LinkPrograms`, `HashBytes`
- **Boolean returns**: `isValid`, `hasError`, `canMerge`

### Interfaces  

- **Single-method**: suffix with `-er` (`Parser`, `Linker`, `Hasher`)
- **Multi-method**: descriptive name (`SymbolResolver`, `ConstraintEngine`)

---

## 3. Error Handling

### Standard Pattern

```go
func ParseFile(path string) (*ir.Program, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("open file: %w", err)
    }
    defer f.Close()

    prog, err := parseStream(f)
    if err != nil {
        return nil, fmt.Errorf("parse stream: %w", err)
    }
    return prog, nil
}
```

<reflection>
Why `%w` instead of `%v`? The `%w` verb wraps errors, enabling `errors.Is()` and `errors.As()` for proper error chain inspection. This is critical for LIA's reproducibility audit trail.
</reflection>

### Custom Errors

```go
type ParseError struct {
    Line int
    Col  int
    Msg  string
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("parse error at %d:%d: %s", e.Line, e.Col, e.Msg)
}
```

---

## 4. Determinism and Canonicalization

### Sorting Before Serialization

```go
func CanonicalizeProgram(p *ir.Program) ([]byte, error) {
    SortAll(p) // Deterministic ordering

    var buf bytes.Buffer
    enc := json.NewEncoder(&buf)
    enc.SetEscapeHTML(false) // Prevent HTML escaping for determinism
    if err := enc.Encode(p); err != nil {
        return nil, err
    }
    return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

func SortAll(p *ir.Program) {
    sort.Slice(p.Modules, func(i, j int) bool {
        return p.Modules[i].Name < p.Modules[j].Name
    })
    // ... sort all slices recursively
}
```

<task_specification>
Every collection (slice, map keys) must be sorted before hashing or serialization to guarantee deterministic output across platforms and Go versions.
</task_specification>

---

## 5. Testing Patterns

### Table-Driven Tests

```go
func TestParsePackRef(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantName string
        wantVer  string
    }{
        {"with version", "HexCore@1.0.0", "HexCore", "1.0.0"},
        {"without version", "MyPack", "MyPack", ""},
        {"beta version", "foo@2.1.3-beta", "foo", "2.1.3-beta"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            name, ver := parsePackRef(tt.input)
            if name != tt.wantName || ver != tt.wantVer {
                t.Errorf("got (%q, %q), want (%q, %q)",
                    name, ver, tt.wantName, tt.wantVer)
            }
        })
    }
}
```

### Test Helpers

```go
func createTempFile(t *testing.T, content string) string {
    t.Helper() // Mark as helper to skip in stack traces
    tmpdir := t.TempDir()
    tmpfile := filepath.Join(tmpdir, "test.lia")
    if err := os.WriteFile(tmpfile, []byte(content), 0o644); err != nil {
        t.Fatalf("failed to create temp file: %v", err)
    }
    return tmpfile
}
```

---

## 6. Performance Patterns

### Avoid Allocations in Hot Paths

```go
// ❌ Bad: Allocates on every call
func HashString(s string) string {
    return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

// ✅ Good: Reuse buffer
func HashString(s string) string {
    h := sha256.Sum256([]byte(s))
    return hex.EncodeToString(h[:])
}
```

### Pre-allocate Slices

```go
// ✅ Good: Pre-allocate if size is known
modules := make([]ir.Module, 0, len(inputs))
for _, input := range inputs {
    modules = append(modules, parseModule(input))
}
```

---

## 7. Concurrency (Future Considerations)

### Channel-Based Pipeline (for parallel parsing)

```go
func ParseFiles(paths []string) ([]*ir.Program, error) {
    type result struct {
        prog *ir.Program
        err  error
    }

    results := make(chan result, len(paths))
    for _, path := range paths {
        go func(p string) {
            prog, err := ParseFile(p)
            results <- result{prog, err}
        }(path)
    }

    progs := make([]*ir.Program, 0, len(paths))
    for range paths {
        r := <-results
        if r.err != nil {
            return nil, r.err
        }
        progs = append(progs, r.prog)
    }
    return progs, nil
}
```

<reflection>
For LIA v0.1, parallel parsing is not yet implemented. However, the deterministic design (canonical sorting, hashing) allows future parallelization without breaking reproducibility.
</reflection>

---

## 8. Documentation

### Godoc Comments

```go
// ParseFile parses a .lia source file into a minimal IR program.
// It returns an error if the file cannot be read or contains invalid syntax.
//
// Example:
//
// prog, err := parser.ParseFile("project.lia")
// if err != nil {
//     log.Fatal(err)
// }
func ParseFile(path string) (*ir.Program, error) {
    // ...
}
```

### Package-Level Documentation

```go
// Package parser implements a minimal LIA parser for v0.1.
//
// The parser reads .lia source files and produces ir.Program structures.
// It does not yet support the full LIA grammar; only a subset for MVP.
//
// Example usage:
//
// prog, err := parser.ParseFile("input.lia")
package parser
```

---

## 9. Common Pitfalls (Specific to LIA)

### ❌ Don't use `map[string]any` in normative IR

```go
// Bad: Non-deterministic iteration order
type Module struct {
    Metadata map[string]any // ❌
}

// Good: Use structs with defined fields
type Module struct {
    Gen *GenMetadata // ✅ Structured
}
```

### ❌ Don't mutate shared state

```go
// Bad: Global mutable state breaks reproducibility
var globalCache = make(map[string]*ir.Program) // ❌

// Good: Pass context explicitly
func ParseWithCache(path string, cache *ProgramCache) (*ir.Program, error) {
    // ✅
}
```

### ❌ Don't ignore hashing failures

```go
// Bad
hash, _ := codec.HashProgram(prog) // ❌ Silent failure

// Good
hash, err := codec.HashProgram(prog)
if err != nil {
    return fmt.Errorf("hash program: %w", err) // ✅
}
```

---

## 10. Tools and Validation

### Pre-Commit Hooks

```bash
go fmt ./...
go vet ./...
golangci-lint run
go test ./internal/... -cover
```

### Static Analysis

- **golangci-lint**: Use `.golangci.yml` with strict settings
- **gosec**: Security scanning
- **gocyclo**: Cyclomatic complexity check (max 15)

---

## 11. Dependencies

### Minimal External Dependencies

Current approved dependencies:

- `github.com/spf13/cobra`: CLI framework
- `github.com/spf13/pflag`: CLI flags
- Standard library for everything else

<task_specification>
Before adding a new dependency, ask:

1. Can this be implemented in stdlib?
2. Is the dependency maintained and stable?
3. Does it support deterministic builds?
</task_specification>

---

## References

- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [LIA Specification](../../docs/en/spec/LIA_V0_1.md)
