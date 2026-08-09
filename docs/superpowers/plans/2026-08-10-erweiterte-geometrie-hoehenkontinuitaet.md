# Höhenkontinuität Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Geometrisch verbundene zweipolige Gleise mit mehr als 0,01 mm Höhenversatz als präzisen,
anklickbaren Warnhinweis erkennen.

**Status:** Lokal vollständig umgesetzt und am 2026-08-10 abgenommen. Kein Push, PR oder Merge.

**Architecture:** Die bestehende Domänenanalyse trägt für zweipolige Geometrien die Anschlusshöhe in
die interne Anschlussdarstellung ein und erzeugt nach einer geometrischen Verbindung optional einen
`elevation_mismatch`-Hinweis. API und Änderungsvorschau verwenden das vorhandene Analysemodell. Die
Frontend-Planprüfung ergänzt lediglich Zähler, Detailtext und Fokusverhalten.

**Tech Stack:** Go, SQLite, React, TypeScript, Vitest, OpenAPI, bestehende RailKeeper-i18n und UI-Komponenten

## Global Constraints

- Nur Geometrien mit genau zwei geordneten Anschlüssen werden geprüft.
- Anschluss 1 verwendet `elevationStartMm`, Anschluss 2 verwendet `elevationEndMm`.
- Eine Differenz bis einschließlich 0,01 mm bleibt ohne Hinweis.
- Der Hinweis ist eine Warnung und verändert Verbindung, Snapping oder Veröffentlichungsregeln nicht.
- Mehrport-Geometrien werden nicht spekulativ bewertet.
- Alle Änderungen bleiben lokal, ohne Push, PR oder Merge.

---

### Task 1: Domänenanalyse für Höhenversatz

**Files:**
- Modify: `backend/internal/domain/track_plan_analysis.go`
- Test: `backend/internal/domain/track_plan_analysis_test.go`

**Interfaces:**
- Consumes: `PlanTrackObject.ElevationStartMM`, `PlanTrackObject.ElevationEndMM`, geordnete `TrackPort`-Liste
- Produces: `TrackPlanIssueElevationMismatch`, `TrackPlanIssue.ElevationDifferenceMM`

- [x] **Step 1: Fehlende und tolerierte Höhenkontinuität als fehlschlagende Tests ergänzen**

```go
func TestAnalyzeTrackPlanReportsElevationMismatchAtConnectedTwoPortGeometry(t *testing.T) {
	first := testG1Object("track-1", 0, 0, 0)
	first.ElevationStartMM, first.ElevationEndMM = 0, 10
	second := testG1Object("track-2", 166, 0, 0)
	second.ElevationStartMM, second.ElevationEndMM = 12, 12

	analysis := AnalyzeTrackPlan([]PlanTrackObject{second, first})
	issues := filterTrackIssues(analysis.Issues, TrackPlanIssueElevationMismatch)
	if len(issues) != 1 || issues[0].ElevationDifferenceMM == nil ||
		math.Abs(*issues[0].ElevationDifferenceMM-2) > 1e-9 {
		t.Fatalf("unexpected elevation mismatch: %#v", issues)
	}
	if !reflect.DeepEqual(issues[0].ObjectIDs, []string{"track-1", "track-2"}) ||
		!reflect.DeepEqual(issues[0].PortIDs, []string{"b", "a"}) {
		t.Fatalf("unexpected affected endpoints: %#v", issues[0])
	}
}

func TestAnalyzeTrackPlanHonorsElevationToleranceAndSkipsMultiPortGeometry(t *testing.T) {
	first := testG1Object("track-1", 0, 0, 0)
	first.ElevationEndMM = 10
	second := testG1Object("track-2", 166, 0, 0)
	second.ElevationStartMM = 10.01
	if issues := filterTrackIssues(AnalyzeTrackPlan([]PlanTrackObject{first, second}).Issues,
		TrackPlanIssueElevationMismatch); len(issues) != 0 {
		t.Fatalf("tolerance boundary produced mismatch: %#v", issues)
	}
	second.Geometry.Geometry.Ports = append(second.Geometry.Geometry.Ports,
		TrackPort{ID: "branch", XMM: 83, YMM: 20, DirectionDegrees: 90})
	second.ElevationStartMM = 20
	if issues := filterTrackIssues(AnalyzeTrackPlan([]PlanTrackObject{first, second}).Issues,
		TrackPlanIssueElevationMismatch); len(issues) != 0 {
		t.Fatalf("multi-port geometry produced speculative mismatch: %#v", issues)
	}
}
```

- [x] **Step 2: Domänentests ausführen und das erwartete Fehlschlagen bestätigen**

Run: `cd backend; go test ./internal/domain -run Elevation`

Expected: Build-Fehler wegen fehlendem `TrackPlanIssueElevationMismatch` und
`ElevationDifferenceMM`.

- [x] **Step 3: Interne Anschlusshöhen und den Warnhinweis implementieren**

```go
const TrackElevationConnectionToleranceMM = 0.01

const TrackPlanIssueElevationMismatch TrackPlanIssueCode = "elevation_mismatch"

type TrackPlanIssue struct {
	Code                  TrackPlanIssueCode     `json:"code"`
	Severity              TrackPlanIssueSeverity `json:"severity"`
	ObjectIDs             []string               `json:"objectIds"`
	PortIDs               []string               `json:"portIds,omitempty"`
	ElevationDifferenceMM *float64               `json:"elevationDifferenceMm,omitempty"`
}

type placedTrackPort struct {
	ObjectID      string
	Port          TrackPort
	ElevationMM   float64
	ElevationKnown bool
}

func trackPortElevation(object PlanTrackObject, portID string) (float64, bool) {
	if len(object.Geometry.Geometry.Ports) != 2 {
		return 0, false
	}
	for index, port := range object.Geometry.Geometry.Ports {
		if port.ID != portID {
			continue
		}
		if index == 0 {
			return object.ElevationStartMM, true
		}
		return object.ElevationEndMM, true
	}
	return 0, false
}
```

Beim Aufbau von `ports` werden `ElevationMM` und `ElevationKnown` gesetzt. Direkt nach dem Anhängen
einer geometrischen Verbindung wird `math.Abs(first.ElevationMM-second.ElevationMM)` berechnet. Wenn
beide Höhen bekannt sind und die Differenz größer als `TrackElevationConnectionToleranceMM` ist, wird
genau ein `TrackPlanIssueElevationMismatch` mit derselben Objekt- und Anschlussreihenfolge angehängt.

- [x] **Step 4: Domänentests und das gesamte Domain-Paket ausführen**

Run: `cd backend; go test ./internal/domain`

Expected: PASS.

- [x] **Step 5: Domänenänderung lokal committen**

```powershell
git add backend/internal/domain/track_plan_analysis.go backend/internal/domain/track_plan_analysis_test.go
git commit -m "feat(planner): detect track elevation mismatches"
```

### Task 2: API-Vertrag und Änderungsvorschau

**Files:**
- Modify: `backend/internal/api/track_planner_handlers_test.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`
- Test: `backend/internal/domain/track_plan_revision_diff_test.go`

**Interfaces:**
- Consumes: `domain.TrackPlanIssueElevationMismatch`, `TrackPlanIssue.ElevationDifferenceMM`
- Produces: JSON-Feld `elevationDifferenceMm` und OpenAPI-Enumwert `elevation_mismatch`

- [x] **Step 1: API- und Vorschautests um den neuen Hinweis ergänzen**

Im Handler-Test werden zwei verbundene Objekte mit Endhöhe 10 mm und Anfangshöhe 12 mm gespeichert.
Die dekodierte Analyse muss genau einen `elevation_mismatch`-Hinweis mit 2 mm Differenz enthalten. Im
Revisionsvergleich wird derselbe Hinweis als hinzugefügt und nach Angleichung als behoben erwartet.

- [x] **Step 2: Betroffene Tests ausführen und Vertragsfehler bestätigen**

Run: `cd backend; go test ./internal/api ./internal/domain -run "TrackPlanner|TrackPlan"`

Expected: Der OpenAPI-Vertrag enthält den neuen Enumwert und das Detailfeld noch nicht vollständig.

- [x] **Step 3: OpenAPI-Schema erweitern**

```yaml
    TrackPlanIssue:
      type: object
      required: [code, severity, objectIds]
      properties:
        code:
          type: string
          enum: [open_end, incompatible_connection, overlap, broken_geometry, elevation_mismatch]
        elevationDifferenceMm:
          type: number
          format: double
          minimum: 0
```

Der gleiche Enumwert wird bei `TrackPlanIssueChange.code` ergänzt. Die vorhandene JSON-Abbildung der
Go-Struktur liefert das optionale Detailfeld ohne neuen Endpunkt.

- [x] **Step 4: API-, Domain- und OpenAPI-Tests ausführen**

Run: `cd backend; go test ./internal/api ./internal/domain`

Expected: PASS.

- [x] **Step 5: Vertrag lokal committen**

```powershell
git add backend/internal/api/track_planner_handlers_test.go backend/internal/api/openapi_contract_test.go `
  backend/internal/domain/track_plan_revision_diff_test.go openapi/railkeeper.yaml
git commit -m "feat(api): expose elevation mismatch details"
```

### Task 3: Planprüfung im Frontend

**Files:**
- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/features/layouts/TrackPlanAnalysisPanel.tsx`
- Modify: `frontend/src/features/layouts/TrackPlanAnalysisPanel.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**
- Consumes: `TrackPlanIssueCode = "elevation_mismatch"`, optionales `elevationDifferenceMm`
- Produces: lokalisierter Zähler, Detailtext und vorhandenes `onSelectObject`-Fokusverhalten

- [x] **Step 1: Fehlschlagenden UI-Test für Zähler, Detail und Fokus ergänzen**

```tsx
const mismatch = {
  code: "elevation_mismatch" as const,
  severity: "warning" as const,
  objectIds: ["track-1", "track-2"],
  portIds: ["b", "a"],
  elevationDifferenceMm: 2
};

expect(screen.getByText("1 Höhenversatz")).toBeInTheDocument();
const issue = screen.getByRole("button", {
  name: "Warnung: Höhenversatz an Gleisverbindung (2,00 mm)"
});
await user.click(issue);
expect(selectObject).toHaveBeenCalledWith("track-1");
```

- [x] **Step 2: UI-Test ausführen und erwartetes Fehlschlagen bestätigen**

Run: `cd frontend; npm.cmd test -- --run src/features/layouts/TrackPlanAnalysisPanel.test.tsx`

Expected: Typ- oder Assertionfehler wegen fehlendem Code, Zähler und Text.

- [x] **Step 3: Typen, Darstellung und Übersetzungen implementieren**

`TrackPlanIssueCode` erhält `elevation_mismatch`, `TrackPlanIssue` erhält
`elevationDifferenceMm?: number`. Der Panel-Header zählt den Code zusätzlich. Der Beschriftungstext
wird mit `Intl.NumberFormat(language === "de" ? "de-DE" : "en-GB", {
minimumFractionDigits: 2, maximumFractionDigits: 2 })` formatiert. Die neuen Schlüssel lauten:

```text
layouts.trackAnalysis.elevationMismatchOne
layouts.trackAnalysis.elevationMismatchMany
layouts.trackAnalysis.issue.elevation_mismatch
layouts.trackAnalysis.issueElevationDetail
```

Deutsch verwendet „{count} Höhenversatz“, „{count} Höhenversätze“ und
„Höhenversatz an Gleisverbindung ({difference} mm)“. Englisch verwendet „{count} elevation mismatch“,
„{count} elevation mismatches“ und „Track connection elevation mismatch ({difference} mm)“.

- [x] **Step 4: Gezielte Tests und TypeScript-Build ausführen**

Run: `cd frontend; npm.cmd test -- --run src/features/layouts/TrackPlanAnalysisPanel.test.tsx`

Run: `cd frontend; npm.cmd run build`

Expected: Beide Befehle PASS.

- [x] **Step 5: Frontend lokal committen**

```powershell
git add frontend/src/shared/apiLayoutsAccessories.ts `
  frontend/src/features/layouts/TrackPlanAnalysisPanel.tsx `
  frontend/src/features/layouts/TrackPlanAnalysisPanel.test.tsx `
  frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat(planner): show elevation continuity warnings"
```

### Task 4: Gesamtprüfung und lokale Abnahme

**Files:**
- Modify: `docs/superpowers/specs/2026-08-10-erweiterte-geometrie-hoehenkontinuitaet-design.md`
- Modify: `docs/superpowers/plans/2026-08-10-erweiterte-geometrie-hoehenkontinuitaet.md`

**Interfaces:**
- Consumes: vollständig implementiertes Paket D
- Produces: reproduzierbares lokales Abnahmeprotokoll

- [x] **Step 1: Vollständige automatische Prüfungen ausführen**

Run: `cd backend; $env:GOCACHE='C:\Users\droth\Documents\GitHub\RailKeeper\.cache\go-build'; go test ./...`

Run: `cd frontend; npm.cmd test -- --run --reporter=verbose --maxWorkers=4`

Run: `cd frontend; npm.cmd run build`

Expected: alle Go-Pakete, alle Vitest-Tests und der Produktionsbuild PASS.

- [x] **Step 2: Server und Browser lokal aktualisieren**

Den vorhandenen Listener auf Port 18083 gezielt neu starten. Danach `/health` prüfen und die bestehende
angemeldete Browser-Sitzung unter `http://127.0.0.1:18083/layouts` vollständig neu laden.

- [x] **Step 3: Höhenversatz im Browser erzeugen und beheben**

Im bestehenden QA-Entwurf zwei verbundene Gleise verwenden. Am ersten Gleis die Anschlusshöhe 9,15 mm
belassen und am zweiten Gleis die zugehörige Anfangshöhe zunächst auf 11,15 mm setzen. Erwartet werden
ein Höhenversatz-Zähler, ein Warnhinweis mit 2,00 mm und Fokus auf das erste betroffene Gleis. Danach
die zweite Höhe auf 9,15 mm setzen. Erwartet werden 0 Höhenversätze und kein entsprechender Hinweis.

- [x] **Step 4: Abnahme dokumentieren und lokal committen**

Designstatus auf „lokal umgesetzt und abgenommen“ setzen, genaue Testergebnisse, Browserwerte,
Serverzustand und lokale Branch-Grenze eintragen. Alle Plan-Checkboxen abhaken.

```powershell
git add docs/superpowers/specs/2026-08-10-erweiterte-geometrie-hoehenkontinuitaet-design.md `
  docs/superpowers/plans/2026-08-10-erweiterte-geometrie-hoehenkontinuitaet.md
git commit -m "docs: record elevation continuity acceptance"
```

- [x] **Step 5: Sauberen lokalen Abschluss prüfen**

Run: `git diff --check; git status --short; git log -5 --oneline`

Expected: keine Diff-Fehler, leeres Status-Listing, Paket-D-Commits nur auf
`dev/issue-36-advanced-geometry` und kein Push.
