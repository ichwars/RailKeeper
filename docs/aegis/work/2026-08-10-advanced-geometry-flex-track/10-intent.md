# Advanced geometry flex-track intent

## Requested outcome

Continue GitHub issue 36 locally with the next bounded Stage 4 package. Package G adds a verified
Tillig TT flex-track product, object-specific smooth geometry, an explicit suggestion and acceptance
workflow, and integration with existing planner analysis.

## Scope

- Tillig 83125 flex-track geometry and source-backed constraints.
- Object-specific cubic Bézier parameters and server-derived effective geometry.
- Optional per-layout minimum flex radius that cannot weaken the product recommendation.
- Read-only server suggestion, explicit persisted acceptance, optimistic locking and audit.
- Existing connection, collision, elevation, grade, clearance, revision, BOM and backup integration.
- German and English app-owned UI, API and OpenAPI alignment.

## Non-goals

- True clothoids or mathematically guaranteed transition curves.
- Offcut, cutting-stock or shared-reservation optimization.
- Automatic changes to neighbouring tracks or complete gap routing.
- Free plan objects, flex turnouts, more catalogues or digital layout control.
- Push, pull request, merge or release.

## Baseline read set and usage

- `AGENTS.md`: acknowledged and binding.
- GitHub issue 36: acknowledged as the Stage 4 scope baseline.
- `docs/superpowers/specs/2026-08-07-anlagen-zubehoer-gleisplaner-design.md`: acknowledged and cited.
- Package A through F Stage 4 specs and plans: acknowledged as the compatibility baseline.
- `backend/internal/domain/track_planner.go`: current geometry owner read.
- `backend/internal/domain/track_plan_analysis.go`: current effective analysis path read.
- `backend/internal/application/track_planner.go`: current validation and use-case owner read.
- `backend/internal/infrastructure/track_planner_repository.go`: current persistence owner read.
- `frontend/src/features/layouts/TrackPlannerCanvas.tsx`: current planner interaction owner read.
- Official Tillig 83125 product page and Tillig track-building guidance: acknowledged and cited.

No required baseline reference is currently missing.

## Impact statement

This package changes a persistent planner object, the geometry-definition kind constraint, backup
compatibility, public API contracts, every geometry analysis, material counting and the primary
planner editor. Migration safety, deterministic geometry, client/server drift and incorrect material
claims are the main risks.

## Execution Readiness View

- Intent lock: deliver only the approved Package G design.
- Scope fence: flex-track foundation and explicit suggestion workflow; Packages H and I stay out.
- Baseline lock: existing fixed G1 plans, lineage, reservations and analyses must remain unchanged.
- Owner constraints: domain derives effective geometry, application owns validation and suggestion,
  infrastructure owns the migration/transactions, frontend renders server contract.
- Compatibility boundary: SQLite remains default, backups 1 through 10 remain restorable, fixed
  geometry JSON and existing plan objects remain valid.
- Retirement boundary: no duplicate persistent geometry or second write endpoint may remain.
- Task batches: design, implementation plan, migration/domain, API/persistence, UI, full acceptance.
- Test obligations: migration preservation, domain golden tests, API/OpenAPI, backup legacy/roundtrip,
  frontend interaction tests, full Go/Vitest/build and local browser acceptance.
- Review gates: written spec approval, local plan execution, fresh verification and clean worktree.
- Drift rule: stop and redesign if true clothoids, cutting-stock inventory, new frameworks or
  automatic multi-object routing become necessary.
- Completion evidence: deterministic fixtures, preserved G1 plan, successful backup compatibility,
  complete automated checks and browser proof with no console errors.
