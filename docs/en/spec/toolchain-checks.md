# Toolchain checks (v0.1)

This document lists the minimum checks that `lia check` and `lia link` must implement to support **profiles** and **packs** such as `ArchBaseline` and `SecurityBaseline`.

## lia check (local validation)

**Objective:** Validate a single `.liao`/`.lial` artifact without resolving external dependencies.

Minimum checklist:

1) **Basic structure**

- `program.version` required.

- `module.name` required and normalized QName.

- `@gen.prompt_ref` requires `@gen.prompt_hash`.

1) **Canonicalization / determinism**

- Orderable slices before canonical serialization.

- Prohibit `map[string]any` in normative IR areas.

1) **Local Rules by Role/Effect**

- Apply local constraints (e.g., `domain` cannot have `io`/`tx`).

- Validate declared effects against roles (e.g., `port` without `io`).

1) **Holes**

- `hole` must have a non-empty `contract`.

- If `repro == strict` and the artifact is `.lial`, no `hole` can remain.

1) **Reproducibility Metadata**

- `@gen` must be structurally valid.

- References to `prompt_ref` must exist in the tape when provided.

1) **Packs (Profiles)**

- Load packs referenced in the project (by default: `./docs/spec/packs` or `./packs`).

- Warn when policies are not applicable in v0.1 (e.g., capabilities/taint not yet modeled).

## lia link (global resolution)

**Objective:** Solve the entire system and produce a deterministic `.lial` file + decision log.

Minimum checklist:

1) **Dependency graph**

- Construct `requires -> provides`.

- Detect missing symbols and collisions.

1) **Global constraints (hard)**

- Apply constraints between modules (e.g., `domain` does not depend on `adapter`).

- Eliminate invalid candidates before scoring.

1) **Deterministic selection**

- Apply preferences (soft) and score.

- Fixed tie-break: semver desc, stability desc, deps asc, lexical.

1) **Explicit merges**

- `choose-one | rename | wrap | adapt` as per policy.

1) **Canonical Decision Log**

- Record choices, scores, and applied constraints.

- Canonical hash of the log.

1) **Reproducibility**

- In `repro=strict`: require full `@gen` on public symbols and emitted decision log.

1) **Profiles (packs)**

- Load packs referenced in the project.

- Unify policies and constraints before the final link.

- In v0.1, apply only the supported subset of the DSL (roles, effects, budgets, deps).

---

## What `ArchBaseline` and `SecurityBaseline` require

- **ArchBaseline**: role enforcement, dependency direction, and effect limits per role.

- **SecurityBaseline**: enforcement of authn/authz policies, input validation at the edge, prohibition of raw SQL, secrets policies, logging, and TLS.

These policies are still pseudo-LIA in v0.1, but they already guide exactly where the toolchain needs to validate and block.
