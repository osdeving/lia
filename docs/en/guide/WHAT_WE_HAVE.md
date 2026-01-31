# What We Have vs What We Don't

This document is a plain checklist for contributors.

## We have (working now)

- CLI entrypoint and commands
- minimal parser for project/module/use pack/repro/tape
- canonical JSON encoder + hash
- pack loader (pseudo-LIA packs)
- minimal policy enforcement
- decision log output
- example project and prompt tape

## We don't have (yet)

- full parser (types, enums, ports, usecases, adapters, wiring, full grammar)
- real symbol resolution and merge in linker
- capabilities / taint
- real lowerers (multi-file codegen, target-specific runtimes)
- plugin system

## How to prove it works

Run `docs/en/guide/GETTING_STARTED.md`.
