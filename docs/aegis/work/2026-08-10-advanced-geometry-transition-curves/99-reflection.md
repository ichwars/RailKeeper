# Advanced geometry transition-curve reflection

## What worked

The effective-geometry seam introduced for free flex paths also supported transition curves without
duplicating analysis logic. A separate path schema kept compatibility explicit, and preview-first
editing made the product-length and layout-radius rules visible before persistence.

Backend-owned integration proved valuable in browser acceptance: left/right routes were exactly
mirrored at every sampled point, while the frontend only rendered the authoritative route. Explicit
`null` fields made conversions auditable and prevented two active path definitions.

## Risks controlled

- Historical flex data stayed untouched through a nullable forward migration.
- Backup compatibility advanced with a bounded legacy normalization.
- Optimistic version checks prevented stale previews from being saved.
- Radius warnings remained visible without silently blocking intentional modelling choices.
- Browser cancellation and reload checks covered accidental writes and persistence regressions.

## Next stage

Package I can add free plan objects independently. It should reuse the planner's coordinate,
selection, versioning and app-owned dialog patterns, but must not enter track connectivity, BOM or
reservation calculations unless the object represents a real inventory-backed track component.
