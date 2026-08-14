# Changelog

[Deutsch](CHANGELOG.de.md)

All notable changes to RailKeeper are documented in this file.

## Unreleased

## [0.1.17.4] - 2026-08-14

### Fixed

- Manufacturer, article number, and storage headers retain their intended widths and no longer
  overlap in the configurable accessory table.

## [0.1.17.3] - 2026-08-14

### Added

- Added repository ownership, contribution, conduct, support, trademark, and third-party notice
  files.
- The accessory table now provides persistent column selection for image, inventory number,
  manufacturer, article number, name, type, gauge, stock, and storage. Inventory number and name
  cannot both be hidden.

### Changed

- New RailKeeper versions are licensed under `AGPL-3.0-only` so modified network-hosted versions
  remain open and provide corresponding source to their users. Existing releases retain their
  previous license terms; AGPL continues to permit commercial use.
- Project funding now links only to GitHub Sponsors and Ko-fi as voluntary tips without benefits,
  paid support, SLA, or special access.
- ECoS integration now reads only locomotive master data, CV values, and static function
  definitions. Write sync remains limited to name, address, and protocol after review and
  confirmation.
- Current speed, direction, active function states, locomotive images, switch objects, routes,
  S88, boosters, and other ECoS object managers were removed from queries and UI contracts.
- Archive, restore, and permanent deletion are now direct, labelled icon actions in table, card,
  and compact views. The previous three-dot menu has been removed.

## [0.1.17.2] - 2026-08-14

### Added

- The layout workspace now provides an interactive digital twin, editable technical positions,
  module ports, track plans, revision previews, material reservations, and track-library exchange.
- Track planning supports verified Tillig geometry, flex tracks, transition curves, elevation
  profiles, free plan objects, clearance checks, grade limits, and connection analysis.
- The accessory overview now offers persistent table and card views plus a compact mobile list.
- Administrators can permanently delete completely unused accessory articles after an explicit
  confirmation.

### Changed

- The layout navigation is enabled again and opens the expanded operational workspace.
- Accessory row and card actions share the same keyboard-accessible menu behavior.

### Fixed

- Imported track manufacturers remain intact and track libraries load consistently.
- Accessory action menus stay visible at table and mobile-card edges, and long card metadata labels
  remain separated.

### Security

- Permanent accessory deletion is enforced server-side for administrators and is rejected whenever
  stock, purchases, movements, reservations, installations, history, assets, or layout references
  exist.

## [0.1.17.1] - 2026-08-13

### Added

- Accessory article search now extracts labelled track specifications such as track system,
  dimensions, direction, roadbed, connections, and digital suitability as individually selectable
  typed suggestions.
- Selecting a gauge automatically fills its configured scale until the scale is edited manually.
- New or previously untagged accessories receive synchronized keyword suggestions from their name,
  manufacturer, article type, and subtype until the keyword field is edited manually.

### Fixed

- The accessory article-search dialog now renders above the editor instead of remaining hidden
  behind its modal layer.
- Gauge multi-selects react to typed option labels and keep Escape scoped to the open option list
  instead of closing the complete article editor.
- Malformed or incompatible track specifications from external pages cannot be selected or imported.

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

[0.1.17.4]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.3...v0.1.17.4
[0.1.17.3]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.2...v0.1.17.3
[0.1.17.2]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.1...v0.1.17.2
[0.1.17.1]: https://github.com/ichwars/RailKeeper/compare/v0.1.17...v0.1.17.1
[0.1.17]: https://github.com/ichwars/RailKeeper/compare/v0.1.16...v0.1.17
[0.1.16]: https://github.com/ichwars/RailKeeper/compare/v0.1.15...v0.1.16
