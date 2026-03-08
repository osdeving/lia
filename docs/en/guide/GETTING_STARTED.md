# Getting Started (MVP)

This is the smallest flow that currently proves the toolchain without depending on `lower`.

## Prereqs

- Go 1.23+

## Run the flow

```bash
# 1) parse .lia -> .liao
go run ./cmd/lia parse examples/full-pipeline/project.lia -o /tmp/full-pipeline.liao

# 2) validate
go run ./cmd/lia check /tmp/full-pipeline.liao --pack-dir ./docs/en/spec/packs

# 3) link .liao -> .lial + decision log
go run ./cmd/lia link /tmp/full-pipeline.liao -o /tmp/full-pipeline.lial --decision-log /tmp/full-pipeline.decision-log.json --pack-dir ./docs/en/spec/packs

# 4) explain
go run ./cmd/lia explain /tmp/full-pipeline.lial --decision-log /tmp/full-pipeline.decision-log.json

# 5) replay
go run ./cmd/lia replay /tmp/full-pipeline.lial --tape ./examples/full-pipeline/prompt-tape.json
```

## What you should see

- `check` ends with `ok`
- `link` writes `/tmp/full-pipeline.lial`
- `explain` prints the hash and decision-log summary
- `replay` reports `modules 3, generated 3, validated 3`

## Current limitations

- no full semantic typechecker
- policy DSL is still partial
- `hole` is not automatically resolved yet
- `lower` is still a stub

For the full status of the implemented language, read `docs/en/guide/LANGUAGE_STATUS.md`.
