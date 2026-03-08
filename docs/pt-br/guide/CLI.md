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

- `lia lower <file.lial> --target java --out-dir out-dir`
  - gera um projeto Java multi-arquivo a partir do `.lial`

- `lia lower <file.lial> --target python -o out.py`
  - lower Python ainda é stub

- `lia replay <file.lial> --tape prompt-tape.json`
  - valida `@gen.prompt_ref` e `@gen.prompt_hash` contra o tape

- `lia gen project --spec project-spec.json --provider ollama|openai-compatible --model <model> --out-dir <dir>`
  - usa IA para gerar um projeto LIA a partir de uma spec JSON e já roda parse/check/link/replay

- `lia gen app --prompt "..." --provider ollama|openai-compatible --model <model> --out-dir <dir>`
  - usa prompt livre, `lia.json` opcional e manifests de packs para gerar um plano interno, produzir LIA e, no target Java, emitir o projeto final

- `lia demo compare --spec project-spec.json --provider ollama|openai-compatible --model <model> --out-dir <dir>`
  - gera `Java direto` e `LIA -> Java`, compila os dois ramos e escreve `COMPARISON.md`

## Carregamento de packs

Comandos que validam ou linkam podem carregar packs:

- caminhos de busca padrão: `./packs` e `./docs/spec/packs` (se existir)
- neste repo, packs ficam em `docs/pt-br/spec/packs` e `docs/en/spec/packs`
- sobrescrever/adicionar caminhos: `--pack-dir <path>` (repetível)

Exemplo:

```bash
lia check out.liao --pack-dir ./docs/pt-br/spec/packs
```
