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
- lowering (Java real, Python stub)
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
    adr/
    en/
      guide/
      spec/
      INDEX.md
      OVERVIEW.md
      ROADMAP.md
    pt-br/
      guide/
      spec/
      INDEX.md
      OVERVIEW.md
      ROADMAP.md
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

Comandos disponíveis:

- `lia parse <file.lia> -o out.liao`
- `lia check <file.liao>`
- `lia link <inputs...> -o out.lial --decision-log out.lial.decision-log.json`
- `lia check ... --pack-dir ./docs/pt-br/spec/packs` (ou `./docs/en/spec/packs`)
- `lia link ... --pack-dir ./docs/pt-br/spec/packs` (ou `./docs/en/spec/packs`)
- `lia explain <file.lial> --decision-log out.lial.decision-log.json`
- `lia lower <file.lial> --target java --out-dir out-dir`
- `lia lower <file.lial> --target python -o out.py`
- `lia replay <file.lial> --tape prompt-tape.json`
- `lia gen project --spec project-spec.json --provider ... --model ... --out-dir out-dir`
- `lia demo compare --spec project-spec.json --provider ... --model ... --out-dir out-dir`

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

Comece por aqui (escolha o idioma):

- **Índice (pt-br):** `docs/pt-br/INDEX.md`
- **Índice (en):** `docs/en/INDEX.md`
- **Visão geral (pt-br):** `docs/pt-br/OVERVIEW.md`
- **Visão geral (en):** `docs/en/OVERVIEW.md`
- **MVP em 10 minutos (pt-br):** `docs/pt-br/guide/GETTING_STARTED.md`

Aprofundando por área:

- **Pipeline:** `docs/pt-br/guide/PIPELINE.md`
- **CLI:** `docs/pt-br/guide/CLI.md`
- **Reprodutibilidade:** `docs/pt-br/guide/REPRO.md`
- **IR/AST:** `docs/pt-br/guide/IR.md`
- **Packs/Policies:** `docs/pt-br/guide/PACKS_AND_POLICIES.md`
- **Linker:** `docs/pt-br/guide/LINKER.md`
- **Checklist rápido:** `docs/pt-br/guide/WHAT_WE_HAVE.md`
- **Arquitetura do sistema:** `docs/pt-br/architecture/SYSTEM.md`
- **O que existe vs não existe:** `docs/pt-br/ROADMAP.md`

Specs e packs:

- Spec inicial (pt-br): `docs/pt-br/spec/LIA_V0_1.md`
- Spec inicial (en): `docs/en/spec/LIA_V0_1.md`
- Rascunho histórico: `lia.md`
- Packs MVP (pt-br): `docs/pt-br/spec/packs/ARCH_BASELINE.lia`, `docs/pt-br/spec/packs/SECURITY_BASELINE.lia`
- Packs MVP (en): `docs/en/spec/packs/ARCH_BASELINE.lia`, `docs/en/spec/packs/SECURITY_BASELINE.lia`
- Checklist do toolchain (pt-br): `docs/pt-br/spec/TOOLCHAIN_CHECKS.md`
- Checklist do toolchain (en): `docs/en/spec/TOOLCHAIN_CHECKS.md`

---

## Próximos passos sugeridos

- Expandir parser (participle/v2 ou gramática formal)
- Formalizar constraints e preferências no linker
- Definir `.liao` e `decision-log` como formatos canônicos
- Implementar `lia.patch` e synth passes com replay
