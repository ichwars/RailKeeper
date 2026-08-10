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

Implementation evidence will be appended after each verified slice.
