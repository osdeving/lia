# Getting Started (MVP)

Este é o menor fluxo que hoje prova o toolchain sem depender de `lower`.

## Pré-requisitos

- Go 1.23+

## Rode o fluxo

```bash
# 1) parse .lia -> .liao
go run ./cmd/lia parse examples/full-pipeline/project.lia -o /tmp/full-pipeline.liao

# 2) validate
go run ./cmd/lia check /tmp/full-pipeline.liao --pack-dir ./docs/pt-br/spec/packs

# 3) link .liao -> .lial + decision log
go run ./cmd/lia link /tmp/full-pipeline.liao -o /tmp/full-pipeline.lial --decision-log /tmp/full-pipeline.decision-log.json --pack-dir ./docs/pt-br/spec/packs

# 4) explain
go run ./cmd/lia explain /tmp/full-pipeline.lial --decision-log /tmp/full-pipeline.decision-log.json

# 5) replay
go run ./cmd/lia replay /tmp/full-pipeline.lial --tape ./examples/full-pipeline/prompt-tape.json
```

## O que você deve ver

- `check` termina com `ok`
- `link` escreve `/tmp/full-pipeline.lial`
- `explain` mostra hash e quantidade de entradas do decision log
- `replay` mostra `modules 3, generated 3, validated 3`

## Limitações atuais

- não existe typechecker completo
- policy DSL ainda é parcial
- `hole` ainda não é resolvido automaticamente
- `lower` continua stub

Para o quadro completo da linguagem suportada hoje, leia `docs/pt-br/guide/LANGUAGE_STATUS.md`.
