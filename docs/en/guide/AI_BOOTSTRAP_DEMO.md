# Demo: Direct Prompt vs LIA

This guide shows the core LIA argument: the value is not “using AI”, but forcing AI generation through a validated intermediate structure.

## Before: ask for Java directly

Example prompt:

```text
Create a Java order microservice with use cases, repository, HTTP adapter, and persistence.
```

Problems:

- the output is already coupled to the target language
- architecture stays implicit in the prompt
- there is no intermediate artifact to validate
- there is no `decision-log`
- there is no `prompt-tape` tied to the final code

## After: ask for a LIA project

You provide a `project-spec.json` and let the AI generate LIA modules.

Example spec:

- [project-spec.json](/home/willams/LIA/lia/examples/ai-bootstrap/project-spec.json)

With Ollama:

```bash
go run ./cmd/lia gen project \
  --spec ./examples/ai-bootstrap/project-spec.json \
  --provider ollama \
  --model qwen2.5-coder:7b \
  --base-url http://localhost:11434 \
  --out-dir /tmp/ai-bootstrap \
  --pack-dir ./docs/en/spec/packs
```

With an OpenAI-compatible API:

```bash
go run ./cmd/lia gen project \
  --spec ./examples/ai-bootstrap/project-spec.json \
  --provider openai-compatible \
  --base-url http://localhost:8000 \
  --api-key-env OPENAI_API_KEY \
  --model gpt-4o-mini \
  --out-dir /tmp/ai-bootstrap \
  --pack-dir ./docs/en/spec/packs
```

## What the command generates

- `/tmp/ai-bootstrap/project.lia`
- `/tmp/ai-bootstrap/project.liao`
- `/tmp/ai-bootstrap/project.lial`
- `/tmp/ai-bootstrap/project.lial.decision-log.json`
- `/tmp/ai-bootstrap/prompt-tape.json`

The command also runs:

- parse
- check
- link
- replay

## What is concretely different

### Direct Java output

- one opaque response
- weak separation between domain, contract, and implementation
- hard to compare two generations semantically
- poor auditability

### LIA output

- explicit modules by role (`domain`, `port`, `usecase`)
- per-module `@gen`
- `prompt-tape.json` with hashes
- canonical `project.liao` and `project.lial`
- `decision-log.json` showing linker choices
- `replay` validating prompt/artifact consistency

## How to position this in a demo

The right message is not “AI generated code”.

The right message is:

1. AI generated an intermediate representation with explicit architectural roles
2. the toolchain validated that representation before any lowering
3. the final build became auditable, replayable, and comparable

## Current limit

The command already bootstraps a LIA project from a structured spec and now there is a real Java lower.

So the correct demo today can be:

- **prompt/spec -> validated LIA project**
- **prompt/spec -> LIA -> compilable Java project**

Not yet:

- **prompt/spec -> production-ready final Java application with semantic typing, runtime conventions, and full policy enforcement**
