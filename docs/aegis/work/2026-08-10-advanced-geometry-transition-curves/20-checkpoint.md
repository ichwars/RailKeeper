# Advanced geometry transition-curve checkpoint

## Implemented stages

- `b1fe7ec`: deterministic Euler-spiral domain geometry and revision comparison.
- `45443d4`: migration 0053, repository mapping, revision cloning and backup version 12.
- `1fa84b1`: versioned preview service, Planner route, CSRF/role tests and OpenAPI schemas.
- `661a3de`: app-owned bilingual dialog, SVG preview, explicit conversion and active-path preservation.

## Key decisions retained

- Linear curvature is represented by length, end radius and left/right direction.
- Simpson integration produces a deterministic route with at most 5 mm segments.
- Overlength is previewable but not applicable; radius below the effective limit remains applicable
  with an explicit warning.
- `null` is used only in frontend write payloads to remove the competing path. Loaded plan objects
  remain strictly typed and contain at most one path.
- Existing snapping, collision, clearance, grade, diff and BOM logic consume effective geometry and
  therefore need no transition-specific duplicate path.

## Boundary

Package H does not add automatic S-curves, full gap routing, flex switches, cutting optimization or
free decorative plan objects. Free plan objects remain Package I.
