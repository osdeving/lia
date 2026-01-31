# Linker (v0.1)

The linker is the heart of LIA. It merges multiple object files into a linked unit.

## Responsibilities

- build the dependency graph (`requires -> provides`)
- filter invalid candidates by hard constraints
- select implementations deterministically (tie-break rules)
- produce a decision log for audit/replay

## Current status

- v0.1 linker is a stub: it merges inputs and emits a minimal decision log
- no real symbol resolution or merge strategies yet

## Decision log

The decision log is canonical JSON with a stable hash. It records:

- chosen candidate
- score and reason
- tie-break information (future)

This is the audit trail used for replay.
