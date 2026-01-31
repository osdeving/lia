# LIA — Linguagem Intermediária Assistida (model-first IR)

> **Status:** pré-alpha / design-first (spec em evolução)
> **Propósito:** tornar geração de software por LLMs **mais confiável, modular, verificável e reproduzível**.

A **LIA (Linguagem Intermediária Assistida)** é **uma linguagem/IR e um sistema**. Ela é desenhada para ser **produzida e consumida principalmente por modelos de linguagem (LLMs)** e agentes. Em vez de gerar código final direto (Java/Python/etc.), a LIA organiza a síntese em **“object files” semânticos** (`.liao`) que podem ser gerados em paralelo e então **linkados** por um *linker* determinístico, com **policies/constraints formais** (arquitetura, segurança, boas práticas) e **trilha de reprodutibilidade** (prompt/model/config).

---

## Visão geral do pipeline

1) **NL → `.liao`** (geração paralela por agentes/LLMs)
2) **Validate** (parse + typecheck + constraints locais)
3) **Link** (resolução de símbolos + políticas + seleção + merge)
4) **Emit `.lial`** (linked unit determinístico)
5) **Lower/Transpile** (para Java/Python/etc.)
6) **Shadow/JIT checks** (opcional)

---

## Reprodutibilidade (guardrails)

- Nada de `map[string]any` nas partes normativas do IR.
- **Structs + slices ordenadas** antes de serializar.
- JSON canônico, sem campos livres em áreas normativas.
- Hashes calculados a partir da representação canônica.

---

## Go-first toolchain

Este repositório é o toolchain da LIA em **Go**, cobrindo:
- AST/IR + serialização canônica
- parser v0.1 (minimalista)
- check/constraints (base)
- linker/solver (grafo + seleção + merge)
- passes internos (verify/transform/synthesize)
- lowering (stubs Java/Python)
- reprodutibilidade (prompt tape + @gen)

---

## Estrutura do repositório

```
lia/
  cmd/
    lia/
      main.go
  internal/
    cli/
    ir/
    parser/
    codec/
    repro/
    check/
    linker/
    passes/
    lower/
      java/
      python/
  pkg/
    api/
  docs/
    guide/
      getting-started.md
      pipeline.md
      cli.md
      repro.md
      ir.md
      packs-and-policies.md
      linker.md
    spec/
      lia-v0.1.md
      packs/
        arch-baseline.lia
        security-baseline.lia
      toolchain-checks.md
    INDEX.md
    overview.md
    roadmap.md
  examples/
    order-service/
      project.lia
      prompt-tape.json
  LICENSE
  README.md
  go.mod
```

`internal/` mantém o core encapsulado; se no futuro você quiser expor um SDK estável, use `pkg/`.

---

## CLI (v0.1)

Comandos disponíveis (stubs iniciais):

- `lia parse <file.lia> -o out.liao`
- `lia check <file.liao>`
- `lia link <inputs...> -o out.lial --decision-log out.lial.decision-log.json`
- `lia check ... --pack-dir ./docs/spec/packs` (opcional)
- `lia link ... --pack-dir ./docs/spec/packs` (opcional)
- `lia explain <file.lial> --decision-log out.lial.decision-log.json`
- `lia lower <file.lial> --target java|python -o out.java|out.py`
- `lia replay <file.lial> --tape prompt-tape.json` (stub)

---

## Build/Run

```bash
go run ./cmd/lia --help
```

```bash
go test ./...
```

---

## Exemplo rápido

```bash
lia parse examples/order-service/project.lia -o /tmp/order.liao
lia check /tmp/order.liao
lia link /tmp/order.liao -o /tmp/order.lial
lia explain /tmp/order.lial --decision-log /tmp/order.lial.decision-log.json
```

---

## Documentação em camadas

Comece por aqui:

- **Índice geral:** `docs/INDEX.md`
- **Visão geral (não técnica):** `docs/overview.md`
- **MVP em 10 minutos:** `docs/guide/getting-started.md`

Aprofundando por área:

- **Pipeline:** `docs/guide/pipeline.md`
- **CLI:** `docs/guide/cli.md`
- **Reprodutibilidade:** `docs/guide/repro.md`
- **IR/AST:** `docs/guide/ir.md`
- **Packs/Policies:** `docs/guide/packs-and-policies.md`
- **Linker:** `docs/guide/linker.md`
- **Checklist rápido:** `docs/guide/what-we-have.md`
- **Arquitetura do sistema:** `docs/architecture/system.md`
- **O que existe vs não existe:** `docs/roadmap.md`

Specs e packs:

- Spec inicial: `docs/spec/lia-v0.1.md`
- Rascunho histórico: `lia.md`
- Packs MVP (pseudo-LIA): `docs/spec/packs/arch-baseline.lia`, `docs/spec/packs/security-baseline.lia`
- Checklist do toolchain: `docs/spec/toolchain-checks.md`

---

## Próximos passos sugeridos

- Expandir parser (participle/v2 ou gramática formal)
- Formalizar constraints e preferências no linker
- Definir `.liao` e `decision-log` como formatos canônicos
- Implementar `lia.patch` e synth passes com replay
