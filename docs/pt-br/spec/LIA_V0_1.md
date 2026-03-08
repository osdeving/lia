# LIA v0.1 — Draft Specification (model-first) + Reprodutibilidade

> **LIA** = *Linguagem Intermediaria Assistida* (nome de trabalho).
> Codinome opcional da linguagem: **Trama** (porque "tece" modulos), mas o nome pode mudar sem alterar o core.

---

## 0) Objetivo

LIA e uma linguagem/IR intermediaria para geracao de software a partir de linguagem natural com foco em:

- **Modularidade linkavel**: "object files" LIA (`.liao`) geraveis em paralelo por multiplos agentes/modelos.
- **Determinismo por construcao**: semantica expressa em constructs formais + verificacao no linker.
- **Qualidade e boas praticas por default**: politicas, constraints e capabilities que impedem estados ruins.
- **Extensibilidade**: plugins como *passes* (transform/verify/synthesize) e packs/stdlib reutilizaveis.
- **Model-first**: a linguagem e desenhada primariamente para consumo/producao por modelos (nao por humanos).
- **Reprodutibilidade**: cada artefato carrega trilha de geracao e decisao (prompt/model/config) + build recipe.

Nao e objetivo (v0.1):

- gerar binario diretamente;
- competir como linguagem geral livre (Java/Python);
- provar correcao de regras de negocio formalmente.

---

## 0.1) Status de implementacao (v0.1)

Esta spec descreve a linguagem **pretendida**. O toolchain atual implementa um subconjunto crescente:

- Parser suporta: `project`, `module`, `use pack`, `repro`, `tape`, mais o core da gramatica (types, enums, ports, usecases/adapters, wiring, constraints, preferencias, effects).
- Implementacao do parser (v0.1): Participle (lexer stateful), conforme ADR 0001.
- Derivacao de simbolos: `provides`/`requires` sao calculados a partir dos modulos.
- Linker resolve `requires -> provides` de forma deterministica (regras v0.1).
- `replay` valida `@gen.prompt_ref` e `@gen.prompt_hash` contra o prompt tape.
- `repro=strict` rejeita holes nao resolvidos no `.lial`.
- Lowerers ainda sao **stubs** (sem codegen real por enquanto).

Veja o **roadmap de implementacao**: `docs/pt-br/ROADMAP.md` (e `docs/en/ROADMAP.md`).

## 1) Artefatos e pipeline

### 1.1 Tipos de arquivo

- `*.lia`  — source LIA (texto authoravel)
- `*.liao` — **LIA object file** (AST canonico + simbolos + constraints + metadados)
- `*.liap` — **LIA pack** (colecao versionada de `.liao` + manifest)
- `*.lial` — **LIA linked unit** ("arquivao final" apos link; ainda LIA, mas resolvido)

### 1.2 Pipeline padrao

1. **NL → geracao de `.liao`** (paralelo)
2. **Validate**: parse + typecheck + constraint checks (por objeto)
3. **Link**: resolver simbolos + aplicar policies + selecionar implementacoes + merge
4. **Lower/Transpile**: `*.lial` → AST do target → emitter (Java/Python/etc.)
5. **Shadow/JIT checks** (opcional): compile/lint/test no target enquanto link/lower acontece

### 1.3 Modo de reprodutibilidade (build profile)

A execucao pode rodar em tres perfis (definidos no `project`):

- `repro = strict`  → determinismo maximo (restricoes mais duras, proibe decisoes nao rastreaveis)
- `repro = pinned`  → dependencias/versionamentos pinados; LLM permitido, mas com "tape" completo
- `repro = best_effort` → produtividade > reprodutibilidade (ainda com logs)

---

## 2) Modelo mental: o que e um object file `.liao`

Um `.liao` representa um ou mais modulos contendo:

- AST canonico
- tabela de simbolos
- tipos
- `requires` e `provides`
- constraints (hard)
- preferencias (soft / scoring)
- metadados RTF (Role-Task-Format)
- **proveniencia e recipe de geracao** (reprodutibilidade)

### 2.1 Identidade e enderecamento por conteudo

Cada `.liao` deve possuir:

- `content_hash`: hash do AST canonico (ex.: SHA-256 do JSON canonico)
- `symbol_hash`: hash da tabela de simbolos
- `api_fingerprint`: hash de exports publicos (para compat/ABI)

Isso permite:

- cache deterministico
- deduplicacao
- tracking de mudancas semanticamente relevantes

### 2.2 Canonicalizacao (regra central)

Para permitir hashing e diffs consistentes, LIA define:

- ordenacao estavel de campos (lexicografica)
- normalizacao de whitespace
- normalizacao de nomes qualificados (QName)
- proibicao de "campos livres" em areas normativas

### 2.3 Proveniencia obrigatoria (reprodutibilidade)

**Todo** modulo (e opcionalmente todo tipo/simbolo publico) deve carregar um bloco `@gen` (metadado estruturado):

- `prompt_ref` (referencia para o prompt usado)
- `prompt_hash` (hash do prompt)
- `model_id` (ex.: `gpt-x.y` ou `llama-…`)
- `model_params` (temperature, top_p, seed quando suportado)
- `context_refs` (packs, docs, regras, arquivos usados)
- `tools_trace_refs` (IDs de chamadas a ferramentas, se houver)
- `generator_pass` (nome/versao do pass que gerou)
- `timestamp` (opcional)

> Observacao: LLMs nem sempre sao deterministicos mesmo com seed. A reprodutibilidade aqui e "**replayable + auditable**": voce consegue repetir a receita, comparar com hashes e entender divergencias.

### 2.4 Prompt Tape (fita de prompts) — artefato de projeto

O projeto pode conter um `prompt_tape` (ou `gen_tape`) versionado:

- mapeia `prompt_ref → prompt_body`
- registra contexto efetivo (packs, versoes, regras, arquivos)
- registra parametros de geracao

O `.liao` so precisa conter `prompt_ref` + hashes; o conteudo do prompt pode ficar centralizado no tape.

---

## 3) Nucleo da linguagem (constructs)

### 3.1 Papeis (roles) de modulo

Papeis explicitos para regras de dependencia e capabilities:

- `domain`
- `usecase`
- `port`
- `adapter`
- `wiring`
- `policy`

Extensao: `role <id>` via packs.

#### 3.1.1 Profiles de arquitetura (nao dogma do core)

Os roles acima **nao tornam a LIA "hexagonal-only"**. Eles sao um vocabulario base.
Arquiteturas como *hexagonal*, *layered*, *clean* e *vertical slice* devem ser
implementadas como **profiles** via packs + policies/constraints, e nao hardcoded no core.

### 3.2 Tipos (minimo)

- primitivos: `Int`, `Bool`, `String`, `Decimal`, `Bytes`, `DateTime`
- ADTs: `enum`, `record`, `option<T>`, `result<T, E>`
- newtypes: `type Email = String where <predicate>`

### 3.3 Efeitos (minimo)

- `pure` (sem IO)
- `io` (pode usar capabilities externas)
- `tx` (boundary transacional)
- `emit` (publica evento — sujeito a policy)

### 3.4 Holes (lacunas) e refinamento monotonico (model-first)

Para permitir geracao incremental e paralela, LIA inclui nos de lacuna **tipados**:

- `hole <name>: <contract>`

Regras:

- holes **nao** podem sobreviver ate `*.lial` (link final) em `repro=strict`.
- o linker pode preencher holes via **synthesize passes** ou selecao de packs.

### 3.5 Multi-candidatos por simbolo (resolucao de conflito)

Um `.liao` pode declarar **candidatos alternativos** para o mesmo slot/simbolo:

- `candidate` com metadados de score e constraints locais.

O linker escolhe deterministamente usando hard constraints + scoring.

---

## 4) Sintaxe (v0.1) — EBNF simplificado

```ebnf
program        := { project | pack | module } ;

project        := "project" Ident "{" { project_item } "}" ;
project_item   := module | policy_decl | constraint_decl | use_decl | repro_decl | tape_decl ;

pack           := "pack" Ident version? "{" { pack_item } "}" ;
pack_item      := module | policy_decl | constraint_decl ;

module         := gen? "module" QName role_decl? "{" { module_item } "}" ;
role_decl      := "as" Ident ;

gen            := "@gen" "{" gen_kv { "," gen_kv } "}" ;

gen_kv         := Ident ":" literal ;

module_item    := type_decl | enum_decl | port_decl | usecase_decl
                | adapter_decl | wiring_decl | constraint_decl | prefer_decl
                | hole_decl | candidate_decl ;

type_decl      := "type" Ident "=" type_ref ("where" expr)? ";" ;

enum_decl      := "enum" Ident "{" Ident { "," Ident } "}" ";"? ;

port_decl      := "port" Ident "{" { method_decl } "}" ;
method_decl    := ("fn" | "method") Ident "(" field_list? ")" return_clause? ";" ;
return_clause  := "->" "(" field_list? ")" ;

usecase_decl   := "usecase" Ident "{" { usecase_item | stmt } "}" ;
adapter_decl   := "adapter" Ident ("implements" QName)? "{" { usecase_item | stmt } "}" ;

usecase_item   := input_block | output_block | effects_decl ;
input_block    := "input" "{" field_list? "}" ";"? ;
output_block   := "output" "{" field_list? "}" ";"? ;
effects_decl   := "effects" "[" Ident { "," Ident } "]" ";"? ;

wiring_decl    := "wiring" Ident "{" { bind_stmt } "}" ;
bind_stmt      := "bind" QName "->" QName ";" ;

constraint_decl:= "constraint" Ident ":" expr ";" ;
prefer_decl    := "prefer" Ident ":" expr ("weight" number)? ";" ;

hole_decl      := "hole" Ident ":" expr ";" ;

candidate_decl := "candidate" QName ("score" number)?
                  ("{" { constraint_decl | ("score" number ";") } "}")? ";"? ;

field_list     := field { "," field } ;
field          := Ident ":" type_ref ;

type_ref       := token { token } ; // parseado como expressao de tipo bruta (sem typecheck)

stmt           := let_stmt | assign_stmt | if_stmt | while_stmt | for_stmt
                | return_stmt | break_stmt | continue_stmt | expr_stmt ;

let_stmt       := "let" Ident "=" expr ";" ;
assign_stmt    := Ident "=" expr ";" ;
if_stmt        := "if" expr "{" { stmt } "}" ("else" "{" { stmt } "}")? ;
while_stmt     := "while" expr "{" { stmt } "}" ;
for_stmt       := "for" Ident "in" expr "{" { stmt } "}" ;
return_stmt    := "return" expr? ";" ;
break_stmt     := "break" ";" ;
continue_stmt  := "continue" ";" ;
expr_stmt      := expr ";" ;

expr           := or_expr ;
or_expr        := and_expr { "||" and_expr } ;
and_expr       := equality { "&&" equality } ;
equality       := comparison { ("==" | "!=") comparison } ;
comparison     := term { ("<" | "<=" | ">" | ">=") term } ;
term           := factor { ("+" | "-") factor } ;
factor         := unary { ("*" | "/" | "%") unary } ;
unary          := ("!" | "-") unary | call ;
call           := primary { call_suffix } ;
call_suffix    := "(" expr_list? ")" | "." Ident | "[" expr "]" ;
expr_list      := expr { "," expr } ;
primary        := Ident | number | string | "true" | "false" | "(" expr ")" | list_literal ;
list_literal   := "[" expr_list? "]" ;
```

---

## 5) Doc/RTF embutido (Role-Task-Format)

Docblocks aceitam tags parseaveis. `@rtf` define um **Prompt Capsule** anexado ao no AST.

Regras:

- Texto de `@rtf` e **nao-normativo**.
- `@rtf` nunca pode relaxar `constraint`/`policy`.
- Em `repro=strict`, `@rtf` so e aceito se estiver associado a um `@gen.prompt_ref` valido.

---

## 6) Policies, constraints, preferencias e budgets

### 6.1 Hard constraints (deterministicas)

Constraints sao expressoes booleanas sobre:

- papeis (`role`)
- dependencias (`imports`)
- efeitos (`effects`)
- capabilities (`capability`)
- (futuro) fluxo de dados

### 6.2 Preferencias (soft; scoring)

Preferencias influenciam selecao entre candidatos validos.

### 6.3 Budgets (anti over-engineering)

Policies podem impor:

- maximo de layers/modulos/deps
- limites de complexidade
- limites de boilerplate

---

## 7) Sistema de simbolos e linking

### 7.1 Identidade de simbolos

Simbolo canonico:

- `QName = <module>::<kind>:<name>`

Cada `.liao` declara:

- `provides` (exports)
- `requires` (imports)

### 7.2 Linker deterministico (v0.1)

1. Load: `.liao` + packs
2. Graph: construir grafo `requires → provides`
3. Filter: remover candidatos que violam hard constraints (subconjunto v0.1)
4. Select: rankear por preferencias e candidate scores
5. Tie-break: regras fixas (score desc, deps asc, id lexical)
6. Merge: `choose-one | rename | wrap | adapt`
7. Verify: revalidar constraints no grafo final
8. Emit: `*.lial` + decision log

### 7.3 Link Trace (proveniencia de decisao)

O linker deve emitir um **Decision Log** contendo:

- selecao por simbolo requerido
- score e criterio de tie-break
- constraints aplicadas

Esse log deve ser canonico e hashavel (replay).

---

## 8) Plugins/passes (extensibilidade oficial)

### 8.1 Tipos de pass

- `verify` (valida; nao altera)
- `transform` (reescreve deterministicamente)
- `synthesize` (preenche holes / gera modulos faltantes)

### 8.2 Contrato de saida (model-first)

Quando um pass usar LLM, ele deve emitir:

- **patch estruturado** (nao "texto livre"), ex.: `lia.patch` (AST delta)
- `@gen` completo (prompt_ref + modelo + params)
- hashes antes/depois

### 8.3 Replay de synth passes

O pass deve conseguir rodar em modo:

- `replay`: reaplica patch gravado
- `regenerate`: tenta reconstruir via prompt (auditable)

Em `repro=strict`, `regenerate` so e aceito se o patch final bater os hashes/constraints esperados.

---

## 9) Shadow/JIT checks (opcional, mas nativo)

Durante link/lower:

- lowering incremental para AST do target
- compile/lint/smoke tests
- feedback rapido para synth passes

O resultado vira diagnosticos que o linker pode usar para preferir candidatos "mais saudaveis".

---

## 10) Lowering/transpilacao (v0.1)

Regras:

- LIA → AST do target → emitter
- mapping por backend (Java/Python)
- policies podem virar interceptors/wrappers/config

---

## 11) Exemplo minimo (com reprodutibilidade)

```lia
project OrderSvc {
  repro strict;
  tape prompt_tape "./gen/prompts.json";

  use pack HexCore@1.0.0;

  @gen { prompt_ref:"p-001", prompt_hash:"…", model_id:"gpt-x", model_params:"temp=0.1" }
  module orders.core as domain {
    enum OrderStatus { NEW, PAID, CANCELLED };
    type OrderId = String where nonEmpty;
  }

  constraint no_domain_io: forbid(effect io) when role == "domain";
}
```

---

## 12) Features adicionais (model-first) para considerar no v0.2

1. **Assumptions & Open Questions**: anotacoes estruturadas `@assumption`, `@open_question` para reduzir alucinacao e facilitar revisao.
2. **Uncertainty budget**: permitir que o modelo declare incerteza e force validacao/testes extras antes do link final.
3. **Semantic diffs**: `lia.patch` como formato oficial para edicoes incrementais.
4. **Safety rails**: proibir que texto livre ("hints/rtf") altere politicas; "policy is code".
5. **Contract tests as artifacts**: permitir declarar testes como parte do IR e exige-los por policy.
6. **Backpressure/Work-stealing**: scheduler de geracao paralela com prioridades por dependency graph.
7. **Content-addressable pack registry**: packs versionados por hash + semver.

---

## 13) Perguntas abertas (para orientar o proximo refinamento)

- O "corpo" de `usecase`/`adapter` sera imperativo minimalista ou declarativo (steps/grafo)?
- Capabilities serao so annotations (v0.1) ou tokens reais (modelo de seguranca forte)?
- O `.liao` canonico sera JSON ou binario (ou ambos)?
- Qual primeiro target (Java/Spring vs Python/FastAPI) para guiar o lowering minimo?
