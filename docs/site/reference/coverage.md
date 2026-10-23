---
title: What is documented and reviewed?
description: What is documented and reviewed?
audience: reference
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# What is documented and reviewed?

All 20 topics now have an English and a German entry page. This is content coverage, not a claim that every device, firmware or edge case has been fully tested.

## Read the evidence

| Evidence | Meaning | Limit |
| --- | --- | --- |
| Source review | Instructions checked against current code or operating rules | Not every variant has been run separately |
| Demo interface | Also illustrated by real screenshots from an isolated test installation | Synthetic data, no connected command station |
| Hardware verification | Only where a chapter explicitly identifies a concrete device test | Applies to that device and firmware |

Pictures were captured on **23 October 2026** with synthetic vehicles. Their CV values are not factory settings. Pages show their version scope and editorial review date. Their Git modification date appears separately at the bottom.

## Topics and evidence

<CoverageTable />

## Which practical workflows have additional evidence?

| Exercise | Actual evidence |
| --- | --- |
| [F0/F1/F3](../guide/vehicles/functions-walkthrough) | Actual local adoption, combined save, provenance and manual protection |
| [Restore](../administration/restore-exercise) | Real export, validation and restore with CVs/functions/image/maintenance; local authentication preserved; invalid version blocked |
| [Search and spare parts](../guide/vehicles/search-walkthrough) | Search with explicitly simulated replies; actual local TXT extraction |
| [Sets](../guide/vehicles/sets) and [workshop](../guide/accessories/workshop-exercise) | Shared set edit, real local stock transactions and maintenance completion |
| [Operating day](../guide/exhibition/practice) | Conflict dialog, corrected entry, saved lock and generated report |
| [Accounts and two-factor](../administration/account-security-example) | Separate Viewer account, role boundary, real TOTP activation and fresh sign-in |

All exercises use synthetic data. QR code and authenticator secret are masked during capture.
No actual backup or credentials are publicly downloadable.

## What remains open?

- CS3 compatibility uses documented response forms and fixtures; a complete project hardware acceptance is not established.
- Z21 and Intellibox 3 do not yet provide an ECoS-equivalent locomotive/CV import. The guide does not invent that capability.
- The decoder library retains explicit model, firmware and source gaps. A catalogue entry does not imply interpretation of every CV on that model.
- Windows/Docker migrations, SMTP delivery, physical cameras/scanners and native printer/PDF dialogs are not accepted by this demo. Additional network-failure and restore scenarios need separate checks.

## Keeping it current

The topic list and evidence come from [docs/coverage.json](https://github.com/ichwars/RailKeeper/blob/main/docs/coverage.json). Automated checks inventory routes, API areas, translation keys and configuration variables. They verify language pairs, metadata, picture files and ownership by handbook topics. Content accuracy and hardware behaviour require separate review.

A workflow change updates its guide and both languages in the same pull request. In `docs`, `npm run check` runs tests, ownership checks and the site build. [Report a gap](https://github.com/ichwars/RailKeeper/discussions/categories/q-a).

[The compatibility table](command-station-compatibility) lists functions, evidence and pending physical tests per connection. Z21 has bounded known-address comparison; Intellibox 3 remains limited to testing and diagnostics.
