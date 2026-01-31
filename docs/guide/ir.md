# IR / AST (v0.1)

The IR is defined as Go structs in `internal/ir/ir.go`.

## Design rules
- Use structs and slices only for normative data.
- Avoid `map[string]any` in canonical areas.
- Sort slices before serialization.

## Key nodes
- `Program` (root)
- `Project`, `Pack`, `Module`
- `TypeDecl`, `EnumDecl`, `PortDecl`, `UsecaseDecl`, `AdapterDecl`
- `ConstraintDecl`, `PreferDecl`, `HoleDecl`, `CandidateDecl`
- `GenMeta`, `RTFMeta`

## Canonical JSON
- implemented in `internal/codec/codec.go`
- output is stable and hashable

## Current status
- Some structures are defined but not fully parsed yet (types, ports, etc.).
- Minimal parser only creates `Project` and `Module` nodes.
