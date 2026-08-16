# Task Intent Draft: Vehicle Search and Spare Parts Guide

## Requested outcome

Continue the bilingual RailKeeper wiki by publishing the next planned coverage topic,
`vehicle-search-spares`, for stable `v0.1.17.6` and integrate it through a reviewed, green pull
request.

## Scope

- English and German user-guide pages for article search, barcode/EAN, web documents, attachment
  extraction, and vehicle spare parts.
- Coverage status, sidebars, landing pages, and related-page transitions.
- Stable-source audit, documentation checks, independent review, PR, green checks, and merge.

## Non-goals

- No application behavior changes.
- No accessory, master-data, settings, deployment, or administration chapter implementation.
- No changes to the dirty local `main` checkout.
- No documentation of behavior newer than `v0.1.17.6`.

## Risk hints

- External suggestions can be mistaken for authoritative model data.
- Search-result application and remote-image persistence occur at different times.
- Document, extraction, and manufacturer-import batches can partially persist.
- OCR is optional and environment-dependent.
- Some visible and translated search functions are not reachable in the stable vehicle UI.

## Baseline read set hint

Required and acknowledged:

- repository `AGENTS.md`;
- `origin/main` documentation structure and `docs/coverage.json`;
- existing core vehicle, media, maintenance, and decoder/CV chapters;
- stable `v0.1.17.6` frontend article-search, document, media, and spare-parts controllers;
- stable shared dialogs, preferences, view models, routes, API adapter, and translations;
- stable backend article-search, remote import, extraction/OCR, spare-parts, migration, and backup
  sources.

Missing baseline refs: none.

## Baseline usage draft

- Acknowledged before design: all refs above.
- Cited by design: exact primary source owners and compatibility boundary.
- Stable source remains authoritative even though the worktree follows newer `main` documentation.

## Impact statement draft

Documentation-only change. It adds one paired user chapter, marks one existing coverage topic as
documented, and extends documentation navigation. Runtime, API, schema, permissions, and release
behavior remain unchanged.

## Execution readiness view

- Intent lock: finish the next complete bilingual user topic.
- Scope fence: vehicle article search, web documents, and spare parts only.
- Baseline lock: behavior claims must match tag `v0.1.17.6`.
- Owner constraints: preserve existing coverage and published-page ownership.
- Compatibility boundary: no claims from later `main` application work.
- Task batches: spec, plan, coverage gate, paired pages, navigation, audit/review, PR/merge.
- Test obligations: documentation test suite, coverage validation, VitePress build, focused scans.
- Review gates: user spec approval, independent review, exact-head green GitHub checks.
- Drift rule: pause if a required workflow belongs to another coverage owner or differs in stable
  source.
- Evidence before completion: commits, green local checks, clean reviews, merged PR, green Pages.
