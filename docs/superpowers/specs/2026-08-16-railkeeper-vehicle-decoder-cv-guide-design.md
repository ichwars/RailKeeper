# RailKeeper Vehicle Decoder, Functions, and CV Data User Guide Design

**Date:** 2026-08-16
**Status:** Approved design
**Documented release:** v0.1.17.6
**Audience:** RailKeeper users

## Context

RailKeeper already documents first setup, the overview, core vehicle records, vehicle media, and
vehicle maintenance in English and German. The next planned user topic in `docs/coverage.json` is
the connected decoder workflow. Stable v0.1.17.6 exposes this workflow through the vehicle editor
tabs **Control**, **Speed Curve**, and **CV**.

These tabs share data and actions. Function mappings can come from JSON or a decoder-file preview,
the speed curve is derived from stored or previewed CV values, and decoder files can independently
suggest metadata, CV values, and functions. One complete page preserves those relationships better
than separate pages would.

## Goals

- Publish a semantically equivalent English and German chapter.
- Document only behavior present in stable RailKeeper v0.1.17.6.
- Explain every visible digital-function, speed-curve, CV-value, and decoder-file workflow.
- Distinguish previews, immediate writes, local exports, file storage, and read-only calculations.
- Make full-record refreshes, non-atomic multi-row actions, and deletion behavior explicit.
- Explain ECoS as an input path without absorbing digital-center administration or write-back.
- Describe role boundaries, validation, stored-language peculiarities, backup scope, and failures.
- Make the chapter discoverable from the sidebar, guide landing pages, and related vehicle pages.
- Keep the documentation coverage contract accurate and fully validated.

## Non-goals

- Do not document later `main` behavior as stable.
- Do not repeat the general vehicle decoder fields already covered by the core vehicle chapter.
- Do not explain ECoS connection setup, locomotive synchronization, or writes to a digital center.
- Do not claim that the Speed Curve tab programs a decoder or command station.
- Do not provide manufacturer-specific programming instructions or recommended CV values.
- Do not explain general vehicle images and attachments as decoder files.
- Do not document search, spare parts, administration, or full backup and restore operation.
- Do not change runtime, API, validation, import, storage, or deletion behavior.
- Do not add screenshots or working links to unpublished pages.

## Stable source of truth

Every factual claim is checked against tag `v0.1.17.6`, especially:

- `frontend/src/features/vehicles/VehicleEditorDialog.tsx`
- `frontend/src/features/vehicles/VehicleFunctionsTab.tsx`
- `frontend/src/features/vehicles/VehicleSpeedCurveTab.tsx`
- `frontend/src/features/vehicles/VehicleCVTab.tsx`
- `frontend/src/features/vehicles/useVehicleFunctionsController.ts`
- `frontend/src/features/vehicles/useVehicleCVController.ts`
- `frontend/src/features/vehicles/useVehicleDecoderFilesController.ts`
- `frontend/src/features/vehicles/cvImport.ts`
- `frontend/src/features/vehicles/speedCurve.ts`
- `frontend/src/features/vehicles/vehicleFiles.ts`
- `frontend/src/features/vehicles/vehicleViewModel.ts`
- `frontend/src/shared/i18n/en.ts`
- `frontend/src/shared/i18n/de.ts`
- `backend/internal/api/routes.go`
- `backend/internal/api/vehicle_decoder_handlers.go`
- `backend/internal/application/vehicle_functions_service.go`
- `backend/internal/application/vehicle_cv_service.go`
- `backend/internal/application/vehicle_validation.go`
- `backend/internal/application/backup.go`
- `backend/internal/application/ecos.go`
- `backend/migrations/0013_vehicle_functions.sql`
- `backend/migrations/0014_vehicle_cv_values.sql`
- `backend/migrations/0020_esu_function_symbols.sql`
- `backend/migrations/0030_vehicle_cv_protocol.sql`
- `backend/migrations/0031_cv8_manufacturers.sql`

Current `main` may be used only for documentation structure and navigation patterns.

## Page pair and metadata

Create:

- `docs/site/guide/vehicles/decoder-cv.md`
- `docs/site/de/guide/vehicles/decoder-cv.md`

The English page uses:

```yaml
---
title: Decoder, functions, and CV data
description: Map digital functions, inspect speed curves, manage CV values, and store decoder files.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---
```

The German page uses:

```yaml
---
title: Decoder, Funktionen und CV-Daten
description: Digitalfunktionen zuordnen, Fahrkurven prüfen, CV-Werte pflegen und Decoder-Dateien speichern.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---
```

The public routes are `/guide/vehicles/decoder-cv` and `/de/guide/vehicles/decoder-cv`.

## Chapter structure

Both pages use the same semantic order:

1. Prerequisites and access rights / Voraussetzungen und Zugriffsrechte
2. Map digital functions F0-F31 / Digitalfunktionen F0-F31 zuordnen
3. Read the speed curve / Fahrkurve lesen
4. Manage CV values manually / CV-Werte manuell verwalten
5. Import and export CV values / CV-Werte importieren und exportieren
6. Preview, apply, and store decoder files / Decoder-Dateien prüfen, übernehmen und speichern
7. Use an ECoS preview as an input path / ECoS-Vorschau als Eingangspfad verwenden
8. Protect data during immediate and multi-row writes / Daten bei direkten und mehrzeiligen Schreibvorgängen schützen
9. Troubleshoot decoder data / Fehler bei Decoderdaten beheben
10. Related pages / Verwandte Seiten
11. Documented RailKeeper version / Dokumentierte RailKeeper-Version

The chapter follows the actual editor-tab order while cross-referencing dependent actions rather
than explaining them twice.

## Prerequisites and roles

The core vehicle record explains the **Digital**, primary decoder number, DT decoder number,
decoder type, and ABC-braking fields. The new chapter links back to that page and starts with the
specialist tabs.

A vehicle must normally have been saved once before function, CV, or decoder-file data can be
persisted because those records need a stored vehicle ID. An ECoS-created unsaved draft is the one
preview boundary: RailKeeper can display imported draft CV values and derive a speed curve, but the
normal write and file actions remain unavailable until the vehicle exists.

Admin, Editor, Viewer, and Planner can read stored functions, CV values, CV history, and decoder
files. Viewer-level access can also export functions and CV values and download decoder files.
Create, update, import, preview-upload, upload, apply-suggestion, and delete operations require
Admin or Editor. The server-side route access is authoritative; disabled controls are only the UI
representation of that boundary.

## Digital functions F0-F31

The **Control** tab presents **Digital functions** for F0 through F31. The summary counts configured
rows, sound rows, and light rows. **Only assigned** filters out rows with no persisted or locally
entered name, symbol, or note.

Each row contains:

- function key F0-F31;
- name;
- symbol;
- mode, stored as `dauer` or `moment` and translated as Continuous/Momentary or Dauer/Moment;
- direction-dependent/inverted switch;
- note;
- immediate save and delete actions.

The selected symbol supplies its label when the name is empty and determines the function type.
Stable stored types are `standard`, `sound`, `licht`, `kupplung`, `rauch`, and
`sonderfunktion`. The English UI translates those meanings, while the stored identifiers remain
unchanged. A new F0 row defaults to the name `Fahrlicht`, the light symbol, and type `licht`; the
other empty rows default to type `standard`. Every row defaults to mode `dauer`.

A new row cannot be saved when name, symbol, and note are all empty. The server accepts only
F0-F31, limits the name to 120 characters, the symbol key to 80 characters, and notes to 1,000
characters, and validates the type and mode lists.

### Function JSON export

Export is available for a saved vehicle with at least one configured row. It downloads
`<inventory-number>-funktionen.json`, or `railkeeper-funktionen.json` when the inventory number is
empty. The JSON includes vehicle inventory number, name, decoder number, and configured mappings.
The decoder value prefers the primary digital decoder number and falls back to the DT decoder
number.

### Function JSON import

Import reads the first selected JSON file. It accepts either a top-level array or a `functions` or
`functionMappings` array, normalizes function keys to uppercase, filters invalid keys, types, and
modes, and writes the valid rows sequentially. Duplicate keys are not removed, so a later mapping
for the same key overwrites the earlier mapping. Import has no row-by-row preview and no
confirmation step. If no valid mappings remain, the stable UI reports a German error even in
English mode.

Saving or deleting one function persists immediately and reloads the full selected vehicle. Delete
has no additional confirmation dialog.

The visible labels **Import** and **Export** remain English in both stable language modes. The
chapter identifies this as a UI peculiarity, not a different workflow.

## Read-only speed curve

The **Speed Curve** tab is a read-only calculation from actual stored CV values or an ECoS draft.
It never writes to a decoder, a digital center, or RailKeeper data.

RailKeeper groups speed-relevant CV values by the combination of decoder profile and protocol.
Users can switch among these profiles. Each summary shows curve mode, CV 29 status, plotted point
count, and forward/reverse trim.

The chapter explains:

- the three-point curve from CV 2 at step 1, CV 6 at step 14, and CV 5 at step 28;
- the 28-point speed table from CV 67 through CV 94;
- bit 4 of CV 29 as the known selector between speed table and three-point curve;
- forward trim from CV 66 and reverse trim from CV 95;
- the chart's real CV values and the lists of missing CVs.

When CV 29 is absent, or its selected curve has no points, RailKeeper selects the most complete
useful curve: a complete 28-point table, then a three-point set with at least two values, then any
available table or three-point data. The page describes the user-visible result without presenting
this fallback as decoder configuration.

## Manual CV values

The **CV** tab summarizes stored values, decoder profiles, and decoder files. Its manual form
contains:

- CV number, integer 1-1024;
- value, integer 0-255;
- optional category;
- optional protocol;
- optional decoder profile;
- optional source decoder file;
- optional description.

The stable category options are stored and displayed in German: `Adresse`, `Fahrverhalten`,
`Motor`, `Licht`, `Sound`, `Funktion`, `Decoder`, and `Sonstiges`. Protocol options include the
stable Motorola, DCC, LGB, and Selectrix choices. A decoder profile is free text. Common profile
names are offered, and profiles already present in CV values or files appear as reusable shortcuts.
RailKeeper does not validate that the free-text profile matches the physical decoder.

The identity used for automatic upsert and import comparison is CV number plus normalized decoder
profile. Protocol is not part of that identity. Adding a value with an existing identity updates the
existing row instead of creating another one. A source file can be assigned only when that file
belongs to the same vehicle.

Save, edit, and delete are immediate API actions. History is added only when an update changes the
numeric value; metadata-only changes create no history item. The stable UI displays the five newest
history entries for a CV row. Delete has no additional confirmation dialog.

## CV import and export

CV import accepts the first selected JSON, CSV, or TXT file. JSON can be a top-level array or an
object with `cvValues`. Text formats use one row per line with semicolon or comma separators in the
order CV number, value, description, category, and decoder profile. A header beginning with `cv` is
ignored. Text import does not populate protocol or source file.

Before writing, RailKeeper displays every parsed row as:

- **new**, selected by default;
- **changed**, selected by default;
- **same**, not selected by default;
- **invalid**, unavailable for selection.

Duplicate identities inside one import are invalid after the first occurrence. **Only new** selects
only new rows; the user can also select all valid rows, select none, or toggle rows individually.
Applying the selection updates matching rows and creates new rows sequentially.

Export downloads `<inventory-number>-cv.json`, or `railkeeper-cv.json` without an inventory number.
The JSON contains vehicle inventory number, name, the same preferred decoder-number fallback as the
function export, and every stored CV value including its metadata and history returned by the
vehicle record.

Some stable validation and completion messages remain German in English mode, and import status
messages inside the preview are stored German text. The English page states this limitation so
users can recognize the same outcomes.

## Decoder-file preview, application, and storage

The decoder-file picker allows multiple files with these extensions:

- JSON, CSV, TXT, and XML;
- Z21;
- ESU and ESUX;
- LokProgrammer;
- ZIP.

The frontend blocks other extensions and script or executable formats. The server also rejects
empty, oversized, or blocked content and applies the configured attachment limit. The normal stable
default is 25 MiB per file, but an operator can configure a stricter or different server limit.

File selection first uploads each file to the preview endpoint only. The preview can display file
type and size, recognized project, decoder, address, type, manufacturer, LokProgrammer metadata,
an extracted preview image, and counts of suggested CV values and functions. Preview alone does not
store the original file on the vehicle.

The chapter distinguishes four independent actions:

1. **Apply suggestion** copies the first recognized metadata suggestion into the profile and note
   fields for a later file save.
2. **Review CVs** converts all detected CV suggestions into the normal CV import preview. Users
   still choose rows before applying them.
3. **Apply functions** immediately writes valid, de-duplicated detected function mappings. Each
   mapping uses the detected name and type, an empty symbol, mode `dauer`, no direction dependency,
   and the preview file name as its note. It does not open the function JSON import or a
   confirmation preview.
4. **Save files** stores the selected original files on the vehicle with the current profile and
   description. Recognized ESUX metadata can fill missing metadata during upload.

These actions are independent. Applying suggested CVs or functions does not save the source file,
and saving a file does not automatically persist suggested CVs or functions. If one file in a
multi-file preview request fails, no file has been persisted and the batch does not produce the
normal preview. Stored files can be downloaded or deleted. File deletion has no additional
confirmation dialog and removes the stored file data when it is no longer referenced.

## ECoS input boundary

An ECoS locomotive draft can provide preview CV values before the vehicle has been saved. The CV
tab displays the first 18 values, reports the number of additional values, and identifies the
source locomotive; the Speed Curve tab can derive its read-only chart from the same draft values.
Once the core vehicle is saved, normal CV, function, and file operations use the stored vehicle.

The chapter does not describe ECoS host configuration, connection tests, raw probes,
synchronization, conflict handling, or confirmed writes to the command station. It names the later
**Digital centers** chapter as the owner of that workflow but does not create a working link while
that chapter remains unpublished.

## Persistence, refreshes, and partial writes

Function saves and deletes, CV saves and deletes, successful CV imports, detected-function
application, decoder-file saves, and decoder-file deletion persist immediately. They are not part
of the vehicle form's **Save changes** action.

After a successful write workflow, the frontend reloads the selected vehicle and replaces the main
vehicle form and specialist editor state with fresh server data. Unrelated unsaved core fields,
function edits, image metadata, or other pending tab changes can therefore be discarded. Both
pages tell users to save or intentionally discard all other pending vehicle changes before any
decoder-data write action.

Function import, selected CV import, detected-function application, and multi-file saving issue
sequential API requests without one encompassing transaction. If a later request fails, earlier
requests remain stored, later rows or files are not attempted, and the normal success refresh does
not run. The user must reload the vehicle, compare stored state with the source, and retry only the
missing items to avoid duplicates or unintended overwrites.

Preview generation, row selection, profile shortcuts, metadata suggestion, and the Speed Curve tab
do not themselves persist data.

## Storage and backup boundary

Vehicle functions, CV values, CV history, decoder-file metadata, and stored decoder-file blobs are
local RailKeeper application data. Functions, CV values, and CV files are included in the
application backup scope. The chapter recommends a current, successfully validated backup before
large imports or cleanup while leaving backup creation, validation, restore limits, and recovery to
the administration documentation.

Local JSON exports are useful exchange and inspection files but are not represented as complete
RailKeeper backups. Function exports do not contain CV values or decoder files, and CV exports do
not contain function mappings or decoder-file blobs.

## Empty and error states

Both language pages cover at least:

| Situation | Documented response |
| --- | --- |
| Vehicle is not saved | Save the core vehicle before persistent function, CV, or file actions. |
| No functions are assigned | Clear **Only assigned**, enter a name, symbol, or note, then save the row. |
| Function import has no valid rows | Check JSON shape, F0-F31 keys, stored type names, and modes. |
| Speed curve is empty | Add or import CV 2/5/6 or CV 67-94 in one profile/protocol group. |
| Speed curve looks unexpected | Check CV 29, profile/protocol grouping, and missing CVs. |
| CV input is rejected | Use an integer CV from 1-1024 and value from 0-255. |
| CV import marks a duplicate invalid | Keep only one CV-number/profile identity in the source file. |
| File preview has no metadata | The file can still be stored; no metadata, CV, or function suggestion is promised. |
| File is rejected | Check extension, content, empty-file status, and the operator's size limit. |
| Multi-row or multi-file action fails | Reload, compare stored results, and retry only missing items. |
| A write discards other edits | Explain the full-record refresh and restore from a suitable backup when needed. |
| A CV value or decoder file was deleted without a prompt | Explain immediate deletion and recovery boundary. |

No page suggests that a later failure rolled back earlier successful sequential requests.

## Navigation and cross-links

Add exact sidebar entries immediately after the maintenance chapter:

```ts
{ text: "Decoder, functions, and CV data", link: "/guide/vehicles/decoder-cv" }
```

```ts
{ text: "Decoder, Funktionen und CV-Daten", link: "/de/guide/vehicles/decoder-cv" }
```

Add concise links to both User Guide landing pages. Add the decoder page to the related-page lists
of the core vehicle, media, and maintenance chapters where it improves the existing transition.
The decoder page links back only to already published pages:

- User Guide overview;
- Vehicle inventory and core records;
- Vehicle images and attachments;
- Vehicle maintenance and condition.

Do not create working links to the planned digital-centers, search-and-spares, or administration
specialist pages. A plain-text boundary can name the future digital-centers chapter.

## Coverage contract

Change only the `vehicle-decoder-cv` topic from `planned` to `documented`. Its existing paths and
owners already match the approved scope:

- translations below `vehicles.cv`, `vehicles.functionMode`, `vehicles.functionType`,
  `vehicles.functions`, and `vehicles.speedCurve`;
- APIs below `/api/v1/vehicles/{id}/functions`, `/api/v1/vehicles/{id}/cv-values`,
  `/api/v1/vehicles/{id}/cv-files`, and `/api/v1/cv-files`.

The implementation plan must first demonstrate the missing-page gate by changing the status before
creating the pages. The expected failure is limited to the absent English and German decoder/CV
files.

## Verification and review

Implementation verification includes:

1. direct stable-tag checks for tabs, roles, fields, validation, defaults, imports, exports, speed
   curve calculation, preview behavior, refreshes, partial writes, deletion, and backup scope;
2. exact frontmatter and matching section-order checks;
3. placeholder, em-dash, whitespace, and link scans;
4. `npm.cmd run check` from `docs`, including all unit tests, coverage validation, and the VitePress
   production build;
5. an independent read-only review of English/German parity, source fidelity, role boundaries,
   persistent versus preview actions, CV identity, speed-curve interpretation, file handling,
   refresh risk, non-atomic actions, deletion, and navigation;
6. correction and re-review of every Critical or Important finding and valid completeness finding;
7. a pull request only from the isolated feature branch;
8. merge only when CI, Trivy, and CodeQL succeed for the exact reviewed head SHA and every review
   thread is resolved.

## Expected outcome

RailKeeper users can configure every stable digital-function field, interpret the derived speed
curve, manage and exchange CV values, and safely preview, apply, store, download, and remove decoder
files. They can distinguish each action's persistence boundary, recognize partial writes, and use
ECoS draft data without mistaking RailKeeper's documentation workspace for a decoder-programming or
digital-center write interface.
