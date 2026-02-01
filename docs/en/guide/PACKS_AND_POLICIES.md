# Packs, Profiles, Policies

## Packs

A pack is a versioned collection of modules and policies.

In v0.1, packs are **pseudo-LIA** files parsed by a minimal pack parser:

- see `docs/en/spec/packs/*.lia` (or `docs/pt-br/spec/packs/*.lia`)
- the parser supports: `pack`, `policy`, `constraint`

## Profiles

Architecture profiles (hexagonal, layered, clean, vertical slice) are implemented as packs.
The core language stays neutral; the profile is policy.

## Policies and constraints

Policies are structured rules that the toolchain can enforce.
In v0.1, only a subset is enforced:

**Enforced in v0.1**

- roles allowlist
- forbid effects by role
- forbid dependencies by role
- budgets: max modules / max deps

**Not enforced yet**

- capabilities (e.g. http.endpoint)
- taint/flow types
- security policies like SQL injection, SSRF, secrets redaction

## ArchBaseline and SecurityBaseline

- `ArchBaseline` contains architecture constraints and budgets.
- `SecurityBaseline` documents policies that are still warnings in v0.1.
