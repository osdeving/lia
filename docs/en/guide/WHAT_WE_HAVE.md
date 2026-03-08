# What We Have vs What We Don't

This document is a plain checklist for contributors.

## We have (working now)

- parser for the supported v0.1 grammar
- canonical `.liao` / `.lial` JSON + hash
- symbol derivation (`provides` / `requires`)
- deterministic linker + decision log
- prompt tape validation through `lia replay`
- canonical prompt tape writing in `internal/llmgen`
- smoke-test fixture under `examples/full-pipeline/`
- real Java lowering from `.lial`
- comparative demo for `direct Java` vs `LIA -> Java`
- internal free-form prompt planner through `lia gen app`
- pack manifests for inference

## We don't have (yet)

- full semantic typechecker
- full policy DSL enforcement
- automatic hole resolution / synthesis
- Python lowering
- plugin runtime

## How to prove it works

Run `docs/en/guide/GETTING_STARTED.md` and `go test ./...`.
