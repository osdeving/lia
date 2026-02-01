# LIA v0.1 — Draft Specification (model-first) + Reproducibility

> **LIA** = *Assisted Intermediate Language* (working name).

> Optional codename for the language: **Trama** (because it “weaves” modules), but the name can change without altering the core.

---

## 0) Objective

LIA is an intermediate language/IR for generating software from natural language with a focus on:

* **Linkable modularity**: LIA “object files” (`.liao`) that can be generated in parallel by multiple agents/models.

* **Construction determinism**: semantics expressed in formal constructs + verification in the linker.

* **Quality and best practices by default**: policies, constraints, and capabilities that prevent bad states.

* **Extensibility**: plugins such as *passes* (transform/verify/synthesize) and reusable packs/stdlibs.

* **Model-first**: the language is primarily designed for consumption/production by models (not by humans).

* **Reproducibility**: each artifact carries a generation and decision trail (prompt/model/config) + build recipe.

Not the objective (v0.1):

* to generate binaries directly;

* to compete as a general-purpose open-source language (Java/Python);

* to formally prove the correctness of business rules.

---

## 0.1) Implementation status (v0.1)

This spec describes the **intended** language. The current toolchain implements only a subset:

* Parser supports: `project`, `module`, `use pack`, `repro`, `tape`.
* Linker and lowerers are **stubs** (no real graph resolution or codegen yet).

See the **implementation roadmap**: `docs/en/roadmap.md` (and `docs/pt-br/roadmap.md`).

## 1) Artifacts and Pipeline

### 1.1 File Types

* `*.lia` — source LIA (authorable text)
* `*.liao` — **LIA object file** (canonical AST + symbols + constraints + metadata)
* `*.liap` — **LIA pack** (versioned collection of `.liao` + manifest)
* `*.lial` — **LIA linked unit** (final file after link; still LIA, but resolved)

### 1.2 Standard Pipeline

1. **NL → generation of `.liao`** (parallel)
2. **Validate**: parse + type check + constraint checks (per object)
3. **Link**: resolve symbols + apply policies + select implementations + merge
4. **Lower/Transpile**: `*.lial` → target AST → emitter (Java/Python/etc.)
5. **Shadow/JIT checks** (optional): compile/lint/test on the target while link/lower happens

### 1.3 Reproducibility mode (build profile)

Execution can run in three profiles (defined in the `project`):

* `repro = strict` → maximum determinism (strictest restrictions, prohibits untraceable decisions)
* `repro = pinned` → pinned dependencies/versions; LLM allowed, but with full "tape"
* `repro = best_effort` → productivity > reproducibility (still with logs)

---

## 2) Mental model: what is a `.liao` object file

A `.liao` represents one or more modules containing:

* Canonical AST
* Symbol table
* Types
* `requires` and `provides`
* Constraints (hard)
* Preferences (soft / scoring)
* RTF (Role-Task-Format) metadata
* **provenance and generation recipe** (reproducibility)

### 2.1 Identity and content-based addressing

Each `.liao` must have:

* `content_hash`: hash of the canonical AST (e.g., SHA-256 of the canonical JSON)
* `symbol_hash`: hash of the symbol table
* `api_fingerprint`: hash of public exports (for compatibility/ABI)

This allows:

* deterministic caching
* deduplication
* tracking of semantically relevant changes

### 2.2 Canonicalization (core rule)

To allow consistent hashing and diffs, LIA defines:

* stable field ordering (lexicographic)
* whitespace normalization
* qualified name normalization (QName)
* prohibition of “free fields” in normative areas

### 2.3 Mandatory provenance (reproducibility)

**Every** module (and optionally every public type/symbol) must carry an `@gen` block (structured metadata):

* `prompt_ref` (reference to the prompt used)
* `prompt_hash` (hash of the prompt)
* `model_id` (e.g., `gpt-x.y` or `llama-…`)
* `model_params` (temperature, top_p, seed when supported)
* `context_refs` (packs, docs, rules, files used)
* `tools_trace_refs` (IDs of tool calls, if any)
* `generator_pass` (name/version of the generated pass)
* `timestamp` (optional)

> Note: LLMs are not always deterministic even with a seed. Reproducibility here is “**replayable + auditable**”: you can repeat the recipe, compare it with hashes, and understand discrepancies.

### 2.4 Prompt Tape — Project Artifact

The project can contain a versioned `prompt_tape` (or `gen_tape`):

* maps `prompt_ref → prompt_body`
* records effective context (packs, versions, rules, files)
* records generation parameters

The `.liao` file only needs to contain `prompt_ref` + hashes; the prompt content can be centered on the tape.

---

## 3) Language Core (constructs)

### 3.1 Module Roles

Explicit roles for dependency rules and capabilities:

* `domain`
* `usecase`
* `port`
* `adapter`
* `wiring`
* `policy`

Extension: `role <id>` via packs.

#### 3.1.1 Architecture Profiles (not core dogma)

The roles above **do not make LIA “hexagonal-only”**. They are a basic vocabulary.

Architectures such as *hexagonal*, *layered*, *clean*, and *vertical slice* should be
implemented as **profiles** via packs + policies/constraints, and not hardcoded in the core.

### 3.2 Types (minimum)

* primitives: `Int`, `Bool`, `S`
