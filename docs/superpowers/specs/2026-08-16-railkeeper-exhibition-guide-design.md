# RailKeeper Exhibition Guide Design

Date: 2026-08-16

Status: Approved in conversation
Documented release: `v0.1.17.6`

## Objective

Publish a complete English and German user-guide section for RailKeeper's stable exhibition
workspace. The section guides exhibition operators through the full working sequence:

1. select an exhibition list and understand its state;
2. add or maintain locomotive entries while the list is open;
3. detect address conflicts and complete operating details;
4. review, sort, and print the finished list;
5. let an authorized administrator prepare, lock, reopen, or remove lists.

The Messe role is the primary audience. Additional Admin capabilities are identified at the point
where they affect the workflow. The guide explains every stable user-facing action in this scope
without implying access to the general inventory or to unpublished layout features.

## Scope

The section owns these stable user workflows:

- exhibition workspace access and isolation from general inventory;
- list selection, status, entry counts, sorting, viewing, and printing;
- creating, editing, locking, unlocking, and deleting exhibition lists;
- creating, editing, and deleting exhibition entries;
- entry owner, locomotive identity, technical, address, day, image, function, and notes fields;
- optional inventory master-data choices available through combined roles;
- DCC and SX address-conflict detection within the selected list;
- day scopes for all days or selected days one through four;
- image upload, image links, preview, replacement, and removal;
- function keys F0 through F31, names, types, and symbols;
- read-only list view and report printing with or without images;
- open and locked states, confirmation steps, empty states, and request failures;
- the exact stable interaction of Messe and Admin authority;
- local persistence and backup ownership of exhibition lists and entries.

## Non-goals

The section does not explain:

- general vehicle inventory, vehicle editing, or vehicle import;
- administration of manufacturers, epochs, railway companies, types, or function symbols;
- user creation, role assignment, sessions, or authentication administration;
- backup and restore operation beyond accurately stating exhibition-data ownership;
- the unpublished layout workspace;
- printer installation, operating-system print settings, or deployment configuration;
- cloud synchronization, public sharing, or later development features.

Already published vehicle or user-guide pages may be linked where they clarify a boundary. Planned
administration and reference pages are not exposed as active navigation destinations.

## Source and compatibility boundary

Behavior claims come from the stable `v0.1.17.6` tag, not from later `main` development. Primary
source owners include:

- `frontend/src/features/exhibition/ExhibitionView.tsx` and its stable tests;
- `frontend/src/styles/exhibition.css` for responsive and print-related presentation behavior;
- exhibition translations in `frontend/src/shared/i18n/en.ts` and `de.ts`;
- stable app routing, navigation, role gating, API adapters, selects, date input, and function
  symbol helpers;
- `backend/internal/application/exhibition.go` and its tests;
- `backend/internal/api/exhibition.go`, route definitions, role tests, and audit behavior;
- exhibition migrations, OpenAPI operations, vehicle exhibition-flag maintenance, and stable backup
  coverage.

Later exhibition changes on `main` must not be described unless they are already present in
`v0.1.17.6`.

## Information architecture

Create a mirrored three-page section:

- overview and operating sequence: `guide/exhibition/index.md` and
  `de/guide/exhibition/index.md`;
- list administration and locking: `guide/exhibition/lists-and-locking.md` and
  `de/guide/exhibition/lists-and-locking.md`;
- entries, validation, and printing: `guide/exhibition/entries-and-printing.md` and
  `de/guide/exhibition/entries-and-printing.md`.

Suggested titles:

- **Exhibition workspace** / **Messearbeitsbereich**;
- **Lists and locking** / **Listen und Sperren**;
- **Entries and printing** / **Einträge und Drucken**.

Every page uses the established user-guide frontmatter contract, names stable version
`v0.1.17.6`, uses matching English and German section order, and states the review date as
2026-08-16.

## Page 1: Exhibition workspace

The overview page introduces the workspace and explains:

- why the Messe role opens a dedicated operational area instead of general inventory;
- the two-panel layout for exhibition lists and entries;
- list columns for designation, date, entry count, status, and actions;
- entry columns, selection, stable sorting, and responsive behavior;
- the complete sequence from list preparation through entry maintenance, locking, and printing;
- which controls are always available, which require an open list, and which require Admin;
- loading, no-list, no-selection, no-entry, and request-error states;
- the relationship between a selected list and the entries shown beside it.

The page acts as the coverage destination and hub for the two focused subpages. It does not repeat
their field-by-field or administrative details.

## Page 2: Lists and locking

The list page documents:

- the requirement for a Messe-capable account to open the workspace;
- the additional Admin authority required for list management;
- list designation and date validation;
- creation and editing through the stable dialog and **Save** action;
- sorting by designation, date, entry count, and lock status;
- read-only list viewing and printing for Messe users;
- locking an open list and unlocking a locked list;
- the effect of a lock on entry creation, editing, and deletion;
- list deletion, its confirmation, and cascading consequences verified against stable behavior;
- server rejection of stale, missing, invalid, unauthorized, or locked resources;
- refresh and retry guidance after a failed operation.

The text states that an Admin role without Messe does not by itself expose the exhibition route in
the stable UI. Administrative operation therefore requires a suitable combined role assignment.

## Page 3: Entries and printing

The entries page documents the stable entry dialog in its three tabs:

### General

- required owner and locomotive designation;
- manufacturer, type, class, epoch, and railway company;
- all-days or selected-day scope for days one through four;
- DT availability, DCC address, SX address, decoder type, adapter or interface, and analog state;
- number, lettering, features, and notes;
- exact behavior when optional master-data lists are unavailable to a pure Messe account.

### Image upload

- one image source per entry;
- local upload or image link;
- preview, replacement, and removal;
- the fact that a selected image remains part of the entry form until **Save** succeeds;
- safe user guidance for external links without presenting remote content as trusted data.

### Function keys

- keys F0 through F31;
- default F0 light behavior;
- function name, type, and optional symbol;
- which configured functions appear in the table, detail view, and print report;
- stable handling of earlier plain-text function-key data where visible.

The page also explains:

- creating and editing entries only while the selected list is open;
- deleting entries only with Admin authority and confirmation;
- DCC and SX duplicate detection within the current list, excluding the entry being edited;
- case- and whitespace-normalized address comparison;
- sorting by owner, locomotive designation, DT state, decoder number, and function keys;
- the read-only detail view;
- the print dialog, **Print with images** option, A4 landscape report, empty report state, function
  symbols, and final browser print dialog.

## Roles and permissions

The section uses one explicit role matrix:

| Capability | Messe | Messe + Admin |
| --- | --- | --- |
| Open the exhibition workspace | Yes | Yes |
| Read lists and entries | Yes | Yes |
| View and print a list | Yes | Yes |
| Add or edit entries in an open list | Yes | Yes |
| Create or edit lists | No | Yes |
| Lock or unlock lists | No | Yes |
| Delete entries from an open list | No | Yes |
| Delete lists | No | Yes |

An account with Admin but without Messe is not described as having normal UI access to the
workspace. Other combined roles may make general inventory master-data options available inside
the entry dialog, but they do not replace Messe as the exhibition-workspace role.

The stable UI and server route checks are both audited. Where visible controls and server authority
differ, the guide treats the server rule as authoritative and names the stable UI limitation.

## Locking, validation, and persistence

Lists and entries are separate immediate server writes. The guide distinguishes:

| State | Persistence |
| --- | --- |
| List or entry form edits | Dialog state until **Save** succeeds |
| Saved list fields or lock state | Immediate server write |
| Saved entry, image data, and function data | Immediate server write |
| View or print options | Temporary browser state |

Required values, date parsing, day normalization, address conflicts, list locks, and role checks are
documented from stable behavior. A duplicate address warning prevents saving until the conflict is
resolved. A lock prevents all entry mutations, including an Admin deletion, until the list is
unlocked.

The guide does not claim that a later reload failure rolls back a successful earlier write. It
instructs users to reload the list and verify the stored state before retrying an uncertain action.

## Empty, partial, and error states

The pages cover at least:

- workspace loading and list-loading failure;
- no lists, no selected list, and no entries;
- invalid or missing designation, date, owner, or locomotive designation;
- missing optional master-data choices for isolated Messe accounts;
- duplicate DCC or SX address;
- invalid, missing, stale, or locked lists;
- save, delete, lock, unlock, detail-load, and print-data load failures;
- a list becoming locked while an entry dialog is open;
- failed image reading or unavailable external image links;
- empty function configuration and legacy function text;
- unauthorized list management or entry deletion;
- browser print cancellation or printer-specific output differences.

Stable German-only errors visible in the English locale are identified as localization limitations,
not translated into behavior that the UI does not provide.

## Storage and backup

Saved exhibition lists and entries belong to local RailKeeper application data and are included in
the stable app backup. Entry images stored directly with an entry and serialized function settings
follow that entry's persistence. Print options and unsaved dialog changes are not backup data.

The guide recommends a current validated backup before deleting a populated exhibition list or
performing extensive corrections. It does not duplicate the planned backup-and-restore manual.

## Navigation and cross-links

Add an **Exhibition** / **Messe** group to both user-guide sidebars after Accessories. The group
links to the overview and its two subpages.

Add the exhibition overview to both User Guide landing pages. Add related-page links among the
three exhibition pages and from relevant already published pages. Link only to existing published
owners, including:

- User Guide overview;
- vehicle inventory and core records where the exhibition boundary needs clarification;
- Accessories only when distinguishing operational lists from accessory allocations.

Do not create active links to planned Settings, role administration, master-data, backup/restore,
deployment, layout, or reference pages.

## Coverage contract

Change only the `exhibition` topic from `planned` to `documented`. Preserve its established
coverage destination and ownership:

- frontend route `/exhibition`;
- translations below `exhibition`;
- API prefixes below `/api/v1/exhibition-lists`.

The primary English and German coverage paths remain the overview pair. The two additional page
pairs are children of that documented topic.

If the stable source inventory exposes a user-facing exhibition API outside the existing prefix
mapping, update the owner map only when required for accurate validation and when it remains inside
the `exhibition` topic. Do not absorb user administration, master data, vehicle, or layout owners.

## Verification and review

Verification includes:

1. a stable-tag audit for route access, lists, entries, fields, defaults, validation, sorting,
   locking, deletion, images, functions, printing, persistence, audit, and backup;
2. matching English and German frontmatter, page structure, tables, warnings, and cross-links;
3. an intentional missing-page validation failure after changing the coverage status;
4. route, navigation, unfinished-marker, internal-link, line-length, and whitespace scans;
5. `npm.cmd run check` from `docs`, including unit tests, coverage validation, and the VitePress
   production build;
6. independent read-only review of stable-source fidelity, bilingual parity, role boundaries,
   lock behavior, duplicate detection, persistence, printing, and recovery guidance;
7. correction and focused re-review of every valid finding;
8. a pull request from the isolated branch only;
9. merge only when all required checks pass for the reviewed head;
10. post-merge verification of all six English and German GitHub Pages routes.

## Expected outcome

Exhibition operators can prepare and use a list without entering the general inventory workspace.
They know which list they are changing, whether it is open, which operating details belong to each
locomotive, how address conflicts are prevented, and how to produce a useful printed report.
Administrators understand that list management requires a Messe-capable account plus Admin and can
lock or remove lists without obscuring the effect on entry maintenance.
