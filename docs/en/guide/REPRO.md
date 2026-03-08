# Reproducibility Model

Reproducibility is a first-class requirement in LIA.

## Provenance (@gen)

Every generated module can carry a `@gen` metadata block with:

- prompt reference + prompt hash
- model ID + params
- context refs (packs, docs)
- tools trace refs
- generator pass name/version

## Prompt tape

A project can store all prompts in a versioned JSON tape:

- maps `prompt_ref` -> `prompt_body`
- records context and params

## Determinism

- canonical JSON serialization
- sorted slices (no maps in normative regions)
- hashes computed on canonical bytes

## Current status (v0.1)

- prompt tape exists as canonical JSON (see `examples/full-pipeline/prompt-tape.json`)
- `check` requires `prompt_hash` and `model_id` when `prompt_ref` is present
- `replay` validates `prompt_ref` and `prompt_hash` against the tape
- `internal/llmgen` writes prompt tape in the canonical project format
