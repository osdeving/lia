# LIA System Architecture (v0.1)

LIA is more than a language; it is a system with multiple subsystems.

## Subsystems

1) **Language/IR**
   - syntax (draft spec)
   - canonical AST
   - symbol model

2) **Toolchain**
   - parser
   - validator (check)
   - linker
   - lowerers

3) **Policy engine**
   - packs
   - constraints and preferences

4) **Reproducibility**
   - `@gen` metadata
   - prompt tape
   - decision logs

## Data flow

`.lia` -> `parser` -> `ir.Program` -> `.liao`

`.liao` -> `check` (policies) -> `.liao`

`.liao` + packs -> `link` -> `.lial` + `decision-log`

`.lial` -> `lower` -> target code

## Current reality
- Several subsystems are present only as MVP skeletons.
- The boundaries are already clear, so we can grow each subsystem independently.
