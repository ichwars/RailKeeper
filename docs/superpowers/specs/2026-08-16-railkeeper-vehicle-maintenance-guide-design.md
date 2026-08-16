# RailKeeper Vehicle Maintenance User Guide Design

**Date:** 2026-08-16
**Status:** Approved design
**Documented release:** v0.1.17.6
**Audience:** RailKeeper users

## Context

RailKeeper already documents first setup, the overview, core vehicle records, and vehicle media in
English and German. The next planned user topic in `docs/coverage.json` is vehicle maintenance.
Stable v0.1.17.6 provides a complete maintenance workspace inside the vehicle editor, but its
persistence and deletion behavior includes details that are easy to miss and can affect other
unsaved vehicle data or linked media.

The new chapter must explain the entire stable maintenance workflow without absorbing the separate
media, decoder/CV, document-search, spare-part, dashboard, or administrative backup workflows.

## Goals

- Publish a semantically equivalent English and German maintenance chapter.
- Document only behavior present in stable RailKeeper v0.1.17.6.
- Explain every visible maintenance field, stored value, validation rule, summary counter, sort
  rule, and write action.
- Make immediate persistence and the full-record refresh after maintenance writes explicit.
- Explain linked-media counts and the safe sequence before deleting a maintenance entry.
- Describe role boundaries, empty states, failures, backup scope, and specialist boundaries.
- Make the chapter discoverable from the sidebar, guide landing pages, and related vehicle pages.
- Keep the coverage contract honest and fully validated.

## Non-goals

- Do not document later `main` behavior as stable.
- Do not explain image and attachment upload workflows again.
- Do not explain decoder/CV, document search, remote imports, or spare-part workflows.
- Do not turn dashboard reminders into a second maintenance chapter.
- Do not document recurring schedules, automatic intervals, notifications, or work orders because
  v0.1.17.6 does not provide them.
- Do not explain administrative backup and restore operation beyond the user-facing safety boundary.
- Do not change runtime, API, validation, storage, or deletion behavior.
- Do not add screenshots or links to unpublished pages.

## Stable source of truth

Every factual claim is checked against tag `v0.1.17.6`, especially:

- `frontend/src/features/vehicles/VehicleMaintenanceTab.tsx`
- `frontend/src/features/vehicles/useVehicleMaintenanceController.ts`
- `frontend/src/features/vehicles/vehicleMaintenance.ts`
- `frontend/src/features/vehicles/vehicleOptions.ts`
- `frontend/src/features/vehicles/VehiclesView.tsx`
- `frontend/src/features/vehicles/useVehicleEditorController.ts`
- `frontend/src/features/vehicles/VehicleUploadsTab.tsx`
- `frontend/src/features/vehicles/useVehicleMediaController.ts`
- `frontend/src/shared/i18n/en.ts`
- `frontend/src/shared/i18n/de.ts`
- `backend/internal/api/routes.go`
- `backend/internal/api/vehicle_operation_handlers.go`
- `backend/internal/application/vehicle_maintenance_service.go`
- `backend/internal/application/vehicle_validation.go`
- `backend/internal/application/vehicle_media.go`
- `backend/migrations/0012_vehicle_maintenance.sql`
- `backend/migrations/0016_maintenance_media_links.sql`

Current `main` may be used only for documentation structure and navigation patterns.

## Page pair and metadata

Create:

- `docs/site/guide/vehicles/maintenance.md`
- `docs/site/de/guide/vehicles/maintenance.md`

The English page uses:

```yaml
---
title: Vehicle maintenance and condition
description: Record, schedule, complete, review, and safely remove vehicle maintenance entries.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---
```

The German page uses:

```yaml
---
title: Fahrzeugwartung und Zustand
description: Fahrzeugwartungen erfassen, planen, abschließen, prüfen und sicher entfernen.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---
```

The public routes are `/guide/vehicles/maintenance` and
`/de/guide/vehicles/maintenance`.

## Chapter structure

Both pages use the same semantic order:

1. Open the Maintenance tab / Tab Wartung öffnen
2. Read the maintenance summary / Wartungsübersicht lesen
3. Add a maintenance entry / Wartungseintrag hinzufügen
4. Fields, values, and validation / Felder, Werte und Validierung
5. Understand due dates and completion / Fälligkeit und Abschluss verstehen
6. Edit or cancel an entry / Eintrag bearbeiten oder abbrechen
7. Mark an entry done / Eintrag als erledigt markieren
8. Check linked media / Verknüpfte Medien prüfen
9. Delete maintenance safely / Wartung sicher löschen
10. Roles, storage, and backup boundaries / Rollen-, Speicher- und Sicherungsgrenzen
11. Empty and error states / Leere und fehlerhafte Zustände
12. Related pages / Verwandte Seiten
13. Documented RailKeeper version / Dokumentierte RailKeeper-Version

## Entry prerequisite and roles

A vehicle must exist before maintenance can be recorded. A new unsaved vehicle displays the stable
empty message and does not offer a working maintenance form until the core vehicle record has been
saved.

Admin, Editor, Viewer, and Planner can inspect maintenance entries. The server permits create,
update, complete, and delete operations only for Admin and Editor. The chapter treats visible
disabled controls as a UI aid, not as the authorization boundary.

## Maintenance summary

The top of the tab shows three counters:

- **Due / Fällig:** entries with a due date on or before the current local date whose status is not
  `erledigt`.
- **Planned/open / Geplant/offen:** every entry whose status is not `erledigt`, including entries
  already counted as due.
- **Done / Erledigt:** entries whose status is `erledigt`.

The due counter is therefore a subset of the planned/open counter, not a separate status total.
The red due presentation comes from the due date and completion state, even if the stored status is
still `geplant`.

The stable list order is:

1. non-completed entries before completed entries;
2. entries with due dates before undated entries within those groups;
3. ascending due date;
4. newest creation first when the earlier keys are equal.

The overview dashboard and vehicle-inventory maintenance filters are transitions to already
published chapters. Their complete workflows are not repeated here.

## Fields and stored values

The creation and edit form contains these fields.

### Type

The stored German values are:

- `Wartung`
- `Reparatur`
- `Umbau`
- `Superung`
- `Reinigung`
- `Schmierung`
- `Decoder-Einbau`
- `Ersatzteiltausch`

The English UI translates these labels but sends the same stored values. The default is `Wartung`.
No arbitrary type can be stored through the API.

### Status

The stable values are `geplant`, `faellig`, and `erledigt`. The UI displays translated labels and
normalizes the German spelling `fällig` to stored `faellig`. The default is `geplant`.

The manually selected `faellig` status and the date-based due highlighting are related but not the
same rule. A planned entry with a due date up to today is counted and highlighted as due.

### Condition

Condition is optional. Allowed stored values are:

- `neuwertig`
- `sehr gut`
- `gut`
- `gebraucht`
- `reparaturbedürftig`

These stable values remain German even in the English interface.

### Dates

**Due on / Fällig am** and **Completed on / Durchgeführt am** are optional date-only fields. The
server accepts valid calendar dates in `YYYY-MM-DD` form and rejects invalid dates.

When a user saves an entry with status `erledigt` and no completion date, the frontend supplies the
current local date. The quick completion action follows the same rule and preserves an existing
completion date.

### Cost

Cost is optional and must be a non-negative decimal amount. Comma and point decimal separators are
accepted. Surrounding spaces and a trailing euro sign are removed before validation. Values that
cannot be parsed as a non-negative number are rejected. The stable UI formats accepted numeric
values as EUR using the German locale.

### Notes

Notes are optional, trimmed by the server, and limited to 4,000 characters.

## Create, edit, cancel, and complete

**Add entry / Eintrag hinzufügen** creates a new row immediately. On success, RailKeeper reloads
the selected vehicle and resets the form to `Wartung` and `geplant`.

The edit action copies all stored entry values into the same form. The primary action becomes
**Save entry / Eintrag speichern** and a cancel action appears. Cancel resets the form and leaves the
stored entry unchanged.

The card action **Done / Erledigt** immediately updates an open entry without opening a separate
confirmation. It preserves type, condition, due date, cost, notes, and an existing completion date.
If the completion date is empty, it uses the current local date.

## Persistence and full-record refresh

Maintenance writes are not part of the vehicle form's **Save changes / Änderungen speichern**
operation. Create, update, quick completion, and delete are immediate API calls.

Every successful maintenance write calls `refreshSelectedVehicle()`. This replaces the selected
vehicle and the main vehicle form with fresh server data and reloads feature-specific editor state.
Consequently, unrelated unsaved core data, image metadata, attachment edits, and other pending tab
changes can be discarded.

The chapter must therefore tell users to save or intentionally discard all other pending vehicle
changes before any maintenance write action.

If a maintenance write fails, the controller shows the server error and does not reset the
maintenance form. The user can correct the values or connection and retry.

## Linked media

Each maintenance card counts linked images and linked attachments separately under
**Linked media / Verknüpfte Medien**.

Image links are managed in the **Uploads** tab. Changing an image link is pending vehicle metadata
and becomes durable only through **Save changes / Änderungen speichern**. The media chapter remains
the authority for that workflow.

Stable v0.1.17.6 can display attachments that already carry a maintenance reference, but the normal
attachment row does not provide a control for assigning or changing that reference. The maintenance
chapter states this limitation instead of inventing a workflow.

## Safe deletion

The delete icon removes the maintenance entry immediately. Stable v0.1.17.6 provides no additional
confirmation dialog for this action.

The backend does not block deletion when the card reports linked media. It also does not delete or
detach those media records. Their stored maintenance identifier can therefore point to a deleted
entry.

Before deleting maintenance, the user must:

1. finish or save all unrelated vehicle edits;
2. inspect the linked image and attachment counts;
3. clear image links in **Uploads** by choosing **No maintenance / Keine Wartung** and saving the
   vehicle;
4. keep the maintenance entry when a linked attachment must remain associated, or deliberately
   remove the attachment after backup and content review because v0.1.17.6 has no attachment-link
   editor;
5. return to the maintenance tab, verify that linked counts are clear, and only then delete.

Deleting maintenance does not delete the linked image or attachment file itself. The risk is the
stale relationship and, for an image, later deletion protection caused by a non-empty maintenance
identifier.

## Storage and backup boundary

Maintenance rows are local RailKeeper application data. They and vehicle uploads belong to the
application backup scope. The chapter recommends a current, successfully validated backup before
substantial cleanup, while leaving export, validation, the 250 MiB restore-upload limit, and restore
operation to the later administration chapter and the published media safety guidance.

The chapter makes no promise about external copies or downloaded files.

## Empty and error states

Both language pages cover at least:

| Situation | Documented response |
| --- | --- |
| Vehicle is not saved | Save the core record before adding maintenance. |
| No maintenance exists | Explain the stable empty state and use **Add entry**. |
| Input is rejected | Check type, status, condition, dates, non-negative cost, and note length. |
| Write action fails | Keep the form open, read the error, check session and connection, then retry. |
| Summary looks outdated | Reload the vehicle after confirming that no unsaved changes remain. |
| Due count differs from status | Explain that due is calculated from date and completion, not only status. |
| Linked media is shown | Resolve links before deletion and use the media chapter for image metadata. |
| Maintenance was deleted without a prompt | Explain immediate deletion and recovery from a suitable backup. |

No page should suggest that an error has rolled back an earlier successful independent action.

## Navigation and cross-links

Add exact sidebar entries immediately after the media chapter:

```ts
{ text: "Vehicle maintenance and condition", link: "/guide/vehicles/maintenance" }
```

```ts
{ text: "Fahrzeugwartung und Zustand", link: "/de/guide/vehicles/maintenance" }
```

Add concise links to both User Guide landing pages. Add the maintenance page to the related-page
lists of the core vehicle, media, and overview chapters where it improves the existing transition.
The maintenance page links back only to already published pages:

- User Guide overview
- Overview, metrics, and data quality
- Vehicle inventory and core records
- Vehicle images and attachments

Do not link to the planned decoder/CV, search-and-spares, or administration pages.

## Coverage contract

Change only the `vehicle-maintenance` topic from `planned` to `documented`. Its existing paths and
owners remain correct:

- translations below `vehicles.maintenance`
- APIs below `/api/v1/vehicles/{id}/maintenance`

The plan must first demonstrate the missing-page gate by changing the status before creating the
pages. The expected failure is limited to the absent English and German maintenance files.

## Verification and review

Implementation verification includes:

1. direct stable-tag checks for translations, route access, field options, validation, summary
   calculation, list ordering, completion date, refresh behavior, media references, and deletion;
2. exact frontmatter and matching section-order checks;
3. placeholder, em-dash, whitespace, and link scans;
4. `npm.cmd run check` from `docs`, including all 19 unit tests, coverage validation, and the
   VitePress production build;
5. an independent read-only review of English/German parity, source fidelity, persistence timing,
   role boundaries, due logic, linked-media safety, deletion, and navigation;
6. correction and re-review of every Critical or Important finding and valid completeness findings;
7. a pull request only from the isolated feature branch;
8. merge only when CI, Trivy, and CodeQL succeed for the exact reviewed head SHA and every review
   thread is resolved.

## Expected outcome

RailKeeper users can understand and safely operate every stable vehicle-maintenance function without
mistaking due calculation for status, losing unrelated unsaved edits, or deleting a maintenance
entry before dealing with linked media. The paired pages remain focused, discoverable, and honest
about v0.1.17.6 limitations.
