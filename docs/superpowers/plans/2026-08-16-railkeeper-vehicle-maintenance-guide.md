# RailKeeper Vehicle Maintenance User Guide Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish a complete, source-verified English and German user-guide chapter for vehicle
maintenance and condition in stable RailKeeper v0.1.17.6.

**Architecture:** Add one paired VitePress Markdown chapter under the existing vehicle guide and
change only the existing maintenance coverage topic from planned to documented. Keep all stable
maintenance operations on one page per language, then add navigation and cross-links only after the
content pair passes the documentation gate.

**Tech Stack:** VitePress 2, Markdown, JSON coverage manifest, Node.js documentation tests, GitHub
Actions.

## Global Constraints

- Document only stable RailKeeper **v0.1.17.6**. Do not describe later `main` behavior as stable.
- Create both `docs/site/guide/vehicles/maintenance.md` and
  `docs/site/de/guide/vehicles/maintenance.md` with semantically equivalent content.
- Use the exact page metadata, public routes, stable UI labels, stored values, and section order from
  the approved design specification.
- Cover entry creation, editing, cancellation, completion, deletion, counters, sorting, validation,
  immediate persistence, full-record refresh, linked media, roles, backup scope, and error states.
- Treat media upload/editing, decoder/CV, document search, spare parts, dashboard operation, and
  administrative backup/restore as boundaries. Do not reproduce those workflows.
- Do not add screenshots or links to unpublished pages.
- Keep lines readable, avoid em dashes, and do not introduce unfinished-content markers.
- Do not change runtime, API, validation, storage, or deletion behavior.
- Do not commit `docs/.vitepress/dist`, dependency caches, local screenshots, or generated output.
- Preserve the dirty local `main`; all work stays in the isolated
  `dev/docs-user-guide-vehicle-maintenance` worktree.
- Merge the GitHub pull request only after CI, Trivy, and CodeQL all succeed for the unchanged,
  reviewed head SHA and every actionable review thread is resolved.

---

### Task 1: Mark maintenance documented and prove the missing-page gate

**Files:**
- Modify: `docs/coverage.json:33`
- Reference: `docs/scripts/validate-docs.mjs`
- Reference: `docs/superpowers/specs/2026-08-16-railkeeper-vehicle-maintenance-guide-design.md`

**Interfaces:**
- Consumes: existing translation owner `vehicles.maintenance` and API owner
  `/api/v1/vehicles/{id}/maintenance` in `docs/coverage.json`.
- Produces: a documented `vehicle-maintenance` contract whose only missing artifacts are its exact
  English and German page paths.

- [ ] **Step 1: Confirm the existing source ownership and page paths**

Run from the repository root:

```powershell
rg -n 'vehicle-maintenance|vehicles\.maintenance|vehicles/\{id\}/maintenance' docs/coverage.json
```

Expected: the topic is `planned`, its paths are `guide/vehicles/maintenance.md` and
`de/guide/vehicles/maintenance.md`, and the existing translation/API prefixes both belong to
`vehicle-maintenance`.

- [ ] **Step 2: Change only the maintenance topic status**

Replace the topic object with exactly:

```json
{
  "id": "vehicle-maintenance",
  "audience": "user",
  "status": "documented",
  "englishPath": "guide/vehicles/maintenance.md",
  "germanPath": "de/guide/vehicles/maintenance.md"
}
```

Do not change either owner prefix.

- [ ] **Step 3: Run the documentation check and verify the intentional failure**

Run:

```powershell
Set-Location docs
npm.cmd run check
```

Expected: all 19 unit tests pass, then coverage validation fails only because
`guide/vehicles/maintenance.md` and `de/guide/vehicles/maintenance.md` are absent. Stop and correct
any JSON, ownership, metadata, or unrelated validation failure before continuing.

Do not commit the intentionally red state.

---

### Task 2: Create the complete English and German maintenance chapter

**Files:**
- Create: `docs/site/guide/vehicles/maintenance.md`
- Create: `docs/site/de/guide/vehicles/maintenance.md`
- Modify: `docs/coverage.json` (already changed and uncommitted in Task 1)
- Reference: `frontend/src/features/vehicles/VehicleMaintenanceTab.tsx` at tag `v0.1.17.6`
- Reference: `frontend/src/features/vehicles/useVehicleMaintenanceController.ts` at tag `v0.1.17.6`
- Reference: `frontend/src/features/vehicles/vehicleMaintenance.ts` at tag `v0.1.17.6`
- Reference: `frontend/src/features/vehicles/vehicleOptions.ts` at tag `v0.1.17.6`
- Reference: `frontend/src/features/vehicles/VehiclesView.tsx` at tag `v0.1.17.6`
- Reference: `frontend/src/features/vehicles/useVehicleEditorController.ts` at tag `v0.1.17.6`
- Reference: `frontend/src/features/vehicles/VehicleUploadsTab.tsx` at tag `v0.1.17.6`
- Reference: `backend/internal/api/routes.go` at tag `v0.1.17.6`
- Reference: `backend/internal/application/vehicle_maintenance_service.go` at tag `v0.1.17.6`
- Reference: `backend/internal/application/vehicle_validation.go` at tag `v0.1.17.6`
- Reference: `backend/internal/application/vehicle_media.go` at tag `v0.1.17.6`
- Reference: `backend/internal/application/backup.go` at tag `v0.1.17.6`
- Reference: `backend/migrations/0016_maintenance_media_links.sql` at tag `v0.1.17.6`

**Interfaces:**
- Consumes: the exact maintenance page paths from Task 1.
- Produces: a validated language pair at `/guide/vehicles/maintenance` and
  `/de/guide/vehicles/maintenance`.

- [ ] **Step 1: Create the English page with complete stable content**

Create `docs/site/guide/vehicles/maintenance.md` with exactly this structure and content:

```markdown
---
title: Vehicle maintenance and condition
description: Record, schedule, complete, review, and safely remove vehicle maintenance entries.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Vehicle maintenance and condition

The **Maintenance** tab records work, repairs, conversions, condition, dates, and costs for one
vehicle. Admin, Editor, Viewer, and Planner can inspect stored entries. Only Admin and Editor can
create, change, complete, or delete them, and the server enforces that boundary.

## Open the Maintenance tab

Open a vehicle from **Vehicle inventory**, choose **Edit**, then select **Maintenance**. The vehicle
must have been saved once because maintenance needs its stored vehicle ID. A new unsaved vehicle
shows **Maintenance entries can be added after the first save.** Save the core record before adding
an entry.

## Read the maintenance summary

The top of the tab shows three counters:

- **Due** counts entries with a due date on or before the current local date whose status is not
  `erledigt`.
- **Planned/open** counts every entry whose status is not `erledigt`, including entries already
  counted as due.
- **Done** counts entries whose status is `erledigt`.

**Due** is therefore a subset of **Planned/open**, not a separate status total. A `geplant` entry
whose date is today or earlier is displayed and counted as due even though its stored status is not
`faellig`.

RailKeeper lists open entries before done entries. Within each group, dated entries come before
undated entries, earlier due dates come first, and equal entries use newest creation first.

## Add a maintenance entry

Save or intentionally discard all other pending vehicle changes first. Enter the maintenance data
and choose **Add entry**. Creation is immediate and does not wait for the vehicle's **Save changes**
action. After success, RailKeeper reloads the complete vehicle and resets the form to type
`Wartung` and status `geplant`.

## Fields, values, and validation

- **Type:** Required. The stored German values are `Wartung`, `Reparatur`, `Umbau`, `Superung`,
  `Reinigung`, `Schmierung`, `Decoder-Einbau`, and `Ersatzteiltausch`. The English labels are
  Maintenance, Repair, Conversion, Detail upgrade, Cleaning, Lubrication, Decoder installation,
  and Spare part replacement. Default: `Wartung`.
- **Status:** Required. Stored values are `geplant`, `faellig`, and `erledigt`; shown in English as
  planned, due, and done. The German spelling `fällig` is normalized to `faellig`. Default:
  `geplant`.
- **Condition:** Optional. Stored values are `neuwertig`, `sehr gut`, `gut`, `gebraucht`, and
  `reparaturbedürftig`. They remain German in stored data.
- **Due on:** Optional valid calendar date in `YYYY-MM-DD` form. It controls date-based due
  calculation.
- **Completed on:** Optional valid calendar date in `YYYY-MM-DD` form. Saving a done entry without
  this date supplies today's local date.
- **Cost:** Optional non-negative decimal amount. Comma and point decimal separators are accepted.
  Surrounding spaces, internal spaces, and one trailing euro sign are removed before validation.
  Accepted numbers are displayed as EUR using the German locale.
- **Notes:** Optional. The server trims the value and accepts at most 4,000 characters.

The server rejects unknown types, statuses, and conditions, invalid dates, negative or non-numeric
costs, and notes above the limit.

## Understand due dates and completion

The stored status `faellig` and date-based due highlighting are related but independent. Selecting
the due status does not add a due date. Conversely, an open entry with a due date on or before today
is highlighted and counted as due even when its stored status remains `geplant`.

An entry stops contributing to **Due** and **Planned/open** only when its status becomes `erledigt`.
The completion date is descriptive; it does not by itself mark the entry done.

## Edit or cancel an entry

Choose **Edit maintenance** on a card to copy its values into the form. The main action becomes
**Save entry**, and **Cancel** appears. Saving updates the entry immediately and reloads the complete
vehicle. **Cancel** resets the maintenance form without changing the stored entry.

## Mark an entry done

Choose **Done** on an open card to complete it immediately. RailKeeper does not ask for separate
confirmation. It preserves the type, condition, due date, cost, notes, and any existing completion
date. If the completion date is empty, RailKeeper uses the current local date.

## Check linked media

When at least one link exists, the maintenance card shows image and attachment counts below
**Linked media**. Manage image links in **Uploads**. Selecting **No maintenance** for an image is
pending metadata until you use the vehicle's **Save changes** action.

Stable v0.1.17.6 can display attachments that already contain a maintenance reference, but the
normal attachment row cannot assign or change that reference. Do not invent a reassignment flow
that the interface does not provide.

## Delete maintenance safely

**Delete maintenance** removes the entry immediately without a confirmation dialog. The backend
does not block deletion for linked media and does not delete or detach those media records. Their
stored maintenance ID can therefore point to a deleted entry.

Before deleting an entry:

1. Save or intentionally discard every other pending vehicle change.
2. Check its linked image and attachment counts.
3. In **Uploads**, select **No maintenance** for every linked image and use **Save changes**.
4. If a linked attachment must retain the association, keep the maintenance entry. Stable
   v0.1.17.6 has no attachment-link editor. Remove an attachment only after backup and content
   review when the file is genuinely no longer needed.
5. Return to **Maintenance**, confirm that **Linked media** no longer appears, then delete the entry.

Deleting maintenance does not delete an image or attachment file. A stale, non-empty maintenance
ID can later prevent image deletion until the link is cleared and saved.

## Roles, storage, and backup boundaries

Admin, Editor, Viewer, and Planner can read maintenance. Server-side create, update, complete, and
delete operations require Admin or Editor. Disabled controls help explain the current mode but do
not replace server authorization.

Maintenance rows and vehicle uploads are local RailKeeper application data and belong to the
application backup scope. Before substantial cleanup, create a current backup and have an Admin
validate it. Use [Vehicle images and attachments](/guide/vehicles/media) for the published media
safety workflow. Backup export, validation, and restore operation belong to administration and are
not repeated here.

## Empty and error states

- **Vehicle is not saved:** Save the core record before adding maintenance.
- **No entry exists:** Use **Add entry** after checking that the correct vehicle is open.
- **Input is rejected:** Check type, status, condition, dates, non-negative cost, and note length.
- **Write action fails:** Keep the form open, read the error, check the session and connection,
  then retry.
- **Summary looks outdated:** Confirm that no unsaved changes remain, then reload the vehicle.
- **Due count differs from status:** Due is calculated from date and completion, not only the
  stored status.
- **Linked media is shown:** Resolve the links before deletion and use the media page for image
  metadata.
- **Entry was deleted without a prompt:** Deletion is immediate. Recovery requires a suitable
  backup.

A failed action does not undo an earlier successful independent maintenance or media action.

## Related pages

- [User Guide overview](/guide/)
- [Overview, metrics, and data quality](/guide/overview/)
- [Vehicle inventory and core records](/guide/vehicles/)
- [Vehicle images and attachments](/guide/vehicles/media)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.17.6** and was last reviewed on 2026-08-16.
```

- [ ] **Step 2: Create the German page with semantic parity and exact UI labels**

Create `docs/site/de/guide/vehicles/maintenance.md` with exactly this structure and content:

```markdown
---
title: Fahrzeugwartung und Zustand
description: Fahrzeugwartungen erfassen, planen, abschließen, prüfen und sicher entfernen.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Fahrzeugwartung und Zustand

Der Tab **Wartung** erfasst Arbeiten, Reparaturen, Umbauten, Zustand, Termine und Kosten an einem
Fahrzeug. Admin, Editor, Viewer und Planner können gespeicherte Einträge ansehen. Nur Admin und
Editor dürfen sie anlegen, ändern, abschließen oder löschen. Der Server erzwingt diese Grenze.

## Tab Wartung öffnen

Öffne ein Fahrzeug im **Fahrzeugbestand**, wähle **Bearbeiten** und dann **Wartung**. Das Fahrzeug
muss bereits einmal gespeichert worden sein, weil Wartungen seine gespeicherte Fahrzeug-ID
benötigen. Ein neues ungespeichertes Fahrzeug zeigt **Wartungseinträge können nach dem ersten
Speichern hinzugefügt werden.** Speichere den Grunddatensatz, bevor du einen Eintrag anlegst.

## Wartungsübersicht lesen

Oben im Tab stehen drei Zähler:

- **Fällig** zählt Einträge mit einem Fälligkeitsdatum bis einschließlich des heutigen lokalen
  Datums, deren Status nicht `erledigt` ist.
- **Geplant/offen** zählt jeden Eintrag, dessen Status nicht `erledigt` ist, einschließlich der
  bereits als fällig gezählten Einträge.
- **Erledigt** zählt Einträge mit dem Status `erledigt`.

**Fällig** ist daher eine Teilmenge von **Geplant/offen**, keine getrennte Statussumme. Ein Eintrag
mit Status `geplant` und einem Datum bis heute wird als fällig angezeigt und gezählt, obwohl sein
gespeicherter Status nicht `faellig` lautet.

RailKeeper zeigt offene Einträge vor erledigten. Innerhalb beider Gruppen stehen Einträge mit Datum
vor Einträgen ohne Datum, frühere Fälligkeitsdaten zuerst und bei gleichen Werten die zuletzt
angelegten Einträge zuerst.

## Wartungseintrag hinzufügen

Speichere oder verwirf zuerst bewusst alle anderen ausstehenden Fahrzeugänderungen. Trage die
Wartungsdaten ein und wähle **Eintrag hinzufügen**. Das Anlegen wirkt sofort und wartet nicht auf
**Änderungen speichern** am Fahrzeug. Nach Erfolg lädt RailKeeper das vollständige Fahrzeug neu und
setzt das Formular auf Art `Wartung` und Status `geplant` zurück.

## Felder, Werte und Validierung

- **Art:** Erforderlich. Gespeicherte Werte sind `Wartung`, `Reparatur`, `Umbau`, `Superung`,
  `Reinigung`, `Schmierung`, `Decoder-Einbau` und `Ersatzteiltausch`. Standard: `Wartung`.
- **Status:** Erforderlich. Gespeicherte Werte sind `geplant`, `faellig` und `erledigt`; angezeigt
  als geplant, fällig und erledigt. Die Schreibweise `fällig` wird beim Speichern zu `faellig`
  normalisiert. Standard: `geplant`.
- **Zustand:** Optional. Gespeicherte Werte sind `neuwertig`, `sehr gut`, `gut`, `gebraucht` und
  `reparaturbedürftig`.
- **Fällig am:** Optionales gültiges Kalenderdatum im Format `YYYY-MM-DD`. Es steuert die
  datumsbasierte Fälligkeitsberechnung.
- **Durchgeführt am:** Optionales gültiges Kalenderdatum im Format `YYYY-MM-DD`. Beim Speichern
  eines erledigten Eintrags ohne dieses Datum wird das heutige lokale Datum eingesetzt.
- **Kosten:** Optionaler nicht negativer Dezimalbetrag. Komma und Punkt werden als Dezimaltrenner
  akzeptiert. Äußere und innere Leerzeichen sowie ein nachgestelltes Eurozeichen werden vor der
  Prüfung entfernt. Gültige Zahlen zeigt die Oberfläche als EUR im deutschen Zahlenformat an.
- **Notizen:** Optional. Der Server entfernt äußere Leerzeichen und akzeptiert höchstens 4.000
  Zeichen.

Der Server weist unbekannte Arten, Status und Zustände, ungültige Daten, negative oder nicht
numerische Kosten sowie zu lange Notizen zurück.

## Fälligkeit und Abschluss verstehen

Der gespeicherte Status `faellig` und die datumsbasierte Fälligkeitsmarkierung hängen zusammen,
sind aber unabhängig. Die Auswahl des Status fällig ergänzt kein Fälligkeitsdatum. Umgekehrt wird
ein offener Eintrag mit einem Fälligkeitsdatum bis heute hervorgehoben und gezählt, auch wenn sein
Status weiter `geplant` lautet.

Ein Eintrag zählt erst dann nicht mehr zu **Fällig** und **Geplant/offen**, wenn sein Status
`erledigt` ist. Das Abschlussdatum beschreibt den Abschluss, markiert den Eintrag allein aber nicht
als erledigt.

## Eintrag bearbeiten oder abbrechen

Wähle **Wartung bearbeiten** an einer Karte, um ihre Werte in das Formular zu übernehmen. Die
Hauptaktion lautet nun **Eintrag speichern**, zusätzlich erscheint **Abbrechen**. Speichern ändert
den Eintrag sofort und lädt das vollständige Fahrzeug neu. **Abbrechen** setzt das Wartungsformular
zurück, ohne den gespeicherten Eintrag zu ändern.

## Eintrag als erledigt markieren

Wähle **Erledigt** an einer offenen Karte, um sie sofort abzuschließen. RailKeeper fragt nicht
zusätzlich nach einer Bestätigung. Art, Zustand, Fälligkeitsdatum, Kosten, Notizen und ein bereits
vorhandenes Abschlussdatum bleiben erhalten. Fehlt das Abschlussdatum, verwendet RailKeeper das
heutige lokale Datum.

## Verknüpfte Medien prüfen

Sobald mindestens eine Verknüpfung besteht, zeigt die Wartungskarte unter **Verknüpfte Medien**
getrennte Zähler für Bilder und Beilagen. Bildverknüpfungen werden im Tab **Uploads** verwaltet. Die
Auswahl **Keine Wartung** an einem Bild bleibt eine ausstehende Metadatenänderung, bis du
**Änderungen speichern** am Fahrzeug verwendest.

RailKeeper v0.1.17.6 kann Beilagen anzeigen, die bereits eine Wartungsreferenz enthalten. Die
normale Beilagenzeile bietet jedoch keine Funktion zum Zuweisen oder Ändern dieser Referenz. Nutze
keinen erfundenen Zuordnungsablauf, den die Oberfläche nicht bereitstellt.

## Wartung sicher löschen

**Wartung löschen** entfernt den Eintrag sofort ohne Bestätigungsdialog. Der Server blockiert das
Löschen bei verknüpften Medien nicht und löscht oder trennt diese Medien ebenfalls nicht. Ihre
gespeicherte Wartungs-ID kann deshalb auf einen gelöschten Eintrag zeigen.

Vor dem Löschen eines Eintrags:

1. Speichere oder verwirf bewusst jede andere ausstehende Fahrzeugänderung.
2. Prüfe die Zähler für verknüpfte Bilder und Beilagen.
3. Wähle im Tab **Uploads** für jedes verknüpfte Bild **Keine Wartung** und nutze
   **Änderungen speichern**.
4. Muss eine verknüpfte Beilage ihre Zuordnung behalten, behalte auch den Wartungseintrag.
   RailKeeper v0.1.17.6 besitzt keinen Editor für Beilagenverknüpfungen. Entferne eine Beilage erst
   nach Sicherungs- und Inhaltsprüfung, wenn die Datei tatsächlich nicht mehr benötigt wird.
5. Kehre zu **Wartung** zurück, prüfe, dass **Verknüpfte Medien** nicht mehr erscheint, und lösche
   erst danach den Eintrag.

Das Löschen der Wartung löscht keine Bild- oder Beilagendatei. Eine veraltete, nicht leere
Wartungs-ID kann das spätere Löschen eines Bildes blockieren, bis die Verknüpfung entfernt und
gespeichert wurde.

## Rollen-, Speicher- und Sicherungsgrenzen

Admin, Editor, Viewer und Planner können Wartungen lesen. Serverseitiges Anlegen, Ändern,
Abschließen und Löschen erfordert Admin oder Editor. Deaktivierte Bedienelemente erklären den
aktuellen Modus, ersetzen aber nicht die serverseitige Berechtigungsprüfung.

Wartungszeilen und Fahrzeugmedien sind lokale RailKeeper-Anwendungsdaten und gehören zum Umfang der
Anwendungssicherung. Erstelle vor umfangreichen Aufräumarbeiten eine aktuelle Sicherung und lasse
sie von einem Admin validieren. Das veröffentlichte Vorgehen für Medien steht unter
[Fahrzeugbilder und Beilagen](/de/guide/vehicles/media). Export, Validierung und Wiederherstellung
gehören zur Administration und werden hier nicht wiederholt.

## Leere und fehlerhafte Zustände

- **Fahrzeug ist nicht gespeichert:** Speichere den Grunddatensatz, bevor du eine Wartung
  hinzufügst.
- **Kein Eintrag vorhanden:** Nutze **Eintrag hinzufügen**, nachdem du das richtige Fahrzeug
  geprüft hast.
- **Eingabe wird abgelehnt:** Prüfe Art, Status, Zustand, Daten, nicht negative Kosten und
  Notizlänge.
- **Schreibaktion schlägt fehl:** Lasse das Formular offen, lies den Fehler, prüfe Sitzung und
  Verbindung und versuche es erneut.
- **Übersicht wirkt veraltet:** Stelle sicher, dass keine ungespeicherten Änderungen bestehen, und
  lade das Fahrzeug neu.
- **Fällig-Zähler passt nicht zum Status:** Fälligkeit wird aus Datum und Abschluss berechnet, nicht
  nur aus dem gespeicherten Status.
- **Verknüpfte Medien werden angezeigt:** Löse die Verknüpfungen vor dem Löschen und nutze die
  Medienseite für Bildmetadaten.
- **Eintrag wurde ohne Nachfrage gelöscht:** Das Löschen wirkt sofort. Wiederherstellung benötigt
  eine geeignete Sicherung.

Eine fehlgeschlagene Aktion macht eine frühere erfolgreiche, unabhängige Wartungs- oder
Medienaktion nicht rückgängig.

## Verwandte Seiten

- [Übersicht des Benutzerhandbuchs](/de/guide/)
- [Übersicht, Kennzahlen und Datenqualität](/de/guide/overview/)
- [Fahrzeugbestand und Grunddaten](/de/guide/vehicles/)
- [Fahrzeugbilder und Beilagen](/de/guide/vehicles/media)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert die stabile RailKeeper-Version **v0.1.17.6** und wurde zuletzt am
16.08.2026 geprüft.
```

- [ ] **Step 3: Verify section parity and content hygiene**

Run from the repository root:

```powershell
rg -n '^## ' docs/site/guide/vehicles/maintenance.md docs/site/de/guide/vehicles/maintenance.md
$unfinished = @('TO' + 'DO', 'T' + 'BD', 'FIX' + 'ME', [char]0x2014)
$pages = @('docs/site/guide/vehicles/maintenance.md', 'docs/site/de/guide/vehicles/maintenance.md')
Select-String -Path $pages -Pattern $unfinished
git diff --check
```

Expected: both pages have the same 13-section semantic order, the unfinished-content scan prints
nothing, and `git diff --check` prints nothing.

- [ ] **Step 4: Run the full documentation check**

Run:

```powershell
Set-Location docs
npm.cmd run check
```

Expected: 19 tests pass, documentation validation passes, and the VitePress production build
completes successfully.

- [ ] **Step 5: Commit the coverage contract and paired pages**

Run from the repository root:

```powershell
git add docs/coverage.json docs/site/guide/vehicles/maintenance.md docs/site/de/guide/vehicles/maintenance.md
git commit -m "docs: add vehicle maintenance user guide"
```

---

### Task 3: Add navigation, landing links, and published cross-links

**Files:**
- Modify: `docs/.vitepress/config.mts:50,119`
- Modify: `docs/site/guide/index.md:30`
- Modify: `docs/site/de/guide/index.md:31`
- Modify: `docs/site/guide/vehicles/index.md:299`
- Modify: `docs/site/de/guide/vehicles/index.md:308`
- Modify: `docs/site/guide/vehicles/media.md:133`
- Modify: `docs/site/de/guide/vehicles/media.md:149`
- Modify: `docs/site/guide/overview/index.md:194`
- Modify: `docs/site/de/guide/overview/index.md:206`

**Interfaces:**
- Consumes: the validated routes `/guide/vehicles/maintenance` and
  `/de/guide/vehicles/maintenance` from Task 2.
- Produces: discoverable maintenance pages connected only to already published guide pages.

- [ ] **Step 1: Add each sidebar entry immediately after the media page**

Add exactly:

```ts
{ text: "Vehicle maintenance and condition", link: "/guide/vehicles/maintenance" }
```

and:

```ts
{ text: "Fahrzeugwartung und Zustand", link: "/de/guide/vehicles/maintenance" }
```

- [ ] **Step 2: Add one concise landing-page transition per language**

Append this paragraph to `docs/site/guide/index.md`:

```markdown

[Vehicle maintenance and condition](/guide/vehicles/maintenance) explains maintenance entries,
due dates, completion, costs, linked media, and safe deletion.
```

Append the semantic counterpart to `docs/site/de/guide/index.md`:

```markdown

[Fahrzeugwartung und Zustand](/de/guide/vehicles/maintenance) erklärt Wartungseinträge,
Fälligkeiten, Abschluss, Kosten, verknüpfte Medien und sicheres Löschen.
```

- [ ] **Step 3: Add maintenance to all six relevant related-page groups**

Add this English link to the related-page lists in the core vehicle, media, and overview chapters:

```markdown
- [Vehicle maintenance and condition](/guide/vehicles/maintenance)
```

Add this German link to the corresponding three lists:

```markdown
- [Fahrzeugwartung und Zustand](/de/guide/vehicles/maintenance)
```

Place each maintenance link next to the existing vehicle-media link where present. In the overview
chapter, place it after the User Guide overview link. Do not link to decoder/CV,
search-and-spares, backup/restore, or another planned page.

- [ ] **Step 4: Run link, whitespace, and full documentation verification**

Run:

```powershell
rg -n 'vehicles/maintenance' docs/.vitepress/config.mts docs/site/guide docs/site/de/guide
rg -n '^- \[' docs/site/guide/vehicles/maintenance.md docs/site/de/guide/vehicles/maintenance.md
git diff --check
Set-Location docs
npm.cmd run check
```

Expected: the maintenance route appears in two sidebar links, two landing links, and six existing
related-page lists. Each new page has four related links. No whitespace errors occur, all 19 tests
pass, validation succeeds, and VitePress builds successfully.

- [ ] **Step 5: Commit navigation and cross-links**

Run from the repository root:

```powershell
git add docs/.vitepress/config.mts docs/site/guide/index.md docs/site/de/guide/index.md
git add docs/site/guide/vehicles/index.md docs/site/de/guide/vehicles/index.md
git add docs/site/guide/vehicles/media.md docs/site/de/guide/vehicles/media.md
git add docs/site/guide/overview/index.md docs/site/de/guide/overview/index.md
git commit -m "docs: link vehicle maintenance guide"
```

---

### Task 4: Audit stable-source fidelity and clear independent review

**Files:**
- Review: every file changed in `origin/main..HEAD`
- Reference: `docs/superpowers/specs/2026-08-16-railkeeper-vehicle-maintenance-guide-design.md`
- Reference: all stable source files listed in Task 2

**Interfaces:**
- Consumes: the complete committed documentation diff.
- Produces: a review-cleared head commit with no Critical or Important findings and no unresolved
  valid completeness finding.

- [ ] **Step 1: Recheck the highest-risk stable behavior directly**

Run from the repository root:

```powershell
git show v0.1.17.6:frontend/src/features/vehicles/VehicleMaintenanceTab.tsx
git show v0.1.17.6:frontend/src/features/vehicles/useVehicleMaintenanceController.ts
git show v0.1.17.6:frontend/src/features/vehicles/vehicleMaintenance.ts
git show v0.1.17.6:frontend/src/features/vehicles/VehiclesView.tsx
git show v0.1.17.6:frontend/src/features/vehicles/useVehicleEditorController.ts
git show v0.1.17.6:frontend/src/features/vehicles/VehicleUploadsTab.tsx
git show v0.1.17.6:backend/internal/api/routes.go
git show v0.1.17.6:backend/internal/application/vehicle_maintenance_service.go
git show v0.1.17.6:backend/internal/application/vehicle_validation.go
git show v0.1.17.6:backend/internal/application/vehicle_media.go
git show v0.1.17.6:backend/internal/application/backup.go
git show v0.1.17.6:backend/migrations/0016_maintenance_media_links.sql
```

Confirm exact fields and stored values, date and cost cleaning, notes limit, completion date,
immediate actions, absence of delete confirmation, `refreshSelectedVehicle()` behavior, due subset
logic, sorting, linked-media counts, lack of a media foreign key, and lack of an attachment-link
editor.

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
Specification: docs/superpowers/specs/2026-08-16-railkeeper-vehicle-maintenance-guide-design.md
Focus: stable v0.1.17.6 fidelity, English/German parity, exact labels and values, roles, due logic,
sorting, persistence timing, refresh data-loss warning, linked-media safety, delete behavior,
coverage, navigation, and unpublished-link boundaries.
```

The reviewer must not mutate the worktree. Fix every Critical and Important finding. Apply valid
Minor corrections when they improve source fidelity, language parity, or completeness.

- [ ] **Step 4: Verify and commit any review corrections**

After corrections, run:

```powershell
Set-Location docs
npm.cmd run check
Set-Location ..
git diff --check origin/main..HEAD
git status --short
```

Expected: 19 tests pass, validation and VitePress build succeed, no diff errors remain, and the
worktree becomes clean after a correction commit such as:

```powershell
git add docs
git commit -m "docs: refine vehicle maintenance guide"
```

Request a focused read-only re-review of all corrections. Do not publish while a Critical,
Important, or valid completeness finding remains.

---

### Task 5: Publish the reviewed branch and merge only when GitHub is green

**Files:**
- No new source files expected.
- Verify: committed branch `dev/docs-user-guide-vehicle-maintenance`.

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

Expected: 19 tests pass, validation and VitePress build succeed, no diff errors or uncommitted files
exist, and the merge base is the expected `origin/main` commit.

- [ ] **Step 2: Push only the feature branch**

Run:

```powershell
git push -u origin dev/docs-user-guide-vehicle-maintenance
```

Do not modify or push local `main`.

- [ ] **Step 3: Create and ready the pull request**

Create a draft pull request against `main` titled:

```text
docs: add bilingual vehicle maintenance guide
```

Use this body:

```markdown
## Summary

- add complete English and German vehicle-maintenance chapters for stable v0.1.17.6
- document fields, counters, sorting, persistence, linked media, and safe deletion
- mark maintenance coverage documented and connect published guide navigation

## Verification

- `npm.cmd run check`
- stable-tag source audit against `v0.1.17.6`
- independent English/German fidelity and safety review

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

Inspect all review conversations and resolve a thread only after its concern is either corrected or
demonstrably inapplicable. If a check fails or a valid finding changes the branch, fix the root
cause, rerun local verification, push the new commit, and restart exact-head verification.

- [ ] **Step 5: Merge with expected-head protection and verify closure**

Immediately before merging, confirm that the pull request is open, non-draft, mergeable, has no
unresolved review thread, and still points to the reviewed SHA. Merge with expected-head protection,
then fetch the pull-request metadata again and require:

```text
state: closed
merged: true
merge_commit_sha: non-empty
```

Leave the isolated worktree in place for traceability and do not modify or push local `main`.
