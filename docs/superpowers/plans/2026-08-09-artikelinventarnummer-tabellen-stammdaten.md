# Artikelinventarnummer, Tabellenstruktur und Stammdatenfelder Implementation Plan

> **Ausführungsregel:** Dieser Plan wird ausschließlich inline mit
> `aegis:executing-plans` umgesetzt. Es werden keine Subagents gestartet.

**Goal:** Jeder Artikelstamm erhält eine automatische Inventarnummer. Die Artikelübersicht folgt
der sortierbaren Tabellenstruktur des Fahrzeugbestands, die Kennzahlen zeigen Summen und dezente
Teilwerte, und der Artikeldialog bindet vorhandene Stammdatenquellen als Auswahlfelder ein.

**Architecture:** Die bestehende Tabelle `inventory_number_schemes` bleibt die einzige Quelle für
laufende Nummern. Eine gemeinsame transaktionale Reservierungsfunktion wird von Fahrzeugen und
Artikeln genutzt. `accessory_products` speichert die vergebene Nummer. Listenquery, API,
OpenAPI und Backup/Restore tragen sie durch. Das Frontend lädt Hersteller, Spurweiten und
Bestandseinheiten über die vorhandene Master-Data-API und verwendet die app-eigenen
Auswahlkomponenten.

**Tech Stack:** Go, SQLite, React 19, TypeScript, Vite, Vitest, Testing Library,
`openapi/railkeeper.yaml`, bestehende RailKeeper-Komponenten und Design-Tokens.

**Baseline/Authority Refs:**

- `docs/superpowers/specs/2026-08-09-artikelinventarnummer-tabellen-stammdaten-design.md`
- `backend/migrations/0008_inventory_numbers.sql`
- `backend/migrations/0017_remove_accessory_inventory_scheme.sql`
- `backend/migrations/0041_article_management_redesign.sql`
- `backend/internal/application/inventory_numbers.go`
- `backend/internal/application/vehicle_persistence.go`
- `backend/internal/application/backup.go`
- `frontend/src/features/vehicles/VehicleInventoryPanel.tsx`

**Compatibility Boundary:** Bestehende Fahrzeuge, Artikel, Einzelstück-Inventarnummern,
Rollen, CSRF, Bestandsstrategien und Backups der Versionen 1 bis 3 bleiben nutzbar. Historische
Migrationen werden nicht verändert. Das Backupformat bleibt bei Version 3, weil das neue Feld
additiv ist und alte Dokumente beim Restore ergänzt werden.

**TDD Route:**

- Mode: off
- Decision: skipped
- Strict authority: not applicable
- Test posture: post-change regression
- Reason: Weder Nutzer noch Projekt verlangen einen strikten RED/GREEN-Zyklus.
- Verification: fokussierte Regressionstests nach jeder Änderung, danach vollständige Backend- und
  Frontendprüfung.

## Plan Basis

### BaselineUsageDraft

- Required baseline refs: freigegebene Spezifikation, Migrationen 0008/0017/0041,
  Inventarnummernservice, Artikelrepository, Listenquery, Backupservice, OpenAPI und
  Fahrzeugbestandsreferenz.
- Delivered context refs: Session-Handover vom 2026-08-09 08:00 und drei Nutzerbilder.
- Acknowledged before plan refs: alle erforderlichen Quellen.
- Cited in plan refs: siehe `Baseline/Authority Refs` und Dateilisten der Tasks.
- Missing refs: keine.
- Decision: continue.

### Requirement Ready Check

- Requirement source refs: Nutzeranforderung und freigegebene Designsitzung vom 2026-08-09.
- Goals and scope refs: freigegebene Spezifikation.
- User / scenario refs: Artikelanlage, Artikelübersicht, Inventarnummerneinstellungen und
  bestehende Artikelmigration.
- Requirement item refs: Kennzahlen, Artikelinventarnummer, sortierbare Tabelle, Auswahl,
  Primärbild und stammdatengebundene Felder.
- Acceptance / verification criteria refs: Abschnitt `Verifikation` der Spezifikation.
- Open blocker questions: keine.
- Decision: ready.

### Change Necessity

- User-visible need: Artikel besitzen keine eigene Inventarnummer, die Tabelle und mehrere
  Formularfelder entsprechen nicht der freigegebenen Arbeitsweise.
- No-change / non-code option: Dokumentation oder Konfiguration können weder Datenpersistenz noch
  automatische Vergabe, Sortierung oder Dropdownbindung herstellen.
- Why code change is necessary: Schema, Transaktion, API, Restore und UI müssen denselben neuen
  Vertrag tragen.
- Minimum change boundary: nächste Migration, vorhandener Nummernowner, Artikelrepository und
  Query, Backupkompatibilität, OpenAPI, Artikel-Frontend und zugehörige Tests.
- Decision: code-change.

### Existence Check

- Proposed new surface: fokussierter Restore-Helfer und fokussierter Hook für Kernstammdaten.
- Existing owner / reuse candidate: `backup.go` und `useArticleEditorController.ts`.
- Why existing surface is insufficient: `backup.go` hat 1114 Zeilen, der Controller 532 Zeilen;
  weitere eigenständige Nummern- bzw. Ladelogik würde zentrale Dateien weiter überladen.
- Creation proof: beide Hilfsdateien kapseln genau eine neue Verantwortung und werden nur vom
  bestehenden Owner aufgerufen.
- Entropy / retirement impact: keine parallelen Nummern- oder API-Owner; die Helfer werden gelöscht,
  falls ihre einzige Aufrufstelle entfällt.
- Decision: add-with-proof.

### Architecture Integrity Lens

- Invariant: Eine Artikelinventarnummer ist eindeutig, verpflichtend und atomar mit der
  Artikelanlage vergeben.
- Canonical owner / contract: `inventory_number_schemes` plus gemeinsame Reservierungslogik in
  `application/inventory_numbers.go`.
- Responsibility overlap: Fahrzeug- und Artikelpersistenz prüfen jeweils nur ihre eigene
  Eindeutigkeit, die Formatierung und Zählerfortschreibung bleiben gemeinsam.
- Higher-level simplification: keine zweite Artikel-Schementabelle und keine clientseitige
  Nummernerzeugung.
- Retirement / falsifier: bestehende manuelle Einzelstücknummern bleiben getrennt; die gelöschte
  Kategorie `Zubehoer` aus Migration 0017 wird nicht reaktiviert.
- Verdict: proceed.

### Plan-Time Complexity Check

- Target files: `backup.go` (1114 Zeilen), `useArticleEditorController.ts` (532),
  `accessory_repository.go` (551), `openapi/railkeeper.yaml` (6449).
- Existing size / shape signals: zentrale Dateien liegen bereits oberhalb der Projektzielgröße.
- Owner fit: Vertragsverdrahtung bleibt in den bestehenden Dateien; eigenständige Restore- und
  Stammdatenlogik wird ausgelagert.
- Add-in-place risk: hoch für Backup und Controller, moderat für Repository und OpenAPI.
- Better file boundary: `backup_article_inventory_numbers.go` und
  `useArticleCoreMasterData.ts`; Queryänderungen bleiben im vorhandenen Queryowner.
- Recommendation: extract helper für Backup und Stammdaten, edit-in-place für Typen, Query,
  OpenAPI und kleine Verdrahtung.

## Global Constraints

- Änderungen ausschließlich inline, keine Subagents.
- Keine neue UI-Bibliothek und keine neuen externen Abhängigkeiten.
- Deutsch und Englisch vollständig pflegen.
- `.superpowers/`, `frontend/dist`, Caches, QA-Daten und Zugangsdaten nicht committen.
- Sammelaktionen bleiben ausdrücklich offen; Auswahl führt keine destruktive Aktion aus.
- Bestehende inaktive Werte bleiben sichtbar, aber nicht neu auswählbar.
- Vor jedem Commit `git diff --check`; Go-Dateien vor Tests mit `gofmt` formatieren.

---

### Task 1: Persistente Artikelinventarnummer und gemeinsamer Nummernowner

**Files:**

- Create: `backend/migrations/0043_article_inventory_numbers.sql`
- Modify: `backend/internal/application/inventory_numbers.go`
- Modify: `backend/internal/application/inventory_numbers_test.go`
- Modify: `backend/internal/application/vehicle_persistence.go`
- Modify: `backend/internal/application/vehicles_test.go`
- Modify: `backend/internal/application/accessories.go`
- Modify: `backend/internal/infrastructure/accessory_repository.go`
- Modify: `backend/internal/infrastructure/accessory_repository_test.go`
- Create or modify: `backend/internal/infrastructure/article_inventory_number_migration_test.go`

**Why:** Der Artikelstamm braucht eine dauerhaft gespeicherte, automatisch und konkurrenzsicher
vergebene Nummer.

**Change Necessity:** Die Nummer muss innerhalb der bestehenden Artikeltransaktion reserviert
werden. Eine UI-generierte oder nachträglich gesetzte Nummer könnte Lücken und Duplikate erzeugen.

**Impact/Compatibility:** Die Kategorie `Artikel` wird neu angelegt. Falls sie lokal bereits
existiert, werden Präfix, Stellenzahl und Zähler nicht überschrieben. Fahrzeugnummern und
Einzelstücknummern bleiben unverändert.

**Steps:**

- [ ] Migration 0043 ergänzt `accessory_products.inventory_number`, legt das Standardschema
  `Artikel` mit `RK-ART` an, nummeriert Altbestände stabil nach `created_at, id`, setzt den Zähler
  hinter die höchste vergebene Nummer und erzwingt Eindeutigkeit sowie Nicht-Leerheit durch Index
  und Insert/Update-Trigger.
- [ ] Migrationstest prüft leere Datenbank, mehrere Altartikel mit identischem Zeitstempel,
  vorhandenes benutzerdefiniertes Artikelschema, stabile Reihenfolge, Zählerstand und Ablehnung von
  leerer bzw. doppelter Nummer.
- [ ] `inventory_numbers.go` stellt die gemeinsame transaktionale Reservierungsfunktion bereit:
  aktive Kategorie lesen, Nummer formatieren, Eindeutigkeitscallback prüfen, Zähler atomar erhöhen
  und nach begrenzten Konfliktversuchen mit fachlichem Fehler abbrechen.
- [ ] `vehicle_persistence.go` verwendet dieselbe Funktion, damit vorhandene Fahrzeugtests beweisen,
  dass Fallbackkategorien und benutzerdefinierte Schemen unverändert funktionieren.
- [ ] `AccessoryProduct` erhält das Responsefeld `InventoryNumber`; Schreibinputs bleiben ohne frei
  setzbare Nummer.
- [ ] `AccessoryRepository.CreateProduct` reserviert `Artikel` innerhalb seiner bestehenden
  Write-Transaktion und schreibt die Nummer mit dem Produkt. Select und Scan lesen das Feld mit.
- [ ] Repositorytests prüfen automatische Nummer, zwei aufeinanderfolgende Anlagen, Rollback ohne
  Zählerverbrauch, deaktiviertes/fehlendes Schema sowie parallele Anlagen ohne Duplikat.

**Verification:**

```powershell
cd backend
gofmt -w internal/application/inventory_numbers.go internal/application/inventory_numbers_test.go `
  internal/application/vehicle_persistence.go internal/application/vehicles_test.go `
  internal/application/accessories.go internal/infrastructure/accessory_repository.go `
  internal/infrastructure/accessory_repository_test.go `
  internal/infrastructure/article_inventory_number_migration_test.go
go test ./internal/application ./internal/infrastructure -count=1
```

Expected: fokussierte Pakete PASS; Fahrzeugnummernverhalten bleibt unverändert.

---

### Task 2: Listenquery, Suche, Sortierung, Bild und API-Vertrag

**Files:**

- Modify: `backend/internal/application/accessory_overview.go`
- Modify: `backend/internal/application/accessory_overview_test.go`
- Modify: `backend/internal/infrastructure/accessory_article_query.go`
- Modify: `backend/internal/infrastructure/accessory_article_query_internal_test.go`
- Modify: `backend/internal/infrastructure/accessory_repository_test.go`
- Modify: `backend/internal/api/accessory_handlers_test.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`

**Why:** Die neue Tabellenstruktur benötigt Inventarnummer und Primärbild in der Liste sowie eine
serverseitige, deterministische Sortierung pro fachlicher Spalte.

**Change Necessity:** Der Client kann weder vollständig noch stabil sortieren, weil Filter und
Aggregationen im Backend liegen. Das Primärbild muss ohne Detailrequest in der Listenantwort
verfügbar sein.

**Impact/Compatibility:** Die bisherigen Sortwerte bleiben als Aliase gültig. Neue Sortwerte sind
`image`, `inventoryNumber`, `manufacturer`, `articleNumber`, `name`, `type`, `gauge`, `stock` und
`storage`. Unbekannte Werte bleiben ein Validierungsfehler.

**Steps:**

- [ ] `AccessoryArticleListItem` erhält `InventoryNumber` und `PrimaryImageURL`; die Suche schließt
  `inventory_number` ein.
- [ ] Die Aggregationsquery ermittelt das Primärbild über `accessory_documents`, liefert eine
  Download-URL und sortiert `image` nach Bild vorhanden bzw. nicht vorhanden.
- [ ] Die Sortmap trennt Inventarnummer, Hersteller, Artikelnummer und Bezeichnung und hängt stets
  `id` als deterministischen Tiebreaker an.
- [ ] Querytests prüfen jede Sortierung in beiden Richtungen, Suche nach Inventarnummer,
  Primärbild-URL, fehlenden Bildplatzhalterwert und stabile Gleichstände.
- [ ] Handler- und OpenAPI-Vertragstests prüfen das neue Pflichtfeld in Detail- und Listenresponses,
  ohne es in `AccessoryProductInput` aufzunehmen.
- [ ] `openapi/railkeeper.yaml` dokumentiert `inventoryNumber`, `primaryImageUrl` und die erweiterten
  Sortparameter konsistent.

**Verification:**

```powershell
cd backend
gofmt -w internal/application/accessory_overview.go internal/application/accessory_overview_test.go `
  internal/infrastructure/accessory_article_query.go `
  internal/infrastructure/accessory_article_query_internal_test.go `
  internal/api/accessory_handlers_test.go internal/api/openapi_contract_test.go
go test ./internal/application ./internal/infrastructure ./internal/api -count=1
```

Expected: Artikelquery, Handler und OpenAPI-Vertrag PASS.

---

### Task 3: Backup- und Restorekompatibilität

**Files:**

- Create: `backend/internal/application/backup_article_inventory_numbers.go`
- Modify: `backend/internal/application/backup.go`
- Modify: `backend/internal/application/backup_test.go`

**Why:** Neue Backups müssen Nummern bewahren; alte Backups dürfen wegen der neuen
Pflichtinvariante nicht scheitern.

**Change Necessity:** Der generische Tabellenrestore ignoriert fehlende Spalten. Ohne gezielte
Vorbereitung würden alte Artikel leere Nummern erhalten und am neuen Datenbankconstraint
scheitern.

**Impact/Compatibility:** Backupversion 3 bleibt bestehen. Vorhandene Nummern werden unverändert
restauriert. Fehlende Nummern werden anhand stabil sortierter Backupzeilen und des lokalen Schemas
mit den niedrigsten freien Werten ergänzt; der lokale Zähler wird mindestens hinter die höchste
verwendete Nummer gesetzt.

**Steps:**

- [ ] Ein fokussierter Restore-Helfer validiert vorhandene Artikelinventarnummern, sammelt belegte
  Werte und ergänzt fehlende Nummern deterministisch vor dem generischen Insert.
- [ ] Gemischte Backups mit vorhandenen und fehlenden Nummern überspringen belegte Werte und
  erzeugen keine Duplikate.
- [ ] Fehlendes oder deaktiviertes Artikelschema macht die Restore-Validierung inkompatibel bzw.
  bricht den Import vor destruktivem Commit ab.
- [ ] `backup.go` ruft den Helfer an der bestehenden Restoregrenze auf; die umfangreiche
  Backupdatei erhält keine neue Detailimplementierung.
- [ ] Tests decken Export/Import mit erhaltener Nummer, Version-2- und Version-3-Dokumente ohne
  Nummer, gemischte Zeilen, wiederholte deterministische Zuweisung und Zählerfortschreibung ab.

**Verification:**

```powershell
cd backend
gofmt -w internal/application/backup_article_inventory_numbers.go `
  internal/application/backup.go internal/application/backup_test.go
go test ./internal/application -run 'Backup|Restore' -count=1
```

Expected: alle Backup- und Restoretests PASS.

---

### Task 4: Frontendvertrag und stammdatengebundener Artikeldialog

**Files:**

- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Create: `frontend/src/features/accessories/useArticleCoreMasterData.ts`
- Create: `frontend/src/features/accessories/useArticleCoreMasterData.test.tsx`
- Modify: `frontend/src/features/accessories/useArticleEditorController.ts`
- Modify: `frontend/src/features/accessories/useArticleEditorController.test.tsx`
- Modify: `frontend/src/features/accessories/ArticleEditorDialog.tsx`
- Modify: `frontend/src/features/accessories/ArticleEditorDialog.test.tsx`
- Modify: `frontend/src/features/accessories/ArticleCoreTab.tsx`
- Modify: `frontend/src/features/accessories/articleEditorModel.ts`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Why:** Hersteller, Spurweiten und Bestandseinheit müssen aus der Datenbank kommen und die
Artikelinventarnummer muss im Dialog klar, aber nicht editierbar erscheinen.

**Change Necessity:** Die derzeitige Hersteller- und Bestandseinheiteneingabe ist Freitext;
Spurweiten stammen aus einer fest codierten Liste. Diese Quellen können nicht auf lokale
Stammdatenänderungen reagieren.

**Impact/Compatibility:** Gespeichert werden weiterhin Herstellerlabel, Spurweitenwerte und
kanonische Bestandseinheitenschlüssel. Inaktive aktuelle Werte werden zusätzlich zu aktiven
Optionen angezeigt. Maßstab und Felder ohne Datenquelle bleiben Freitext.

**Steps:**

- [ ] Frontendtypen ergänzen `inventoryNumber` und `primaryImageUrl`; Sorttypen entsprechen dem
  erweiterten Backendvertrag.
- [ ] `useArticleCoreMasterData` lädt `manufacturer`, `gauge` und `stock_unit` unabhängig, hält
  Loading- und Fehlerzustände je Quelle, ignoriert veraltete Requests und bietet gezielte Retry-
  Funktionen.
- [ ] Der Controller startet den Hook pro Dialogsession und reicht ein gebündeltes
  `coreMasterData`-Objekt weiter, ohne die bestehenden Ressourcen- und Dirty-State-Flows zu ändern.
- [ ] `ArticleCoreTab` zeigt die Inventarnummer schreibgeschützt; im Create-Modus erscheint der
  lokalisierte Hinweis auf automatische Vergabe.
- [ ] Hersteller wird `AppSelect` mit aktiven Labels als persistierten Werten, Spurweite bleibt
  `AppMultiSelect` aus aktiven Gauge-Einträgen, Bestandseinheit wird `AppSelect` mit kanonischem Key
  und lokalisierter Anzeige über `masterDataDisplayLabel`.
- [ ] Aktuelle inaktive oder historische Werte werden als deaktivierte Option erhalten. Ein
  Ladefehler deaktiviert nur das betroffene Feld, zeigt eine lokalisierte Meldung und lässt Retry
  zu.
- [ ] Tests prüfen Datenquellen, persistierte Werte, Deutsch/Englisch, inaktive Altwerte,
  unabhängige Fehler, Retry, Request-Rennen, Dirty-State und die nicht editierbare Inventarnummer.

**Verification:**

```powershell
cd frontend
npm.cmd run test:run -- src/features/accessories/useArticleCoreMasterData.test.tsx `
  src/features/accessories/useArticleEditorController.test.tsx `
  src/features/accessories/ArticleEditorDialog.test.tsx
npm.cmd run build
```

Expected: fokussierte Tests und TypeScript-Produktionsbuild PASS.

---

### Task 5: Kennzahlenhierarchie, sortierbare Tabelle und Auswahl

**Files:**

- Modify: `frontend/src/features/accessories/ArticleMetrics.tsx`
- Modify: `frontend/src/features/accessories/ArticleTable.tsx`
- Modify: `frontend/src/features/accessories/ArticleTable.test.tsx`
- Modify: `frontend/src/features/accessories/AccessoriesView.tsx`
- Modify: `frontend/src/features/accessories/AccessoriesView.test.tsx`
- Modify: `frontend/src/features/accessories/useArticleOverview.ts`
- Modify: `frontend/src/features/accessories/useArticleOverview.test.tsx`
- Modify: `frontend/src/features/accessories/accessoriesResponsive.test.ts`
- Modify: `frontend/src/styles/accessories.css`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Why:** Die Übersicht soll wie der Fahrzeugbestand scannbar sein und alle fachlichen Spalten
serverseitig sortieren.

**Change Necessity:** Die aktuelle Tabelle bündelt mehrere Felder, besitzt weder Bild noch
Inventarnummer oder Auswahl und kann die geforderten Einzelspalten nicht sortieren.

**Impact/Compatibility:** Ansicht, Bearbeiten, Archivieren und Wiederherstellen bleiben erhalten.
Auswahl ist rein lokal und wird bei neuen Ergebnissen auf sichtbare IDs begrenzt. Es entstehen
keine Sammelaktionen.

**Steps:**

- [ ] `ArticleMetrics` trennt je Karte `summary` und `detail`: Artikel, frei, gebunden und Hinweise
  stehen groß; Arten, Lagerorte, reserviert/eingebaut und unvollständig stehen als dezentes `em`
  darunter. Bestehende Filterbuttons und Aktivzustände bleiben erhalten.
- [ ] `AccessoriesView` hält die sichtbare Mehrfachauswahl, schneidet sie nach jedem neuen Resultat
  auf geladene IDs zu und übergibt Auswahl sowie Toggle-Callbacks an die Tabelle.
- [ ] `ArticleTable` rendert exakt die freigegebene Spaltenreihenfolge: Auswahl, Bild,
  Inventarnummer, Hersteller, Artikelnummer, Bezeichnung, Artikelart/Unterart, Spurweite, Bestand,
  Lagerort, Aktionen.
- [ ] Kopfcheckbox unterstützt checked, unchecked und indeterminate, wählt alle sichtbaren Zeilen
  und besitzt deutsche sowie englische barrierefreie Beschriftungen. Zeilencheckboxen enthalten
  die Artikelbezeichnung im Accessible Name.
- [ ] Primärbild nutzt `inventory-thumb`, fehlendes Bild den Bestandsplatzhalter. Inventarnummer,
  Hersteller, Artikelnummer, Bezeichnung, Typ, Spurweite, Bestand, Lagerort und Bild sind über den
  Tabellenkopf sortierbar; Auswahl und Aktionen nicht.
- [ ] `useArticleOverview` verwendet `inventoryNumber` als Standardsortierung und behält das
  bestehende Umschalten zwischen auf- und absteigend bei.
- [ ] CSS ersetzt `nth-child`-Annahmen durch benannte Spaltenklassen, hält Tabellenzellen kompakt,
  stellt Summen-/Detailhierarchie her und setzt eine kontrollierte Mindestbreite für mobile
  horizontale Führung.
- [ ] Tests prüfen DOM-Reihenfolge, alle Sortcallbacks, Bildzustände, Einzel-/Gesamtauswahl,
  indeterminate, Auswahlbereinigung, Kennzahlentexte, Responsive-Minimum und fehlende
  Sammelaktionsbuttons.

**Verification:**

```powershell
cd frontend
npm.cmd run test:run -- src/features/accessories/ArticleTable.test.tsx `
  src/features/accessories/AccessoriesView.test.tsx `
  src/features/accessories/useArticleOverview.test.tsx `
  src/features/accessories/accessoriesResponsive.test.ts
npm.cmd run test:run
npm.cmd run build
```

Expected: fokussierte und vollständige Frontendtests sowie Build PASS.

---

### Task 6: Gesamtverifikation, Browser-QA und Dokumentabschluss

**Files:**

- Modify: `docs/superpowers/specs/2026-08-09-artikelinventarnummer-tabellen-stammdaten-design.md`
- Modify: `docs/superpowers/plans/2026-08-09-artikelinventarnummer-tabellen-stammdaten.md`

**Why:** Migration, Backup und UI bilden einen gemeinsamen Datenvertrag und müssen zusammen
abgenommen werden.

**Steps:**

- [ ] Alle geänderten Go-Dateien mit `gofmt` formatieren und `git diff --check` ausführen.
- [ ] Vollständigen Backendlauf ausführen:

  ```powershell
  cd backend
  go test ./... -count=1
  ```

- [ ] Vollständigen Frontendlauf ausführen:

  ```powershell
  cd frontend
  npm.cmd run test:run
  npm.cmd run build
  ```

- [ ] Lokalen Server mit demselben QA-Datensatz neu starten und `/health` sowie
  `/api/v1/accessory-products` prüfen.
- [ ] Browser-QA auf `/accessories`: Dark und Light, Desktop und 390 × 844 Pixel, Kennzahlen,
  sämtliche Sortierköpfe, Auswahl, Bild/Platzhalter, horizontale Tabellenführung, Dialog-Dropdowns,
  Inventarnummer nach Neuanlage, inaktive Altwerte und lange deutsche Texte.
- [ ] Prüfen, dass keine Konsolenfehler, kein Dokumentüberlauf und keine ungewollte Sammelaktion
  vorhanden sind.
- [ ] Spezifikationsstatus auf `Implemented and verified` setzen, Plancheckboxen abschließen und
  den offenen Folgepunkt „Sammelaktionen fehlen“ unverändert erhalten.
- [ ] Nur Quellcode, Migration, Tests, OpenAPI, Spezifikation und Plan committen. `.superpowers/`,
  `frontend/dist`, `.cache` und lokale Daten bleiben untracked bzw. ignoriert.

## Risks and Rollback

- **Migrationsrisiko:** Nummernbackfill darf keine Altartikel verlieren. Vor jedem destructive
  rebuild wird die trigger-/indexbasierte additive Migration bevorzugt und per Migrationstest
  bewiesen.
- **Restore-Risiko:** Alte Backups fehlen das Pflichtfeld. Der Restore-Helfer läuft vor dem ersten
  Insert und innerhalb derselben Transaktion; ein Fehler lässt Datenbank und Uploadswap unverändert.
- **Parallelitätsrisiko:** SQLite-Schreibreservierung und Zählerupdate bleiben in einer
  Transaktion; Paralleltests sind Pflicht.
- **UI-Risiko:** Elf Spalten erhöhen die Mindestbreite. Mobile nutzt kontrollierten Tabellen-Scroll,
  nicht Dokumentüberlauf oder versteckte Pflichtspalten.
- **Stammdatenrisiko:** Historische Werte dürfen nicht verschwinden. Aktueller Wert wird unabhängig
  vom Aktivstatus erhalten.
- **Rollback:** Code- und UI-Commits sind revertierbar. Migrationen bleiben vorwärtsgerichtet; ein
  Rollback erfolgt über App-Backup bzw. Datenbankkopie, nicht durch Umschreiben von Migration 0043.

## Execution Readiness View

- Intent Lock: freigegebene vier Nutzeranforderungen plus bestätigte Trennung von Artikelstamm und
  Einzelstück.
- Scope Fence: Inventarnummer, Übersicht, Kennzahlen, Stammdatenfelder und zugehörige Verträge;
  keine Sammelaktionen.
- Baseline Lock: freigegebene Spezifikation und bestehende Fahrzeug-/Artikelowner.
- Approved Behavior: automatische `RK-ART`-Vergabe, deterministischer Altbestand, freigegebene
  Spaltenfolge, sortierbare Datenköpfe, DB-Dropdowns und Fahrzeugbestands-Kennzahlenhierarchie.
- Owner / Contract Constraints: Schemenverwaltung bleibt Nummernowner; OpenAPI, Backend und
  Frontend bleiben synchron.
- Compatibility Boundary: alte Backups, Fahrzeuge, Einzelstücke, Rollen und lokale SQLite-Daten.
- Retirement Boundary: keine Wiederbelebung von `Zubehoer`; fest codierte Spurweitenliste und freie
  Hersteller-/Bestandseinheiteneingaben entfallen.
- Task Batches: Persistenz, Query/API, Backup, Editor, Übersicht, Gesamt-QA.
- Test Obligations: fokussierte Regression pro Batch, dann vollständige Go- und Frontendläufe.
- Review Gates: Diffprüfung vor Commit, Backup-/Migrationsevidence vor UI-Abschluss, Browser-QA vor
  Completion Claim.
- Drift / Rewind Rules: Neue Stammdatenart, neue Sammelaktion, Schema-Rebuild oder Änderung am
  Backupformat erfordert Rückkehr zum Design.
- Evidence Required Before Completion: grüne Volltests, erfolgreicher Build, Healthcheck,
  Browser-QA und sauberer Git-Diff.
- Advisory Boundary: Method-Pack-Ausführungshilfe, keine autoritative Completion-Erlaubnis.

## Execution Route

- Decision: inline.
- Evidence: Daniel hat Inline-Arbeit ausdrücklich verlangt; Subagents sind untersagt.
- Fallback: kein paralleler Agentenpfad; Aufgaben werden sequenziell mit Checkpoints umgesetzt.
- User confirmation required: no.
