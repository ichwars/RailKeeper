# Changelog

[Deutsch](CHANGELOG.de.md)

All notable changes to RailKeeper are documented in this file.

## [0.1.17] - 2026-08-13

### Added

- Article-data and barcode search are now available directly while creating or editing accessories,
  with source previews and explicit field-by-field selection before values are applied.
- Selected product images from article-search results can be imported as private accessory
  documents after saving the article.
- GitHub sponsorship links now include Buy Me a Coffee and PayPal.

### Changed

- The former `Article Overview` area is now named `Accessories` while its existing `/accessories`
  route remains stable.
- Shared article-search and barcode dialogs now use app-owned modal, keyboard, focus-trap, and focus
  restoration behavior in both vehicle and accessory workflows.
- Updated `modernc.org/sqlite` to 1.56.0, `lucide-react` to 1.29.0, and Vite to 8.2.1.

### Fixed

- Unknown or inactive manufacturer and gauge suggestions can no longer be selected when the
  corresponding master-data value is unavailable.
- Remote image-import retries are idempotent, preserve existing primary images, and return an
  already committed document without downloading it again.
- Escape and Tab remain scoped to the active nested search dialog instead of reaching the accessory
  editor behind it.

### Security

- Remote accessory images are restricted to public HTTP(S) targets with URL, DNS, redirect, MIME,
  and attachment-size validation before they are stored.

## [0.1.16] - 2026-08-09

### Added

- Complete article catalogue and inventory with generated inventory numbers, quantity or individual
  tracking, controlled attributes, purchases, and stock units.
- Storage-location stock, transactional stock movements, reservations, installations, condition
  history, removals, documents, and consolidated usage history.
- Sortable, selection-first article overview with product images and operational stock summaries.
- Layout, module, setup, plan-variant, and plan-revision foundations with version-conflict protection.
- Backup format version 3 for the extended article data while retaining import support for backup
  versions 1 and 2.

### Changed

- Article forms now use configured master-data dropdowns for manufacturers, gauges, article types,
  subtypes, stock units, and controlled additional fields.
- Article overview metrics use prominent totals with quieter supporting values.
- Article editing, stock booking, reservation, installation, and track-specific forms use compact,
  aligned multi-column layouts.
- Article master data is consolidated under `Settings > Data` with separate general and article
  groups.
- The main-navigation item is named `Layout` and remains visible but temporarily disabled while the
  workspace is refined. Direct `/layouts` access and the layout API remain available.

### Fixed

- Preserved transactional stock and allocation invariants across retries, transfers, reservations,
  installations, removals, and restore operations.
- Preserved inactive controlled attributes and historical values in editing and backups.
- Hardened article inputs, document validation, search queries, legacy updates, and focus restoration
  after confirmation dialogs.
- Retained backup version 1 and 2 compatibility without weakening version 3 completeness checks.

### Security

- Tightened API validation and protected master-data import invariants.
- Kept backup restore preflight conservative and authentication data excluded from exports.

[0.1.17]: https://github.com/ichwars/RailKeeper/compare/v0.1.16...v0.1.17
[0.1.16]: https://github.com/ichwars/RailKeeper/compare/v0.1.15...v0.1.16
