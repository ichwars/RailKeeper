# RailKeeper Vehicle Decoder, Functions, and CV Data User Guide Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish a complete, source-verified English and German user-guide chapter for vehicle
decoder functions, speed curves, CV values, and decoder files in stable RailKeeper v0.1.17.6.

**Architecture:** Add one paired VitePress Markdown chapter under the existing vehicle guide and
change only the existing decoder/CV coverage topic from planned to documented. Keep the connected
Control, Speed Curve, CV, decoder-file, and ECoS-preview workflows together, then add navigation and
cross-links only after the page pair passes the documentation gate.

**Tech Stack:** VitePress 2, Markdown, JSON coverage manifest, Node.js documentation tests, GitHub
Actions.

## Global Constraints

- Document only stable RailKeeper **v0.1.17.6**. Do not describe later `main` behavior as stable.
- Create both `docs/site/guide/vehicles/decoder-cv.md` and
  `docs/site/de/guide/vehicles/decoder-cv.md` with semantically equivalent content.
- Use the exact page metadata, public routes, stable UI labels, stored values, and semantic section
  order from the approved specification.
- Cover F0-F31, function JSON exchange, the read-only speed curve, manual CV data, CV import/export,
  history, decoder-file preview/application/storage, ECoS draft input, roles, backup, and failures.
- Distinguish every preview-only, local-only, immediate-write, sequential-write, and download action.
- Treat core decoder fields, digital-center setup/write-back, general media, search/spares, and
  administration as boundaries. Do not reproduce those workflows.
- Do not add screenshots or working links to unpublished pages.
- Keep lines readable, avoid em dashes, and do not introduce unfinished-content markers.
- Do not change runtime, API, validation, import, storage, or deletion behavior.
- Do not commit `docs/.vitepress/dist`, dependency caches, local screenshots, or generated output.
- Preserve the dirty local `main`. All work stays in the isolated
  `dev/docs-user-guide-decoder-cv` worktree.
- Merge only after CI, Trivy, and CodeQL succeed for the exact reviewed head SHA and every
  actionable review thread is resolved.

---

### Task 1: Mark decoder/CV documented and prove the missing-page gate

**Files:**
- Modify: `docs/coverage.json:40`
- Reference: `docs/scripts/validate-docs.mjs`
- Reference:
  `docs/superpowers/specs/2026-08-16-railkeeper-vehicle-decoder-cv-guide-design.md`

**Interfaces:**
- Consumes: existing translation owners `vehicles.cv`, `vehicles.functionMode`,
  `vehicles.functionType`, `vehicles.functions`, and `vehicles.speedCurve` plus the existing
  function/CV API owners in `docs/coverage.json`.
- Produces: a documented `vehicle-decoder-cv` contract whose only missing artifacts are its exact
  English and German page paths.

- [ ] **Step 1: Confirm the existing source ownership and page paths**

Run from the repository root:

```powershell
rg -n -A 25 -B 3 'vehicle-decoder-cv|vehicles\.cv|vehicles\.function|vehicles\.speedCurve' docs/coverage.json
rg -n 'vehicles/\{id\}/(functions|cv-values|cv-files)|/api/v1/cv-files' docs/coverage.json
```

Expected: the topic is `planned`, its paths are `guide/vehicles/decoder-cv.md` and
`de/guide/vehicles/decoder-cv.md`, and all listed owners already belong to
`vehicle-decoder-cv`.

- [ ] **Step 2: Change only the decoder/CV topic status**

Replace the topic object with exactly:

```json
{
  "id": "vehicle-decoder-cv",
  "audience": "user",
  "status": "documented",
  "englishPath": "guide/vehicles/decoder-cv.md",
  "germanPath": "de/guide/vehicles/decoder-cv.md"
}
```

Do not change any translation or API owner.

- [ ] **Step 3: Run the documentation check and verify the intentional failure**

Run:

```powershell
Set-Location docs
npm.cmd run check
```

Expected: all 19 unit tests pass, then coverage validation fails only because
`guide/vehicles/decoder-cv.md` and `de/guide/vehicles/decoder-cv.md` are absent. Stop and correct
any JSON, ownership, metadata, or unrelated validation failure before continuing.

Do not commit the intentionally red state.

---

### Task 2: Create the complete English and German decoder/CV chapter

**Files:**
- Create: `docs/site/guide/vehicles/decoder-cv.md`
- Create: `docs/site/de/guide/vehicles/decoder-cv.md`
- Modify: `docs/coverage.json` (already changed and uncommitted in Task 1)
- Reference: all stable files named in the approved design specification

**Interfaces:**
- Consumes: the exact page paths from Task 1 and the stable behavior at tag `v0.1.17.6`.
- Produces: a validated language pair at `/guide/vehicles/decoder-cv` and
  `/de/guide/vehicles/decoder-cv`.

- [ ] **Step 1: Create the complete English page**

Create `docs/site/guide/vehicles/decoder-cv.md` with exactly:

```markdown
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
```

- [ ] **Step 2: Create the German page with semantic parity**

Create `docs/site/de/guide/vehicles/decoder-cv.md` with exactly:

```markdown
---
title: Decoder, Funktionen und CV-Daten
description: Digitalfunktionen zuordnen, Fahrkurven prüfen, CV-Werte pflegen und Decoder-Dateien speichern.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Decoder, Funktionen und CV-Daten

RailKeeper speichert Funktionszuordnungen, Fahrkurvendaten, CV-Werte und Decoder-Projektdateien am
Fahrzeug. Der zusammengehörige Ablauf befindet sich in **Steuerung**, **Fahrkurve** und **CV**.

## Voraussetzungen und Zugriffsrechte

Öffne ein Fahrzeug im **Fahrzeugbestand**, wähle **Bearbeiten** und dann den benötigten Tab.
Allgemeine Felder wie **Digital**, Decodernummer, Decodertyp und ABC-Bremsung erklärt
[Fahrzeugbestand und Grunddaten](/de/guide/vehicles/).

Ein Fahrzeug muss normalerweise einmal gespeichert sein, bevor RailKeeper Funktionen, CV-Werte
oder Dateien speichern kann. Ein ungespeicherter ECoS-Entwurf kann CVs und eine abgeleitete Kurve
anzeigen. Seine normalen Schreibaktionen bleiben jedoch deaktiviert.

Admin, Editor, Viewer und Planner können gespeicherte Decoderdaten ansehen. Auf Viewer-Ebene sind
auch Funktions- und CV-Export sowie der Download von Decoder-Dateien möglich. Nur Admin und Editor
dürfen Daten speichern, importieren, übernehmen, hochladen oder löschen. Der Server erzwingt diese
Grenze.

::: warning Andere Änderungen zuerst speichern
Jede erfolgreiche Schreibaktion an Decoderdaten lädt das vollständige ausgewählte Fahrzeug neu.
Speichere oder verwirf bewusst ausstehende Grunddaten, Funktionsänderungen, Bildmetadaten und
Änderungen anderer Tabs vor einer Schreibaktion.
:::

## Digitalfunktionen F0-F31 zuordnen

Öffne **Steuerung**. **Digitalfunktionen** enthält eine Zeile für jede Taste von F0 bis F31. Die
Übersicht zählt belegte, Sound- und Lichtfunktionen. Aktiviere **Nur belegte**, um ungenutzte Zeilen
auszublenden.

Jede Zeile enthält:

- **Funktionsname**
- **Symbol**
- **Betriebsart**
- **Invertiert**
- **Notiz**
- **Speichern** und **Löschen**

Die Auswahl eines Symbols kann einen leeren Namen füllen und leitet den gespeicherten Funktionstyp
ab. Der Typ ist in dieser stabilen Ansicht kein eigenes Bedienelement.

| Gespeicherter Typ | Deutsche Bedeutung |
| --- | --- |
| `standard` | Standard |
| `sound` | Sound |
| `licht` | Licht |
| `kupplung` | Kupplung |
| `rauch` | Rauch |
| `sonderfunktion` | Sonderfunktion |

Die Betriebsarten werden als `dauer` und `moment` gespeichert und als **Dauer** und **Moment**
angezeigt. **Invertiert** speichert das richtungsabhängige beziehungsweise invertierte Merkmal der
Zeile.

F0 beginnt mit dem Namen `Fahrlicht`, dem Lichtsymbol und dem Typ `licht`. Andere neue Zeilen
beginnen mit `standard`. Jede neue Zeile verwendet zunächst `dauer`. Eine neue Zeile benötigt
mindestens Name, Symbol oder Notiz, bevor sie gespeichert werden kann. Der lokale F0-Standard zählt
daher bereits vor **F0 speichern** als belegt.

Der Server akzeptiert nur F0-F31, bekannte Typen und Betriebsarten, Namen bis 120 Zeichen,
Symbolschlüssel bis 80 Zeichen und Notizen bis 1.000 Zeichen. Speichern oder Löschen einer Zeile
wirkt sofort, lädt das vollständige Fahrzeug neu und besitzt keine zusätzliche Löschbestätigung.

### Funktionen importieren und exportieren

**Export** lädt `<inventarnummer>-funktionen.json` herunter. Ohne Inventarnummer lautet der Name
`railkeeper-funktionen.json`. Die Datei enthält Inventarnummer, Fahrzeugname, Decodernummer und alle
belegten Zuordnungen. Die Decodernummer verwendet zuerst die primäre digitale Nummer und ersatzweise
die DT-Decodernummer. Der Export verwendet die aktuellen Zeilen einschließlich ungespeicherter
lokaler Funktionsänderungen, speichert sie aber nicht in RailKeeper.

**Import** liest die erste ausgewählte JSON-Datei. Er akzeptiert ein Array auf oberster Ebene oder
ein Array unter `functions` beziehungsweise `functionMappings`. Funktionstasten werden in
Großbuchstaben umgewandelt. Zeilen mit ungültigen Tasten, Typen oder Betriebsarten werden
übersprungen. Gültige Zeilen werden ohne Vorschau oder Bestätigung nacheinander geschrieben.
Doppelte Tasten bleiben in der Reihenfolge, daher überschreibt eine spätere Zeile die frühere.

Schlägt eine Anfrage fehl, bleiben frühere Zeilen gespeichert, spätere werden nicht versucht und
die normale Aktualisierung läuft nicht. Lade das Fahrzeug neu, vergleiche die Zuordnungen und
wiederhole nur fehlende Zeilen. Die stabilen Schaltflächen **Import** und **Export** sowie manche
Importfehler bleiben unabhängig von der Oberflächensprache englisch oder deutsch.

## Fahrkurve lesen

Öffne **Fahrkurve**. Dieser Tab ist **Nur lesen**. Er berechnet eine Geschwindigkeitskennlinie aus
gespeicherten CV-Werten oder einem ECoS-Entwurf und schreibt niemals in RailKeeper, einen Decoder
oder eine Digitalzentrale.

RailKeeper gruppiert relevante CVs nach Decoderprofil und Protokoll. Wähle ein Profil, um Folgendes
zu sehen:

- Anzahl relevanter CVs in dieser Gruppe;
- Kurvenart;
- Zustand von CV 29;
- Anzahl dargestellter Punkte;
- Vorwärts-/Rückwärtstrimmung;
- Diagramm und zugrunde liegende CV-Listen;
- fehlende CVs.

Die **3-Punkt-Kurve** verwendet CV 2 bei Fahrstufe 1, CV 6 bei Stufe 14 und CV 5 bei Stufe 28. Die
**28-Punkt-Speedtable** verwendet CV 67 bis CV 94. CV 66 liefert die Vorwärts- und CV 95 die
Rückwärtstrimmung.

Ist CV 29 bekannt, wählt Bit 4 die 28-Punkt-Tabelle oder 3-Punkt-Kurve. Enthält die gewählte Kurve
keine Punkte oder ist CV 29 unbekannt, verwendet RailKeeper die sinnvollsten verfügbaren Daten:
eine vollständige 28-Punkt-Tabelle, mindestens zwei 3-Punkt-Werte, beliebige Tabellenwerte und
danach einen beliebigen 3-Punkt-Wert. Diese Auswahl ändert nur die Anzeige.

## CV-Werte manuell verwalten

Öffne **CV**. Die Übersicht zeigt die Anzahl der **CV-Werte**, **Profile** und **Dateien**.

Das manuelle Formular enthält:

| Feld | Regel |
| --- | --- |
| CV-Nummer | Erforderliche Ganzzahl von 1 bis 1024 |
| Wert | Erforderliche Ganzzahl von 0 bis 255 |
| Kategorie | Optionale gespeicherte deutsche Kategorie |
| Protokoll | Optionales Protokoll |
| Decoderprofil | Optionaler Freitext |
| Quelldatei | Optionale Decoder-Datei dieses Fahrzeugs |
| Beschreibung | Optionaler Text |

Stabile Kategorien sind `Adresse`, `Fahrverhalten`, `Motor`, `Licht`, `Sound`, `Funktion`,
`Decoder` und `Sonstiges`.

Protokolle sind `Motorola 14`, `Motorola 27`, `Motorola 28`, `Motorola FX 14`, `DCC 14`,
`DCC 28`, `DCC 128`, `LGB` und `Selectrix`.

Häufige Profilvorschläge sind ESU LokPilot 5, ESU LokSound 5, Zimo MS, Zimo MX, D&H SD, D&H DH,
Märklin mLD3, Märklin mSD3 und Lenz Standard+. Bereits in CV-Werten oder Dateien verwendete Profile
erscheinen als Schnellwahl. Ein Profil ist beschreibender Freitext und keine Prüfung des
physischen Decoders.

RailKeeper identifiziert eine CV-Zeile durch CV-Nummer und normalisiertes Decoderprofil. Das
Protokoll gehört nicht zu dieser Identität. **CV hinzufügen** aktualisiert eine vorhandene passende
Zeile, anstatt ein Duplikat anzulegen. **CV speichern**, **CV bearbeiten** und **CV löschen** wirken
sofort und laden das vollständige Fahrzeug neu. Löschen besitzt keine zusätzliche Bestätigung.

Ändert eine Aktualisierung den Zahlenwert, legt RailKeeper einen Historieneintrag an. Reine
Metadatenänderungen erzeugen keinen. Die stabile Oberfläche zeigt die fünf neuesten
Historieneinträge einer Zeile.

## CV-Werte importieren und exportieren

Der CV-**Import** liest die erste ausgewählte JSON-, CSV- oder TXT-Datei.

- JSON darf ein Array auf oberster Ebene oder ein Objekt mit `cvValues` sein.
- CSV oder TXT verwendet eine Zeile je Wert in der Reihenfolge CV-Nummer, Wert, Beschreibung,
  Kategorie und Decoderprofil.
- Als Trenner sind Semikolon und Komma möglich.
- Eine mit `cv` beginnende Zeile gilt als Kopfzeile.
- Beim Textimport bleiben Protokoll und Quelldatei leer.

Die Vorschau markiert jede Zeile als **neu**, **geändert**, **gleich** oder **ungültig**. Neue und
geänderte Zeilen sind zunächst ausgewählt, gleiche und ungültige nicht. Doppelte Kombinationen aus
CV-Nummer und Profil sind nach dem ersten Vorkommen ungültig.

Nutze **Nur neue**, **Alle auswählen**, **Keine auswählen** oder einzelne Kontrollkästchen und
danach **Ausgewählte Felder übernehmen**. Ausgewählte Zeilen werden nacheinander geschrieben. Ein
später Fehler macht frühere Zeilen nicht rückgängig und verhindert die normale Aktualisierung. Lade
neu, vergleiche und wiederhole nur fehlende Zeilen.

**Export** lädt `<inventarnummer>-cv.json` oder `railkeeper-cv.json` herunter. Die Datei enthält
Fahrzeugidentität, bevorzugte Decodernummer und alle mit dem Fahrzeug gelieferten CV-Datensätze
einschließlich Metadaten und Historie. Das aktuelle ungespeicherte CV-Formular,
Funktionszuordnungen und Decoder-Dateiinhalte sind nicht enthalten.

Einige stabile Vorschau-, Status- und Validierungsmeldungen bleiben in der englischen Oberfläche
deutsch. Die sichtbaren CV-Schaltflächen **Import** und **Export** bleiben in beiden Sprachmodi
englisch.

## Decoder-Dateien prüfen, übernehmen und speichern

Trage unter **CV-Dateien** ein optionales Decoderprofil und eine Bemerkung ein und wähle
**CV-Datei hochladen**. Mehrere Dateien sind möglich.

Unterstützte Endungen:

- JSON, CSV, TXT und XML
- Z21
- ESU und ESUX
- LokProgrammer
- ZIP

Die normale Grenze beträgt 25 MiB je Datei. Ein Betreiber kann eine andere Beilagengrenze
konfigurieren. RailKeeper weist nicht unterstützte Endungen, blockierte ausführbare oder
skriptartige Inhalte, leere und zu große Dateien zurück.

Die Auswahl erzeugt zuerst eine **Upload-Vorschau**. Sie kann Größe, MIME-Typ, Vorschaubild,
Projekt, Decoder, Adresse, Typ, Hersteller, LokProgrammer-Metadaten sowie die Anzahl erkannter
CV-Werte und Funktionstasten anzeigen. Eine Vorschau speichert die Originaldatei nicht.

Die Vorschauaktionen sind unabhängig:

1. **Vorschlag übernehmen** kopiert das erste erkannte Profil und die Beschreibung in die noch
   ungespeicherten Dateifelder.
2. **CVs prüfen** übergibt erkannte Werte an die normale CV-Importvorschau. Erst das Übernehmen
   ausgewählter CV-Zeilen schreibt Daten. Ein erkanntes Profil gewinnt, andernfalls verwendet
   RailKeeper nacheinander aktuelles Dateiprofil, erkannten Decoder oder erkannten Projektnamen.
3. **Funktionen übernehmen** schreibt gültige erkannte Funktionen sofort. Sie verwenden erkannten
   Namen und Typ, ein leeres Symbol, Betriebsart `dauer`, keine Richtungsabhängigkeit und den
   Vorschaudateinamen als Notiz.
   Doppelte Tasten werden zusammengeführt, wobei die später erkannte Zuordnung gewinnt.
4. **Dateien speichern** speichert die ausgewählten Originaldateien mit aktuellem Profil und
   aktueller Bemerkung.

Das Übernehmen von CVs oder Funktionen speichert nicht die Originaldatei. Das Speichern von Dateien
übernimmt nicht automatisch erkannte CVs oder Funktionen. Erkannte ESU/LokProgrammer-Metadaten
können beim Speichern leer gelassene Felder füllen.

Schlägt eine Datei während der Vorschauerzeugung fehl, wurde keine Datei gespeichert und der Stapel
erzeugt nicht die normale Vorschau. **Funktionen übernehmen** und **Dateien speichern** senden
Anfragen nacheinander. Ein später Fehler lässt frühere Ergebnisse gespeichert und verhindert die
normale Aktualisierung. Lade neu und vergleiche vor einem erneuten Versuch.

Gespeicherte Dateien zeigen Originalname, Profil, MIME-Typ, Größe und Beschreibung. **Download**
liefert die Originaldatei. **Löschen** entfernt sie sofort ohne weitere Bestätigung und entfernt die
gespeicherten Dateidaten, wenn keine Referenz verbleibt.

Ein CV-Wert kann unter **Quelldatei** eine Decoder-Datei nennen. Bearbeite vor dem Löschen einer
Datei die CV-Zeilen, prüfe dieses Feld, wähle für jede passende Zeile **Ohne Datei** und speichere
die Änderung. Die stabile CV-Tabelle zeigt die Quellenzuordnung nicht. Das Löschen einer Datei
löscht weder ihre CV-Werte noch deren gespeicherte Quellen-ID. Ohne diese Reihenfolge können
veraltete Referenzen verbleiben.

## ECoS-Vorschau als Eingangspfad verwenden

Ein ungespeicherter ECoS-Lokomotiventwurf kann CV-Werte liefern, bevor das Fahrzeug existiert. Der
Tab **CV** zeigt die ersten 18 Werte, die Anzahl weiterer Werte und die Quelllokomotive.
**Fahrkurve** kann daraus seine schreibgeschützte Anzeige ableiten.

Nach dem Speichern des Grunddatensatzes verwenden normale Funktions-, CV- und Dateiaktionen das
gespeicherte Fahrzeug. Dieses Kapitel erklärt keine ECoS-Verbindung, Rohprüfung, Synchronisierung,
Konfliktbehandlung oder Schreibvorgänge zur Digitalzentrale. Diese Abläufe gehören zum geplanten
Kapitel Digitalzentralen.

## Daten bei Schreibvorgängen schützen

| Aktion | Speichert Daten | Lädt das vollständige Fahrzeug neu |
| --- | --- | --- |
| Funktionsfeld ohne Speichern bearbeiten | Nein | Nein |
| Einzelne Funktion speichern oder löschen | Sofort | Nach Erfolg |
| Funktions-JSON importieren | Nacheinander | Nur nach vollständigem Erfolg |
| Funktions-JSON exportieren | Nein | Nein |
| Fahrkurve ansehen | Nein | Nein |
| CV-Importvorschau erzeugen oder auswählen | Nein | Nein |
| Ausgewählte CV-Zeilen übernehmen | Nacheinander | Nur nach vollständigem Erfolg |
| Einzelnen CV hinzufügen, speichern oder löschen | Sofort | Nach Erfolg |
| CV-JSON exportieren | Nein | Nein |
| Decoder-Dateivorschau erzeugen | Nein | Nein |
| Metadatenvorschlag übernehmen | Nein | Nein |
| Erkannte CVs prüfen | Nein, bis Zeilen übernommen werden | Nein |
| Erkannte Funktionen übernehmen | Nacheinander | Nur nach vollständigem Erfolg |
| Decoder-Dateien speichern | Nacheinander | Nur nach vollständigem Erfolg |
| Decoder-Datei herunterladen | Nein | Nein |
| Decoder-Datei löschen | Sofort | Nach Erfolg |

Funktionen, CV-Werte, CV-Historie, Decoder-Dateimetadaten und Decoder-Dateiinhalte gehören zur
lokalen RailKeeper-Anwendungssicherung. Erstelle und validiere vor großen Importen oder
Aufräumarbeiten eine aktuelle Sicherung. Funktions- und CV-JSON-Exporte sind Austauschdateien,
keine vollständigen RailKeeper-Sicherungen.

## Fehler bei Decoderdaten beheben

| Situation | Reaktion |
| --- | --- |
| Fahrzeug ist nicht gespeichert | Speichere den Grunddatensatz vor dauerhaften Funktions-, CV- oder Dateiaktionen. |
| Keine Funktion unter **Nur belegte** | Deaktiviere den Filter, trage Name, Symbol oder Notiz ein und speichere. |
| Funktionsimport findet nichts Gültiges | Prüfe JSON-Struktur, F0-F31, gespeicherte Typnamen und Betriebsarten. |
| Fahrkurve ist leer | Erfasse oder importiere CV 2/5/6 oder CV 67-94 in einer Profil-/Protokollgruppe. |
| Kurvenauswahl wirkt falsch | Prüfe CV 29, Profil-/Protokollgruppe und fehlende CVs. |
| CV-Eingabe wird abgelehnt | Nutze eine ganzzahlige CV von 1-1024 und einen Wert von 0-255. |
| CV-Import meldet ein Duplikat | Behalte eine Kombination aus CV-Nummer und Profil in der Quelle. |
| Dateivorschau enthält keine Metadaten | Die Datei kann trotzdem gespeichert werden, es gibt aber keinen Vorschlag. |
| Datei wird abgelehnt | Prüfe Endung, Inhalt, leere Datei und die Größengrenze des Betreibers. |
| Datei ist noch CV-Quelle | Entferne **Quelldatei** an passenden CV-Zeilen vor dem Löschen. |
| Stapel schlägt teilweise fehl | Lade neu, vergleiche gespeicherte Ergebnisse und wiederhole nur fehlende Elemente. |
| Andere Änderungen verschwinden | Neuladen ersetzt Ungespeichertes. Nutze bei Bedarf eine Sicherung. |
| CV oder Datei wurde ohne Nachfrage gelöscht | Löschen wirkt sofort. Wiederherstellung benötigt eine Sicherung. |

Keine später fehlgeschlagene Anfrage macht eine frühere erfolgreiche Anfrage derselben
nacheinander ausgeführten Aktion rückgängig.

## Verwandte Seiten

- [Übersicht des Benutzerhandbuchs](/de/guide/)
- [Fahrzeugbestand und Grunddaten](/de/guide/vehicles/)
- [Fahrzeugbilder und Beilagen](/de/guide/vehicles/media)
- [Fahrzeugwartung und Zustand](/de/guide/vehicles/maintenance)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert die stabile RailKeeper-Version **v0.1.17.6** und wurde zuletzt am
16.08.2026 geprüft.
```

- [ ] **Step 3: Verify page parity and content hygiene**

Run from the repository root:

```powershell
rg -n '^## ' docs/site/guide/vehicles/decoder-cv.md docs/site/de/guide/vehicles/decoder-cv.md
rg -n 'F0|CV 29|67|94|ECoS|25 MiB|sequential' docs/site/guide/vehicles/decoder-cv.md
rg -n 'F0|CV 29|67|94|ECoS|25 MiB|nacheinander' docs/site/de/guide/vehicles/decoder-cv.md
$unfinished = @('TO' + 'DO', 'T' + 'BD', 'FIX' + 'ME', [char]0x2014)
$pages = @('docs/site/guide/vehicles/decoder-cv.md', 'docs/site/de/guide/vehicles/decoder-cv.md')
Select-String -Path $pages -Pattern $unfinished
git diff --check
```

Expected: both pages have the same 11-section semantic order, the key-fact scan finds matching
coverage, the unfinished-content scan prints nothing, and `git diff --check` prints nothing.

- [ ] **Step 4: Run the full documentation check**

Run:

```powershell
Set-Location docs
npm.cmd run check
```

Expected: all 19 tests pass, documentation validation passes, and the VitePress production build
completes successfully.

- [ ] **Step 5: Commit the coverage contract and page pair**

Run from the repository root:

```powershell
git add docs/coverage.json docs/site/guide/vehicles/decoder-cv.md
git add docs/site/de/guide/vehicles/decoder-cv.md
git commit -m "docs: add vehicle decoder and CV guide"
```

---

### Task 3: Add navigation, landing links, and published cross-links

**Files:**
- Modify: `docs/.vitepress/config.mts:51,121`
- Modify: `docs/site/guide/index.md:33`
- Modify: `docs/site/de/guide/index.md:34`
- Modify: `docs/site/guide/vehicles/index.md:299`
- Modify: `docs/site/de/guide/vehicles/index.md:308`
- Modify: `docs/site/guide/vehicles/media.md:133`
- Modify: `docs/site/de/guide/vehicles/media.md:149`
- Modify: `docs/site/guide/vehicles/maintenance.md:132`
- Modify: `docs/site/de/guide/vehicles/maintenance.md:144`

**Interfaces:**
- Consumes: the validated routes `/guide/vehicles/decoder-cv` and
  `/de/guide/vehicles/decoder-cv` from Task 2.
- Produces: discoverable decoder/CV pages linked only from published user-guide pages.

- [ ] **Step 1: Add each sidebar entry after maintenance**

In the English User Guide sidebar, add:

```ts
{ text: "Decoder, functions, and CV data", link: "/guide/vehicles/decoder-cv" }
```

In the German User Guide sidebar, add:

```ts
{ text: "Decoder, Funktionen und CV-Daten", link: "/de/guide/vehicles/decoder-cv" }
```

- [ ] **Step 2: Add one landing-page transition per language**

Append after the maintenance paragraph in `docs/site/guide/index.md`:

```markdown

[Decoder, functions, and CV data](/guide/vehicles/decoder-cv) covers F0-F31 mappings, the read-only
speed curve, CV values and exchange, decoder-file previews, and safe persistence.
```

Append the semantic counterpart in `docs/site/de/guide/index.md`:

```markdown

[Decoder, Funktionen und CV-Daten](/de/guide/vehicles/decoder-cv) behandelt Zuordnungen F0-F31,
die schreibgeschützte Fahrkurve, CV-Werte und Austausch, Decoder-Dateivorschauen und sicheres
Speichern.
```

- [ ] **Step 3: Add decoder/CV to six related-page groups**

Add this link to the related-page lists in the English core vehicle, media, and maintenance pages:

```markdown
- [Decoder, functions, and CV data](/guide/vehicles/decoder-cv)
```

Add this link to the corresponding three German pages:

```markdown
- [Decoder, Funktionen und CV-Daten](/de/guide/vehicles/decoder-cv)
```

Place it after the maintenance link in the core and media pages, and after the media link in the
maintenance page. Do not link to Digital centers, search and spares, or another planned page.

- [ ] **Step 4: Run navigation, link, and full documentation verification**

Run:

```powershell
rg -n 'vehicles/decoder-cv' docs/.vitepress/config.mts docs/site/guide docs/site/de/guide
rg -n '^- \[' docs/site/guide/vehicles/decoder-cv.md docs/site/de/guide/vehicles/decoder-cv.md
git diff --check
Set-Location docs
npm.cmd run check
```

Expected: the route appears in two sidebar links, two landing links, six existing related-page
lists, and the four links on each new page. No whitespace error occurs, all 19 tests pass,
validation succeeds, and VitePress builds successfully.

- [ ] **Step 5: Commit navigation and cross-links**

Run from the repository root:

```powershell
git add docs/.vitepress/config.mts docs/site/guide/index.md docs/site/de/guide/index.md
git add docs/site/guide/vehicles/index.md docs/site/de/guide/vehicles/index.md
git add docs/site/guide/vehicles/media.md docs/site/de/guide/vehicles/media.md
git add docs/site/guide/vehicles/maintenance.md docs/site/de/guide/vehicles/maintenance.md
git commit -m "docs: link vehicle decoder and CV guide"
```

---

### Task 4: Audit stable-source fidelity and clear independent review

**Files:**
- Review: every file changed in `origin/main..HEAD`
- Reference:
  `docs/superpowers/specs/2026-08-16-railkeeper-vehicle-decoder-cv-guide-design.md`
- Reference: all stable source files listed in that specification

**Interfaces:**
- Consumes: the complete committed documentation diff.
- Produces: a review-cleared head commit with no Critical or Important finding and no unresolved
  valid completeness finding.

- [ ] **Step 1: Recheck the highest-risk stable behavior**

Run from the repository root:

```powershell
git show v0.1.17.6:frontend/src/features/vehicles/VehicleFunctionsTab.tsx
git show v0.1.17.6:frontend/src/features/vehicles/VehicleSpeedCurveTab.tsx
git show v0.1.17.6:frontend/src/features/vehicles/VehicleCVTab.tsx
git show v0.1.17.6:frontend/src/features/vehicles/useVehicleFunctionsController.ts
git show v0.1.17.6:frontend/src/features/vehicles/useVehicleCVController.ts
git show v0.1.17.6:frontend/src/features/vehicles/useVehicleDecoderFilesController.ts
git show v0.1.17.6:frontend/src/features/vehicles/cvImport.ts
git show v0.1.17.6:frontend/src/features/vehicles/speedCurve.ts
git show v0.1.17.6:frontend/src/features/vehicles/vehicleFiles.ts
git show v0.1.17.6:backend/internal/api/routes.go
git show v0.1.17.6:backend/internal/api/vehicle_decoder_handlers.go
git show v0.1.17.6:backend/internal/application/vehicle_functions_service.go
git show v0.1.17.6:backend/internal/application/vehicle_cv_service.go
git show v0.1.17.6:backend/internal/application/vehicle_validation.go
git show v0.1.17.6:backend/internal/application/backup.go
```

Confirm F0 defaults, F0-F31 limits, type and mode values, imports/exports, CV identity, CV history,
speed-curve selection, file formats and limits, the four preview actions, ECoS preview limits,
roles, immediate refreshes, partial sequential writes, stale source identifiers after file deletion,
and backup scope.

- [ ] **Step 2: Inspect the full diff and repository state**

Run:

```powershell
git diff --check origin/main..HEAD
git diff --stat origin/main..HEAD
git status --short
git log --oneline origin/main..HEAD
```

Expected: no whitespace errors, only the specification, implementation plan, coverage, page pair,
navigation, landing, and related-page changes are present, and the worktree is clean.

- [ ] **Step 3: Request an independent read-only review**

Use the `requesting-code-review` workflow with:

```text
Base: output of git rev-parse origin/main
Head: output of git rev-parse HEAD
Specification: docs/superpowers/specs/2026-08-16-railkeeper-vehicle-decoder-cv-guide-design.md
Focus: stable v0.1.17.6 fidelity, English/German parity, exact labels and stored values, roles,
F0 defaults, function import/export, speed-curve interpretation, CV identity and history, CV
import/export, decoder-file previews and independent actions, ECoS boundary, full-record refresh,
non-atomic actions, source-file cleanup before deletion, backup scope, coverage, navigation, and
unpublished-link boundaries.
```

The reviewer must not mutate the worktree. Fix every Critical and Important finding. Apply valid
Minor corrections when they improve source fidelity, language parity, safety, or completeness.

- [ ] **Step 4: Verify and commit review corrections**

After corrections, run:

```powershell
Set-Location docs
npm.cmd run check
Set-Location ..
git diff --check origin/main..HEAD
git status --short
```

Expected: all 19 tests pass, validation and VitePress build succeed, no diff error remains, and the
worktree becomes clean after a correction commit such as:

```powershell
git add docs
git commit -m "docs: refine vehicle decoder and CV guide"
```

Request a focused read-only re-review of every correction. Do not publish while a Critical,
Important, or valid completeness finding remains.

---

### Task 5: Publish and merge only when the exact reviewed head is green

**Files:**
- No new source files expected.
- Verify: committed branch `dev/docs-user-guide-decoder-cv`.

**Interfaces:**
- Consumes: a clean, independently reviewed branch with fresh local verification.
- Produces: a merged pull request on `main`, guarded by the exact reviewed head SHA.

- [ ] **Step 1: Run fresh pre-push verification**

Run:

```powershell
Set-Location docs
npm.cmd run check
Set-Location ..
git diff --check origin/main..HEAD
git status --short
git rev-parse HEAD
git merge-base HEAD origin/main
```

Expected: all 19 tests pass, validation and VitePress build succeed, no diff error or uncommitted
file exists, and the merge base is the expected `origin/main` commit.

- [ ] **Step 2: Push only the feature branch**

Run:

```powershell
git push -u origin dev/docs-user-guide-decoder-cv
```

Do not modify or push local `main`.

- [ ] **Step 3: Create and ready the pull request**

Create a draft pull request against `main` titled:

```text
docs: add bilingual vehicle decoder and CV guide
```

Use this body:

```markdown
## Summary

- add complete English and German decoder/CV chapters for stable v0.1.17.6
- document functions, speed curves, CV exchange and history, decoder files, and ECoS preview
- mark decoder/CV coverage documented and connect published guide navigation

## Verification

- `npm.cmd run check`
- stable-tag source audit against `v0.1.17.6`
- independent English/German fidelity, persistence, and safety review

No runtime or API behavior changes are included.
```

Mark the pull request ready only after confirming its remote head SHA equals the locally reviewed
SHA.

- [ ] **Step 4: Monitor every required GitHub check and review thread**

For the exact head SHA, require:

```text
CI: success
Trivy: success
CodeQL: success
```

Inspect all review conversations and resolve a thread only after its concern is corrected or
demonstrably inapplicable. If a check fails or a valid finding changes the branch, fix the root
cause, rerun local verification, push the new commit, and restart exact-head verification.

- [ ] **Step 5: Merge with expected-head protection and verify closure**

Immediately before merging, confirm the pull request is open, non-draft, mergeable, has no
unresolved review thread, and still points to the reviewed SHA. Merge with expected-head
protection, then fetch the pull-request metadata again and require:

```text
state: closed
merged: true
merge_commit_sha: non-empty
```

Leave the isolated worktree in place for traceability and do not modify or push local `main`.
