# Language Status (v0.1)

This document describes what LIA currently supports in code, what already works end to end, and what is still partial.

## What works end to end

- `lia parse`: `.lia` text to canonical `.liao` JSON.
- `lia check`: structural validation, effects, derived symbols, and a subset of packs/policies.
- `lia link`: deterministic `requires -> provides` resolution with a decision log.
- `lia explain`: hash, module counts, and decision-log summary.
- `lia replay`: validates `.lial` against `prompt-tape.json` using `@gen.prompt_ref` and `@gen.prompt_hash`.
- `internal/llmgen`: writes canonical prompt tapes and parses LIA returned by the provider.
- `lia gen project`: bootstraps a LIA project from `project-spec.json` plus a compatible provider.

## Keywords supported today

### Top level

- `project`
- `pack`
- `module`

### Inside `project`

- `repro`
- `tape`
- `use pack`
- `policy`
- `constraint`

### Inside `module`

- `@gen`
- `as`
- `type`
- `enum`
- `port`
- `fn`
- `method`
- `usecase`
- `adapter`
- `implements`
- `input`
- `output`
- `effects`
- `wiring`
- `bind`
- `prefer`
- `weight`
- `hole`
- `candidate`
- `score`
- `constraint`

### Statements

- `let`
- `return`
- `if`
- `else`
- `while`
- `for`
- `in`
- `break`
- `continue`

### Expressions and operators

- literals: number, string, bool, list
- call: `fn(x)`
- member: `obj.field`
- index: `items[0]`
- operators: `||`, `&&`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `+`, `-`, `*`, `/`, `%`, `!`

## Semantics that are actually implemented

- `provides` and `requires` are derived for `type`, `enum`, `port`, `usecase`, `adapter`, and `wiring`.
- accepted `effects`: `pure`, `io`, `tx`, `emit`.
- `pure` cannot be combined with any other effect.
- `repro=strict` and `repro=pinned` require a project `tape`.
- `repro=strict` rejects unresolved `hole` declarations in `.lial`.
- `replay` validates prompt-ref presence in the tape and checks `prompt_hash`.
- the linker uses `candidate` score, requester preferences, and deterministic tie-breaks.

## What is still partial

- There is no semantic typechecker yet; `type_ref` is still treated as a raw form.
- `policy` and `constraint` are `raw_expr`; only a subset is enforced today.
- `hole` is parsed and validated in `strict`, but there is still no automatic hole resolution.
- `candidate` already affects scoring, but there is no full synthesis/substitution mechanism yet.
- `lower` is still a stub for Java/Python.

## How to test the current state

### Repository automated test suite

```bash
go test ./...
```

### Manual pipeline smoke test

Use the fixture under `examples/full-pipeline/`:

```bash
go run ./cmd/lia parse examples/full-pipeline/project.lia -o /tmp/full-pipeline.liao
go run ./cmd/lia check /tmp/full-pipeline.liao --pack-dir ./docs/en/spec/packs
go run ./cmd/lia link /tmp/full-pipeline.liao -o /tmp/full-pipeline.lial --decision-log /tmp/full-pipeline.decision-log.json --pack-dir ./docs/en/spec/packs
go run ./cmd/lia explain /tmp/full-pipeline.lial --decision-log /tmp/full-pipeline.decision-log.json
go run ./cmd/lia replay /tmp/full-pipeline.lial --tape ./examples/full-pipeline/prompt-tape.json
```

Expected result:

- `check` ends with `ok`
- `link` emits `.lial` and `decision-log.json`
- `replay` reports `modules 3, generated 3, validated 3`

## Using it with AI today

If you ask a model to generate LIA, keep it inside this supported subset and run the validation flow above.

Practical recommendation:

- generate one `module ...` at a time
- include `@gen` with `prompt_ref`, `prompt_hash`, and `model_id`
- always validate with `parse -> check -> link -> replay`
