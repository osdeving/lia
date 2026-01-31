# LIA v0.1 — Draft Specification (model‑first) + Reprodutibilidade

> **LIA** = *Linguagem Intermediária Assistida* (nome de trabalho).
> Codinome opcional da linguagem: **Trama** (porque “tece” módulos), mas o nome pode mudar sem alterar o core.

---

## 0) Objetivo

LIA é uma linguagem/IR intermediária para geração de software a partir de linguagem natural com foco em:

* **Modularidade linkável**: “object files” LIA (`.liao`) geráveis em paralelo por múltiplos agentes/modelos.
* **Determinismo por construção**: semântica expressa em constructs formais + verificação no linker.
* **Qualidade e boas práticas por default**: políticas, constraints e capabilities que impedem estados ruins.
* **Extensibilidade**: plugins como *passes* (transform/verify/synthesize) e packs/stdlib reutilizáveis.
* **Model‑first**: a linguagem é desenhada primariamente para consumo/produção por modelos (não por humanos).
* **Reprodutibilidade**: cada artefato carrega trilha de geração e decisão (prompt/model/config) + build recipe.

Não é objetivo (v0.1):

* gerar binário diretamente;
* competir como linguagem geral livre (Java/Python);
* provar correção de regras de negócio formalmente.

---

## 1) Artefatos e pipeline

### 1.1 Tipos de arquivo

* `*.lia`  — source LIA (texto authorável)
* `*.liao` — **LIA object file** (AST canônico + símbolos + constraints + metadados)
* `*.liap` — **LIA pack** (coleção versionada de `.liao` + manifest)
* `*.lial` — **LIA linked unit** (“arquivão final” após link; ainda LIA, mas resolvido)

### 1.2 Pipeline padrão

1. **NL → geração de `.liao`** (paralelo)
2. **Validate**: parse + typecheck + constraint checks (por objeto)
3. **Link**: resolver símbolos + aplicar policies + selecionar implementações + merge
4. **Lower/Transpile**: `*.lial` → AST do target → emitter (Java/Python/etc.)
5. **Shadow/JIT checks** (opcional): compile/lint/test no target enquanto link/lower acontece

### 1.3 Modo de reprodutibilidade (build profile)

A execução pode rodar em três perfis (definidos no `project`):

* `repro = strict`  → determinismo máximo (restrições mais duras, proíbe decisões não rastreáveis)
* `repro = pinned`  → dependências/versionamentos pinados; LLM permitido, mas com “tape” completo
* `repro = best_effort` → produtividade > reprodutibilidade (ainda com logs)

---

## 2) Modelo mental: o que é um object file `.liao`

Um `.liao` representa um ou mais módulos contendo:

* AST canônico
* tabela de símbolos
* tipos
* `requires` e `provides`
* constraints (hard)
* preferências (soft / scoring)
* metadados RTF (Role‑Task‑Format)
* **proveniência e recipe de geração** (reprodutibilidade)

### 2.1 Identidade e endereçamento por conteúdo

Cada `.liao` deve possuir:

* `content_hash`: hash do AST canônico (ex.: SHA‑256 do JSON canônico)
* `symbol_hash`: hash da tabela de símbolos
* `api_fingerprint`: hash de exports públicos (para compat/ABI)

Isso permite:

* cache determinístico
* deduplicação
* tracking de mudanças semanticamente relevantes

### 2.2 Canonicalização (regra central)

Para permitir hashing e diffs consistentes, LIA define:

* ordenação estável de campos (lexicográfica)
* normalização de whitespace
* normalização de nomes qualificados (QName)
* proibição de “campos livres” em áreas normativas

### 2.3 Proveniência obrigatória (reprodutibilidade)

**Todo** módulo (e opcionalmente todo tipo/símbolo público) deve carregar um bloco `@gen` (metadado estruturado):

* `prompt_ref` (referência para o prompt usado)
* `prompt_hash` (hash do prompt)
* `model_id` (ex.: `gpt‑x.y` ou `llama‑…`)
* `model_params` (temperature, top_p, seed quando suportado)
* `context_refs` (packs, docs, regras, arquivos usados)
* `tools_trace_refs` (IDs de chamadas a ferramentas, se houver)
* `generator_pass` (nome/versão do pass que gerou)
* `timestamp` (opcional)

> Observação: LLMs nem sempre são determinísticos mesmo com seed. A reprodutibilidade aqui é “**replayable + auditable**”: você consegue repetir a receita, comparar com hashes e entender divergências.

### 2.4 Prompt Tape (fita de prompts) — artefato de projeto

O projeto pode conter um `prompt_tape` (ou `gen_tape`) versionado:

* mapeia `prompt_ref → prompt_body`
* registra contexto efetivo (packs, versões, regras, arquivos)
* registra parâmetros de geração

O `.liao` só precisa conter `prompt_ref` + hashes; o conteúdo do prompt pode ficar centralizado no tape.

---

## 3) Núcleo da linguagem (constructs)

### 3.1 Papéis (roles) de módulo

Papéis explícitos para regras de dependência e capabilities:

* `domain`
* `usecase`
* `port`
* `adapter`
* `wiring`
* `policy`

Extensão: `role <id>` via packs.

#### 3.1.1 Profiles de arquitetura (não dogma do core)

Os roles acima **não tornam a LIA “hexagonal-only”**. Eles são um vocabulário base.
Arquiteturas como *hexagonal*, *layered*, *clean* e *vertical slice* devem ser
implementadas como **profiles** via packs + policies/constraints, e não hardcoded no core.

### 3.2 Tipos (mínimo)

* primitivos: `Int`, `Bool`, `String`, `Decimal`, `Bytes`, `DateTime`
* ADTs: `enum`, `record`, `option<T>`, `result<T, E>`
* newtypes: `type Email = String where <predicate>`

### 3.3 Efeitos (mínimo)

* `pure` (sem IO)
* `io` (pode usar capabilities externas)
* `tx` (boundary transacional)
* `emit` (publica evento — sujeito a policy)

### 3.4 “Holes” (lacunas) e refinamento monotônico (model‑first)

Para permitir geração incremental e paralela, LIA inclui nós de lacuna **tipados**:

* `hole <name>: <contract>`

Exemplos de contrato:

* “preciso de um adapter que implemente `OrderRepositoryPort`”
* “preciso de mapping DTO↔Domain”

Regras:

* holes **não** podem sobreviver até `*.lial` (link final) em `repro=strict`.
* o linker pode preencher holes via **synthesize passes** ou seleção de packs.

### 3.5 Multi‑candidatos por símbolo (para resolução de conflito)

Um `.liao` pode declarar **candidatos alternativos** para o mesmo slot/símbolo:

* `candidate` com metadados de score e constraints locais.

O linker escolhe determinísticamente usando hard constraints + scoring.

---

## 4) Sintaxe (v0.1) — EBNF simplificado

```ebnf
program        := { project | pack | module } ;

project        := "project" Ident "{" { project_item } "}" ;
project_item   := policy_decl | import_decl | use_decl | repro_decl | tape_decl ;

pack           := "pack" Ident version? "{" { pack_item } "}" ;
pack_item      := module | policy_decl | constraint_decl ;

module         := doc? gen? "module" QName role_decl? "{" { module_item } "}" ;
role_decl      := "as" Ident ;

gen            := "@gen" "{" gen_kv { "," gen_kv } "}" ;

module_item    := type_decl | enum_decl | port_decl | usecase_decl
                | adapter_decl | wiring_decl | constraint_decl | prefer_decl
                | hole_decl | candidate_decl ;

hole_decl      := "hole" Ident ":" contract_expr ";" ;

// docblocks

doc            := "/**" { doc_line } "*/" ;
doc_line       := "@" Ident doc_payload | text ;
```

---

## 5) Doc/RTF embutido (Role‑Task‑Format) — parte oficial

### 5.1 Conceito

Docblocks aceitam tags parseáveis. `@rtf` define um **Prompt Capsule** anexado ao nó AST.

### 5.2 Tag `@rtf`

Formato mínimo:

* `role` (quem)
* `task` (o que)
* `format` (como entregar)

Campos recomendados (model‑first):

* `inputs` (lista)
* `must` (lista de invariantes)
* `avoid` (lista de anti‑padrões)
* `tests` (lista de testes desejados)
* `repair` (roteiro de correção quando falhar)
* `merge` (sugestão de estratégia; não enfraquece constraints)

### 5.3 Regras de segurança

* Texto de `@rtf` é **não‑normativo**.
* `@rtf` nunca pode relaxar `constraint`/`policy`.
* Em `repro=strict`, `@rtf` só é aceito se estiver associado a um `@gen.prompt_ref` válido.

---

## 6) Policies, constraints, preferências e budgets

### 6.1 Hard constraints (determinísticas)

Constraints são expressões booleanas sobre:

* papéis (`role`)
* dependências (`imports`)
* efeitos (`effects`)
* capabilities (`capability`)
* (futuro) fluxo de dados

Exemplos:

* proibir IO em domain
* proibir dependência domain→adapter

### 6.2 Preferências (soft; scoring)

Preferências influenciam seleção entre candidatos válidos.

### 6.3 Budgets (anti over‑engineering)

Policies podem impor:

* máximo de layers/módulos/deps
* limites de complexidade
* limites de boilerplate

> Em um projeto “simple”, budgets evitam que a stdlib de padrões exploda o shape.

---

## 7) Sistema de símbolos e linking

### 7.1 Identidade de símbolos

Símbolo canônico:

* `QName = <pack?>::<module>::<symbol>`

Cada `.liao` declara:

* `provides` (exports)
* `requires` (imports)

### 7.2 Linker determinístico (v0.1)

1. Load: `.liao` + packs
2. Graph: construir grafo `requires→provides`
3. Filter: remover candidatos que violam hard constraints
4. Select: rankear por preferências (score determinístico)
5. Tie‑break: regras fixas (semver desc, stability desc, deps asc, id lexical)
6. Merge: `choose-one | rename | wrap | adapt`
7. Verify: revalidar constraints no grafo final
8. Emit: `*.lial` + relatório (trace)

### 7.3 Link Trace (proveniência de decisão)

O linker deve emitir um **Decision Log** canônico contendo:

* seleção de cada símbolo (candidato escolhido)
* score por critério
* constraints aplicadas
* tie‑break usado
* merges executados

Esse log deve ser hashável e versionado (replay).

---

## 8) Plugins/passes (extensibilidade oficial)

### 8.1 Tipos de pass

* `verify` (valida; não altera)
* `transform` (reescreve deterministicamente)
* `synthesize` (preenche holes / gera módulos faltantes)

### 8.2 Contrato de saída (model‑first)

Quando um pass usar LLM, ele deve emitir:

* **patch estruturado** (não “texto livre”), ex.: `lia.patch` (AST delta)
* `@gen` completo (prompt_ref + modelo + params)
* hashes antes/depois

### 8.3 Replay de synth passes

O pass deve conseguir rodar em modo:

* `replay`: reaplica patch gravado
* `regenerate`: tenta reconstruir via prompt (auditable)

Em `repro=strict`, `regenerate` só é aceito se o patch final bater os hashes/constraints esperados.

---

## 9) Shadow/JIT checks (opcional, mas nativo)

Durante link/lower:

* lowering incremental para AST do target
* compile/lint/smoke tests
* feedback rápido para synth passes

O resultado vira diagnósticos que o linker pode usar para preferir candidatos “mais saudáveis”.

---

## 10) Lowering/transpilação (v0.1)

Regras:

* LIA → AST do target → emitter
* mapping por backend (Java/Python)
* policies podem virar interceptors/wrappers/config

---

## 11) Exemplo mínimo (com reprodutibilidade)

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

## 12) Features adicionais (model‑first) para considerar no v0.2

1. **Assumptions & Open Questions**: anotações estruturadas `@assumption`, `@open_question` para reduzir alucinação e facilitar revisão.
2. **Uncertainty budget**: permitir que o modelo declare incerteza e force validação/testes extras antes do link final.
3. **Semantic diffs**: `lia.patch` como formato oficial para edições incrementais (melhor para agentes do que reescrever arquivos).
4. **Safety rails**: proibir que texto livre (“hints/rtf”) altere políticas; “policy is code”.
5. **Contract tests as artifacts**: permitir declarar testes como parte do IR e exigí-los por policy.
6. **Backpressure/Work‑stealing**: scheduler de geração paralela com prioridades por dependency graph.
7. **Content-addressable pack registry**: packs versionados por hash + semver.

---

## 13) Perguntas abertas (para orientar o próximo refinamento)

* O “corpo” de `usecase`/`adapter` será imperativo minimalista ou declarativo (steps/grafo)?
* Capabilities serão só annotations (v0.1) ou tokens reais (modelo de segurança forte)?
* O `.liao` canônico será JSON ou binário (ou ambos)?
* Qual primeiro target (Java/Spring vs Python/FastAPI) para guiar o lowering mínimo?
