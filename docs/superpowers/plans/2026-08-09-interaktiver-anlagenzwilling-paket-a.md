# Interaktiver Anlagenzwilling, Paket A Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Technische Positionen einer Anlageneinheit werden versionssicher gespeichert, über API und
typisierten Client verwaltet und im Register `Technik` mit app-eigenen Dialogen gepflegt.

**Architecture:** Eine neue normalisierte Positionstabelle ergänzt das Stage-1-Anlagenmodell. Ein
fokussiertes Repository und ein eigener Application-Baustein halten die bestehenden zentralen
Anlagendateien klein. Paket A liefert eine vertikale Funktion von Migration und Backup bis zur
deutsch/englischen Oberfläche; Statuslayer und SVG-Gesamtansicht folgen in Paket B.

**Tech Stack:** Go, SQLite, React 19, TypeScript 7 strict, Vite 8, Vitest 4, Testing Library,
OpenAPI 3, bestehende RailKeeper-UI-Komponenten.

## Global Constraints

- Alle Änderungen bleiben lokal auf `dev/issue-35-layout-twin`; kein Push und kein PR.
- Technische Positionen dokumentieren nur, sie senden keine Steuerbefehle.
- Viewer, Editor, Planner und Admin dürfen lesen; nur Planner und Admin dürfen Positionen schreiben.
- Messe erhält keinen allgemeinen Anlagenzugriff.
- Alle Schreibzugriffe bleiben CSRF-geschützt, serverseitig autorisiert und auditiert.
- Koordinaten verwenden Millimeter; Drehungen werden auf `0 <= value < 360` normalisiert.
- Positionen werden archiviert statt physisch gelöscht.
- Backup-Version 4 enthält jede neue Stage-2-Tabelle; Versionen 1 bis 3 bleiben importierbar.
- Deutsche und englische Texte bleiben synchron.
- App-eigene Controls und Dialoge sind verpflichtend; keine sichtbar nativen Browser-Controls.
- Backend, Frontend-Client und `openapi/railkeeper.yaml` müssen denselben Vertrag verwenden.

---

## File Map

- Create `backend/migrations/0044_layout_twin_positions.sql`: Kontur-, Positions- und
  Zuordnungstabellen.
- Create `backend/internal/domain/layout_twin.go`: gültige technische Positionstypen.
- Create `backend/internal/domain/layout_twin_test.go`: Typvalidierung.
- Create `backend/internal/infrastructure/layout_position_repository.go`: Positionspersistenz und Audit.
- Create `backend/internal/infrastructure/layout_position_repository_test.go`: CRUD, Konflikte und Referenzen.
- Create `backend/internal/application/layout_positions.go`: DTOs, Validierung und Service-Methoden.
- Create `backend/internal/application/layout_positions_test.go`: Bereinigung und Grenzfälle.
- Modify `backend/internal/application/layouts.go`: Repository-Schnittstelle um drei Methoden ergänzen.
- Create `backend/internal/api/layout_position_handlers.go`: Listen, Anlegen und Aktualisieren.
- Create `backend/internal/api/layout_position_handlers_test.go`: Rollen, CSRF, Vertrag und Fehler.
- Modify `backend/internal/api/routes.go`: drei Positionsrouten registrieren.
- Modify `backend/internal/api/layout_handlers.go`: Positionsfehler auf Problemantworten abbilden.
- Modify `backend/internal/application/backup.go`: Backup-Version 4 und Tabellenreihenfolge.
- Modify `backend/internal/application/backup_test.go`: Version-4-Roundtrip und Legacy-Kompatibilität.
- Modify `backend/internal/infrastructure/layout_accessory_schema_test.go`: Migrationstabellen prüfen.
- Modify `openapi/railkeeper.yaml`: Positionsschemas und Pfade ergänzen.
- Modify `backend/internal/api/openapi_contract_test.go`: Positionsendpunkte im Vertrag prüfen.
- Modify `frontend/src/shared/apiLayoutsAccessories.ts`: Typen und Clientmethoden ergänzen.
- Modify `frontend/src/shared/apiLayoutsAccessories.test.ts`: Pfade und Nutzdaten prüfen.
- Create `frontend/src/features/layouts/LayoutPositionDialog.tsx`: app-eigener Anlege-/Bearbeitendialog.
- Create `frontend/src/features/layouts/LayoutPositionDialog.test.tsx`: Controls, Submit, Fokus und Verwerfen.
- Create `frontend/src/features/layouts/LayoutTechnicalPositionsPanel.tsx`: Einheitsauswahl und Positionstabelle.
- Create `frontend/src/features/layouts/LayoutTechnicalPositionsPanel.test.tsx`: Laden, Rollen und Konflikte.
- Modify `frontend/src/features/layouts/LayoutWorkspace.tsx`: Register `Technik` aktivieren.
- Modify `frontend/src/features/layouts/LayoutsView.test.tsx`: Technik-Integration absichern.
- Modify `frontend/src/shared/i18n/de.ts`: deutsche Positionstexte.
- Modify `frontend/src/shared/i18n/en.ts`: englische Positionstexte.
- Modify `frontend/src/styles/layouts.css`: dichte Positionsliste und Dialograster.

---

### Task 1: Schema und Fachtypen

**Files:**
- Create: `backend/migrations/0044_layout_twin_positions.sql`
- Create: `backend/internal/domain/layout_twin.go`
- Create: `backend/internal/domain/layout_twin_test.go`
- Modify: `backend/internal/infrastructure/layout_accessory_schema_test.go`

**Interfaces:**
- Produces: `domain.LayoutTechnicalPositionKind` mit `Valid() bool`.
- Produces: die Tabellen `layout_unit_outline_points`, `layout_technical_positions`,
  `accessory_reservation_positions` und `accessory_installation_positions`.
- Consumes: `layout_units`, `accessory_products`, `accessory_reservations` und
  `accessory_installations` aus den Migrationen 0039 bis 0041.

- [ ] **Step 1: Fehlende Schema- und Domaintests schreiben**

Erweitere den Schematest um die vier Tabellen und prüfe Fremdschlüssel sowie die
Drehungsbeschränkung. Ergänze im Domaintest:

```go
func TestLayoutTechnicalPositionKinds(t *testing.T) {
	valid := []LayoutTechnicalPositionKind{
		LayoutPositionTurnout, LayoutPositionSignal, LayoutPositionFeedback,
		LayoutPositionDecoder, LayoutPositionLighting, LayoutPositionPower,
		LayoutPositionSensor, LayoutPositionOther,
	}
	for _, kind := range valid {
		if !kind.Valid() { t.Fatalf("expected %q to be valid", kind) }
	}
	if LayoutTechnicalPositionKind("command").Valid() {
		t.Fatal("control commands must not be valid technical positions")
	}
}
```

- [ ] **Step 2: Tests ausführen und erwartetes Fehlschlagen bestätigen**

Run: `cd backend; go test ./internal/domain ./internal/infrastructure -run "LayoutTechnicalPosition|LayoutAccessorySchema"`

Expected: FAIL, weil Migration 0044 und der Fachtyp fehlen.

- [ ] **Step 3: Migration und Fachtyp implementieren**

Die Positionstabelle verwendet diese Kernstruktur:

```sql
CREATE TABLE layout_technical_positions (
  id TEXT PRIMARY KEY,
  layout_unit_id TEXT NOT NULL,
  label TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN (
    'turnout', 'signal', 'feedback', 'decoder', 'lighting', 'power', 'sensor', 'other'
  )),
  position_x_mm REAL NOT NULL,
  position_y_mm REAL NOT NULL,
  rotation_degrees REAL NOT NULL DEFAULT 0
    CHECK (rotation_degrees >= 0 AND rotation_degrees < 360),
  product_id TEXT,
  description TEXT NOT NULL DEFAULT '',
  version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
  archived INTEGER NOT NULL DEFAULT 0 CHECK (archived IN (0, 1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (layout_unit_id) REFERENCES layout_units(id) ON DELETE RESTRICT,
  FOREIGN KEY (product_id) REFERENCES accessory_products(id) ON DELETE RESTRICT
);
```

Konturpunkte verwenden `(layout_unit_id, point_index)` als Primärschlüssel. Die beiden
Zuordnungstabellen verwenden `reservation_id` beziehungsweise `installation_id` als Primärschlüssel
und `position_id` als nicht eindeutigen Fremdschlüssel, damit eine Position mehrere Artikel bündeln
kann.

- [ ] **Step 4: Tests ausführen und gofmt anwenden**

Run: `gofmt -w backend/internal/domain/layout_twin.go backend/internal/domain/layout_twin_test.go`

Run: `cd backend; go test ./internal/domain ./internal/infrastructure -run "LayoutTechnicalPosition|LayoutAccessorySchema"`

Expected: PASS.

- [ ] **Step 5: Commit erstellen**

```powershell
git add backend/migrations/0044_layout_twin_positions.sql backend/internal/domain/layout_twin.go `
  backend/internal/domain/layout_twin_test.go backend/internal/infrastructure/layout_accessory_schema_test.go
git commit -m "feat(layouts): add technical position schema"
```

---

### Task 2: Positionsrepository und Application-Service

**Files:**
- Create: `backend/internal/infrastructure/layout_position_repository.go`
- Create: `backend/internal/infrastructure/layout_position_repository_test.go`
- Create: `backend/internal/application/layout_positions.go`
- Create: `backend/internal/application/layout_positions_test.go`
- Modify: `backend/internal/application/layouts.go`

**Interfaces:**
- Produces: `LayoutTechnicalPosition`, `CreateLayoutTechnicalPositionInput`,
  `UpdateLayoutTechnicalPositionInput`.
- Produces: `ListTechnicalPositions`, `CreateTechnicalPosition` und `UpdateTechnicalPosition` auf
  `LayoutService` und `LayoutRepository`.
- Produces: `ErrLayoutPositionNotFound`, `ErrLayoutPositionVersionConflict` und
  `ErrLayoutPositionProductNotFound`.

- [ ] **Step 1: Fehlende Service- und Repositorytests schreiben**

Die Tests müssen mindestens abdecken:

```go
func TestCreateTechnicalPositionNormalizesRotationAndTrimsText(t *testing.T) {
	position, err := service.CreateTechnicalPosition(ctx, "unit-1", CreateLayoutTechnicalPositionInput{
		Label: "  Signal A  ", Kind: domain.LayoutPositionSignal,
		PositionXMM: 120.5, PositionYMM: -4, RotationDegrees: 450,
		Description: "  Einfahrt  ",
	}, "planner")
	if err != nil { t.Fatal(err) }
	if position.RotationDegrees != 90 || position.Label != "Signal A" || position.Description != "Einfahrt" {
		t.Fatalf("unexpected normalized position: %#v", position)
	}
}
```

Repositorytests prüfen Sortierung, optionale Produktreferenz, Auditereignisse, fehlende Einheit,
fehlendes Produkt, Archivierung und einen 409-fähigen Versionskonflikt.

- [ ] **Step 2: Tests ausführen und erwartetes Fehlschlagen bestätigen**

Run: `cd backend; go test ./internal/application ./internal/infrastructure -run "TechnicalPosition"`

Expected: FAIL, weil Typen und Methoden fehlen.

- [ ] **Step 3: Application-Typen und Validierung implementieren**

`layout_positions.go` enthält fokussiert:

```go
type LayoutTechnicalPosition struct {
	ID string `json:"id"`
	LayoutUnitID string `json:"layoutUnitId"`
	Label string `json:"label"`
	Kind domain.LayoutTechnicalPositionKind `json:"kind"`
	PositionXMM float64 `json:"positionXMm"`
	PositionYMM float64 `json:"positionYMm"`
	RotationDegrees float64 `json:"rotationDegrees"`
	ProductID string `json:"productId,omitempty"`
	Description string `json:"description,omitempty"`
	Version int `json:"version"`
	Archived bool `json:"archived"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
```

Leere IDs, leere Bezeichnungen, ungültige Typen, nicht endliche Koordinaten und Versionen unter 1
werden mit `ErrLayoutValidation` abgelehnt. Drehungen werden über die vorhandene
`normalizeRotation`-Funktion normalisiert.

- [ ] **Step 4: Repository in eigener Datei implementieren**

Verwende Transaktionen für Create und Update. Prüfe Einheit und optionales Produkt vor dem Insert.
Update verwendet `WHERE id=? AND version=?`, erhöht `version` und schreibt
`LayoutTechnicalPositionCreated` beziehungsweise `LayoutTechnicalPositionUpdated` ins Auditlog.

- [ ] **Step 5: gofmt und zielgerichtete Tests ausführen**

Run: `gofmt -w backend/internal/application/layout_positions.go backend/internal/application/layout_positions_test.go backend/internal/infrastructure/layout_position_repository.go backend/internal/infrastructure/layout_position_repository_test.go backend/internal/application/layouts.go`

Run: `cd backend; go test ./internal/application ./internal/infrastructure -run "TechnicalPosition"`

Expected: PASS.

- [ ] **Step 6: Commit erstellen**

```powershell
git add backend/internal/application/layout_positions.go backend/internal/application/layout_positions_test.go `
  backend/internal/application/layouts.go backend/internal/infrastructure/layout_position_repository.go `
  backend/internal/infrastructure/layout_position_repository_test.go
git commit -m "feat(layouts): persist technical positions"
```

---

### Task 3: HTTP-API und Rollenmatrix

**Files:**
- Create: `backend/internal/api/layout_position_handlers.go`
- Create: `backend/internal/api/layout_position_handlers_test.go`
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/internal/api/layout_handlers.go`

**Interfaces:**
- Produces: `GET` und `POST /api/v1/layout-units/{id}/technical-positions`.
- Produces: `PUT /api/v1/layout-technical-positions/{id}`.
- Consumes: Service-Methoden aus Task 2.

- [ ] **Step 1: API-Tests für Rollen, CSRF und Fehler schreiben**

Prüfe GET für Admin, Planner, Editor und Viewer mit Status 200 sowie Messe mit 403. Prüfe POST und
PUT für Planner/Admin, Editor/Viewer mit 403 und fehlendes CSRF mit 403. Prüfe ungültige Eingaben mit
400, fehlende Position mit 404 und alte Version mit 409.

- [ ] **Step 2: Tests ausführen und erwartetes Fehlschlagen bestätigen**

Run: `cd backend; go test ./internal/api -run "TechnicalPositions"`

Expected: FAIL mit 404 auf den noch nicht registrierten Routen.

- [ ] **Step 3: Handler und Routen implementieren**

Registriere:

```go
{http.MethodGet, "/api/v1/layout-units/{id}/technical-positions", routeAccessViewer, (*App).listLayoutTechnicalPositions, nil},
{http.MethodPost, "/api/v1/layout-units/{id}/technical-positions", routeAccessPlanner, (*App).createLayoutTechnicalPosition, nil},
{http.MethodPut, "/api/v1/layout-technical-positions/{id}", routeAccessPlanner, (*App).updateLayoutTechnicalPosition, nil},
```

Mappe Positions-Not-Found auf 404 und Positions-Versionskonflikt auf 409. Nutze für ungültige Daten
weiterhin `layout_validation`.

- [ ] **Step 4: gofmt und API-Tests ausführen**

Run: `gofmt -w backend/internal/api/layout_position_handlers.go backend/internal/api/layout_position_handlers_test.go backend/internal/api/routes.go backend/internal/api/layout_handlers.go`

Run: `cd backend; go test ./internal/api -run "TechnicalPositions"`

Expected: PASS.

- [ ] **Step 5: Commit erstellen**

```powershell
git add backend/internal/api/layout_position_handlers.go backend/internal/api/layout_position_handlers_test.go `
  backend/internal/api/routes.go backend/internal/api/layout_handlers.go
git commit -m "feat(api): expose layout technical positions"
```

---

### Task 4: Backup-Version 4

**Files:**
- Modify: `backend/internal/application/backup.go`
- Modify: `backend/internal/application/backup_test.go`

**Interfaces:**
- Produces: `backupVersion = 4`.
- Produces: vollständigen Export und Restore aller vier Tabellen aus Migration 0044.
- Preserves: Importkompatibilität für Backup-Versionen 1, 2 und 3.

- [ ] **Step 1: Fehlende Backup-Tests schreiben**

Ergänze einen Roundtrip mit Konturpunkt, technischer Position und beiden Zuordnungstabellen. Erweitere
die versionsabhängigen Testtabellen so, dass Version 4 alle vier Tabellen verlangt und Version 3 ohne
diese Tabellen importierbar bleibt.

- [ ] **Step 2: Backup-Tests ausführen und erwartetes Fehlschlagen bestätigen**

Run: `cd backend; go test ./internal/application -run "BackupVersionFour|VersionThreeWithoutVersionFour"`

Expected: FAIL, weil der Export noch Version 3 verwendet und die Tabellen fehlen.

- [ ] **Step 3: Tabellenreihenfolge und Versionsrichtlinie implementieren**

Ordne `layout_unit_outline_points` und `layout_technical_positions` direkt nach `layout_units` ein.
Ordne `accessory_reservation_positions` nach Reservierungen und `accessory_installation_positions`
nach Installationen ein. Alle vier Tabellen erhalten `{introduced: 4, required: 4}`.

- [ ] **Step 4: Backup-Tests vollständig ausführen**

Run: `cd backend; go test ./internal/application -run "Backup"`

Expected: PASS.

- [ ] **Step 5: Commit erstellen**

```powershell
git add backend/internal/application/backup.go backend/internal/application/backup_test.go
git commit -m "feat(backup): include layout twin positions"
```

---

### Task 5: OpenAPI und typisierter Frontend-Client

**Files:**
- Modify: `openapi/railkeeper.yaml`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/shared/apiLayoutsAccessories.test.ts`

**Interfaces:**
- Produces: `LayoutTechnicalPositionKind`, `LayoutTechnicalPosition`,
  `LayoutTechnicalPositionInput`, `LayoutTechnicalPositionUpdateInput`.
- Produces: `layoutTechnicalPositions(unitId)`, `createLayoutTechnicalPosition(unitId, input)` und
  `updateLayoutTechnicalPosition(id, input)`.

- [ ] **Step 1: Vertrags- und Clienttests schreiben**

Der Clienttest erwartet exakt:

```ts
await api.layoutTechnicalPositions("unit/1");
await api.createLayoutTechnicalPosition("unit/1", position);
await api.updateLayoutTechnicalPosition("position/1", { ...position, expectedVersion: 2 });
```

und die kodierten Pfade `/layout-units/unit%2F1/technical-positions` sowie
`/layout-technical-positions/position%2F1`.

- [ ] **Step 2: Tests ausführen und erwartetes Fehlschlagen bestätigen**

Run: `cd frontend; npm.cmd run test:run -- src/shared/apiLayoutsAccessories.test.ts`

Run: `cd backend; go test ./internal/api -run "OpenAPI"`

Expected: FAIL, weil Methoden und Vertrag fehlen.

- [ ] **Step 3: Typen, Methoden und OpenAPI-Schemas ergänzen**

Der Write-Input enthält `label`, `kind`, `positionXMm`, `positionYMm`, `rotationDegrees`, optional
`productId`, optional `description` und optional `archived`. Der Update-Input ergänzt
`expectedVersion` als Pflichtfeld.

- [ ] **Step 4: Client- und Vertragstests ausführen**

Run: `cd frontend; npm.cmd run test:run -- src/shared/apiLayoutsAccessories.test.ts`

Run: `cd backend; go test ./internal/api -run "OpenAPI"`

Expected: PASS.

- [ ] **Step 5: Commit erstellen**

```powershell
git add openapi/railkeeper.yaml backend/internal/api/openapi_contract_test.go `
  frontend/src/shared/apiLayoutsAccessories.ts frontend/src/shared/apiLayoutsAccessories.test.ts
git commit -m "feat(api): define technical position contract"
```

---

### Task 6: App-eigener Positionsdialog und Technikregister

**Files:**
- Create: `frontend/src/features/layouts/LayoutPositionDialog.tsx`
- Create: `frontend/src/features/layouts/LayoutPositionDialog.test.tsx`
- Create: `frontend/src/features/layouts/LayoutTechnicalPositionsPanel.tsx`
- Create: `frontend/src/features/layouts/LayoutTechnicalPositionsPanel.test.tsx`
- Modify: `frontend/src/features/layouts/LayoutWorkspace.tsx`
- Modify: `frontend/src/features/layouts/LayoutsView.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/layouts.css`

**Interfaces:**
- Produces: `LayoutPositionDialog` für `create` und `edit`.
- Produces: `LayoutTechnicalPositionsPanel({ units, canPlan })`.
- Consumes: Clientmethoden und Typen aus Task 5 sowie vorhandene `AppTextInput`, `AppTextArea`,
  `AppCheckbox`, `AppSelect` und `LayoutConfirmDialog`.

- [ ] **Step 1: Dialog- und Paneltests schreiben**

Tests decken ab:

- initialer Fokus auf Bezeichnung,
- app-eigene Auswahl für Typ, Einheit und Produkt,
- normaler Submit mit Zahlenwerten,
- Warnung vor dem Verwerfen eines geänderten Entwurfs,
- Laden und Umschalten der Anlageneinheit,
- Anlegen, Bearbeiten und Archivieren,
- Entwurf bleibt bei API-Fehler erhalten,
- Serverstand-Neuladen bei 409,
- keine Schreibaktionen für Viewer/Editor,
- deutsche und englische Beschriftungen.

- [ ] **Step 2: Tests ausführen und erwartetes Fehlschlagen bestätigen**

Run: `cd frontend; npm.cmd run test:run -- src/features/layouts/LayoutPositionDialog.test.tsx src/features/layouts/LayoutTechnicalPositionsPanel.test.tsx src/features/layouts/LayoutsView.test.tsx`

Expected: FAIL, weil Dialog und Panel fehlen und das Technikregister noch den Stage-Hinweis zeigt.

- [ ] **Step 3: Positionsdialog implementieren**

Der Dialog folgt `LayoutFormDialog`: Portal, Fokusfalle, Escape, Fokusrückgabe, app-eigener
Verwerf-Dialog und gesperrte Aktionen beim Speichern. Koordinaten und Rotation verwenden
`AppTextInput type="number"`; Typ und Produkt verwenden `AppSelect`.

- [ ] **Step 4: Technikpanel implementieren**

Das Panel lädt Positionen für die ausgewählte aktive Einheit. Die Tabelle beginnt mit Bezeichnung und
zeigt Typ, Koordinate, Produkt, Status und Aktion. Der Panelkopf enthält Einheitsauswahl und
`Position anlegen`. Bei leerer Einheit oder leerer Positionsliste erscheinen getrennte Leerzustände.

- [ ] **Step 5: Workspace, Übersetzungen und Styles integrieren**

Ersetze für `tab === "technology"` den Deferred-Block durch:

```tsx
<LayoutTechnicalPositionsPanel units={units} canPlan={canPlan} />
```

Halte Tabellen, Dialog, lange deutsche Texte, schmale Ansicht und beide Themes mit vorhandenen Tokens
kompakt und ruhig.

- [ ] **Step 6: Zielgerichtete Frontendtests ausführen**

Run: `cd frontend; npm.cmd run test:run -- src/features/layouts/LayoutPositionDialog.test.tsx src/features/layouts/LayoutTechnicalPositionsPanel.test.tsx src/features/layouts/LayoutsView.test.tsx src/shared/i18n.test.ts`

Expected: PASS.

- [ ] **Step 7: Commit erstellen**

```powershell
git add frontend/src/features/layouts/LayoutPositionDialog.tsx `
  frontend/src/features/layouts/LayoutPositionDialog.test.tsx `
  frontend/src/features/layouts/LayoutTechnicalPositionsPanel.tsx `
  frontend/src/features/layouts/LayoutTechnicalPositionsPanel.test.tsx `
  frontend/src/features/layouts/LayoutWorkspace.tsx frontend/src/features/layouts/LayoutsView.test.tsx `
  frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts frontend/src/styles/layouts.css
git commit -m "feat(layouts): manage technical positions"
```

---

### Task 7: Gesamtprüfung und lokale Browser-Abnahme

**Files:**
- Modify only when a failing test or visual finding requires a narrow correction.

**Interfaces:**
- Consumes: completed Paket A.
- Produces: locally verified Stage-2 vertical slice.

- [ ] **Step 1: Go-Format und Backendtests ausführen**

Run: `gofmt -w backend/internal/domain/layout_twin.go backend/internal/application/layout_positions.go backend/internal/infrastructure/layout_position_repository.go backend/internal/api/layout_position_handlers.go`

Run: `cd backend; go test ./...`

Expected: PASS.

- [ ] **Step 2: Frontendtests und Build ausführen**

Run: `cd frontend; npm.cmd run test:run -- --reporter=dot`

Run: `cd frontend; npm.cmd run build`

Expected: alle Tests und der Build PASS.

- [ ] **Step 3: Git- und Vertragsstand prüfen**

Run: `git diff --check; git status --short`

Expected: keine Whitespacefehler, kein `frontend/dist`, keine Daten, Caches oder Zugangsdaten.

- [ ] **Step 4: Lokalen Server und Browser prüfen**

Prüfe `http://127.0.0.1:18083/layouts` mit bestehender QA-Sitzung:

1. Technikregister öffnet ohne Fehler.
2. Einheit wählen, Position anlegen, bearbeiten und archivieren.
3. Dialoge verwenden app-eigene Controls, Fokusfalle und Verwerfbestätigung.
4. Viewer/Editor sehen keine Schreibaktionen.
5. Deutsch/Englisch, Hell/Dunkel, Desktop und schmale Ansicht funktionieren.
6. Browser-Konsole enthält keine Fehler oder Warnungen.

- [ ] **Step 5: Abnahmefunde testgetrieben korrigieren**

Für jeden Fund zuerst den engsten fehlschlagenden Test ergänzen, dann nur die verantwortliche Stelle
korrigieren und anschließend Task 7 vollständig wiederholen.

- [ ] **Step 6: Finalen lokalen Prüfstand committen, falls nötig**

```powershell
git add backend frontend/src openapi/railkeeper.yaml
git commit -m "fix(layouts): finalize technical positions"
```

Kein Push, kein PR und kein Merge ohne einen späteren ausdrücklichen Auftrag.
