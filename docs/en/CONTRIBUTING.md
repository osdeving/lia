# Contribution and Workflow Guide

Welcome to the **LIA** development. As this is a critical AI infrastructure project, we follow rigorous engineering standards.

## 1. Development Cycle

1. **Branching**: Create branches from `develop`.

- `feat/...` for new features.
- `fix/...` for fixes.
- `spec/...` for changes to the LIA specification.

1. **Commits**: Use [Conventional Commits](https://www.conventionalcommits.org/v1.0.0/).

- Ex: `feat(linker): add support for selective symbol merging`

1. **Testing**: Every PR must pass `go test ./...`.

1. **Linting**: Code must pass `golangci-lint` (check with `golangci-lint run`).

## 2. Pull Request (PR) Rules

- **Reproducibility**: If you change the Parser or Linker, update `examples/order-service`.

- **Docs First**: Change the specification in `docs/en/spec` and `docs/pt-br/spec` before changing the Go code.

- **@gen Blocks**: Any AI-generated artifact in the repository *must* have provenance metadata.

## 3. Code Generation Workflow

AI agents working in LIA should follow this flow:

1. Read the specification in `docs/en/spec` (and keep `docs/pt-br/spec` in sync).

2. Generate `.lia` or `.liao` files.

3. Validate using `lia check`.

4. Ensure that the local `prompt-tape.json` is up-to-date.

5. Dockerization

To run the toolchain in an isolated environment:

```bash
docker build -t lia-toolchain .

docker run --rm lia-toolchain parse examples/order-service/project.lia
```
