# Stage 1 completion intent

## Requested outcome

Complete every remaining Stage 1 work package in the approved layouts and accessories plan:
#39, #41, #42, #43, and #44. Publish bilingual GitHub pull requests, monitor CI and security
checks, merge verified work, maintain issues and roadmap, and complete the final acceptance audit.

## Scope

- Reservations and immutable installation lifecycle.
- OpenAPI alignment and typed frontend API access.
- German and English accessory workspace.
- German and English layout workspace.
- Backup versioning, backward-compatible restore, role smoke tests, and final acceptance.

## Non-goals

- Graphical track planner or Tillig TT geometry.
- Flex-track or elevation calculations.
- Digital control of a layout.
- Cloud sync, multi-tenant hosting, or public sharing.

## Baseline read set and usage

- `AGENTS.md`: acknowledged and binding.
- `docs/superpowers/specs/2026-08-07-anlagen-zubehoer-gleisplaner-design.md`: approved product baseline.
- `docs/superpowers/plans/2026-08-07-anlagen-zubehoer-etappe-1.md`: approved execution baseline.
- GitHub roadmap #31, parent stage #34, and work packages #39/#41/#42/#43/#44: public status baseline.
- Existing Go, React, SQLite, backup, API, i18n, and role patterns: read per active slice.

No required baseline reference is currently missing.

## Impact statement

This work changes persistent inventory allocation behavior, public API contracts, role-protected
workflows, frontend navigation and data management, and backup compatibility. Data loss,
authorization drift, API mismatch, and UI regressions are the primary risks.

## Execution Readiness View

- Intent lock: deliver approved Stage 1 only.
- Scope fence: #39, #41, #42, #43, #44.
- Compatibility boundary: SQLite remains default, backup v1 remains restorable, local auth data is
  never imported or overwritten.
- Owner constraints: handlers stay thin, application owns business rules, infrastructure owns
  transactions, feature folders own UI, OpenAPI and frontend client remain aligned.
- Retirement boundary: no temporary compatibility path may remain without an explicit trigger.
- Task batches: one or more reviewable PRs per work package.
- Test obligations: targeted regression tests, full Go tests and vet, frontend tests/build, CI,
  CodeQL, and Trivy.
- Review gates: local verification, public PR, all required GitHub checks green, then squash merge.
- Drift rule: pause or revise the slice if a new architecture, cloud dependency, control feature,
  or graphical planner requirement appears.
- Completion evidence: every issue closed, parent checklist complete, final backup and role tests
  green, clean worktree, and verified main commit.

## TDD route guard

- Mode: off.
- Decision: regression-first tests are used where practical, without claiming strict TDD authority.
- Test posture: every behavioral change needs focused automated coverage before merge.
- Verification: package tests first, then complete project checks.
