# Advanced geometry transition-curve intent

## Goal

Complete Package H of issue 36 locally by adding true Euler transition curves to existing TILLIG
83125 flex objects without changing the meaning of free Bézier paths or rigid track geometry.

## Constraints

- Keep `flexPath` and `transitionPath` mutually exclusive.
- Derive all effective geometry in the backend domain.
- Preserve one physical 83125 per plan object and the 664 mm maximum usable length.
- Use the stricter radius of product and layout limits.
- Keep preview separate from an explicit, versioned save.
- Preserve German and English UI, backup compatibility and the OpenAPI contract.
- Keep every change local on `dev/issue-36-advanced-geometry`.

## Acceptance target

Mirrored left and right Euler paths, deterministic persistence and restore, app-owned editing,
blocking overlength, visible radius warning, path preservation during pose/height changes, complete
automated verification and browser evidence.
