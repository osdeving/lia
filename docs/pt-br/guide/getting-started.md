# Getting Started (MVP)

This is the minimal end-to-end run that proves the toolchain works.

## Prereqs

- Go 1.23+ (or current system Go)

## Run the MVP

```bash
# 1) parse .lia -> .liao
go run ./cmd/lia parse examples/order-service/project.lia -o /tmp/order.liao

# 2) validate (use --pack-dir se quiser packs)
go run ./cmd/lia check /tmp/order.liao --pack-dir ./docs/pt-br/spec/packs

# 3) link .liao -> .lial + decision log
go run ./cmd/lia link /tmp/order.liao -o /tmp/order.lial --pack-dir ./docs/pt-br/spec/packs

# 4) lower (stub) -> Java
go run ./cmd/lia lower /tmp/order.lial --target java -o /tmp/order.java
```

## What you should see

- `check` prints warnings for policies not enforced (expected in v0.1) and then `ok`.
- `link` emits `/tmp/order.lial` and `/tmp/order.lial.decision-log.json`.
- `lower` emits a stub file with module count.

## Known limitations in MVP

- Parser é minimalista (subset de project/module/usecase; tipos/ports/adapters ainda não).
- Linker is a stub (no real symbol resolution yet).
- SecurityBaseline rules are not enforced (warnings only).

Se quiser o modelo conceitual completo, leia `docs/pt-br/guide/pipeline.md`.
