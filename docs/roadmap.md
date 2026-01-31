# Roadmap / Status

This doc describes what exists today vs what is planned.

## Exists (MVP)
- CLI with parse/check/link/explain/lower
- IR structs + canonical JSON + hashing
- pack loader (pseudo-LIA) and minimal policy engine
- reproducibility scaffolding (`@gen`, prompt tape)
- example project and prompt tape

## Partially implemented
- policy enforcement (subset only)
- linker (stub, no graph resolution)
- parser (very small subset)
- lowerers (stubs)

## Not implemented yet
- full grammar and parser
- full symbol table and `requires/provides`
- scoring + tie-break selection
- synth/transform passes
- capabilities and taint/dataflow
- real lowering to Java/Python
- pack registry / semver resolution

## Immediate next milestones
1) parse `provides/requires` + effects
2) build dependency graph and enforce role deps
3) implement deterministic selection + decision log detail
4) add first real lowering target
