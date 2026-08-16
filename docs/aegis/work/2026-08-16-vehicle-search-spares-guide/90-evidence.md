# Evidence Bundle Draft

## Baseline

- `origin/main` refreshed before topic selection.
- Isolated branch and worktree created without changing local `main`.
- Documentation dependencies installed only in the isolated worktree.
- Baseline `npm.cmd run check`: 19 tests passed, coverage validation passed, VitePress production
  build completed.

## Stable-source evidence

- Normal search criteria, EAN-only exception, local field/image application, and selection defaults
  confirmed in the stable vehicle article-search controller and dialog.
- Camera security, permission, digit normalization, and minimum detected length confirmed in the
  stable barcode dialog.
- Ten-second search timeout and ten-result cap confirmed in the stable article-search service.
- Document discovery, URL-based import detection, category inference, sequential imports, and
  refresh behavior confirmed in the stable document controller and remote attachment handler.
- Attachment extraction order, 12 MiB read bound, 80-suggestion cap, optional OCR, automatic
  sequential import, and deduplication confirmed in stable API and application sources.
- Manual spare-part validation, identity, conservative create merge, immediate actions, lookup cap,
  availability checking, and Piko/Roco import order confirmed in stable frontend and service code.
- Viewer read/search routes and Editor write routes confirmed in stable route specifications.
- Stored spare parts and media confirmed in application backup scope.

## Review state

Design discussion, written-spec self-review, and written-spec user review completed. The detailed
implementation plan passed a spec-coverage, placeholder, type/name, line-hygiene, and scope review.
`npm.cmd run check` passed again with 19 tests, coverage validation, and the VitePress build.

Implementation commits added the paired chapter and coverage, then navigation and cross-links.
Stable-source audit confirmed search, camera, remote import, extraction, spare-part, role, and
backup behavior against `v0.1.17.6`.

Independent read-only review found no Critical or Important issue. Three Minor findings were
corrected: legacy source-default migration, extraction success/failure refresh and tab timing, and
clear German wording for trimmed whitespace. Focused re-review reported no remaining finding and
declared the branch ready after commit and fresh verification.
