---
title: Decoder, functions, and CV data
description: Map digital functions, inspect speed curves, manage CV values, and store decoder files.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Decoder, functions, and CV data

RailKeeper keeps function mappings, speed-curve data, CV values, and decoder project files with a
vehicle. The connected workflow is available in **Control**, **Speed curve**, and **CV**.

## Prerequisites and access rights

Open a vehicle in **Vehicle inventory**, choose **Edit**, then select the required tab. General
fields such as **Digital**, decoder number, decoder type, and ABC braking are covered by
[Vehicle inventory and core records](/guide/vehicles/).

A vehicle must normally be saved once before RailKeeper can persist functions, CV values, or files.
An unsaved ECoS draft can preview CVs and a derived curve, but its normal write actions remain
disabled.

Admin, Editor, Viewer, and Planner can inspect stored decoder data. Viewer-level access can also
export functions and CV values and download decoder files. Only Admin and Editor can save, import,
apply, upload, or delete data. The server enforces this boundary.

::: warning Save other edits first
Every successful decoder-data write reloads the complete selected vehicle. Save or intentionally
discard pending core fields, function edits, image metadata, and changes on other tabs before a
write action.
:::

## Map digital functions F0-F31

Open **Control**. **Digital functions** provides one row for every key from F0 through F31. The
summary counts assigned, sound, and light functions. Enable **Only assigned** to hide unused rows.

Each row contains:

- **Function name**
- **Symbol**
- **Mode**
- **Inverted**
- **Note**
- **Save** and **Delete**

Selecting a symbol can fill an empty name and infers the stored function type. The type is not a
separate control in this stable view.

| Stored type | English meaning |
| --- | --- |
| `standard` | Standard |
| `sound` | Sound |
| `licht` | Light |
| `kupplung` | Coupler |
| `rauch` | Smoke |
| `sonderfunktion` | Special function |

Modes are stored as `dauer` and `moment` and displayed as **Continuous** and **Momentary**. The
**Inverted** switch stores the row's direction-dependent/inverted flag.

F0 starts with name `Fahrlicht`, the light symbol, and type `licht`. Other new rows start with type
`standard`. Every new row starts in mode `dauer`. A new row needs at least a name, symbol, or note
before it can be saved. The local F0 default therefore counts as assigned even before **Save F0**.

The server accepts only F0-F31, known types and modes, names up to 120 characters, symbol keys up to
80 characters, and notes up to 1,000 characters. Saving or deleting one row acts immediately,
reloads the complete vehicle, and has no additional delete confirmation.

### Import and export functions

**Export** downloads `<inventory-number>-funktionen.json`. Without an inventory number, the name is
`railkeeper-funktionen.json`. It contains vehicle inventory number, name, decoder number, and all
assigned mappings. The decoder number prefers the primary digital number and falls back to the DT
decoder number. Export uses the current rows, including unsaved local function edits, but does not
save them in RailKeeper.

**Import** reads the first selected JSON file. It accepts a top-level array or a `functions` or
`functionMappings` array. Function keys are changed to uppercase. Rows with invalid keys, types, or
modes are ignored. Valid rows are written in sequence without a preview or confirmation. Duplicate
keys remain in the sequence, so a later row for the same key overwrites the earlier row.

If one request fails, earlier rows remain stored, later rows are not attempted, and the normal
refresh does not run. Reload the vehicle, compare the stored mappings, and retry only missing rows.
The stable buttons **Import** and **Export**, and some import errors, remain English or German
regardless of the selected interface language.

## Read the speed curve

Open **Speed curve**. This tab is **Read only**. It calculates a speed characteristic from stored CV
values or an ECoS draft and never writes to RailKeeper, a decoder, or a command station.

RailKeeper groups relevant CVs by decoder profile and protocol. Select a profile to view:

- the number of relevant CVs in that group;
- curve mode;
- CV 29 state;
- plotted point count;
- forward/reverse trim;
- the chart and the underlying CV lists;
- missing CVs.

The **3-point curve** uses CV 2 at speed step 1, CV 6 at step 14, and CV 5 at step 28. The
**28-point speed table** uses CV 67 through CV 94. CV 66 supplies forward trim and CV 95 reverse
trim.

When CV 29 is known, bit 4 selects the 28-point table or 3-point curve. If the selected curve has no
points, or CV 29 is unknown, RailKeeper falls back to the most useful available data: a complete
28-point table, at least two 3-point values, any table values, then any 3-point value. This fallback
changes only the display.

## Manage CV values manually

Open **CV**. The summary shows the number of **CV values**, **Profiles**, and **Files**.

The manual form contains:

| Field | Rule |
| --- | --- |
| CV number | Required integer from 1 through 1024 |
| Value | Required integer from 0 through 255 |
| Category | Optional stored German category |
| Protocol | Optional protocol |
| Decoder profile | Optional free text |
| Source file | Optional decoder file belonging to this vehicle |
| Description | Optional text |

Stable categories are `Adresse`, `Fahrverhalten`, `Motor`, `Licht`, `Sound`, `Funktion`,
`Decoder`, and `Sonstiges`. They remain German in the English interface.

Protocol choices are `Motorola 14`, `Motorola 27`, `Motorola 28`, `Motorola FX 14`, `DCC 14`,
`DCC 28`, `DCC 128`, `LGB`, and `Selectrix`.

Common profile suggestions are ESU LokPilot 5, ESU LokSound 5, Zimo MS, Zimo MX, D&H SD, D&H DH,
Märklin mLD3, Märklin mSD3, and Lenz Standard+. Profiles already used by CV values or files appear
as shortcuts. A profile is descriptive free text, not validation of the physical decoder.

RailKeeper identifies a CV row by CV number plus normalized decoder profile. Protocol is not part
of that identity. **Add CV** updates an existing matching row instead of creating a duplicate.
**Save CV**, **Edit CV**, and **Delete CV** act immediately and reload the complete vehicle.
Deletion has no additional confirmation.

When an update changes the numeric value, RailKeeper adds a history record. Metadata-only changes
do not add history. The stable interface displays the five newest history entries for a row.

## Import and export CV values

CV **Import** reads the first selected JSON, CSV, or TXT file.

- JSON may be a top-level array or an object with `cvValues`.
- CSV or TXT uses one row per line in the order CV number, value, description, category, and
  decoder profile.
- Text separators may be semicolons or commas.
- A line beginning with `cv` is treated as a header.
- Text import leaves protocol and source file empty.

The preview marks each row as **new**, **changed**, **same**, or **invalid**. New and changed rows
start selected. Same and invalid rows do not. Duplicate CV-number/profile identities after the
first occurrence are invalid.

Use **Only new**, **Select all**, **Select none**, or individual checkboxes, then **Apply selected
fields**. Selected rows are written sequentially. A later failure does not roll back earlier rows
and prevents the normal refresh. Reload, compare, and retry only missing rows.

**Export** downloads `<inventory-number>-cv.json` or `railkeeper-cv.json`. It contains the vehicle
identity, preferred decoder number, and all CV records returned with the vehicle, including
metadata and history. It excludes the current unsaved CV form, function mappings, and decoder-file
contents.

Some stable preview status and validation messages remain German in the English interface. The
visible CV toolbar labels **Import** and **Export** remain English in both language modes.

## Preview, apply, and store decoder files

Under **CV files**, enter an optional decoder profile and note, then choose **Upload CV file**.
Multiple files are allowed.

Supported extensions are:

- JSON, CSV, TXT, and XML
- Z21
- ESU and ESUX
- LokProgrammer
- ZIP

The normal limit is 25 MiB per file. An operator can configure another attachment limit. RailKeeper
rejects unsupported extensions, blocked executable or script content, empty files, and files above
the server limit.

Selection first creates an **Upload preview**. It can show size, MIME type, a preview image,
project, decoder, address, type, manufacturer, LokProgrammer metadata, and counts of detected CV
values and function keys. A preview does not store the original file.

The preview actions are independent:

1. **Apply suggestion** copies the first recognized profile and description into the unsaved file
   fields.
2. **Review CVs** sends detected values to the normal CV import preview. Nothing is written until
   selected CV rows are applied. A detected profile wins; otherwise RailKeeper uses the current
   file profile, detected decoder, or detected project name in that order.
3. **Apply functions** immediately writes valid detected functions. They use the detected name and
   type, an empty symbol, mode `dauer`, no direction dependency, and the preview filename as note.
   Duplicate keys are consolidated, with the later detected mapping winning.
4. **Save files** stores the selected original files with the current profile and note.

Applying CVs or functions does not save the original file. Saving files does not automatically
apply detected CVs or functions. Recognized ESU/LokProgrammer metadata can fill fields that were
left empty during file saving.

If one file in preview generation fails, no file has been stored and the batch does not produce the
normal preview. **Apply functions** and **Save files** send sequential requests. A later failure
leaves earlier results stored and prevents the normal refresh. Reload and compare before retrying.

Stored files show original name, profile, MIME type, size, and description. **Download** retrieves
the original file. **Delete** removes it immediately without another confirmation and removes the
stored file data when no reference remains.

A CV value can name a decoder file under **Source file**. Before deleting a file, edit CV rows,
inspect that field, choose **No file** for every matching row, and save each change. The stable CV
table does not display the source assignment. Deleting a file does not delete its CV values or
clear their stored source identifier, so skipping this sequence can leave stale references.

## Use an ECoS preview as an input path

An unsaved ECoS locomotive draft can supply CV values before the vehicle exists. The **CV** tab
shows the first 18 values, the count of additional values, and the source locomotive. **Speed
curve** can derive its read-only display from the same draft.

After the core vehicle is saved, normal function, CV, and file actions use the stored vehicle. This
chapter does not cover ECoS connection setup, raw probes, synchronization, conflict handling, or
writes to a command station. Those operations belong to the planned Digital centers chapter.

## Protect data during writes

| Action | Persists data | Reloads the complete vehicle |
| --- | --- | --- |
| Edit a function field without Save | No | No |
| Save or delete one function | Immediately | After success |
| Import function JSON | Sequentially | Only after full success |
| Export function JSON | No | No |
| View the speed curve | No | No |
| Build or select a CV import preview | No | No |
| Apply selected CV rows | Sequentially | Only after full success |
| Add, save, or delete one CV | Immediately | After success |
| Export CV JSON | No | No |
| Build a decoder-file preview | No | No |
| Apply a metadata suggestion | No | No |
| Review detected CVs | No, until rows are applied | No |
| Apply detected functions | Sequentially | Only after full success |
| Save decoder files | Sequentially | Only after full success |
| Download a decoder file | No | No |
| Delete a decoder file | Immediately | After success |

Functions, CV values, CV history, decoder-file metadata, and decoder-file data belong to the local
RailKeeper application backup. Create and validate a current backup before large imports or cleanup.
Function and CV JSON exports are exchange files, not complete RailKeeper backups.

## Troubleshoot decoder data

| Situation | Response |
| --- | --- |
| Vehicle is not saved | Save the core record before persistent function, CV, or file actions. |
| No function appears under **Only assigned** | Disable the filter, enter a name, symbol, or note, and save. |
| Function import finds no valid mappings | Check the JSON shape, F0-F31 keys, stored type names, and modes. |
| Speed curve is empty | Add or import CV 2/5/6 or CV 67-94 in one profile/protocol group. |
| Curve selection looks wrong | Check CV 29, profile/protocol grouping, and missing CVs. |
| CV input is rejected | Use an integer CV from 1-1024 and a value from 0-255. |
| CV import reports a duplicate | Keep one CV-number/profile identity in the source. |
| File preview has no metadata | The file may still be stored, but no suggestion is available. |
| File is rejected | Check extension, content, empty-file status, and the operator's size limit. |
| A file is still a CV source | Clear **Source file** on matching CV rows before deleting it. |
| A batch partly fails | Reload, compare stored results, then retry only missing items. |
| Other edits disappear after a write | Reload replaces unsaved data. Recover from a backup if needed. |
| A CV or file was deleted without a prompt | Deletion is immediate. Recovery requires a suitable backup. |

No failed later request rolls back an earlier successful request in the same sequential action.

## Related pages

- [User Guide overview](/guide/)
- [Vehicle inventory and core records](/guide/vehicles/)
- [Vehicle images and attachments](/guide/vehicles/media)
- [Vehicle maintenance and condition](/guide/vehicles/maintenance)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.17.6** and was last reviewed on 2026-08-16.
