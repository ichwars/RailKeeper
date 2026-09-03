# Changelog

[Deutsch](CHANGELOG.de.md)

All notable changes to RailKeeper are documented in this file.

## Unreleased

## [0.1.20.4] - 2026-09-03

### Changed

- Exhibitions have a separate A4 landscape print document. It includes every saved entry in the
  selected event with images, model and control data, operating days, availability, status, complete
  function assignments, and notes. Screen filters do not limit the report; navigation and controls
  are excluded.
- Z21 provides a read-only preview of known locomotive addresses with time and item limits.
  Intellibox transport selection and Z21 diagnostics distinguish available capabilities from
  planned ones more clearly. These flows do not write to the command station.

### Fixed

- Printing waits for images and remains disabled while a newly selected exhibition is loading.
  Legacy free-text function assignments are preserved. Hidden row actions no longer cause
  page overflow in the exhibition view.
- ECoS rejects incomplete compatibility replies and distinguishes them from complete replies
  with firmware that has not yet been verified.
- Concurrent edits and autosaves in the layout twin preserve other unsaved drafts. Position
  history, conflict reloads, and stale replies after switching modules were corrected.
- The track planner safeguards dragging, gauge validation, material analysis, and transactional
  snapping. Module-port previews account for unsaved placements; dimension changes cannot leave
  existing ports outside module bounds.

## [0.1.20.3] - 2026-08-28

### Added

- Märklin CS3 and CS3 Plus can read their locomotive rosters through the known read-only endpoints
  into the Digital centers workspace. Diagnostics, comparison, conflict review, and RailKeeper
  vehicle creation or assignment remain separate from writes to the command station.
- Vehicle CSV files recognize complete sets, set memberships, and individual member labels. The
  import preview presents set changes explicitly before applying them.
- Master-data records can be bulk-deactivated through an accessible multi-selection workflow.

### Changed

- Vehicle and accessory tables use persistent, resizable columns. The mobile vehicle view,
  required-field cues, contextual exhibition controls, typography, design tokens, and reduced
  motion handling were consolidated.
- Article-search status messages distinguish real progress more clearly from indeterminate
  processing phases. Temporary working documents were removed from `docs/`, and documentation
  governance was clarified.

### Fixed

- Set member labels and unmapped existing values survive CSV replacement. Set rows and selection,
  data, and action columns remain aligned after resizing.
- CS3 vehicle drafts and external mappings preserve the `cs3` provider instead of being stored
  incorrectly as ECoS mappings.

### Security

- CS3 HTTP targets are restricted to private LAN addresses before any request. Loopback,
  link-local, public, and mixed DNS targets are rejected, validated DNS results are pinned for the
  request, and redirects remain disabled.
- The runtime image enforces updated Alpine OpenSSL packages.

## [0.1.20.2] - 2026-08-24

### Changed

- Import and export profiles show direction, status, areas, and format in one consistently aligned
  table.
- Start actions use compact, accessible icons. Long CSV mappings and review results remain fully
  scrollable and usable throughout the import dialogs.
- Go, SQLite, frontend, and GitHub Actions dependencies were updated.

### Fixed

- Removed table offsets and cramped lower sections in the import/export workspace.
- Import reviews can proceed through confirmation even with extensive CSV mappings.

### Security

- Administrators can permanently delete cancelled transfer jobs only after confirmation. The
  restriction is enforced server-side and the action is recorded in the audit log.

## [0.1.20.1] - 2026-08-24

### Added

- Vehicle sets can have a dedicated main image used in inventory, mobile views, and set dialogs,
  with a focused upload workflow for managing it.
- The Digital centers workspace can safely adopt reviewed ECoS locomotives and, after explicit
  authorization, create missing locomotives on the device through a controlled workflow.

### Changed

- Import and export profiles are listed together with explicit direction, areas, format, and
  status. Vehicle CSV files support all 62 transfer fields and persisted, user-reviewed column
  mappings.
- Navigation and personal workspace visibility synchronize more consistently.
- Backups use format version 20 and include vehicle-set main-image assignments.

### Fixed

- Partial and legacy vehicle CSV files preserve every unmapped existing field during replacement.
  Incomplete mappings, unknown boolean values, and contradictory profile data are rejected before
  apply.
- CSV protection markers survive export and reimport without data loss. Version-1 JSON packages
  are processed strictly through their historical schema.
- ECoS reads, vehicle assignment, write conflicts, and result handling are hardened against stale,
  incomplete, or ambiguous device state.

### Security

- ECoS writes remain limited to explicitly authorized differences and are verified on the device
  after writing.
- CSV formula injection, multipart upload order, and transactional import apply are handled without
  silent data mutation.

## [0.1.20] - 2026-08-22

### Added

- The redesigned overview brings together inventory metrics, maintenance, value trends, and stock
  composition in compact cards and charts. Detail dialogs and direct filter links lead from each
  metric to the affected vehicles and accessories.
- Persistent data-transfer profiles and an auditable job history support versioned exports, import
  previews, conflict decisions, and transactional application of reviewed vehicle, accessory, and
  exhibition data.
- The **Digital centers** workspace provides durable sessions, locomotive worklists, desired-versus-
  actual comparisons, and passive live telemetry. Writes require an explicit, time-limited grant
  followed by final confirmation.
- Exhibition operation has a dedicated responsive workspace for lists, entries, lock states, and
  deliberately approved conflict exceptions.
- The independently authored **RailKeeper Werkstatt-Linie** replaces all bundled function
  graphics. Its 94 symbols provide active, inactive, and print palettes for the 86 known ECoS
  function-description codes plus neutral fallback types.

### Changed

- Data transfer, digital centers, and exhibition use compact, bilingual, responsive interfaces
  with app-owned controls and persistently stored work state.
- Backups use format version 19, include data-transfer profiles and exhibition conflict exceptions,
  and preserve custom symbol keys. Restoring an older backup keeps the currently installed bundled
  symbol library protected.
- ECoS codes remain solely as neutral compatibility metadata. RailKeeper includes no manufacturer
  function graphics or graphic source files.

### Security

- Server-side authorization enforces role and exhibition boundaries for the new workspaces. Import
  preview, path validation, atomic export claiming, and transactional application treat transfer
  files as untrusted input.
- Digital centers remain passive by default. Writes are limited to reviewed differences,
  short-lived grants, and explicit user confirmation.

## [0.1.19.2] - 2026-08-19

### Changed

- Vehicle sets are collapsed by default in the inventory. Mixed member categories and types are
  summarized explicitly, while the hierarchy, selection controls, and set dialogs use the compact
  aligned layout.
- Article-search results prioritize candidates with the most extracted fields, use score as a
  stable tie-breaker, and present more restrained result actions on desktop.

### Fixed

- Set data can be edited and saved through the full set editor with the same validated dropdowns
  used by individual vehicles.
- Set values now act as creation defaults instead of overwriting member records. Individual set
  vehicles can keep and edit their own category, type, model, technical, storage, and condition
  data without changing the canonical set.
- Set summary and editor dialogs, member alignment, checkboxes, hierarchy guides, and responsive
  layouts render consistently on desktop and mobile.

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

[0.1.20.2]: https://github.com/ichwars/RailKeeper/compare/v0.1.20.1...v0.1.20.2
[0.1.20.1]: https://github.com/ichwars/RailKeeper/compare/v0.1.20...v0.1.20.1
[0.1.20]: https://github.com/ichwars/RailKeeper/compare/v0.1.19.2...v0.1.20
[0.1.19.2]: https://github.com/ichwars/RailKeeper/compare/v0.1.19.1...v0.1.19.2
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
