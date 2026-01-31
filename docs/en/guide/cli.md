# CLI Reference (v0.1)

## Commands

- `lia parse <file.lia> -o out.liao`
  - minimal parser; outputs canonical JSON

- `lia check <file.liao|file.lial>`
  - runs local validation + pack rules (subset)

- `lia link <inputs...> -o out.lial --decision-log out.lial.decision-log.json`
  - merges inputs and emits a decision log

- `lia explain <file.lial> --decision-log out.lial.decision-log.json`
  - prints hash + summary

- `lia lower <file.lial> --target java|python -o out.java|out.py`
  - stub lowerers

- `lia replay <file.lial> --tape prompt-tape.json`
  - placeholder for future replay

## Pack loading

Commands that validate or link can load packs:

- default search paths: `./packs` and `./docs/spec/packs` (if present)
- in this repo, packs live under `docs/en/spec/packs` and `docs/pt-br/spec/packs`
- override/add paths: `--pack-dir <path>` (repeatable)

Example:

```bash
lia check out.liao --pack-dir ./docs/en/spec/packs
```
