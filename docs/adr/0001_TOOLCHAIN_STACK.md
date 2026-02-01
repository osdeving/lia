# ADR 0001 — Toolchain Stack (Commodity vs Core)

## Status
Accepted (initial direction)

## Context

LIA is a model‑first toolchain. Some parts are **commodity** (parsing, policy engines, canonical JSON), while others are **core IP** (IR, deterministic linker, reproducibility model, packs).
We want speed and reliability by adopting proven libraries where the domain is not the differentiator.

## Decision

Adopt a **commodity‑first** stack for non‑core concerns and keep core semantics in‑house.

### Commodity stack (adopt)

- **Parser (v0.1–v0.2)**: Participle (MIT) for rapid grammar iteration in Go.
- **Syntax tooling (DX)**: Tree‑sitter (MIT) for highlighting/indent/folding (future LSP).
- **Canonical JSON**: JCS (RFC 8785); Go impl: `gowebpki/jcs` (Apache‑2.0).
- **Graph algorithms**: Gonum (BSD‑3) for DAG/SCC/toposort in the linker.
- **Constraints**: CUE (Apache‑2.0) for structural validation.
- **Policy engine**: OPA/Rego (Apache‑2.0) for security/compliance gates.
- **Expression DSL**: CEL (Apache‑2.0) for safe, non‑Turing expressions.
- **Plugins/passes**: WASM + wazero (Apache‑2.0) for sandboxed plugins.
- **LLM interface**: MCP (MIT) with Go SDK for elicitation.
- **CLI**: Cobra (Apache‑2.0) (already adopted).
- **Release**: GoReleaser (MIT) (already adopted).

### Core (keep in‑house)

- **IR schema and invariants** (types, symbols, constraints, provenance).
- **Deterministic linker** (selection, scoring, decision log).
- **Reproducibility model** (tapes, hashing, provenance).
- **Packs and architectural/security semantics**.

## Consequences

- Faster time‑to‑feature for parser/tooling without sacrificing control of core semantics.
- Clear separation: commodity dependencies can be swapped if they conflict with determinism or licensing.
- Toolchain upgrades are staged by milestone (v0.1/v0.2/v0.3).

## Alternatives considered

- **Hand‑rolled everything**: rejected due to slow iteration and higher bug surface.
- **Monolithic framework**: rejected to avoid lock‑in and to keep core logic explicit.

## Milestone mapping

- **v0.1**: Participle, JCS, Gonum (basic usage), Cobra.
- **v0.2**: CUE, OPA, CEL, MCP.
- **v0.3**: wazero plugins + tree‑sitter/LSP.
