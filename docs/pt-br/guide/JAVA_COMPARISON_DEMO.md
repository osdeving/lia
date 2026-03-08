# Demo Java vs LIA

Esta demo existe para mostrar algo concreto para apresentacao:

- um projeto Spring Boot de referencia feito manualmente
- um projeto Spring Boot pedido direto para a IA
- o mesmo briefing passando pelo fluxo `LIA -> Java`
- o mesmo `project.lia` baixado em tres perfis: `plain`, `spring-boot` e `quarkus`

## Artefatos do repositório

- spec neutra da demo: `examples/demo-compare/project-spec.json`
- referencia Java manual (plain): `examples/java-reference/orders-service/`
- referencia Java manual (Spring Boot): `examples/java-reference/orders-service-spring-boot/`

## Como rodar

```bash
go run ./cmd/lia demo compare \
  --spec ./examples/demo-compare/project-spec.json \
  --provider openai-compatible \
  --base-url https://api.openai.com \
  --model gpt-4o-mini \
  --temperature 0.1 \
  --compare-profile spring-boot \
  --lia-java-profiles plain,spring-boot,quarkus \
  --out-dir /tmp/java-compare-demo \
  --pack-dir ./docs/pt-br/spec/packs
```

## O que sai

- `/tmp/java-compare-demo/COMPARISON.md`
- `/tmp/java-compare-demo/direct-java/`
- `/tmp/java-compare-demo/direct-java.compile.txt`
- `/tmp/java-compare-demo/lia-artifacts/project.lia`
- `/tmp/java-compare-demo/lia-artifacts/prompt-tape.json`
- `/tmp/java-compare-demo/lia-artifacts/project.lial.decision-log.json`
- `/tmp/java-compare-demo/lia-java-plain/`
- `/tmp/java-compare-demo/lia-java-plain.compile.txt`
- `/tmp/java-compare-demo/lia-java-spring-boot/`
- `/tmp/java-compare-demo/lia-java-spring-boot.compile.txt`
- `/tmp/java-compare-demo/lia-java-quarkus/`
- `/tmp/java-compare-demo/lia-java-quarkus.compile.txt`
- `/tmp/java-compare-demo/reference-java.compile.txt`

## Como ler a comparacao

O argumento central da demo nao e “a IA escreve Java melhor”.

O argumento e:

1. o mesmo briefing, quando passa por uma estrutura intermediaria tipada e modular, fica mais deterministico
2. o ramo LIA ganha artefatos auditaveis (`project.lia`, tape, decision log)
3. o lower consegue preservar a estrutura do LIA em perfis diferentes sem inventar arquitetura depois
4. a comparacao destaca o que ficou faltando no branch sem LIA

No run validado neste repositório:

- `Reference Java (spring-boot)`: compila
- `Direct Java (spring-boot)`: nao compila
- `LIA -> Java (spring-boot)`: compila
- `LIA -> Java (plain)`: compila
- `LIA -> Java (quarkus)`: compila

O arquivo `COMPARISON.md` agora tambem lista:

- quais artefatos o branch direto nao produziu
- quais elementos o fluxo LIA preservou
- os tres lowers do mesmo `project.lia`
- os reports de lowering usados para explicar a fidelidade do Java final

## Profiles do lower

Hoje o target Java aceita tres perfis deterministas:

- `plain`: Java puro com composicao manual
- `spring-boot`: `@SpringBootApplication` + `@Configuration` + `@Bean`
- `quarkus`: `@QuarkusMain` + producers CDI

Em todos os casos, o mapeamento semantico e o mesmo:

- `type` -> `record`
- `enum` -> `enum`
- `port` -> `interface`
- `usecase` -> classe com dependencias explicitas
- `wiring bind` -> composicao explicita do profile escolhido

## Limite atual

- o lower Java ainda usa placeholders conservadores para tipos referenciados e nao declarados
- adapters continuam com lowering comportamental parcial
- Python continua stub
- ainda nao existe typechecker semantico nem runtime alvo completo
