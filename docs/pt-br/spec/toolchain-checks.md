# Toolchain checks (v0.1)

Este documento lista os checks mínimos que `lia check` e `lia link` devem implementar para suportar **profiles** e **packs** como `ArchBaseline` e `SecurityBaseline`.

## lia check (validação local)

**Objetivo:** validar um único artefato `.liao`/`.lial` sem resolver dependências externas.

Checklist mínimo:

1) **Estrutura básica**
   - `program.version` obrigatório.
   - `module.name` obrigatório e QName normalizado.
   - `@gen.prompt_ref` exige `@gen.prompt_hash`.

2) **Canonicalização / determinismo**
   - Slices ordenáveis antes da serialização canônica.
   - Proibir `map[string]any` em áreas normativas do IR.

3) **Regras locais por role/effect**
   - Aplicar constraints locais (ex.: `domain` não pode `io`/`tx`).
   - Validar efeitos declarados vs role (ex.: `port` sem `io`).

4) **Buracos (holes)**
   - `hole` deve ter `contract` não vazio.
   - Se `repro == strict` e o artefato for `.lial`, nenhum `hole` pode permanecer.

5) **Metadados de reprodutibilidade**
   - `@gen` deve ser estruturalmente válido.
   - Referências a `prompt_ref` devem existir no tape quando fornecido.

6) **Packs (profiles)**
   - Carregar packs referenciados no project (por default: `./docs/spec/packs` ou `./packs`).
   - Avisar quando políticas não são aplicáveis no v0.1 (ex.: capabilities/taint ainda não modelados).

## lia link (resolução global)

**Objetivo:** resolver o sistema inteiro e produzir `.lial` determinístico + decision log.

Checklist mínimo:

1) **Grafo de dependências**
   - Construir `requires -> provides`.
   - Detectar símbolos ausentes e colisões.

2) **Constraints globais (hard)**
   - Aplicar constraints entre módulos (ex.: `domain` não depende de `adapter`).
   - Eliminar candidatos inválidos antes do scoring.

3) **Seleção determinística**
   - Aplicar preferências (soft) e score.
   - Tie-break fixo: semver desc, stability desc, deps asc, lexical.

4) **Merges explícitos**
   - `choose-one | rename | wrap | adapt` conforme política.

5) **Decision log canônico**
   - Registrar escolhas, scores e constraints aplicadas.
   - Hash canônico do log.

6) **Reprodutibilidade**
   - Em `repro=strict`: exigir `@gen` completo nos símbolos públicos e decision log emitido.

7) **Profiles (packs)**
   - Carregar packs referenciados no project.
   - Unificar políticas e constraints antes do link final.
   - Em v0.1, aplicar apenas o subconjunto suportado do DSL (roles, efeitos, budgets, deps).

---

## O que `ArchBaseline` e `SecurityBaseline` exigem

- **ArchBaseline**: enforcement de roles, direção de dependências e limites de efeitos por role.
- **SecurityBaseline**: enforcement de políticas de authn/authz, validação de input na borda, proibição de SQL raw, políticas de secrets, logging e TLS.

Essas políticas ainda são pseudo-LIA no v0.1, mas já guiam exatamente onde o toolchain precisa validar e bloquear.
