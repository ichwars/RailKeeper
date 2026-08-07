# Stage 1 completion reflection

## Outcome

The approved Stage 1 foundation is implemented across persistence, application services, HTTP and
OpenAPI contracts, typed frontend access, bilingual accessory and layout workspaces, backup version
2, role boundaries, and operational acceptance. No graphical planner or digital-control behavior was
introduced.

## What materially improved the result

- Reviewable slices kept application, persistence, API, typed client, accessory UI, and layout UI
  changes independently verifiable.
- Real browser and HTTP acceptance found two issues that unit compilation alone would not expose: an
  unstable plan-loading callback and an unauthorized Messe master-data request.
- A complete Stage 1 backup fixture exposed the immediate SQLite self-reference constraint during
  restore. Deferring foreign keys within the existing transaction fixed the canonical restore path
  without weakening commit-time integrity.
- Public issues and bilingual PRs kept the user-visible roadmap synchronized with delivered work.

## Compatibility and retirement result

- Backup version 1 remains supported explicitly; version 2 is the only current export format.
- Authentication, roles, sessions, password data, local settings, and audit/security state remain
  outside application backup and restore.
- No temporary repository, API, UI, or backup fallback remains.
- Graphical track planning is the next stage, not a hidden partial implementation in Stage 1.

## Final decision

Local acceptance is complete. The work may close after the final bilingual PR passes CI, CodeQL, and
Trivy, merges to `main`, and GitHub issues #44 and #34 plus roadmap #31 are updated.
