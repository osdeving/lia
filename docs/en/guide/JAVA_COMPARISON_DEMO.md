# Java vs LIA Demo

This demo is meant to show something concrete in a presentation:

- a handcrafted Java reference project
- a project requested directly in Java from the model
- the same brief going through `LIA -> Java`

## Repository artifacts

- neutral demo spec: `examples/demo-compare/project-spec.json`
- handcrafted Java reference: `examples/java-reference/orders-service/`

## How to run it

```bash
go run ./cmd/lia demo compare \
  --spec ./examples/demo-compare/project-spec.json \
  --provider openai-compatible \
  --base-url https://api.openai.com \
  --model gpt-4o-mini \
  --temperature 0.1 \
  --out-dir /tmp/java-compare-demo \
  --reference-dir ./examples/java-reference/orders-service \
  --pack-dir ./docs/pt-br/spec/packs
```

## What it writes

- `/tmp/java-compare-demo/COMPARISON.md`
- `/tmp/java-compare-demo/direct-java/`
- `/tmp/java-compare-demo/direct-java.compile.txt`
- `/tmp/java-compare-demo/lia-artifacts/project.lia`
- `/tmp/java-compare-demo/lia-artifacts/prompt-tape.json`
- `/tmp/java-compare-demo/lia-artifacts/project.lial.decision-log.json`
- `/tmp/java-compare-demo/lia-java/`
- `/tmp/java-compare-demo/lia-java.compile.txt`
- `/tmp/java-compare-demo/reference-java.compile.txt`

## How to read the comparison

The point is not “AI writes Java better”.

The point is:

1. the same brief becomes more deterministic when it goes through a typed, modular intermediate form
2. the LIA branch gains auditable artifacts (`project.lia`, tape, decision log)
3. the lower can still emit compilable Java even when the model leaves small semantic gaps

In the validated run for this repository:

- `Reference Java`: compiles
- `Direct Java`: does not compile
- `LIA -> Java`: compiles

## Current limit

- the Java lower still uses conservative placeholders for referenced-but-undeclared types
- Python is still a stub
- there is still no semantic typechecker or full target runtime
