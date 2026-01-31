# Referência CLI (v0.1)

## Comandos

- `lia parse <file.lia> -o out.liao`
  - parser mínimo; emite JSON canônico

- `lia check <file.liao|file.lial>`
  - executa validação local + regras de pack (subconjunto)

- `lia link <inputs...> -o out.lial --decision-log out.lial.decision-log.json`
  - mescla inputs e emite um decision log

- `lia explain <file.lial> --decision-log out.lial.decision-log.json`
  - imprime hash + resumo

- `lia lower <file.lial> --target java|python -o out.java|out.py`
  - lowerers stub

- `lia replay <file.lial> --tape prompt-tape.json`
  - placeholder para replay futuro

## Carregamento de packs

Comandos que validam ou linkam podem carregar packs:

- caminhos de busca padrão: `./packs` e `./docs/spec/packs` (se existir)
- neste repo, packs ficam em `docs/pt-br/spec/packs` e `docs/en/spec/packs`
- sobrescrever/adicionar caminhos: `--pack-dir <path>` (repetível)

Exemplo:

```bash
lia check out.liao --pack-dir ./docs/pt-br/spec/packs
```
