# Changelog

All notable changes to the LIA Toolchain project will be documented in this file.

## [0.1.0-dev] - 2026-03-15

### Added

- **Language Expressiveness (T2)**
  - Centralized Keyword Registry for robust lexical structure (`internal/parser/keywords.go`).
  - Error-Tolerant parsing mode (`ParseWithRecovery`) capable of salvaging partial ASTs gracefully during syntax errors.
  - LLM Syntax Normalization pre-pass (`NormalizeLLMOutput`) designed to clean raw AI inputs by injecting semicolons and stripping raw markdown fences.
  - Native parsing and IR generation for `RecordDecl` constructs.

- **Backend Architecture Decoupling (T1)**
  - Comprehensive pluggable targeting architecture via the `lower.Backend` interface.
  - Reusable foundational utilities for multiple backends (identifier sanitization, type mapping).
  - Adopted existing Java logic to the interface (`internal/lower/java/backend.go`).
  - Implemented a complete, operational Python generator translating `Record`, `Enum`, `Type`, `Port` and `Usecase` declarations into frozen Python dataclasses, Enum objects, ABC interfaces and concrete classes.

- **Type System Expansion (T6)**
  - Fully decoupled backend Record representations.
  - Generates exact Java 16+ `record` type artifacts for LIA `RecordDecl` entries mapped back directly via semantic parsing.

- **Syntax & Grammar Formalization (T3)**
  - Detailed EBNF contextual standards documented within `docs/en/spec/GRAMMAR.ebnf`.
  - Comprehensive source-to-IR line-column positional tracker for precise debugging diagnostics map.
  - Built-in fuzz testing integration (`fuzz_test.go`) covering dynamic payload crashes.

- **Passes & Transformations (T5)**
  - Strict sequential pipeline registry framework (`Runner`) inside `internal/passes/runner.go`.
  - Checkers and solvers: `VerifyHolesPass`, `SynthFillHolesPass`, and `TransformDeduplicatePass`.

- **Developer Experience & Tooling (T7/T9)**
  - Generic Language Server stub binary available at `cmd/lia-lsp/main.go`.
  - Live interactive parse loop via the CLI (`lia repl`) to instantly trace block failures natively.
  - Automatic LLM output formatting commands (`lia fmt`) applied as standard CLI tooling.
  
### Changed

- **CLI Separation (T8)**
  - CLI application execution fully segregated from Cobra API flags via deterministic option structs across `lower.go`, `parse.go`, `link.go` and `check.go` routines.
  - Modularized `ast.go` pipeline boundaries by extracting IR generation algorithms to an isolated `convert.go` scope.

### Improved

- **Testing Coverage (T4)**
  - Implemented Golden file payload evaluations for AST regressions inside `parser_golden_test.go`.
  - Scaled module engine test coverages up to 100%, handling empty versions, missing bounds, null packages and edge case profiles thoroughly structurally.
