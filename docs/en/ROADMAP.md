# Roadmap / Status — Language Completeness

This doc is the implementation plan for making LIA a more complete language/IR. It is **docs‑first** and should be updated before code changes.

## Principles

- Spec first (update `docs/en/spec` + `docs/pt-br/spec` before code).
- Determinism and reproducibility are non‑negotiable.
- Small, verifiable milestones with explicit acceptance criteria.

## Related ADRs

- `docs/adr/0001_TOOLCHAIN_STACK.md` — Toolchain stack (commodity vs core)

## Current state (v0.1 on develop)

- CLI: parse/check/link/explain/lower (stubs in parts).
- Parser: **very small subset** (`project`, `module`, `use pack`, `repro`, `tape`).
- IR structs + canonical JSON + hashing.
- Pack loader (pseudo‑LIA packs) + minimal policy engine.
- Linker: stub (no graph resolution/selection).
- Lowerers: stubs.

## What “more complete” means

A more complete LIA should support:

- A **full core grammar** (types, enums, ports, usecases, adapters, wiring, constraints, preferences, effects).
- A **semantic layer** (symbols, requires/provides, effect validation).
- A **deterministic linker** (graph resolution, selection, decision log).
- **Enforced policies** (roles/effects/deps; then capabilities/taint).
- **At least one real lowering target** with runtime conventions.

## Milestones (phased)

### M1 — Core grammar + AST

**Scope**
- Add grammar for: types, enums, ports, usecases, adapters, wiring, constraints, preferences, effects.
- Define minimal expression/statement grammar for usecase bodies (if imperative).
- Expand IR to represent these constructs.

**Done when**
- Parser accepts the canonical examples in `examples/`.
- Spec EBNF is updated in EN/PT.
- Unit tests cover parser edge cases.

### M2 — Semantic IR (symbols + effects)

**Scope**
- `requires/provides` generated from modules.
- Effect declarations validated against roles.
- Canonicalization rules updated as needed.

**Done when**
- `lia check` reports precise diagnostics for missing/invalid symbols.
- Deterministic hashes for identical inputs.

### M3 — Deterministic linker

**Scope**
- Dependency graph construction.
- Collision detection and candidate selection.
- Fixed tie‑break rules and decision log details.

**Done when**
- Linker resolves multiple `.liao` into a stable `.lial` with reproducible decision log.
- Tests for selection + tie‑break rules.

### M4 — Policy enforcement (profiles)

**Scope**
- Enforce role/effect/dependency constraints from packs.
- Budgets (max modules, max deps) enforced.

**Done when**
- Violations are hard errors in `lia check/link`.
- Pack constraints are applied consistently.

### M5 — First real lowering target

**Scope**
- One target (Python or Java) with executable output and minimal runtime conventions.
- Mapping for usecases, ports/adapters, and basic IO.

**Done when**
- `lia lower` emits runnable output for a non‑trivial example.
- Golden tests validate codegen output.

### M6 — Tooling + reproducibility hardening

**Scope**
- Strict `@gen` + prompt tape validation in `repro=strict`.
- `lia replay` behavior defined and enforced.

**Done when**
- Rebuilds are reproducible under `repro=strict`.
- Decision log + prompt tape fully traceable.

## Out of scope (for now)

- Full formal verification of business rules.
- Marketplace/registry for packs.
- External plugin runtime (WASM) for passes.

## Immediate next steps

1) Expand the grammar and parser (M1).
2) Introduce symbol table + requires/provides (M2).
3) Deterministic linker selection (M3).
