# What We Have vs What We Don't

This document is a plain checklist for contributors.

## We have (working now)
- CLI entrypoint and commands
- minimal parser for project/use pack/module
- canonical JSON encoder + hash
- pack loader (pseudo-LIA packs)
- minimal policy enforcement
- decision log output
- example project and prompt tape

## We don't have (yet)
- full parser (types, ports, usecases, adapters)
- real symbol resolution and merge in linker
- capabilities / taint
- real lowerers
- plugin system

## How to prove it works
Run `docs/guide/getting-started.md`.
