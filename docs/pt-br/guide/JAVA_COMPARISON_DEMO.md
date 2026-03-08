# Demo Java vs LIA

Esta demo existe para mostrar algo concreto para apresentacao:

- um projeto Java de referencia feito manualmente
- um projeto pedido direto em Java para a IA
- o mesmo briefing passando pelo fluxo `LIA -> Java`

## Artefatos do repositório

- spec neutra da demo: `examples/demo-compare/project-spec.json`
- referencia Java manual: `examples/java-reference/orders-service/`

## Como rodar

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

## O que sai

- `/tmp/java-compare-demo/COMPARISON.md`
- `/tmp/java-compare-demo/direct-java/`
- `/tmp/java-compare-demo/direct-java.compile.txt`
- `/tmp/java-compare-demo/lia-artifacts/project.lia`
- `/tmp/java-compare-demo/lia-artifacts/prompt-tape.json`
- `/tmp/java-compare-demo/lia-artifacts/project.lial.decision-log.json`
- `/tmp/java-compare-demo/lia-java/`
- `/tmp/java-compare-demo/lia-java.compile.txt`
- `/tmp/java-compare-demo/reference-java.compile.txt`

## Como ler a comparacao

O argumento central da demo nao e “a IA escreve Java melhor”.

O argumento e:

1. o mesmo briefing, quando passa por uma estrutura intermediaria tipada e modular, fica mais deterministico
2. o ramo LIA ganha artefatos auditaveis (`project.lia`, tape, decision log)
3. o lower consegue gerar um projeto Java compilavel mesmo quando o modelo deixa pequenos buracos semanticos

No run validado neste repositório:

- `Reference Java`: compila
- `Direct Java`: nao compila
- `LIA -> Java`: compila

## Limite atual

- o lower Java ainda usa placeholders conservadores para tipos referenciados e nao declarados
- Python continua stub
- ainda nao existe typechecker semantico nem runtime alvo completo
