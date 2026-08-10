# Advanced geometry transition-curve evidence

## Implementation evidence

- `b1fe7ec feat(planner): derive transition curve geometry`
- `45443d4 feat(planner): persist transition curve paths`
- `1fa84b1 feat(api): preview transition curve paths`
- `661a3de feat(planner): edit transition curve paths`
- Migration 0053 adds nullable `transition_path_json` without rewriting existing Bézier data.
- Backup version 12 preserves transition paths and normalizes version-11 documents to `NULL`.
- OpenAPI includes path, preview input/output and transition fields on plan objects and writes.

## Automated verification

- `go test ./...`: all seven backend packages passed.
- Targeted frontend run: 2 files and 10 tests passed.
- `npm.cmd test -- --run`: 69 files and 366 tests passed.
- `npm.cmd run build`: 2.176 modules transformed successfully.
- `git diff --check`: no whitespace errors before documentation commit.
- Non-blocking build notices: future Vite native-config-loader behavior and large chunk warning.

## Runtime evidence

- Local URL: `http://127.0.0.1:18083/layouts`.
- `/health`: HTTP 200 with `{"status":"ok"}`.
- Listener after restart: PID 14512.
- Authenticated user: `codex-test`.
- Current migration and frontend build were loaded after replacing the prior listener PID 25984.

## Browser acceptance

- Left preview at 500 mm and 700 mm: effective length 500,00 mm, minimum radius 700,00 mm,
  applicable, 101 sampled points, endpoint approximately `493,66 / 58,98` mm.
- Right preview with the same inputs: all sampled X coordinates match and all Y coordinates are the
  exact sign mirror, endpoint approximately `493,66 / -58,98` mm.
- Length 665 mm against the 664 mm product length: explicit warning and disabled apply action.
- Radius 600 mm against the 700 mm layout limit: explicit warning and enabled apply action.
- Saved left path survived a full page reload with the same 500,00 mm, 700,00 mm and endpoint.
- Previewing right and cancelling preserved the saved left path.
- BOM remained one TILLIG 83125 for the single flex object.
- A freshly opened `/layouts` tab reported no console errors.

## Publication boundary

All implementation, data changes and evidence remain local on `dev/issue-36-advanced-geometry`.
No push, pull request, merge or release occurred.
