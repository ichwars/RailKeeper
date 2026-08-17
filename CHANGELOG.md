# Changelog

[Deutsch](CHANGELOG.de.md)

All notable changes to RailKeeper are documented in this file.

## Unreleased

## [0.1.19.1] - 2026-08-17

### Changed

- The three-step vehicle creation dialog now follows the agreed compact desktop and mobile layout,
  including clearer step navigation, restrained actions, responsive article results, and a focused
  review of imported data.
- Sets use a durable, automatically assigned `RK-SET` inventory number and appear as canonical rows
  with ordered member vehicles across table, card, and mobile inventory views.

### Fixed

- Set creation preserves reviewed article fields and images, validates every member, keeps drafts
  isolated per exact username, and respects configured article-search sources.
- Set grouping, sorting, optional columns, filtered member access, duplication, editing, deletion,
  backup restore, and inventory valuation remain consistent across partial and legacy data paths.
- Compact action placement, preview placeholders, column alignment, search sizing, hierarchy guides,
  and narrow-screen behavior now match the approved inventory layout.

## [0.1.19] - 2026-08-16

### Added

- Vehicle creation now guides users through type and core data, optional barcode or web search, and
  the remaining vehicle data in three steps.
- Vehicles that belong together can be created as a set with shared article, acquisition, storage,
  and condition data plus individual inventory and vehicle numbers. Sets appear as expandable
  groups in table and mobile inventory views.
- The bilingual handbook now also documents import and export, general settings, and management of
  general and article-specific master data.

### Changed

- Set creation validates every member before external image retrieval. Shared set data remains
  protected while individual members are edited, and set prices contribute exactly once to
  inventory valuation.
- The OpenAPI contract, backups, and restores include the set model and its ordered members.

### Fixed

- Selected article images survive set creation; ECoS drafts continue through the complete
  persistence path for mappings, CV values, and functions.
- Vehicle quick menus remain fully accessible at limited viewport heights.
- The master-data action column stays visible at the right table edge without causing page-level
  horizontal overflow.

## [0.1.18] - 2026-08-16

### Added

- The vehicle inventory now provides per-user column visibility and ordering, additional inventory
  filters, expanded search, and compact expandable mobile cards.
- Vehicles now record acquisition, purchase price and date, storage, condition, and packaging.
  Vehicle and accessory list and purchase values are evaluated separately and summarized clearly
  on the overview.
- Master data can be deactivated, reactivated, and deleted only when no domain references remain.
  Origin and lifecycle state survive exports and backups.
- The Windows standalone build can show available versions and download the matching ZIP directly.
  Persistent data paths, validated SQLite snapshots, and automatic pre-migration backups protect
  existing installations.
- The bilingual RailKeeper handbook now documents setup, overview, vehicles, maintenance, media,
  decoder/CV data, search, accessories, and exhibition operation.

### Changed

- CSV import and export cover every scalar vehicle field. Unambiguous German and English headers,
  additional aliases, and the existing change preview provide a lossless RailKeeper CSV round trip.
- The vehicle inventory supports `PluX12`; action menus remain visible and keyboard-accessible in
  table and mobile views.

### Fixed

- Overview metrics remain on a single row on wide screens.
- Vehicle quick menus are no longer obscured at table or viewport edges.
- Windows updates separate the program package from user data instead of risking data stored inside
  an overwritten installation directory.

### Security

- A consistent SQLite snapshot is created before database migrations; unsafe or ambiguous Windows
  data paths block startup with an actionable diagnostic.
- Master-data imports and changes preserve referenced historical values and enforce lifecycle rules
  on the server.

## [0.1.17.6] - 2026-08-15

### Changed

- Digital-center setup now guides users through four explicit steps from adapter selection and
  configuration to connection testing and deliberate activation.
- Connection testing, diagnostics, technical details, ECoS messages, and the safety scope now share
  one operational workspace.
- Adapters can be activated, deactivated, and removed completely. The ECoS live monitor remains
  gated by a successful test and an active connection.
- The redesigned workflow adapts to narrow screens without horizontal overflow.

## [0.1.17.5] - 2026-08-15

### Changed

- The layout workspace remains available by direct URL but is no longer shown in the main
  navigation.

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

[0.1.19.1]: https://github.com/ichwars/RailKeeper/compare/v0.1.19...v0.1.19.1
[0.1.19]: https://github.com/ichwars/RailKeeper/compare/v0.1.18...v0.1.19
[0.1.18]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.6...v0.1.18
[0.1.17.6]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.5...v0.1.17.6
[0.1.17.5]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.4...v0.1.17.5
[0.1.17.4]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.3...v0.1.17.4
[0.1.17.3]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.2...v0.1.17.3
[0.1.17.2]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.1...v0.1.17.2
[0.1.17.1]: https://github.com/ichwars/RailKeeper/compare/v0.1.17...v0.1.17.1
[0.1.17]: https://github.com/ichwars/RailKeeper/compare/v0.1.16...v0.1.17
[0.1.16]: https://github.com/ichwars/RailKeeper/compare/v0.1.15...v0.1.16
