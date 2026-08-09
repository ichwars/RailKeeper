# Changelog

[Deutsch](CHANGELOG.de.md)

All notable changes to RailKeeper are documented in this file.

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

[0.1.16]: https://github.com/ichwars/RailKeeper/compare/v0.1.15...v0.1.16
