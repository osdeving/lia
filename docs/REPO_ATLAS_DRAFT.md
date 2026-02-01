# Repo Atlas (draft)

This file is a temporary map of the repository. It groups concepts and lists every file with a short, practical explanation (what it is, what it does, how it is used). Keep this updated while exploring the repo.

## 0) High-level purpose

LIA is a model-first toolchain + spec. The pipeline is:

- .lia (source) -> parse -> .liao (canonical IR)
- .liao -> check (policies/constraints)
- .liao -> link -> .lial (linked unit + decision log)
- .lial -> lower -> target code (java/python, etc)

Key goals: reproducibility (canonical JSON + hashes + @gen), linkability (module merge), and policy enforcement (roles/effects/constraints).

## 1) Concepts grouped (what exists vs how it works)

### 1.1 Spec and language
- LIA v0.1 spec lives in docs (EN/PT). The grammar is intentionally small, then expanded by toolchain passes.
- `lia.md` is a top-level spec summary snapshot used by humans and prompts.

### 1.2 IR + canonicalization
- The IR is defined in `internal/ir/ir.go`.
- Canonical JSON + hashing are in `internal/codec/codec.go`.
- Determinism: all relevant collections are sorted before serialization.

### 1.3 Policies and packs
- Packs are pseudo-LIA examples under `docs/*/spec/packs/`.
- The policy engine is in `internal/policy/engine.go`.
- Pack parsing/loading is in `internal/packs/`.

### 1.4 CLI
- Cobra-based CLI under `internal/cli/` and `cmd/lia/main.go`.
- Commands: parse, check, link, explain, replay, lower, packs.

### 1.5 Linker
- The linker merges modules, emits a decision log, and applies policy rules.

### 1.6 Lowering
- Lowers `.lial` into Python/Java stubs.

### 1.7 Repro and provenance
- `@gen` metadata is mandatory for AI-generated artifacts.
- Prompt tape is stored in examples (`prompt-tape.json`).


## 2) Repository map (every file)

### 2.1 Root files
- `AGENT.md` - rules for AI agents: docs-first, tests, coverage, branch naming, @gen.
- `LICENSE` - project license.
- `README.md` - project entry point, quick usage.
- `Dockerfile` - container build to run toolchain isolated.
- `.dockerignore` - excludes files from Docker build context.
- `.gitignore` - ignore rules (includes `out/`).
- `.golangci.yml` - golangci-lint configuration.
- `.goreleaser.yaml` - GoReleaser configuration.
- `.agent/` - local agent metadata/config (if used by tooling).
- `go.mod` - Go module definition and dependencies.
- `go.sum` - Go dependency checksums.
- `lia.md` - high-level LIA spec snapshot, used as guidance.
- `coverage` - local coverage output file (generated).

### 2.2 GitHub metadata
- `.github/COMMIT_TEMPLATE.md` - conventional commit template.
- `.github/pull_request_template.md` - PR checklist and structure.
- `.github/workflows/ci.yml` - CI pipeline (lint, tests, build).
- `.github/workflows/release.yml` - GoReleaser pipeline on tags.
- `.github/appmod/` and `.github/appmod/appcat/` - empty placeholder dirs.

### 2.3 Command binaries
- `cmd/lia/main.go` - CLI entrypoint, wires commands.

### 2.4 Internal packages (core)
- `internal/ir/ir.go` - IR/AST definitions, including statements/expressions.
- `internal/codec/codec.go` - canonical JSON + hashing + sorting.
- `internal/codec/codec_test.go` - codec tests.
- `internal/parser/parser.go` - `.lia` parser (minimal grammar + usecase body).
- `internal/parser/parser_test.go` - parser tests (coverage >85%).
- `internal/check/check.go` - structural validation and policy checks.
- `internal/linker/linker.go` - linker + decision log + policy application.
- `internal/linker/linker_test.go` - linker tests (coverage >85%).
- `internal/policy/engine.go` - minimal policy/constraint DSL executor.
- `internal/packs/loader.go` - pack loader (files -> IR).
- `internal/packs/parse.go` - pack parsing helpers.
- `internal/repro/repro.go` - repro helpers (prompt hash, metadata).
- `internal/passes/passes.go` - pass registry scaffolding (v0.1).
- `internal/llmgen/provider.go` - LLM provider interface.
- `internal/llmgen/generator.go` - generator implementation.
- `internal/llmgen/generator_test.go` - llmgen tests.
- `internal/lower/python/python.go` - Python lowerer for usecase body.
- `internal/lower/java/java.go` - Java lowerer stub.

### 2.5 CLI subcommands
- `internal/cli/root.go` - root Cobra command.
- `internal/cli/common.go` - shared helpers (flags, IO).
- `internal/cli/parse.go` - `lia parse` implementation.
- `internal/cli/check.go` - `lia check` implementation.
- `internal/cli/link.go` - `lia link` implementation.
- `internal/cli/explain.go` - `lia explain` implementation.
- `internal/cli/replay.go` - `lia replay` implementation (placeholder).
- `internal/cli/lower.go` - `lia lower` implementation.
- `internal/cli/packs.go` - `lia packs` implementation.
- `internal/cli/diag.go` - CLI diagnostics output helpers.

### 2.6 Examples
- `examples/order-service/project.lia` - baseline example project.
- `examples/order-service/prompt-tape.json` - prompt tape for order-service.

### 2.7 Docs (English)
- `docs/REPO_ATLAS_DRAFT.md` - temporary repository map (this file).
- `docs/en/INDEX.md` - doc index (entry points).
- `docs/en/OVERVIEW.md` - non-technical overview.
- `docs/en/LIA_MANIFESTO.md` - model-first manifesto.
- `docs/en/ROADMAP.md` - roadmap status.
- `docs/en/PROCESS.md` - change process (roadmap -> ADR -> spec -> code).
- `docs/en/CONTRIBUTING.md` - contribution workflow.
- `docs/en/TESTING_GUIDE.md` - how to run tests.
- `docs/en/TESTING_STRATEGY.md` - testing strategy, LLM-aware.
- `docs/en/architecture/SYSTEM.md` - system architecture overview.
- `docs/en/guide/GETTING_STARTED.md` - quick start.
- `docs/en/guide/PIPELINE.md` - pipeline artifacts.
- `docs/en/guide/CLI.md` - CLI reference.
- `docs/en/guide/REPRO.md` - repro model.
- `docs/en/guide/IR.md` - IR/AST shape.
- `docs/en/guide/LINKER.md` - linker responsibilities.
- `docs/en/guide/PACKS_AND_POLICIES.md` - packs, policies, constraints.
- `docs/en/guide/WHAT_WE_HAVE.md` - what exists vs not.
- `docs/en/spec/LIA_V0_1.md` - spec and grammar.
- `docs/en/spec/TOOLCHAIN_CHECKS.md` - validation checklist.
- `docs/en/spec/packs/ARCH_BASELINE.lia` - ArchBaseline pack (pseudo-LIA).
- `docs/en/spec/packs/SECURITY_BASELINE.lia` - SecurityBaseline pack (pseudo-LIA).

### 2.8 Docs (PT-BR)
- `docs/pt-br/INDEX.md` - indice da documentacao.
- `docs/pt-br/OVERVIEW.md` - visao geral.
- `docs/pt-br/LIA_MANIFESTO.md` - manifesto.
- `docs/pt-br/ROADMAP.md` - roadmap.
- `docs/pt-br/PROCESS.md` - processo de mudancas.
- `docs/pt-br/CONTRIBUTING.md` - contribuicao.
- `docs/pt-br/TESTING_GUIDE.md` - guia de testes.
- `docs/pt-br/TESTING_STRATEGY.md` - estrategia de testes.
- `docs/pt-br/architecture/SYSTEM.md` - arquitetura do sistema.
- `docs/pt-br/guide/GETTING_STARTED.md` - guia rapido.
- `docs/pt-br/guide/PIPELINE.md` - pipeline e artefatos.
- `docs/pt-br/guide/CLI.md` - CLI.
- `docs/pt-br/guide/REPRO.md` - reproducao.
- `docs/pt-br/guide/IR.md` - IR/AST.
- `docs/pt-br/guide/LINKER.md` - linker.
- `docs/pt-br/guide/PACKS_AND_POLICIES.md` - packs e politicas.
- `docs/pt-br/guide/WHAT_WE_HAVE.md` - o que existe.
- `docs/pt-br/spec/LIA_V0_1.md` - spec e gramatica.
- `docs/pt-br/spec/TOOLCHAIN_CHECKS.md` - checklist.
- `docs/pt-br/spec/packs/ARCH_BASELINE.lia` - pack arch.
- `docs/pt-br/spec/packs/SECURITY_BASELINE.lia` - pack security.

### 2.9 ADRs
- `docs/adr/README.md` - ADR index and conventions.

## 3) Points of attention / gotchas

- Docs are duplicated in EN and PT-BR. Keep them aligned.
- `docs/*/PROCESS.md` defines the roadmap + ADR + PR flow; CONTRIBUTING is lighter.
- Parser/Linker must keep >85% coverage (AGENT.md).
- Any AI-generated artifact must include @gen and have prompt tape entries.
- Canonical JSON is required for reproducibility; avoid map iteration.

## 4) Next steps (for this draft)

- Fill missing details if any file changes.
- Keep this draft updated while exploring.
