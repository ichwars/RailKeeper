# Advanced geometry flex-track evidence

## Design evidence

- GitHub issue 36 lists flex-track optimization, transition curves and additional free plan objects;
  the latter two are explicitly separated into later packages.
- Tillig product page for 83125 identifies the wooden-sleeper flex track and an approximate length of
  664 mm for TT, 1:120 and 12 mm gauge.
- Tillig track-building guidance recommends flex track for radii above R543 and as compensation
  pieces.
- Current `TrackGeometryKind` excludes `flex`; migration 0045 constrains the same values in SQLite.
- Current plan objects reference one immutable geometry definition and carry pose plus two heights.
- Current analysis, snapping, collision, grade and clearance paths consume definition routes and
  definition length directly, so effective geometry must be introduced at their shared domain seam.
- Current reservation and BOM behavior is object-based, supporting the explicit Package G rule that
  one flex object consumes one physical piece.

## Current verification baseline

- Package F commit: `7d370b2 feat(planner): show track clearance warnings`.
- `go test ./...`: 7 packages passed before Package G design work.
- `npm.cmd test -- --run`: 66 files and 356 tests passed before Package G design work.
- `npm.cmd run build`: 2,173 modules passed before Package G design work.
- Local runtime: `/health` returned 200 on port 18083; browser remained logged in as `codex-test`.

## Final implementation evidence

- `f1e9588`: migration 0052, TILLIG 83125, layout limit, flex-path column and backup version 11.
- `60c2cd2`: deterministic effective Bézier route, length, radius and shared analysis consumption.
- `3047b43`: versioned preview endpoint, persistence, clone behavior and OpenAPI contract.
- `757823d`: app-owned minimum-radius input, payload, profile display and bilingual copy.
- `1e1c279`: app-owned Flex editor, local preview, server confirmation, effective canvas, inspector,
  snapping and analysis warning.
- `62a489f`: regression test and service validation for non-derivable extreme previews.

## Final automated verification

- `go test ./...`: all seven backend packages passed after `62a489f`.
- `npm.cmd test -- --run`: 68 files and 363 tests passed on the final frontend implementation.
- `npm.cmd run build`: 2.173 modules built successfully.
- `git diff --check`: no whitespace errors.
- Known non-blocking build output: Vite native-config-loader future notice and large-chunk warning.

## Browser acceptance

- Runtime: `127.0.0.1:18083`, listener PID 25984, `/health` HTTP 200, user `codex-test`.
- Layout limit: 700,00 mm persisted and survived reload.
- Valid preview: endpoint 500/100 mm at 20°, length 512,55 mm, radius 1.118,29 mm, applicable.
- Radius warning: endpoint 300/100 mm at 90°, length 370,51 mm, radius 86,85 mm against 700,00 mm,
  visible and deliberately applicable.
- Length rejection: endpoint 700/0 mm at 0°, length 700,00 mm against 664,00 mm, apply disabled.
- Invalid extreme input remains in the dialog with a validation error and no app crash.
- Saved valid path survived reload with the same length and radius.
- BOM contains one TILLIG 83125 for the single flex object.
- A newly opened browser tab reported an empty console log.

## Publication boundary

All evidence and implementation remain local on `dev/issue-36-advanced-geometry`. No push, pull
request, merge or release was performed.
