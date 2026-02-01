# Processo de Mudanças e Workflow LIA

Este documento define o ciclo de vida completo de mudanças no projeto, desde a concepção (Roadmap) até a entrega (Release).

## 1. Roadmap e Planejamento

O Roadmap (`ROADMAP.md`) é a fonte da verdade para a direção estratégica.

- **Curadoria**: Mantido pelos *Core Maintainers*.
- **Atualização**: Revisado trimestralmente ou após *Milestone Releases*.
- **Propostas**: Qualquer contribuidor pode sugerir itens via *Issues* com a label `type:roadmap`.

### Ciclo de Feature (RFC/ADR)

Para mudanças significativas (arquitetura, mudanças na linguagem LIA), seguimos o fluxo **Specs-First**:

1. **RFC (Request for Comments)**:
   - Crie uma issue `RFC: [Título]` descrevendo o problema e proposta.
   - Discussão aberta com a comunidade.
2. **ADR (Architecture Decision Record)**:
   - Se aprovado, formalize em `docs/adr/` (um arquivo por decisão).
   - Define o contrato antes do código.
3. **Spec Update**:
   - Atualize `docs/pt-br/spec/` e `docs/en/spec/` com a nova gramática/comportamento.
4. **Implementação**:
   - Só inicie o código após a Spec estar aprovada.

## 2. Pull Request (PR) Standard

### Template de PR

Todo PR deve conter na descrição:

```markdown
## Tipo
- [ ] Feat
- [ ] Fix
- [ ] Docs
- [ ] Spec
- [ ] Test
- [ ] Refactor
- [ ] Perf

## Contexto
O que mudou e por que? (Link para Issue/RFC)

## Checklist
- [ ] Spec atualizada (se aplicável)
- [ ] Testes adicionados (`go test ./...`)
- [ ] Linting passou (`golangci-lint run`)
- [ ] Reprodutibilidade verificada (se gerado por IA)
- [ ] Coverage não diminuiu
```

### Política de Review

- **Aprovação Obrigatória**: Pelo menos 1 *Core Maintainer*.
- **CI Checks**: Todos os checks (Test, Lint, Build) devem passar (verde).
- **Sem Regressão**: Coverage não deve diminuir.

## 3. Branches e Estratégia Git

Seguimos um modelo simplificado do **Gitflow**:

- `main`: Produção (estável, tags `vX.Y.Z`).
- `develop`: Integração (instável, base para features).
- `feat/...`: Desenvolvimento de features (baseada em `develop`).
- `fix/...`: Correção de bugs.
- `release/...`: Preparação para nova versão (congelamento).

**Regra de Ouro**: PRs de features devem sempre apontar para `develop`.

## 4. Release Cycle

1. **Feature Freeze**: Merge de features bloqueado na `develop`.
2. **Release Candidate**: Tag `v0.1.0-rc1`.
3. **Validação**: Testes manuais e de reprodução em massa.
4. **Tag Final**: Tag `v0.1.0` na `main`.
5. **Automação**: CI dispara GoReleaser para gerar artefatos.

---

**Referências**:

- [AGENT.md](../AGENT.md): Guidelines para IA.
- [CONTRIBUTING.md](CONTRIBUTING.md): Guia técnico para devs.
