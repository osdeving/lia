# LIA Overview

LIA (Linguagem Intermediaria Assistida) is a **model-first intermediate representation** and a **toolchain** for generating software safely and reproducibly.

Think of it as two things working together:
1) **A language/IR** that models software structure in a deterministic, linkable format.
2) **A system** (CLI + passes + linker + reproducibility) that turns LLM output into artifacts you can audit.

## Why it exists
Direct "prompt -> code" generation is fragile:
- it is hard to make architecture consistent across modules,
- it is easy to break boundaries or mix layers,
- it is difficult to reproduce or audit decisions,
- parallel generation is unreliable without stable interfaces.

LIA moves these concerns into a compiler-like pipeline:
- structured object files (`.liao`),
- verification (constraints/policies),
- linking (deterministic selection/merge),
- reproducibility (prompt tape + hashes + decision logs).

## What it is
- A deterministic IR with canonical JSON.
- A toolchain that can parse, validate, link, and lower artifacts.
- A foundation for architecture profiles (hexagonal, layered, etc.) as **packs/policies**, not hardcoded dogma.

## What it is not (yet)
- A complete programming language with full syntax.
- A full compiler to native binaries.
- A complete security engine (policies are still partially enforced).

## Core principles
- **Reproducibility first:** hashes and provenance are required.
- **Determinism by construction:** canonical representation, stable ordering.
- **Modularity:** object files link together safely.
- **Policy-driven architecture:** rules are code, not text.

If you want a 10-minute runnable MVP, go to `docs/en/guide/getting-started.md`.
