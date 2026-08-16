# Todo Checkpoint Draft

## Current state

- Topic selected: `vehicle-search-spares`.
- Recommended workflow-first structure selected by the user.
- Three design sections approved by the user.
- Isolated worktree created from current `origin/main` on branch
  `dev/docs-user-guide-search-spares`.
- Baseline documentation check is green.

## Active slice

Write, self-review, and commit the complete design specification and continuity records.

## Todo map

- [x] Determine the next coverage topic.
- [x] Audit stable frontend, backend, routes, translations, storage, and backup boundaries.
- [x] Present alternatives and obtain design approval.
- [ ] Commit and obtain user review of the written specification.
- [ ] Write and commit the detailed implementation plan.
- [ ] Implement coverage, paired pages, navigation, and cross-links.
- [ ] Verify, independently review, correct, and re-review.
- [ ] Push, open PR, await exact-head checks, merge, and verify publication.

## Evidence refs

- Stable tag: `v0.1.17.6`.
- Worktree base: `2e243c66971ef2979892b592852bacd18693c6d9`.
- Baseline: `cd docs; npm.cmd run check`, 19 tests passed, coverage validation passed, VitePress
  build passed.
- Approved design discussion: workflow-first approach plus data-safety and review sections.

## Blocked on

Nothing for the active slice. User review is the next required gate after the spec commit.

## Resume state hint

Read `10-intent.md`, this checkpoint, the committed design specification, and current worktree
status. Do not resume from the dirty local `main` checkout.

## Drift check draft

- Original intent served: yes.
- Scope fence respected: yes.
- Stable compatibility boundary respected: yes.
- New runtime owner or adapter introduced: no.
- Evidence sufficient for next step: yes.
- Decision: continue.
