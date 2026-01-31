# Change Process and Workflow

This document defines the complete lifecycle of changes in the project, from conception (Roadmap) to delivery (Release).

## 1. Roadmap and Planning

The Roadmap (`roadmap.md`) is the source of truth for strategic direction.

- **Curation**: Maintained by *Core Maintainers*.
- **Update**: Reviewed quarterly or after *Milestone Releases*.
- **Proposals**: Any contributor can suggest items via *Issues* with the label `type:roadmap`.

### Feature Cycle (RFC/ADR)

For significant changes (architecture, LIA language changes), we follow the **Specs-First** flow:

1. **RFC (Request for Comments)**:
   - Create an issue `RFC: [Title]` describing the problem and proposal.
   - Open discussion with the community.
2. **ADR (Architecture Decision Record)**:
   - If approved, formalize in `docs/adr/` (one file per decision).
   - Defines the contract before code.
3. **Spec Update**:
   - Update `docs/en/spec/` and `docs/pt-br/spec/` with the new grammar/behavior.
4. **Implementation**:
   - Start coding only after Spec is approved.

## 2. Pull Request (PR) Standard

### PR Template

Every PR description must contain:

```markdown
## Type
- [ ] Feat
- [ ] Fix
- [ ] Docs
- [ ] Spec
- [ ] Test
- [ ] Refactor
- [ ] Perf

## Context
What changed and why? (Link to Issue/RFC)

## Checklist
- [ ] Spec updated (if applicable)
- [ ] Tests added (`go test ./...`)
- [ ] Linting passed (`golangci-lint run`)
- [ ] Reproducibility verified (if AI generated)
- [ ] Coverage did not decrease
```

### Review Policy

- **Mandatory Approval**: At least 1 *Core Maintainer*.
- **CI Checks**: All checks (Test, Lint, Build) must pass (green).
- **No Regression**: Coverage should not decrease.

## 3. Branches and Git Strategy

We follow a simplified **Gitflow** model:

- `main`: Production (stable, tags `vX.Y.Z`).
- `develop`: Integration (unstable, base for features).
- `feat/...`: Feature development (based on `develop`).
- `fix/...`: Bug fixes.
- `release/...`: Preparation for new version (freeze).

**Golden Rule**: Feature PRs must always point to `develop`.

## 4. Release Cycle

1. **Feature Freeze**: Feature merges blocked on `develop`.
2. **Release Candidate**: Tag `v0.1.0-rc1`.
3. **Validation**: Manual tests and mass reproduction.
4. **Final Tag**: Tag `v0.1.0` on `main`.
5. **Automation**: CI triggers GoReleaser to generate artifacts.

---

**References**:

- [AGENT.md](../../AGENT.md): AI Guidelines.
- [CONTRIBUTING.md](CONTRIBUTING.md): Technical guide for devs.
