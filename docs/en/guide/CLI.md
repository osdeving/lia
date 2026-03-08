# CLI Reference (v0.1)

## Commands

- `lia parse <file.lia> -o out.liao`
  - parses the supported v0.1 grammar and writes canonical JSON

- `lia check <file.liao|file.lial>`
  - runs structural, effects, reproducibility, and subset pack validation

- `lia link <inputs...> -o out.lial --decision-log out.lial.decision-log.json`
  - resolves symbols deterministically and emits a decision log

- `lia explain <file.lial> --decision-log out.lial.decision-log.json`
  - prints hash + summary

- `lia lower <file.lial> --target java --out-dir out-dir`
  - emits a multi-file Java project from `.lial`

- `lia lower <file.lial> --target python -o out.py`
  - Python lowering is still a stub

- `lia replay <file.lial> --tape prompt-tape.json`
  - validates `@gen.prompt_ref` and `@gen.prompt_hash` against the tape

- `lia gen project --spec project-spec.json --provider ollama|openai-compatible --model <model> --out-dir <dir>`
  - uses AI to generate a LIA project from a JSON spec and already runs parse/check/link/replay

- `lia demo compare --spec project-spec.json --provider ollama|openai-compatible --model <model> --out-dir <dir>`
  - generates `direct Java` and `LIA -> Java`, compiles both branches, and writes `COMPARISON.md`

## Pack loading

Commands that validate or link can load packs:

- default search paths: `./packs` and `./docs/spec/packs` (if present)
- in this repo, packs live under `docs/en/spec/packs` and `docs/pt-br/spec/packs`
- override/add paths: `--pack-dir <path>` (repeatable)

Example:

```bash
lia check out.liao --pack-dir ./docs/en/spec/packs
```
