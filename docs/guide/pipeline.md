# Pipeline and Artifacts

LIA is a compiler-like pipeline. Each step produces a deterministic artifact.

## Artifacts
- `.lia`  — source (human or model-generated)
- `.liao` — object file (canonical AST + symbols + constraints + provenance)
- `.liap` — pack (versioned collection of `.liao` + manifest)
- `.lial` — linked unit (resolved, ready to lower)

## Pipeline
1) **Parse**
   - `.lia` -> `ir.Program` -> `.liao` (canonical JSON)

2) **Check**
   - local validation + policy/constraint subset
   - should be deterministic

3) **Link**
   - merge multiple `.liao`
   - apply global constraints
   - emit **decision log** for replay

4) **Lower**
   - `.lial` -> target AST -> emitter
   - in v0.1, this is a stub (placeholders)

## Reproducibility
- hashes are computed from canonical JSON
- prompt tape stores prompt bodies + hashes
- decision log stores linker choices + hash

See `docs/guide/repro.md` for details.
