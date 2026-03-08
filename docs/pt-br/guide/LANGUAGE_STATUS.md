# Estado da Linguagem (v0.1)

Este documento descreve o que a LIA suporta hoje no código, o que já funciona de ponta a ponta e o que ainda é parcial.

## O que já funciona de ponta a ponta

- `lia parse`: texto `.lia` para `.liao` em JSON canônico.
- `lia check`: validação estrutural, efeitos, símbolos derivados e subset de packs/policies.
- `lia link`: resolução determinística de `requires -> provides`, com `decision-log`.
- `lia explain`: resumo de hash, contagem de módulos e decision log.
- `lia replay`: validação do `.lial` contra `prompt-tape.json` usando `@gen.prompt_ref` e `@gen.prompt_hash`.
- `internal/llmgen`: grava prompt tape no formato canônico e parseia o LIA devolvido pelo provider.
- `lia gen project`: bootstrap de projeto LIA a partir de `project-spec.json` + provider compatível.
- `lia lower --target java`: gera projeto Java multi-arquivo compilável a partir do `.lial`.
- `lia demo compare`: compara `Java direto` vs `LIA -> Java` com o mesmo briefing.

## Palavras-chave suportadas hoje

### Top-level

- `project`
- `pack`
- `module`

### Dentro de `project`

- `repro`
- `tape`
- `use pack`
- `policy`
- `constraint`

### Dentro de `module`

- `@gen`
- `as`
- `type`
- `enum`
- `port`
- `fn`
- `method`
- `usecase`
- `adapter`
- `implements`
- `input`
- `output`
- `effects`
- `wiring`
- `bind`
- `prefer`
- `weight`
- `hole`
- `candidate`
- `score`
- `constraint`

### Statements

- `let`
- `return`
- `if`
- `else`
- `while`
- `for`
- `in`
- `break`
- `continue`

### Expressões e operadores

- literais: número, string, bool, lista
- chamada: `fn(x)`
- membro: `obj.field`
- índice: `items[0]`
- operadores: `||`, `&&`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `+`, `-`, `*`, `/`, `%`, `!`

## Semântica implementada de fato

- `provides` e `requires` são derivados para `type`, `enum`, `port`, `usecase`, `adapter` e `wiring`.
- `effects` aceitos: `pure`, `io`, `tx`, `emit`.
- `pure` não pode ser combinado com outros efeitos.
- `repro=strict` e `repro=pinned` exigem `tape` no `project`.
- `repro=strict` não permite `hole` sobreviver ao `.lial`.
- `replay` valida presença do `prompt_ref` no tape e confere `prompt_hash`.
- O linker usa score de `candidate`, preferências do requester e desempate determinístico.

## O que ainda é parcial

- Não existe typechecker semântico; `type_ref` ainda é tratado como forma bruta.
- `policy` e `constraint` são `raw_expr`; só um subconjunto é realmente aplicado.
- `hole` é parseado e validado em `strict`, mas ainda não existe resolução automática de holes.
- `candidate` já influencia score, mas ainda não existe mecanismo completo de síntese/substituição.
- `lower` já é real para Java, mas Python continua stub.
- o lower Java ainda usa placeholders conservadores quando o LIA referencia tipos não declarados.

## Como testar o estado atual

### Teste automatizado do repositório

```bash
go test ./...
```

### Smoke test manual do pipeline

Use o fixture `examples/full-pipeline/`:

```bash
go run ./cmd/lia parse examples/full-pipeline/project.lia -o /tmp/full-pipeline.liao
go run ./cmd/lia check /tmp/full-pipeline.liao --pack-dir ./docs/pt-br/spec/packs
go run ./cmd/lia link /tmp/full-pipeline.liao -o /tmp/full-pipeline.lial --decision-log /tmp/full-pipeline.decision-log.json --pack-dir ./docs/pt-br/spec/packs
go run ./cmd/lia explain /tmp/full-pipeline.lial --decision-log /tmp/full-pipeline.decision-log.json
go run ./cmd/lia replay /tmp/full-pipeline.lial --tape ./examples/full-pipeline/prompt-tape.json
```

Resultado esperado:

- `check` termina com `ok`
- `link` gera `.lial` e `decision-log.json`
- `replay` reporta `modules 3, generated 3, validated 3`

### Smoke test do lower Java

```bash
go run ./cmd/lia lower /tmp/full-pipeline.lial --target java --out-dir /tmp/full-pipeline-java
javac $(find /tmp/full-pipeline-java/src/main/java -name '*.java')
```

Resultado esperado:

- o comando `lower` escreve um projeto Java em `/tmp/full-pipeline-java`
- `javac` termina sem erro

### Demo comparativa Java direto vs LIA

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

Artefatos esperados:

- `COMPARISON.md`
- `direct-java/`
- `lia-artifacts/project.lia`
- `lia-artifacts/prompt-tape.json`
- `lia-artifacts/project.lial.decision-log.json`
- `lia-java/`

## Como usar com IA agora

Se for pedir LIA para um modelo, peça para ele gerar apenas constructs desta lista e depois valide no fluxo acima.

Recomendação prática:

- gerar um único `module ...` por vez
- incluir `@gen` com `prompt_ref`, `prompt_hash` e `model_id`
- validar sempre com `parse -> check -> link -> replay`
