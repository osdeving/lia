# ADR 0001: Toolchain Stack Choices (Parser, Canonicalization, Policy, Plugins)

- Status: proposed (playground)
- Date: 2026-01-31

## Context
LIA is a language + toolchain with strong determinism, reproducibility, and policy enforcement. We want to avoid reinventing commodity components and focus on the core semantics: IR, linker determinism, reproducibility, and policy enforcement.

## Decision
We adopt the following stack (phased):

**Parsing / AST**
- Use **Participle** (MIT) for rapid Go-native parser development.
- Maintain a **Tree-sitter** grammar in parallel for editor tooling (highlighting, indent, future LSP).

**Canonicalization / Hashing**
- Use **JCS (RFC 8785)** for JSON canonicalization.
- Hash with SHA-256 (Go stdlib).

**Linker Graph / Algorithms**
- Use **Gonum** for graph algorithms (DAG/SCC/toposort).
- Keep selection/merge logic and decision log in-house.

**Constraints / Policies**
- Use **CUE** for structural validation and constraints.
- Use **OPA/Rego** later for richer security/compliance policies.
- Use **CEL** later for lightweight embedded expressions when stable.

**Plugins / Passes**
- Use **WASM + wazero** for sandboxed plugins and passes.
- Define a stable ABI for analyze/rewrite/score/emit.

**LLM Interface**
- Implement **MCP** server so the compiler/linker can pause for elicitation.

## Rationale
- Participle accelerates early compiler development.
- Tree-sitter enables IDE support without coupling to the compiler.
- JCS avoids custom canonicalization bugs.
- Gonum provides robust graph algorithms.
- CUE/OPA/CEL allow policy evolution without baking everything into LIA.
- WASM enables cross-platform, sandboxed extension ecosystem.
- MCP provides a standard path for interactive generation and elicitation.

## Consequences
- Initial complexity is front-loaded in tooling choices but reduces long-term maintenance.
- The LIA core remains small and deterministic.
- Policy evolution can happen without language changes.

## Implementation Plan (phased)
1) Participle + AST; JCS canonical JSON; minimal linker.
2) Gonum integration for dependency graph.
3) CUE-based structural validation.
4) MCP server + elicitation hooks.
5) OPA/Rego for security policies.
6) WASM plugin ABI + registry.

## Status Notes
- This ADR is **playground-only** until validated in develop.
