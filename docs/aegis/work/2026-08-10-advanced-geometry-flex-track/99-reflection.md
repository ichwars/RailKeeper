# Advanced geometry flex-track reflection

## Outcome

Package G now provides a complete object-level flex-track workflow without weakening RailKeeper's
existing planner invariants. The immutable library owns product length and recommendation, the
object owns only compact path parameters, and the domain remains the single owner of effective
geometry. Preview and persistence remain separate operations.

## What worked

- Introducing effective geometry at the shared domain seam kept snapping, overlap, grade, clearance,
  revision comparison and BOM behavior consistent.
- Red-green tests across domain, application, infrastructure, API and UI prevented backup and fixed
  track regressions.
- Browser acceptance found two environment or boundary failures that isolated tests could not show:
  a stale server binary and an incomplete preview for an extreme coordinate.
- Replacing the stale listener before judging persistence avoided an unnecessary product patch.
- Rejecting an invalid derived suggestion in the application layer preserves a complete API contract
  and prevents a malformed response from terminating the React tree.

## Compatibility and scope

Legacy backups through version 10 remain importable, fixed G1 plans are unchanged and each flex
object still maps to one physical piece. The package does not introduce cutting, leftovers,
multi-object routing or automatic neighbor changes.

The cubic Bézier model is intentionally not called a clothoid or true transition curve. Package H
must introduce transition-curve semantics on top of the effective-geometry seam without changing
the meaning of existing Flex-path schema version 1. Free plan objects remain a separate Package I.

## Verification conclusion

Full backend and frontend suites, the production build, runtime health and the complete browser flow
passed. The new-tab console was clean. Package G is accepted locally on
`dev/issue-36-advanced-geometry`; no push, pull request or merge occurred.
