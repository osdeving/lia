# Referência CLI (v0.1)

## Comandos

- `lia parse <file.lia> -o out.liao`
  - parseia a gramática v0.1 suportada e emite JSON canônico

- `lia check <file.liao|file.lial>`
  - executa validação estrutural, efeitos, repro e subset de packs

- `lia link <inputs...> -o out.lial --decision-log out.lial.decision-log.json`
  - resolve símbolos de forma determinística e emite decision log

- `lia explain <file.lial> --decision-log out.lial.decision-log.json`
  - imprime hash + resumo

- `lia lower <file.lial> --target java --profile plain|spring-boot|quarkus --out-dir out-dir`
  - gera um projeto Java multi-arquivo a partir do `.lial`
  - `plain`: Java puro com composicao manual
  - `spring-boot`: gera bootstrap/configuracao Spring Boot
  - `quarkus`: gera bootstrap/producers CDI para Quarkus

- `lia lower <file.lial> --target python -o out.py`
  - lower Python ainda é stub

- `lia replay <file.lial> --tape prompt-tape.json`
  - valida `@gen.prompt_ref` e `@gen.prompt_hash` contra o tape

- `lia gen project --spec project-spec.json --provider ollama|openai-compatible --model <model> --out-dir <dir>`
  - usa IA para gerar um projeto LIA a partir de uma spec JSON e já roda parse/check/link/replay

- `lia gen app --prompt "..." --provider ollama|openai-compatible --model <model> --out-dir <dir> [--java-profile plain|spring-boot|quarkus] [--java-profiles plain,spring-boot,quarkus]`
  - usa prompt livre, `lia.json` opcional e manifests de packs para gerar um plano interno, produzir LIA e, no target Java, emitir o projeto final
  - com `--java-profiles`, o mesmo `project.lia` pode ser baixado em varios perfis no mesmo run

- `lia demo compare --spec project-spec.json --provider ollama|openai-compatible --model <model> --out-dir <dir> [--compare-profile spring-boot] [--lia-java-profiles plain,spring-boot,quarkus]`
  - gera `Java direto` e `LIA -> Java`, compila os ramos relevantes e escreve `COMPARISON.md`
  - por padrao, compara o perfil `spring-boot` e tambem emite lowers `plain`, `spring-boot` e `quarkus` para o mesmo LIA

## Carregamento de packs

Comandos que validam ou linkam podem carregar packs:

- caminhos de busca padrão: `./packs` e `./docs/spec/packs` (se existir)
- neste repo, packs ficam em `docs/pt-br/spec/packs` e `docs/en/spec/packs`
- sobrescrever/adicionar caminhos: `--pack-dir <path>` (repetível)

Exemplo:

```bash
lia check out.liao --pack-dir ./docs/pt-br/spec/packs
```
