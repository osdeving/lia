# Roadmap / Status — Completude da Linguagem

Este documento é o plano de implementação para tornar a LIA uma linguagem/IR mais completa. Ele é **docs‑first** e deve ser atualizado antes do código.

## Princípios

- Spec first (atualizar `docs/pt-br/spec` + `docs/en/spec` antes do código).
- Determinismo e reprodutibilidade são inegociáveis.
- Marcos pequenos, verificáveis e com critérios claros.

## Estado atual (v0.1 na develop)

- CLI: parse/check/link/explain/lower (stubs em partes).
- Parser: **subconjunto muito pequeno** (`project`, `module`, `use pack`, `repro`, `tape`).
- Estruturas IR + JSON canônico + hashing.
- Carregador de packs (pseudo‑LIA) + motor mínimo de políticas.
- Linker: stub (sem resolução de grafo/seleção).
- Lowerers: stubs.

## O que significa “mais completo”

Uma LIA mais completa deve suportar:

- **Gramática core completa** (types, enums, ports, usecases, adapters, wiring, constraints, preferences, effects).
- **Camada semântica** (symbols, requires/provides, validação de effects).
- **Linker determinístico** (grafo, seleção, decision log).
- **Políticas aplicadas** (roles/effects/deps; depois capabilities/taint).
- **Ao menos um lowering real** com convenções de runtime.

## Marcos (fases)

### M1 — Gramática core + AST

**Escopo**
- Adicionar gramática para: types, enums, ports, usecases, adapters, wiring, constraints, preferences, effects.
- Definir gramática mínima de expressões/statement para bodies de usecase (se imperativo).
- Expandir IR para representar esses constructs.

**Done when**
- Parser aceita os exemplos canônicos em `examples/`.
- EBNF da spec atualizado em EN/PT.
- Testes unitários cobrindo edge cases do parser.

### M2 — IR semântico (symbols + effects)

**Escopo**
- `requires/provides` derivado dos módulos.
- Effects validados contra roles.
- Regras de canonicalização ajustadas.

**Done when**
- `lia check` reporta diagnósticos precisos para símbolos faltantes/invalidos.
- Hashes determinísticos para entradas idênticas.

### M3 — Linker determinístico

**Escopo**
- Construção do grafo de dependências.
- Detecção de colisões e seleção de candidatos.
- Tie‑break fixo e decision log detalhado.

**Done when**
- Linker resolve múltiplos `.liao` em `.lial` estável com decision log reproduzível.
- Testes para seleção + regras de tie‑break.

### M4 — Enforcement de políticas (profiles)

**Escopo**
- Enforçar constraints de role/effect/deps vindas dos packs.
- Budgets (max modules, max deps) aplicados.

**Done when**
- Violações viram erro duro no `lia check/link`.
- Constraints de packs aplicadas de forma consistente.

### M5 — Primeiro lowering real

**Escopo**
- Um target (Python ou Java) com saída executável e convenções mínimas de runtime.
- Mapeamento para usecases, ports/adapters e IO básico.

**Done when**
- `lia lower` gera output executável para exemplo não‑trivial.
- Golden tests validam codegen.

### M6 — Tooling + reprodutibilidade

**Escopo**
- Validação estrita de `@gen` + prompt tape em `repro=strict`.
- `lia replay` definido e aplicado.

**Done when**
- Rebuilds reproduzíveis sob `repro=strict`.
- Decision log + prompt tape totalmente rastreáveis.

## Fora de escopo (por enquanto)

- Prova formal completa de regras de negócio.
- Marketplace/registry de packs.
- Runtime de plugins externos (WASM) para passes.

## Próximos passos imediatos

1) Expandir gramática e parser (M1).
2) Introduzir symbol table + requires/provides (M2).
3) Linker determinístico (M3).
