# LIA v0.1 — Draft Specification (model-first) + Reproducibility

> **LIA** = *Assisted Intermediate Language* (working name).

> Optional codename for the language: **Trama** (because it “weaves” modules), but the name can change without altering the core.

---

## 0) Objective

LIA is an intermediate language/IR for generating software from natural language with a focus on:

- **Linkable modularity**: LIA “object files” (`.liao`) that can be generated in parallel by multiple agents/models.
- **Construction determinism**: semantics expressed in formal constructs + verification in the linker.
- **Quality and best practices by default**: policies, constraints, and capabilities that prevent bad states.
- **Extensibility**: plugins such as *passes* (transform/verify/synthesize) and reusable packs/stdlibs.
- **Model-first**: the language is primarily designed for consumption/production by models (not by humans).
- **Reproducibility**: each artifact carries a generation and decision trail (prompt/model/config) + build recipe.

Not the objective (v0.1):

- to generate binaries directly;
- to compete as a general-purpose open-source language (Java/Python);
- to formally prove the correctness of business rules.

---

## 0.1) Implementation status (v0.1)

This spec describes the **intended** language. The current toolchain implements a growing subset:

- Parser supports: `project`, `module`, `use pack`, `repro`, `tape`, plus core grammar (types, enums, ports, usecases/adapters, wiring, constraints, preferences, effects).
- Parser implementation (v0.1): Participle (stateful lexer), per ADR 0001.
- Symbol derivation: `provides`/`requires` are computed from modules.
- Linker resolves `requires -> provides` deterministically (v0.1 rules).
- `replay` validates `@gen.prompt_ref` and `@gen.prompt_hash` against the prompt tape.
- `repro=strict` rejects unresolved holes in `.lial`.
- Lowerers are still **stubs** (no real codegen yet).

See the **implementation roadmap**: `docs/en/ROADMAP.md` (and `docs/pt-br/ROADMAP.md`).

## 1) Artifacts and Pipeline

### 1.1 File Types

- `*.lia` — source LIA (authorable text)
- `*.liao` — **LIA object file** (canonical AST + symbols + constraints + metadata)
- `*.liap` — **LIA pack** (versioned collection of `.liao` + manifest)
- `*.lial` — **LIA linked unit** (final file after link; still LIA, but resolved)

### 1.2 Standard Pipeline

1. **NL → generation of `.liao`** (parallel)
2. **Validate**: parse + type check + constraint checks (per object)
3. **Link**: resolve symbols + apply policies + select implementations + merge
4. **Lower/Transpile**: `*.lial` → target AST → emitter (Java/Python/etc.)
5. **Shadow/JIT checks** (optional): compile/lint/test on the target while link/lower happens

### 1.3 Reproducibility mode (build profile)

Execution can run in three profiles (defined in the `project`):

- `repro = strict` → maximum determinism (strictest restrictions, prohibits untraceable decisions)
- `repro = pinned` → pinned dependencies/versions; LLM allowed, but with full "tape"
- `repro = best_effort` → productivity > reproducibility (still with logs)

---

## 2) Mental model: what is a `.liao` object file

A `.liao` represents one or more modules containing:

- Canonical AST
- Symbol table
- Types
- `requires` and `provides`
- Constraints (hard)
- Preferences (soft / scoring)
- RTF (Role-Task-Format) metadata
- **provenance and generation recipe** (reproducibility)

### 2.1 Identity and content-based addressing

Each `.liao` must have:

- `content_hash`: hash of the canonical AST (e.g., SHA-256 of the canonical JSON)
- `symbol_hash`: hash of the symbol table
- `api_fingerprint`: hash of public exports (for compatibility/ABI)

This allows:

- deterministic caching
- deduplication
- tracking of semantically relevant changes

### 2.2 Canonicalization (core rule)

To allow consistent hashing and diffs, LIA defines:

- stable field ordering (lexicographic)
- whitespace normalization
- qualified name normalization (QName)
- prohibition of “free fields” in normative areas

### 2.3 Mandatory provenance (reproducibility)

**Every** module (and optionally every public type/symbol) must carry an `@gen` block (structured metadata):

- `prompt_ref` (reference to the prompt used)
- `prompt_hash` (hash of the prompt)
- `model_id` (e.g., `gpt-x.y` or `llama-…`)
- `model_params` (temperature, top_p, seed when supported)
- `context_refs` (packs, docs, rules, files used)
- `tools_trace_refs` (IDs of tool calls, if any)
- `generator_pass` (name/version of the generated pass)
- `timestamp` (optional)

> Note: LLMs are not always deterministic even with a seed. Reproducibility here is “**replayable + auditable**”: you can repeat the recipe, compare it with hashes, and understand discrepancies.

### 2.4 Prompt Tape — Project Artifact

The project can contain a versioned `prompt_tape` (or `gen_tape`):

- maps `prompt_ref → prompt_body`
- records effective context (packs, versions, rules, files)
- records generation parameters

The `.liao` file only needs to contain `prompt_ref` + hashes; the prompt content can be centered on the tape.

---

## 3) Language Core (constructs)

### 3.1 Module Roles

Explicit roles for dependency rules and capabilities:

- `domain`
- `usecase`
- `port`
- `adapter`
- `wiring`
- `policy`

Extension: `role <id>` via packs.

#### 3.1.1 Architecture Profiles (not core dogma)

The roles above **do not make LIA “hexagonal-only”**. They are a basic vocabulary.

Architectures such as *hexagonal*, *layered*, *clean*, and *vertical slice* should be
implemented as **profiles** via packs + policies/constraints, and not hardcoded in the core.

### 3.2 Types (minimum)

- primitives: `Int`, `Bool`, `String`, `Decimal`, `Bytes`, `DateTime`
- ADTs: `enum`, `record`, `option<T>`, `result<T, E>`
- newtypes: `type Email = String where <predicate>`

### 3.3 Effects (minimum)

- `pure` (no IO)
- `io` (may use external capabilities)
- `tx` (transaction boundary)
- `emit` (publishes event — policy-governed)

### 3.4 Holes and monotonic refinement (model-first)

To allow incremental and parallel generation, LIA includes **typed holes**:

- `hole <name>: <contract>`

Rules:

- holes **cannot** survive until `*.lial` (final link) in `repro=strict`.
- the linker can fill holes via **synthesize passes** or pack selection.

### 3.5 Multi-candidates per symbol (conflict resolution)

A `.liao` can declare **alternative candidates** for the same slot/symbol:

- `candidate` with metadata score and local constraints.

The linker chooses deterministically using hard constraints + scoring.

---

## 4) Syntax (v0.1) — simplified EBNF

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

type_ref       := token { token } ; // parsed as a raw type expression (not yet type-checked)

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

## 5) Embedded Doc/RTF (Role-Task-Format)

Docblocks accept parseable tags. `@rtf` defines a **Prompt Capsule** attached to the AST.

Rules:

- `@rtf` text is **non-normative**.
- `@rtf` can never relax `constraint`/`policy`.
- In `repro=strict`, `@rtf` is only accepted if associated with a valid `@gen.prompt_ref`.

---

## 6) Policies, constraints, preferences and budgets

### 6.1 Hard constraints (deterministic)

Constraints are boolean expressions about:

- roles (`role`)
- dependencies (`imports`)
- effects (`effects`)
- capabilities (`capability`)
- (future) dataflow

### 6.2 Preferences (soft; scoring)

Preferences influence selection between valid candidates.

### 6.3 Budgets (anti over-engineering)

Policies can impose:

- maximum number of modules/deps
- complexity limits
- boilerplate caps

---

## 7) Symbol system and linking

### 7.1 Symbol identity

Canonical symbol form:

- `QName = <module>::<kind>:<name>`

Each `.liao` declares:

- `provides` (exports)
- `requires` (imports)

### 7.2 Deterministic linker (v0.1)

1. Load: `.liao` + packs
2. Graph: build `requires → provides`
3. Filter: remove candidates that violate hard constraints (v0.1 subset)
4. Select: rank by preferences and candidate scores
5. Tie-break: fixed rules (score desc, deps asc, lexical id)
6. Merge: `choose-one | rename | wrap | adapt`
7. Verify: revalidate constraints in the final graph
8. Emit: `*.lial` + decision log

### 7.3 Link Trace (decision provenance)

The linker must emit a **Decision Log** containing:

- selection per required symbol
- score and tie-break rationale
- constraints applied

This log is canonical and hashable.

---

## 8) Plugins/passes (official extensibility)

### 8.1 Pass kinds

- `verify` (validate; no change)
- `transform` (deterministic rewrite)
- `synthesize` (fill holes / generate missing modules)

### 8.2 Output contract (model-first)

When a pass uses an LLM, it must emit:

- **structured patch** (not free text), e.g. `lia.patch` (AST delta)
- full `@gen` (prompt_ref + model + params)
- before/after hashes

### 8.3 Replay

The pass must support:

- `replay`: apply recorded patch
- `regenerate`: try to rebuild via prompt (auditable)

In `repro=strict`, `regenerate` is accepted only if final hashes and constraints match.

---

## 9) Shadow/JIT checks (optional)

During link/lower:

- incremental lowering to target AST
- compile/lint/smoke tests
- quick feedback for synth passes

The result becomes diagnostics that the linker may use to prefer healthier candidates.

---

## 10) Lowering/transpilation (v0.1)

Rules:

- LIA → target AST → emitter
- mapping per backend (Java/Python)
- policies may become interceptors/wrappers/config

---

## 11) Minimal example (with reproducibility)

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

## 12) Additional features to consider in v0.2

1. **Assumptions & Open Questions**: structured annotations `@assumption`, `@open_question` to reduce hallucination and ease review.
2. **Uncertainty budget**: allow the model to declare uncertainty and force extra validation/tests before final link.
3. **Semantic diffs**: `lia.patch` as the official format for incremental edits.
4. **Safety rails**: prohibit free text ("hints/rtf") from altering policies; “policy is code”.
5. **Contract tests as artifacts**: declare tests as part of IR and require them by policy.
6. **Backpressure/Work-stealing**: parallel generation scheduler with dependency-aware priorities.
7. **Content-addressable pack registry**: packs versioned by hash + semver.

---

## 13) Open questions (for the next refinement)

- Will the `usecase`/`adapter` body be imperative minimalism or declarative steps/graph?
- Will capabilities be annotations (v0.1) or real tokens (strong security model)?
- Will canonical `.liao` be JSON or binary (or both)?
- Which target should guide the minimal lowering (Java/Spring vs Python/FastAPI)?
