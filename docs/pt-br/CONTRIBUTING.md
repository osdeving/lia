# Guia de Contribuição e Workflow

Bem-vindo ao desenvolvimento do **LIA**. Como este é um projeto crítico de infraestrutura para IA, seguimos padrões de engenharia rigorosos.

## 1. Ciclo de Desenvolvimento

1. **Branching**: Crie branches a partir da `develop`.
    - `feat/...` para novas funcionalidades.
    - `fix/...` para correções.
    - `spec/...` para mudanças na especificação LIA.
2. **Commits**: Use [Conventional Commits](https://www.conventionalcommits.org/v1.0.0/).
    - Ex: `feat(linker): add support for selective symbol merging`
3. **Testes**: Todo PR deve passar em `go test ./...`.
4. **Linting**: O código deve passar no `golangci-lint` (verifique com `golangci-lint run`).

## 2. Regras de Pull Request (PR)

- **Reprodutibilidade**: Se você mudar o Parser ou Linker, atualize o `examples/order-service`.
- **Docs First**: Mude a especificação em `docs/spec` antes de mudar o código Go.
- **Blocos @gen**: Qualquer artefato gerado por IA no repositório *deve* ter metadados de proveniência.

## 3. Workflow de Geração de Código

Agentes de IA trabalhando no LIA devem seguir este fluxo:

1. Ler a especificação em `docs/spec`.
2. Gerar arquivos `.lia` ou `.liao`.
3. Validar usando `lia check`.
4. Garantir que o `prompt-tape.json` local está atualizado.

## 4. Dockerização

Para rodar o toolchain em um ambiente isolado:

```bash
docker build -t lia-toolchain .
docker run --rm lia-toolchain parse examples/order-service/project.lia
```
