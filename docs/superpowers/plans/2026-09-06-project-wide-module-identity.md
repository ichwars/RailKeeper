# Projektweite Modulidentität Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this
> plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. Work completely inline and
> do not dispatch subagents.

**Goal:** RailKeeper erhält eine projektweite Modulbibliothek mit getrennten Modulvorlagen,
versionierten technischen Revisionen und physischen Modulexemplaren, einschließlich verlustfreier
Migration, Bestandsprüfung, Anlagenwiederverwendung, OpenAPI sowie Backup und Restore.

**Architecture:** Die Migration ergänzt das bestehende Modell additiv. `layout_units` bleibt in
Stufe 1 die technische Kompatibilitätsschicht für bestehende Konturen, Ports, Gleispläne und
Zubehörbezüge. Die neuen Tabellen `module_templates`, `module_template_revisions`,
`module_instances` und `layout_module_template_usages` bilden die kanonischen projektweiten
Identitäten. Bestehende IDs werden in den getrennten Tabellen wiederverwendet, sodass die Zuordnung
deterministisch bleibt. Anlagenkonfigurationen erhalten zusätzlich die kanonische
`module_template_revision_id`; ihr bestehender `unit_id` bleibt bis Stufe 4 erhalten.

**Tech Stack:** Go, `net/http`, SQLite mit vorwärtsgerichteten SQL-Migrationen, React 19,
TypeScript 7, Vite 8, Vitest, Testing Library, OpenAPI 3.1.

## Global Constraints

- Dieser Plan implementiert ausschließlich **Stufe 1: Projektweite Modulidentität**.
- Die Umsetzung erfolgt vollständig inline und ohne Subagenten.
- Vor sichtbaren Frontendänderungen müssen beide Referenzdateien vorhanden sein und exakt diese
  SHA-256-Werte besitzen:
  `3CAB446FD1FB6D3DC24391E7FF6C050CCEFF841EC5F429CF410BBF943522ABA5` und
  `C387A13776ED3C839BEACAEED0D1B7A186C2749D56CF9BD29D5926E6EA8F1DA2`.
- Stufe 1 baut noch keinen Form-, Kopf- oder Normeditor. Die verbindliche Variante-2-Arbeitsfläche
  bleibt für Stufe 3 unverändert maßgeblich.
- Migrierte Datensätze erhalten weder geratene Normen noch behauptete Ausstellungsverfügbarkeit.
- Jedes bestehende `LayoutUnit` wird genau einer Vorlage, einer ersten Revision und einem Exemplar
  mit dem Status **Bestand prüfen** zugeordnet.
- Die erste migrierte Revision bleibt `draft`, weil ihre bestehende technische Definition heute
  veränderlich ist. Erst die ausdrückliche Freigabe macht sie unveränderlich.
- Ein Exemplar darf erst als geprüft bestätigt werden, wenn seine Revision `released` ist, eine
  eindeutige Kennung besitzt und Zustand sowie Verfügbarkeit ausdrücklich gewählt wurden.
- Bestehende Anlagenkonfigurationen behalten Position, Reihenfolge, Drehung und
  Gleisplanrevision. Sie referenzieren zusätzlich die migrierte Vorlagenrevision, aber kein
  physisches Exemplar.
- Admin und Planner dürfen Vorlagen, Revisionen, Exemplare und Anlagenzuordnungen ändern. Admin,
  Planner, Editor und Viewer dürfen sie lesen. Messe erhält keinen allgemeinen Bibliothekszugriff.
- Schreibende Endpunkte bleiben durch die bestehende CSRF- und Rollen-Middleware geschützt.
- Deutsche und englische UI-Texte stammen aus der gemeinsamen i18n-Struktur und werden vor
  Abschluss auf Rechtschreibung, Terminologie und Platzhaltergleichheit geprüft.
- `widthMm` und `heightMm` bleiben in Stufe 1 erhalten. Stabile Kontursegmente,
  Verbindungsseiten, Normprofile, Portbindungen und `Z`-Platzierungen gehören nicht in diesen Plan.
- SQLite bleibt die lokale Quelle der Wahrheit. Es werden keine Cloud-, Multi-Tenant- oder
  öffentlichen Freigabefunktionen ergänzt.

---

## Dateistruktur und Verantwortlichkeiten

### Neue Backenddateien

- `backend/migrations/0065_project_wide_module_identity.sql`: Tabellen, Indizes, deterministische
  Datenmigration, Konfigurationsreferenz und Unveränderlichkeits-Trigger.
- `backend/internal/domain/module_library.go`: ausschließlich Modulbibliotheks-Enums und deren
  Gültigkeitsprüfung.
- `backend/internal/domain/module_library_test.go`: vollständige Enum-Matrix.
- `backend/internal/application/module_library.go`: API-neutrale Datentypen, Filter,
  Repository-Port, Service und Validierung.
- `backend/internal/application/module_library_test.go`: Servicevalidierung und Normalisierung.
- `backend/internal/infrastructure/module_library_repository.go`: paginierte Abfragen und
  transaktionale Schreibvorgänge inklusive Audit.
- `backend/internal/infrastructure/module_library_repository_test.go`: Persistenz,
  Seitennavigation, Konflikte, Freigabe, Bestandsprüfung und Anlagenzuordnung.
- `backend/internal/infrastructure/module_identity_migration_test.go`: Migration eines realistischen
  Bestandsdatensatzes durch Migration 0065.
- `backend/internal/api/module_library_handlers.go`: Query-Parsing, JSON-Handler und fachliche
  Fehlerabbildung.
- `backend/internal/api/module_library_handlers_test.go`: HTTP-Vertrag, Rollen und CSRF.
- `backend/internal/application/backup_module_library.go`: Restore-Vorprüfung und Backfill alter
  Backups ohne Modulbibliothek.

### Geänderte Backenddateien

- `backend/internal/application/layouts.go:45-96,137-166`: Legacy-Antworten um kanonische IDs
  ergänzen, Konfigurationen auf Revisionen verweisen lassen.
- `backend/internal/infrastructure/layout_repository.go:118-333,397-573,586-704`: projektweite
  Anlagenverwendung, Legacy-Erstellung und Konfigurationsreferenzen synchronisieren.
- `backend/internal/infrastructure/layout_outline_repository.go:1-105`: Triggerfehler freigegebener
  Revisionen in den fachlichen Unveränderlichkeitsfehler übersetzen.
- `backend/internal/infrastructure/layout_position_repository.go:1-180`: denselben Schutz für
  technische Positionen anwenden.
- `backend/internal/infrastructure/layout_twin_repository.go:21-116`: Zugehörigkeit über
  `layout_module_template_usages` statt über Eigentum prüfen.
- `backend/internal/infrastructure/track_planner_repository.go:55-70,698-714`: Spurweite aus der
  Vorlagenrevision lesen, nicht aus der Ursprungsanlage.
- `backend/internal/api/router.go:20-92`: `ModuleLibraryService` in Konfiguration und App aufnehmen.
- `backend/internal/api/layout_handlers.go:171-205`: Unveränderlichkeit auch für kompatible
  Layout-Unit-Endpunkte als HTTP 409 ausgeben.
- `backend/internal/api/routes.go:190-220`: Bibliotheksrouten mit Viewer- beziehungsweise
  Planner-Zugriff registrieren.
- `backend/cmd/railkeeper/startup.go:220-268`: Repository und Service verdrahten.
- `backend/internal/api/openapi_contract_test.go:170-205`: neue Schemas und Operationen prüfen.
- `backend/internal/api/routes_security_test.go:15-103`: Bibliotheksrouten in die vorhandene
  Rollenmatrix einbeziehen.
- `backend/internal/application/backup.go:23-205,326-410,678-875`: Backup-Version 21 und neue
  Tabellen registrieren, Legacy-Backfill aufrufen.
- `backend/internal/application/backup_test.go:591-657,1937-2128`: aktueller Roundtrip und Restore
  eines Backups der Version 20.
- `openapi/railkeeper.yaml:3636-3768,8660-8890`: neue Schemas, Filterparameter, Antworten und
  Endpunkte dokumentieren.

### Neue Frontenddateien

- `frontend/src/features/layouts/modules/ModuleLibraryPanel.tsx`: gemeinsame globale und
  anlagenbezogene Bibliotheksansicht.
- `frontend/src/features/layouts/modules/ModuleLibraryPanel.test.tsx`: Suche, Filter, Pagination,
  Auswahl, Anlagenzuordnung und Rollen.
- `frontend/src/features/layouts/modules/ModuleTemplateDialog.tsx`: Vorlage mit erster Revision
  anlegen sowie neue Entwurfsrevision erstellen.
- `frontend/src/features/layouts/modules/ModuleInstanceDialog.tsx`: physisches Exemplar anlegen
  und bearbeiten.
- `frontend/src/features/layouts/modules/ModuleInventoryReviewDialog.tsx`: explizite
  Bestandsprüfung mit Kennung, Zustand und Verfügbarkeit.
- `frontend/src/features/layouts/modules/ModuleInventoryReviewDialog.test.tsx`: Pflichtangaben,
  Freigabesperre, Konflikt und Fokusführung.
- `frontend/src/features/layouts/ModuleLibraryView.tsx`: eigenständige Route `/modules`.
- `frontend/src/styles/module-library.css`: ausschließlich Bibliotheksliste, Detailbereich und
  Dialoggeometrie.

### Geänderte Frontenddateien

- `frontend/src/shared/apiLayoutsAccessories.ts:1-240,1120-1185`: Typen, Filter und API-Methoden.
- `frontend/src/shared/apiLayoutsAccessories.test.ts:12-130`: URL-, Body- und CSRF-Vertrag.
- `frontend/src/app/App.tsx:8-62,190-202`: eigenständige Modulbibliotheksroute ohne neuen
  Ausstellungszugriff.
- `frontend/src/app/navigationAvailability.ts:3-13`: `modules` als erlaubte interne Route.
- `frontend/src/features/layouts/LayoutsView.tsx:1-115`: Link zur projektweiten Bibliothek.
- `frontend/src/features/layouts/LayoutWorkspace.tsx:1-170`: anlagenbezogenen Modulbereich auf
  dieselbe Bibliothekskomponente umstellen.
- `frontend/src/features/layouts/LayoutModulesPanel.tsx:1-89`: schmaler Adapter für den
  Anlagenkontext, keine zweite Datenhaltung.
- `frontend/src/shared/i18n/de.ts:2504-2605`: deutsche Bibliotheks- und Dialogtexte.
- `frontend/src/shared/i18n/en.ts:2504-2605`: englische Bibliotheks- und Dialogtexte.
- `frontend/src/shared/i18n.test.ts:41-103`: Schlüsselgleichheit und festgelegte Fachbegriffe.
- `frontend/src/app/styles.css:20-23`: fokussierte CSS-Datei importieren.

---

### Task 1: Fachliche Enum-Typen festlegen

**Files:**
- Create: `backend/internal/domain/module_library.go`
- Create: `backend/internal/domain/module_library_test.go`

**Interfaces:**
- Consumes: `domain.LayoutUnitKind` aus `backend/internal/domain/layout.go`.
- Produces: `ModuleTemplateRevisionStatus`, `ModuleInstanceCondition`,
  `ModuleInstanceAvailability`, `ModuleInventoryReviewStatus` und jeweils `Valid() bool`.

- [ ] **Step 1: Fehlenden Enum-Vertrag als Test festhalten**

```go
package domain_test

import (
	"testing"

	"railkeeper/backend/internal/domain"
)

func TestModuleLibraryEnumsAcceptOnlyDefinedValues(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"draft revision", domain.ModuleTemplateRevisionDraft.Valid()},
		{"released revision", domain.ModuleTemplateRevisionReleased.Valid()},
		{"archived revision", domain.ModuleTemplateRevisionArchived.Valid()},
		{"unknown condition", domain.ModuleInstanceConditionUnknown.Valid()},
		{"ready condition", domain.ModuleInstanceConditionReady.Valid()},
		{"maintenance condition", domain.ModuleInstanceConditionMaintenanceDue.Valid()},
		{"defective condition", domain.ModuleInstanceConditionDefective.Valid()},
		{"unchecked availability", domain.ModuleInstanceAvailabilityUnchecked.Valid()},
		{"available availability", domain.ModuleInstanceAvailabilityAvailable.Valid()},
		{"unavailable availability", domain.ModuleInstanceAvailabilityUnavailable.Valid()},
		{"review required", domain.ModuleInventoryReviewRequired.Valid()},
		{"review confirmed", domain.ModuleInventoryReviewConfirmed.Valid()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !test.valid {
				t.Fatal("documented enum value must be valid")
			}
		})
	}
	if domain.ModuleTemplateRevisionStatus("published").Valid() ||
		domain.ModuleInstanceCondition("stored").Valid() ||
		domain.ModuleInstanceAvailability("reserved").Valid() ||
		domain.ModuleInventoryReviewStatus("approved").Valid() {
		t.Fatal("undocumented enum value accepted")
	}
}
```

- [ ] **Step 2: Test ausführen und den erwarteten Compilerfehler prüfen**

Run: `cd backend; go test ./internal/domain -run TestModuleLibraryEnumsAcceptOnlyDefinedValues`

Expected: FAIL mit `undefined: domain.ModuleTemplateRevisionDraft`.

- [ ] **Step 3: Die vier geschlossenen Enum-Mengen implementieren**

```go
package domain

type ModuleTemplateRevisionStatus string

const (
	ModuleTemplateRevisionDraft    ModuleTemplateRevisionStatus = "draft"
	ModuleTemplateRevisionReleased ModuleTemplateRevisionStatus = "released"
	ModuleTemplateRevisionArchived ModuleTemplateRevisionStatus = "archived"
)

func (status ModuleTemplateRevisionStatus) Valid() bool {
	return status == ModuleTemplateRevisionDraft || status == ModuleTemplateRevisionReleased ||
		status == ModuleTemplateRevisionArchived
}

type ModuleInstanceCondition string

const (
	ModuleInstanceConditionUnknown        ModuleInstanceCondition = "unknown"
	ModuleInstanceConditionReady          ModuleInstanceCondition = "ready"
	ModuleInstanceConditionMaintenanceDue ModuleInstanceCondition = "maintenance_due"
	ModuleInstanceConditionDefective      ModuleInstanceCondition = "defective"
)

func (condition ModuleInstanceCondition) Valid() bool {
	return condition == ModuleInstanceConditionUnknown || condition == ModuleInstanceConditionReady ||
		condition == ModuleInstanceConditionMaintenanceDue ||
		condition == ModuleInstanceConditionDefective
}

type ModuleInstanceAvailability string

const (
	ModuleInstanceAvailabilityUnchecked   ModuleInstanceAvailability = "unchecked"
	ModuleInstanceAvailabilityAvailable   ModuleInstanceAvailability = "available"
	ModuleInstanceAvailabilityUnavailable ModuleInstanceAvailability = "unavailable"
)

func (availability ModuleInstanceAvailability) Valid() bool {
	return availability == ModuleInstanceAvailabilityUnchecked ||
		availability == ModuleInstanceAvailabilityAvailable ||
		availability == ModuleInstanceAvailabilityUnavailable
}

type ModuleInventoryReviewStatus string

const (
	ModuleInventoryReviewRequired  ModuleInventoryReviewStatus = "review_required"
	ModuleInventoryReviewConfirmed ModuleInventoryReviewStatus = "confirmed"
)

func (status ModuleInventoryReviewStatus) Valid() bool {
	return status == ModuleInventoryReviewRequired || status == ModuleInventoryReviewConfirmed
}
```

- [ ] **Step 4: Domain-Test vollständig ausführen**

Run: `cd backend; go test ./internal/domain`

Expected: PASS.

- [ ] **Step 5: Commit erstellen**

```powershell
git add backend/internal/domain/module_library.go backend/internal/domain/module_library_test.go
git commit -m "feat(modules): define module library states"
```

### Task 2: Additive Datenmigration und Unveränderlichkeit

**Files:**
- Create: `backend/migrations/0065_project_wide_module_identity.sql`
- Create: `backend/internal/infrastructure/module_identity_migration_test.go`
- Modify: `backend/internal/infrastructure/layout_accessory_schema_test.go:12-52`

**Interfaces:**
- Consumes: `layouts`, `layout_units`, `layout_configuration_units`, `layout_unit_ports`,
  `layout_unit_outline_points`, `layout_technical_positions` aus Migrationen 0039, 0044 und 0048.
- Produces: die vier kanonischen Tabellen und
  `layout_configuration_units.module_template_revision_id`.

- [ ] **Step 1: Einen realistischen Bestandsdatensatz vor Migration 0065 testen**

```go
func TestProjectWideModuleIdentityMigrationPreservesLayoutUse(t *testing.T) {
	root := t.TempDir()
	partial := filepath.Join(root, "migrations-through-0064")
	if err := os.Mkdir(partial, 0o700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), partial,
		"0064_replace_bundled_function_symbols.sql")
	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, partial); err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`INSERT INTO layouts(id,name,kind,gauge,scale,created_at,updated_at)
		 VALUES('layout-1','Vereinsanlage','club','TT','1:120','now','now')`,
		`INSERT INTO layout_units(id,layout_id,name,kind,owner_label,width_mm,height_mm,created_at,updated_at)
		 VALUES('unit-1','layout-1','Bahnhof Süd','module','MEC',1200,500,'now','now')`,
		`INSERT INTO layout_configurations(id,layout_id,name,created_at,updated_at)
		 VALUES('configuration-1','layout-1','Standard','now','now')`,
		`INSERT INTO layout_configuration_units(configuration_id,unit_id,position_x_mm,position_y_mm,
		 rotation_degrees,sort_order) VALUES('configuration-1','unit-1',1500,250,90,3)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	var templateName, revisionStatus, owner, reviewStatus, revisionID string
	if err := db.QueryRow(`SELECT template.name, revision.status, instance.owner_label,
	 instance.review_status, configuration.module_template_revision_id
	 FROM module_templates template
	 JOIN module_template_revisions revision ON revision.module_template_id=template.id
	 JOIN module_instances instance ON instance.module_template_revision_id=revision.id
	 JOIN layout_configuration_units configuration ON configuration.unit_id=revision.source_layout_unit_id
	 WHERE template.id='unit-1'`).Scan(
		&templateName, &revisionStatus, &owner, &reviewStatus, &revisionID,
	); err != nil {
		t.Fatal(err)
	}
	if templateName != "Bahnhof Süd" || revisionStatus != "draft" || owner != "MEC" ||
		reviewStatus != "review_required" || revisionID != "unit-1" {
		t.Fatalf("unexpected migrated identity: %q %q %q %q %q",
			templateName, revisionStatus, owner, reviewStatus, revisionID)
	}
	assertForeignKeyCheck(t, db)
}
```

- [ ] **Step 2: Migrationstest ausführen und die fehlende Tabelle bestätigen**

Run: `cd backend; go test ./internal/infrastructure -run TestProjectWideModuleIdentityMigrationPreservesLayoutUse`

Expected: FAIL mit `no such table: module_templates`.

- [ ] **Step 3: Tabellen, Backfill, Indizes und Trigger in Migration 0065 anlegen**

```sql
CREATE TABLE module_templates (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
  archived INTEGER NOT NULL DEFAULT 0 CHECK (archived IN (0, 1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX ix_module_templates_name ON module_templates(name COLLATE NOCASE, id);
CREATE INDEX ix_module_templates_archived ON module_templates(archived, name COLLATE NOCASE, id);

CREATE TABLE module_template_revisions (
  id TEXT PRIMARY KEY,
  module_template_id TEXT NOT NULL,
  revision_number INTEGER NOT NULL CHECK (revision_number > 0),
  base_revision_id TEXT,
  status TEXT NOT NULL CHECK (status IN ('draft', 'released', 'archived')),
  kind TEXT NOT NULL CHECK (kind IN ('baseboard', 'module', 'segment', 'area')),
  gauge TEXT NOT NULL,
  width_mm REAL NOT NULL DEFAULT 0 CHECK (width_mm >= 0),
  height_mm REAL NOT NULL DEFAULT 0 CHECK (height_mm >= 0),
  source_layout_unit_id TEXT,
  version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
  released_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (module_template_id) REFERENCES module_templates(id) ON DELETE RESTRICT,
  FOREIGN KEY (base_revision_id) REFERENCES module_template_revisions(id) ON DELETE RESTRICT,
  FOREIGN KEY (source_layout_unit_id) REFERENCES layout_units(id) ON DELETE RESTRICT,
  UNIQUE (module_template_id, revision_number),
  UNIQUE (source_layout_unit_id),
  CHECK ((status = 'draft' AND released_at IS NULL) OR
         (status IN ('released', 'archived') AND released_at IS NOT NULL))
);

CREATE UNIQUE INDEX ux_module_template_revisions_draft
  ON module_template_revisions(module_template_id) WHERE status='draft';
CREATE INDEX ix_module_template_revisions_template
  ON module_template_revisions(module_template_id, revision_number DESC);
CREATE INDEX ix_module_template_revisions_gauge
  ON module_template_revisions(gauge COLLATE NOCASE, status);

CREATE TABLE module_instances (
  id TEXT PRIMARY KEY,
  module_template_revision_id TEXT NOT NULL,
  inventory_number TEXT NOT NULL DEFAULT '',
  owner_label TEXT NOT NULL DEFAULT '',
  condition_state TEXT NOT NULL DEFAULT 'unknown'
    CHECK (condition_state IN ('unknown', 'ready', 'maintenance_due', 'defective')),
  availability_status TEXT NOT NULL DEFAULT 'unchecked'
    CHECK (availability_status IN ('unchecked', 'available', 'unavailable')),
  review_status TEXT NOT NULL DEFAULT 'review_required'
    CHECK (review_status IN ('review_required', 'confirmed')),
  location_label TEXT NOT NULL DEFAULT '',
  notes TEXT NOT NULL DEFAULT '',
  version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
  archived INTEGER NOT NULL DEFAULT 0 CHECK (archived IN (0, 1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (module_template_revision_id) REFERENCES module_template_revisions(id) ON DELETE RESTRICT,
  CHECK (review_status <> 'confirmed' OR
         (inventory_number <> '' AND condition_state <> 'unknown' AND availability_status <> 'unchecked'))
);

CREATE UNIQUE INDEX ux_module_instances_inventory_number
  ON module_instances(inventory_number COLLATE NOCASE) WHERE inventory_number <> '';
CREATE INDEX ix_module_instances_revision
  ON module_instances(module_template_revision_id, archived, inventory_number COLLATE NOCASE);
CREATE INDEX ix_module_instances_review
  ON module_instances(review_status, availability_status, condition_state);

CREATE TABLE layout_module_template_usages (
  layout_id TEXT NOT NULL,
  module_template_id TEXT NOT NULL,
  origin TEXT NOT NULL CHECK (origin IN ('migrated', 'planned')),
  created_at TEXT NOT NULL,
  PRIMARY KEY (layout_id, module_template_id),
  FOREIGN KEY (layout_id) REFERENCES layouts(id) ON DELETE RESTRICT,
  FOREIGN KEY (module_template_id) REFERENCES module_templates(id) ON DELETE RESTRICT
);

CREATE INDEX ix_layout_module_template_usages_template
  ON layout_module_template_usages(module_template_id, layout_id);

ALTER TABLE layout_configuration_units
  ADD COLUMN module_template_revision_id TEXT REFERENCES module_template_revisions(id) ON DELETE RESTRICT;

INSERT INTO module_templates(id, name, description, version, archived, created_at, updated_at)
SELECT id, name, '', 1, archived, created_at, updated_at FROM layout_units;

INSERT INTO module_template_revisions(
  id, module_template_id, revision_number, base_revision_id, status, kind, gauge, width_mm, height_mm,
  source_layout_unit_id, version, released_at, created_at, updated_at
)
SELECT unit.id, unit.id, 1, NULL, 'draft', unit.kind, layout.gauge,
       COALESCE(unit.width_mm, 0), COALESCE(unit.height_mm, 0), unit.id, 1, NULL,
       unit.created_at, unit.updated_at
FROM layout_units unit
JOIN layouts layout ON layout.id=unit.layout_id;

INSERT INTO module_instances(
  id, module_template_revision_id, inventory_number, owner_label, condition_state,
  availability_status, review_status, location_label, notes, version, archived, created_at, updated_at
)
SELECT id, id, '', owner_label, 'unknown', 'unchecked', 'review_required', '', '', 1,
       archived, created_at, updated_at
FROM layout_units;

INSERT INTO layout_module_template_usages(layout_id, module_template_id, origin, created_at)
SELECT layout_id, id, 'migrated', created_at FROM layout_units;

UPDATE layout_configuration_units
SET module_template_revision_id=unit_id
WHERE module_template_revision_id IS NULL;

CREATE TRIGGER validate_module_revision_base_insert
BEFORE INSERT ON module_template_revisions
WHEN NEW.base_revision_id IS NOT NULL AND NOT EXISTS (
  SELECT 1 FROM module_template_revisions base
  WHERE base.id=NEW.base_revision_id
    AND base.module_template_id=NEW.module_template_id
    AND base.status IN ('released', 'archived')
    AND base.revision_number<NEW.revision_number
)
BEGIN
  SELECT RAISE(ABORT, 'module_revision_invalid_base');
END;

CREATE TRIGGER validate_module_revision_base_update
BEFORE UPDATE OF base_revision_id, module_template_id, revision_number
ON module_template_revisions
WHEN NEW.base_revision_id IS NOT NULL AND NOT EXISTS (
  SELECT 1 FROM module_template_revisions base
  WHERE base.id=NEW.base_revision_id
    AND base.module_template_id=NEW.module_template_id
    AND base.status IN ('released', 'archived')
    AND base.revision_number<NEW.revision_number
)
BEGIN
  SELECT RAISE(ABORT, 'module_revision_invalid_base');
END;

CREATE TRIGGER validate_module_instance_review_insert
BEFORE INSERT ON module_instances
WHEN NEW.review_status='confirmed' AND NOT EXISTS (
  SELECT 1 FROM module_template_revisions revision
  WHERE revision.id=NEW.module_template_revision_id AND revision.status='released'
)
BEGIN
  SELECT RAISE(ABORT, 'module_instance_revision_not_released');
END;

CREATE TRIGGER validate_module_instance_review_update
BEFORE UPDATE ON module_instances
WHEN NEW.review_status='confirmed'
  AND (NEW.review_status IS NOT OLD.review_status OR
       NEW.module_template_revision_id IS NOT OLD.module_template_revision_id)
  AND NOT EXISTS (
    SELECT 1 FROM module_template_revisions revision
    WHERE revision.id=NEW.module_template_revision_id AND revision.status='released'
  )
BEGIN
  SELECT RAISE(ABORT, 'module_instance_revision_not_released');
END;

CREATE TRIGGER guard_released_module_revision_update
BEFORE UPDATE ON module_template_revisions
WHEN OLD.status IN ('released', 'archived') AND (
  NEW.status NOT IN ('released', 'archived') OR
  NEW.module_template_id IS NOT OLD.module_template_id OR
  NEW.revision_number IS NOT OLD.revision_number OR
  NEW.base_revision_id IS NOT OLD.base_revision_id OR
  NEW.kind IS NOT OLD.kind OR NEW.gauge IS NOT OLD.gauge OR
  NEW.width_mm IS NOT OLD.width_mm OR NEW.height_mm IS NOT OLD.height_mm OR
  NEW.source_layout_unit_id IS NOT OLD.source_layout_unit_id OR
  NEW.released_at IS NOT OLD.released_at
)
BEGIN
  SELECT RAISE(ABORT, 'module_revision_immutable');
END;

CREATE TRIGGER guard_released_module_revision_delete
BEFORE DELETE ON module_template_revisions
WHEN OLD.status IN ('released', 'archived')
BEGIN
  SELECT RAISE(ABORT, 'module_revision_immutable');
END;

CREATE TRIGGER guard_released_layout_unit_update
BEFORE UPDATE ON layout_units
WHEN EXISTS (
  SELECT 1 FROM module_template_revisions
  WHERE source_layout_unit_id=OLD.id AND status IN ('released', 'archived')
)
BEGIN
  SELECT RAISE(ABORT, 'module_revision_immutable');
END;

CREATE TRIGGER guard_released_layout_unit_delete
BEFORE DELETE ON layout_units
WHEN EXISTS (
  SELECT 1 FROM module_template_revisions
  WHERE source_layout_unit_id=OLD.id AND status IN ('released', 'archived')
)
BEGIN
  SELECT RAISE(ABORT, 'module_revision_immutable');
END;

CREATE TRIGGER guard_released_layout_unit_port_insert
BEFORE INSERT ON layout_unit_ports
WHEN EXISTS (
  SELECT 1 FROM module_template_revisions
  WHERE source_layout_unit_id=NEW.layout_unit_id AND status IN ('released', 'archived')
)
BEGIN
  SELECT RAISE(ABORT, 'module_revision_immutable');
END;

CREATE TRIGGER guard_released_layout_unit_port_update
BEFORE UPDATE ON layout_unit_ports
WHEN EXISTS (
  SELECT 1 FROM module_template_revisions
  WHERE source_layout_unit_id=OLD.layout_unit_id AND status IN ('released', 'archived')
)
OR EXISTS (
  SELECT 1 FROM module_template_revisions
  WHERE source_layout_unit_id=NEW.layout_unit_id AND status IN ('released', 'archived')
)
BEGIN
  SELECT RAISE(ABORT, 'module_revision_immutable');
END;

CREATE TRIGGER guard_released_layout_unit_port_delete
BEFORE DELETE ON layout_unit_ports
WHEN EXISTS (
  SELECT 1 FROM module_template_revisions
  WHERE source_layout_unit_id=OLD.layout_unit_id AND status IN ('released', 'archived')
)
BEGIN
  SELECT RAISE(ABORT, 'module_revision_immutable');
END;

CREATE TRIGGER guard_released_layout_unit_outline_insert
BEFORE INSERT ON layout_unit_outline_points
WHEN EXISTS (
  SELECT 1 FROM module_template_revisions
  WHERE source_layout_unit_id=NEW.layout_unit_id AND status IN ('released', 'archived')
)
BEGIN
  SELECT RAISE(ABORT, 'module_revision_immutable');
END;

CREATE TRIGGER guard_released_layout_unit_outline_update
BEFORE UPDATE ON layout_unit_outline_points
WHEN EXISTS (
  SELECT 1 FROM module_template_revisions
  WHERE source_layout_unit_id=OLD.layout_unit_id AND status IN ('released', 'archived')
)
OR EXISTS (
  SELECT 1 FROM module_template_revisions
  WHERE source_layout_unit_id=NEW.layout_unit_id AND status IN ('released', 'archived')
)
BEGIN
  SELECT RAISE(ABORT, 'module_revision_immutable');
END;

CREATE TRIGGER guard_released_layout_unit_outline_delete
BEFORE DELETE ON layout_unit_outline_points
WHEN EXISTS (
  SELECT 1 FROM module_template_revisions
  WHERE source_layout_unit_id=OLD.layout_unit_id AND status IN ('released', 'archived')
)
BEGIN
  SELECT RAISE(ABORT, 'module_revision_immutable');
END;

CREATE TRIGGER guard_released_layout_technical_position_insert
BEFORE INSERT ON layout_technical_positions
WHEN EXISTS (
  SELECT 1 FROM module_template_revisions
  WHERE source_layout_unit_id=NEW.layout_unit_id AND status IN ('released', 'archived')
)
BEGIN
  SELECT RAISE(ABORT, 'module_revision_immutable');
END;

CREATE TRIGGER guard_released_layout_technical_position_update
BEFORE UPDATE ON layout_technical_positions
WHEN EXISTS (
  SELECT 1 FROM module_template_revisions
  WHERE source_layout_unit_id=OLD.layout_unit_id AND status IN ('released', 'archived')
)
OR EXISTS (
  SELECT 1 FROM module_template_revisions
  WHERE source_layout_unit_id=NEW.layout_unit_id AND status IN ('released', 'archived')
)
BEGIN
  SELECT RAISE(ABORT, 'module_revision_immutable');
END;

CREATE TRIGGER guard_released_layout_technical_position_delete
BEFORE DELETE ON layout_technical_positions
WHEN EXISTS (
  SELECT 1 FROM module_template_revisions
  WHERE source_layout_unit_id=OLD.layout_unit_id AND status IN ('released', 'archived')
)
BEGIN
  SELECT RAISE(ABORT, 'module_revision_immutable');
END;
```

Der Migrationstest ergänzt je einen direkten Insert-, Update- und Löschversuch nach Freigabe und
erwartet für jede geschützte Tabelle `module_revision_immutable`.

- [ ] **Step 4: Migration und Fremdschlüssel prüfen**

Run: `cd backend; go test ./internal/infrastructure -run 'TestProjectWideModuleIdentityMigration|TestLayoutAccessoryMigration'`

Expected: PASS, einschließlich `PRAGMA foreign_key_check` ohne Zeilen.

- [ ] **Step 5: Commit erstellen**

```powershell
git add backend/migrations/0065_project_wide_module_identity.sql `
  backend/internal/infrastructure/module_identity_migration_test.go `
  backend/internal/infrastructure/layout_accessory_schema_test.go
git commit -m "feat(modules): migrate project-wide module identities"
```

### Task 3: Application-Service und Validierungsregeln

**Files:**
- Create: `backend/internal/application/module_library.go`
- Create: `backend/internal/application/module_library_test.go`

**Interfaces:**
- Consumes: Enums aus Task 1 und `domain.LayoutUnitKind`.
- Produces: `ModuleLibraryRepository`, `ModuleLibraryService`, Seiten- und Eingabetypen für
  Repository, HTTP und Frontendvertrag.

- [ ] **Step 1: Servicevalidierung mit einem Repository-Spy festhalten**

```go
type moduleLibraryRepositorySpy struct {
	ModuleLibraryRepository
	templateFilter ModuleTemplateFilter
}

func (spy *moduleLibraryRepositorySpy) ListTemplates(
	_ context.Context,
	filter ModuleTemplateFilter,
) (*ModuleTemplatePage, error) {
	spy.templateFilter = filter
	return &ModuleTemplatePage{Items: []ModuleTemplate{}, Page: filter.Page, PageSize: filter.PageSize}, nil
}

func TestModuleLibraryServiceNormalizesFiltersAndRequiresExplicitReview(t *testing.T) {
	spy := &moduleLibraryRepositorySpy{}
	service := NewModuleLibraryService(spy)
	_, err := service.ListTemplates(t.Context(), ModuleTemplateFilter{Query: "  Bahnhof  ", Page: -1, PageSize: 500})
	if err != nil {
		t.Fatal(err)
	}
	if spy.templateFilter.Query != "Bahnhof" || spy.templateFilter.Page != 1 || spy.templateFilter.PageSize != 100 {
		t.Fatalf("unexpected normalized filter: %#v", spy.templateFilter)
	}
	_, err = service.ReviewInstance(t.Context(), "instance-1", ReviewModuleInstanceInput{
		InventoryNumber: " M-001 ", Condition: domain.ModuleInstanceConditionUnknown,
		Availability: domain.ModuleInstanceAvailabilityAvailable, ExpectedVersion: 1,
	}, "planner")
	if !errors.Is(err, ErrModuleLibraryValidation) {
		t.Fatalf("expected explicit condition validation, got %v", err)
	}
}
```

- [ ] **Step 2: Test ausführen und fehlende Typen bestätigen**

Run: `cd backend; go test ./internal/application -run TestModuleLibraryService`

Expected: FAIL mit `undefined: ModuleTemplateRevision`.

- [ ] **Step 3: Öffentliche Typen und Repository-Port definieren**

```go
var (
	ErrModuleLibraryValidation = errors.New("module library validation failed")
	ErrModuleTemplateNotFound  = errors.New("module template not found")
	ErrModuleRevisionNotFound  = errors.New("module template revision not found")
	ErrModuleInstanceNotFound  = errors.New("module instance not found")
	ErrModuleVersionConflict   = errors.New("module library version conflict")
	ErrModuleRevisionImmutable = errors.New("module template revision is immutable")
	ErrModuleTemplateInUse     = errors.New("module template is used by a layout")
)

type ModuleTemplateRevision struct {
	ID                 string                              `json:"id"`
	ModuleTemplateID   string                              `json:"moduleTemplateId"`
	RevisionNumber     int                                 `json:"revisionNumber"`
	BaseRevisionID     string                              `json:"baseRevisionId,omitempty"`
	Status             domain.ModuleTemplateRevisionStatus `json:"status"`
	Kind               domain.LayoutUnitKind               `json:"kind"`
	Gauge              string                              `json:"gauge"`
	WidthMM            float64                             `json:"widthMm"`
	HeightMM           float64                             `json:"heightMm"`
	SourceLayoutUnitID string                              `json:"sourceLayoutUnitId,omitempty"`
	Version            int                                 `json:"version"`
	ReleasedAt         string                              `json:"releasedAt,omitempty"`
	CreatedAt          string                              `json:"createdAt"`
	UpdatedAt          string                              `json:"updatedAt"`
}

type ModuleTemplate struct {
	ID                  string                         `json:"id"`
	Name                string                         `json:"name"`
	Description         string                         `json:"description,omitempty"`
	Version             int                            `json:"version"`
	Archived            bool                           `json:"archived"`
	LatestRevision      *ModuleTemplateRevision        `json:"latestRevision,omitempty"`
	InstanceCount       int                            `json:"instanceCount"`
	ReviewRequiredCount int                            `json:"reviewRequiredCount"`
	CreatedAt           string                         `json:"createdAt"`
	UpdatedAt           string                         `json:"updatedAt"`
}

type ModuleInstance struct {
	ID                       string                           `json:"id"`
	ModuleTemplateRevisionID string                           `json:"moduleTemplateRevisionId"`
	ModuleTemplateID         string                           `json:"moduleTemplateId"`
	TemplateName             string                           `json:"templateName"`
	RevisionNumber           int                              `json:"revisionNumber"`
	InventoryNumber          string                           `json:"inventoryNumber,omitempty"`
	OwnerLabel               string                           `json:"ownerLabel,omitempty"`
	Condition                domain.ModuleInstanceCondition   `json:"condition"`
	Availability             domain.ModuleInstanceAvailability `json:"availability"`
	ReviewStatus             domain.ModuleInventoryReviewStatus `json:"reviewStatus"`
	LocationLabel            string                           `json:"locationLabel,omitempty"`
	Notes                    string                           `json:"notes,omitempty"`
	Version                  int                              `json:"version"`
	Archived                 bool                             `json:"archived"`
	CreatedAt                string                           `json:"createdAt"`
	UpdatedAt                string                           `json:"updatedAt"`
}

type ModuleTemplatePage struct {
	Items      []ModuleTemplate `json:"items"`
	Page       int              `json:"page"`
	PageSize   int              `json:"pageSize"`
	TotalItems int              `json:"totalItems"`
	TotalPages int              `json:"totalPages"`
}

type ModuleTemplateRevisionPage struct {
	Items      []ModuleTemplateRevision `json:"items"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"pageSize"`
	TotalItems int                      `json:"totalItems"`
	TotalPages int                      `json:"totalPages"`
}

type ModuleInstancePage struct {
	Items      []ModuleInstance `json:"items"`
	Page       int              `json:"page"`
	PageSize   int              `json:"pageSize"`
	TotalItems int              `json:"totalItems"`
	TotalPages int              `json:"totalPages"`
}

type ModuleTemplateFilter struct {
	Query    string
	LayoutID string
	Gauge    string
	Archived *bool
	Page     int
	PageSize int
}

type ModuleTemplateRevisionFilter struct {
	ModuleTemplateID string
	Status           domain.ModuleTemplateRevisionStatus
	Page             int
	PageSize         int
}

type ModuleInstanceFilter struct {
	Query        string
	TemplateID   string
	OwnerLabel   string
	Condition    domain.ModuleInstanceCondition
	Availability domain.ModuleInstanceAvailability
	ReviewStatus domain.ModuleInventoryReviewStatus
	Archived     *bool
	Page         int
	PageSize     int
}

type CreateModuleTemplateRevisionInput struct {
	BaseRevisionID string                `json:"baseRevisionId"`
	Kind           domain.LayoutUnitKind `json:"kind"`
	Gauge          string                `json:"gauge"`
	WidthMM        float64               `json:"widthMm"`
	HeightMM       float64               `json:"heightMm"`
}

type UpdateModuleTemplateRevisionInput struct {
	Kind            domain.LayoutUnitKind `json:"kind"`
	Gauge           string                `json:"gauge"`
	WidthMM         float64               `json:"widthMm"`
	HeightMM        float64               `json:"heightMm"`
	ExpectedVersion int                   `json:"expectedVersion"`
}

type CreateModuleTemplateInput struct {
	Name            string                            `json:"name"`
	Description     string                            `json:"description"`
	InitialRevision CreateModuleTemplateRevisionInput `json:"initialRevision"`
}

type UpdateModuleTemplateInput struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	Archived        bool   `json:"archived"`
	ExpectedVersion int    `json:"expectedVersion"`
}

type ReleaseModuleTemplateRevisionInput struct {
	ExpectedVersion int `json:"expectedVersion"`
}

type CreateModuleInstanceInput struct {
	ModuleTemplateRevisionID string                            `json:"moduleTemplateRevisionId"`
	InventoryNumber          string                            `json:"inventoryNumber"`
	OwnerLabel               string                            `json:"ownerLabel"`
	Condition                domain.ModuleInstanceCondition   `json:"condition"`
	Availability             domain.ModuleInstanceAvailability `json:"availability"`
	LocationLabel            string                            `json:"locationLabel"`
	Notes                    string                            `json:"notes"`
	Archived                 bool                              `json:"archived"`
}

type UpdateModuleInstanceInput struct {
	CreateModuleInstanceInput
	ExpectedVersion int `json:"expectedVersion"`
}

type ReviewModuleInstanceInput struct {
	InventoryNumber string                            `json:"inventoryNumber"`
	OwnerLabel      string                            `json:"ownerLabel"`
	Condition       domain.ModuleInstanceCondition   `json:"condition"`
	Availability    domain.ModuleInstanceAvailability `json:"availability"`
	LocationLabel   string                            `json:"locationLabel"`
	Notes           string                            `json:"notes"`
	ExpectedVersion int                               `json:"expectedVersion"`
}

type ModuleLibraryRepository interface {
	ListTemplates(context.Context, ModuleTemplateFilter) (*ModuleTemplatePage, error)
	GetTemplate(context.Context, string) (*ModuleTemplate, error)
	CreateTemplate(context.Context, CreateModuleTemplateInput, string) (*ModuleTemplate, error)
	UpdateTemplate(context.Context, string, UpdateModuleTemplateInput, string) (*ModuleTemplate, error)
	ListRevisions(context.Context, ModuleTemplateRevisionFilter) (*ModuleTemplateRevisionPage, error)
	CreateRevision(context.Context, string, CreateModuleTemplateRevisionInput, string) (*ModuleTemplateRevision, error)
	UpdateRevision(context.Context, string, UpdateModuleTemplateRevisionInput, string) (*ModuleTemplateRevision, error)
	ReleaseRevision(context.Context, string, int, string) (*ModuleTemplateRevision, error)
	ListInstances(context.Context, ModuleInstanceFilter) (*ModuleInstancePage, error)
	GetInstance(context.Context, string) (*ModuleInstance, error)
	CreateInstance(context.Context, CreateModuleInstanceInput, string) (*ModuleInstance, error)
	UpdateInstance(context.Context, string, UpdateModuleInstanceInput, string) (*ModuleInstance, error)
	ReviewInstance(context.Context, string, ReviewModuleInstanceInput, string) (*ModuleInstance, error)
	LinkTemplateToLayout(context.Context, string, string, string) error
	UnlinkTemplateFromLayout(context.Context, string, string, string) error
}

type ModuleLibraryService struct {
	repository ModuleLibraryRepository
}

func NewModuleLibraryService(repository ModuleLibraryRepository) *ModuleLibraryService {
	return &ModuleLibraryService{repository: repository}
}
```

`ModuleLibraryService` hält genau diesen Repository-Port.

- [ ] **Step 4: Service mit festen Regeln implementieren**

```go
func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func validModuleRevisionInput(input CreateModuleTemplateRevisionInput) bool {
	return input.Kind.Valid() && strings.TrimSpace(input.Gauge) != "" &&
		finite(input.WidthMM) && input.WidthMM >= 0 &&
		finite(input.HeightMM) && input.HeightMM >= 0
}

func validModuleReviewInput(input ReviewModuleInstanceInput) bool {
	return strings.TrimSpace(input.InventoryNumber) != "" &&
		input.Condition.Valid() && input.Condition != domain.ModuleInstanceConditionUnknown &&
		input.Availability.Valid() &&
		input.Availability != domain.ModuleInstanceAvailabilityUnchecked &&
		input.ExpectedVersion > 0
}

func (s *ModuleLibraryService) ReviewInstance(
	ctx context.Context,
	id string,
	input ReviewModuleInstanceInput,
	actor string,
) (*ModuleInstance, error) {
	id = strings.TrimSpace(id)
	input.InventoryNumber = strings.TrimSpace(input.InventoryNumber)
	input.OwnerLabel = strings.TrimSpace(input.OwnerLabel)
	input.LocationLabel = strings.TrimSpace(input.LocationLabel)
	input.Notes = strings.TrimSpace(input.Notes)
	if id == "" || !validModuleReviewInput(input) {
		return nil, ErrModuleLibraryValidation
	}
	return s.repository.ReviewInstance(ctx, id, input, actor)
}

func (s *ModuleLibraryService) ListTemplates(
	ctx context.Context,
	filter ModuleTemplateFilter,
) (*ModuleTemplatePage, error) {
	filter.Query = strings.TrimSpace(filter.Query)
	filter.LayoutID = strings.TrimSpace(filter.LayoutID)
	filter.Gauge = strings.TrimSpace(filter.Gauge)
	filter.Page, filter.PageSize = normalizePage(filter.Page, filter.PageSize)
	return s.repository.ListTemplates(ctx, filter)
}

func (s *ModuleLibraryService) GetTemplate(ctx context.Context, id string) (*ModuleTemplate, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrModuleLibraryValidation
	}
	return s.repository.GetTemplate(ctx, id)
}

func (s *ModuleLibraryService) CreateTemplate(
	ctx context.Context,
	input CreateModuleTemplateInput,
	actor string,
) (*ModuleTemplate, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.InitialRevision.BaseRevisionID = ""
	input.InitialRevision.Gauge = strings.TrimSpace(input.InitialRevision.Gauge)
	if input.Name == "" || !validModuleRevisionInput(input.InitialRevision) {
		return nil, ErrModuleLibraryValidation
	}
	return s.repository.CreateTemplate(ctx, input, actor)
}

func (s *ModuleLibraryService) UpdateTemplate(
	ctx context.Context,
	id string,
	input UpdateModuleTemplateInput,
	actor string,
) (*ModuleTemplate, error) {
	id = strings.TrimSpace(id)
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if id == "" || input.Name == "" || input.ExpectedVersion < 1 {
		return nil, ErrModuleLibraryValidation
	}
	return s.repository.UpdateTemplate(ctx, id, input, actor)
}

func (s *ModuleLibraryService) ListRevisions(
	ctx context.Context,
	templateID string,
	filter ModuleTemplateRevisionFilter,
) (*ModuleTemplateRevisionPage, error) {
	filter.ModuleTemplateID = strings.TrimSpace(templateID)
	if filter.ModuleTemplateID == "" || (filter.Status != "" && !filter.Status.Valid()) {
		return nil, ErrModuleLibraryValidation
	}
	filter.Page, filter.PageSize = normalizePage(filter.Page, filter.PageSize)
	return s.repository.ListRevisions(ctx, filter)
}

func (s *ModuleLibraryService) CreateRevision(
	ctx context.Context,
	templateID string,
	input CreateModuleTemplateRevisionInput,
	actor string,
) (*ModuleTemplateRevision, error) {
	templateID = strings.TrimSpace(templateID)
	input.BaseRevisionID = strings.TrimSpace(input.BaseRevisionID)
	input.Gauge = strings.TrimSpace(input.Gauge)
	if templateID == "" || input.BaseRevisionID == "" || !validModuleRevisionInput(input) {
		return nil, ErrModuleLibraryValidation
	}
	return s.repository.CreateRevision(ctx, templateID, input, actor)
}

func (s *ModuleLibraryService) UpdateRevision(
	ctx context.Context,
	id string,
	input UpdateModuleTemplateRevisionInput,
	actor string,
) (*ModuleTemplateRevision, error) {
	id = strings.TrimSpace(id)
	input.Gauge = strings.TrimSpace(input.Gauge)
	if id == "" || input.ExpectedVersion < 1 || !validModuleRevisionInput(
		CreateModuleTemplateRevisionInput{
			Kind: input.Kind, Gauge: input.Gauge, WidthMM: input.WidthMM, HeightMM: input.HeightMM,
		},
	) {
		return nil, ErrModuleLibraryValidation
	}
	return s.repository.UpdateRevision(ctx, id, input, actor)
}

func (s *ModuleLibraryService) ReleaseRevision(
	ctx context.Context,
	id string,
	expectedVersion int,
	actor string,
) (*ModuleTemplateRevision, error) {
	id = strings.TrimSpace(id)
	if id == "" || expectedVersion < 1 {
		return nil, ErrModuleLibraryValidation
	}
	return s.repository.ReleaseRevision(ctx, id, expectedVersion, actor)
}

func cleanModuleInstanceInput(input CreateModuleInstanceInput) CreateModuleInstanceInput {
	input.ModuleTemplateRevisionID = strings.TrimSpace(input.ModuleTemplateRevisionID)
	input.InventoryNumber = strings.TrimSpace(input.InventoryNumber)
	input.OwnerLabel = strings.TrimSpace(input.OwnerLabel)
	input.LocationLabel = strings.TrimSpace(input.LocationLabel)
	input.Notes = strings.TrimSpace(input.Notes)
	if input.Condition == "" {
		input.Condition = domain.ModuleInstanceConditionUnknown
	}
	if input.Availability == "" {
		input.Availability = domain.ModuleInstanceAvailabilityUnchecked
	}
	return input
}

func validModuleInstanceInput(input CreateModuleInstanceInput) bool {
	return input.ModuleTemplateRevisionID != "" && input.Condition.Valid() && input.Availability.Valid()
}

func (s *ModuleLibraryService) ListInstances(
	ctx context.Context,
	filter ModuleInstanceFilter,
) (*ModuleInstancePage, error) {
	filter.Query = strings.TrimSpace(filter.Query)
	filter.TemplateID = strings.TrimSpace(filter.TemplateID)
	filter.OwnerLabel = strings.TrimSpace(filter.OwnerLabel)
	if filter.Condition != "" && !filter.Condition.Valid() ||
		filter.Availability != "" && !filter.Availability.Valid() ||
		filter.ReviewStatus != "" && !filter.ReviewStatus.Valid() {
		return nil, ErrModuleLibraryValidation
	}
	filter.Page, filter.PageSize = normalizePage(filter.Page, filter.PageSize)
	return s.repository.ListInstances(ctx, filter)
}

func (s *ModuleLibraryService) CreateInstance(
	ctx context.Context,
	input CreateModuleInstanceInput,
	actor string,
) (*ModuleInstance, error) {
	input = cleanModuleInstanceInput(input)
	if !validModuleInstanceInput(input) {
		return nil, ErrModuleLibraryValidation
	}
	return s.repository.CreateInstance(ctx, input, actor)
}

func (s *ModuleLibraryService) UpdateInstance(
	ctx context.Context,
	id string,
	input UpdateModuleInstanceInput,
	actor string,
) (*ModuleInstance, error) {
	id = strings.TrimSpace(id)
	input.CreateModuleInstanceInput = cleanModuleInstanceInput(input.CreateModuleInstanceInput)
	if id == "" || input.ExpectedVersion < 1 || !validModuleInstanceInput(input.CreateModuleInstanceInput) {
		return nil, ErrModuleLibraryValidation
	}
	return s.repository.UpdateInstance(ctx, id, input, actor)
}

func (s *ModuleLibraryService) LinkTemplateToLayout(
	ctx context.Context,
	layoutID string,
	templateID string,
	actor string,
) error {
	layoutID = strings.TrimSpace(layoutID)
	templateID = strings.TrimSpace(templateID)
	if layoutID == "" || templateID == "" {
		return ErrModuleLibraryValidation
	}
	return s.repository.LinkTemplateToLayout(ctx, layoutID, templateID, actor)
}

func (s *ModuleLibraryService) UnlinkTemplateFromLayout(
	ctx context.Context,
	layoutID string,
	templateID string,
	actor string,
) error {
	layoutID = strings.TrimSpace(layoutID)
	templateID = strings.TrimSpace(templateID)
	if layoutID == "" || templateID == "" {
		return ErrModuleLibraryValidation
	}
	return s.repository.UnlinkTemplateFromLayout(ctx, layoutID, templateID, actor)
}
```

`GetInstance` bleibt ausschließlich eine Repository-Hilfsmethode für Rückgaben nach Mutationen und
wird in Stufe 1 nicht als eigener HTTP-Endpunkt veröffentlicht.

- [ ] **Step 5: Service-Tests ausführen**

Run: `cd backend; go test ./internal/application -run TestModuleLibraryService`

Expected: PASS für Filter, Revisionseingaben, Exemplare, Bestandsprüfung und leere IDs.

- [ ] **Step 6: Commit erstellen**

```powershell
git add backend/internal/application/module_library.go backend/internal/application/module_library_test.go
git commit -m "feat(modules): add module library service"
```

### Task 4: SQLite-Repository, Pagination, Audit und Konflikte

**Files:**
- Create: `backend/internal/infrastructure/module_library_repository.go`
- Create: `backend/internal/infrastructure/module_library_repository_test.go`

**Interfaces:**
- Consumes: `application.ModuleLibraryRepository` und alle Typen aus Task 3.
- Produces: `NewModuleLibraryRepository(*sql.DB) *ModuleLibraryRepository`.

- [ ] **Step 1: Persistenzverhalten zuerst testen**

```go
func moduleLibraryTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestModuleLibraryRepositoryReleasesRevisionAndReviewsInstance(t *testing.T) {
	db := moduleLibraryTestDB(t)
	repository := infrastructure.NewModuleLibraryRepository(db)
	service := application.NewModuleLibraryService(repository)
	template, err := service.CreateTemplate(t.Context(), application.CreateModuleTemplateInput{
		Name: "Abzweig Süd",
		InitialRevision: application.CreateModuleTemplateRevisionInput{
			Kind: domain.LayoutUnitKindModule, Gauge: "TT", WidthMM: 2400, HeightMM: 850,
		},
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	revision, err := service.ReleaseRevision(t.Context(), template.LatestRevision.ID, 1, "planner")
	if err != nil {
		t.Fatal(err)
	}
	instance, err := service.CreateInstance(t.Context(), application.CreateModuleInstanceInput{
		ModuleTemplateRevisionID: revision.ID,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	reviewed, err := service.ReviewInstance(t.Context(), instance.ID,
		application.ReviewModuleInstanceInput{
			InventoryNumber: "M-001", OwnerLabel: "MEC", Condition: domain.ModuleInstanceConditionReady,
			Availability: domain.ModuleInstanceAvailabilityAvailable, ExpectedVersion: instance.Version,
		}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if reviewed.ReviewStatus != domain.ModuleInventoryReviewConfirmed || reviewed.InventoryNumber != "M-001" {
		t.Fatalf("unexpected reviewed instance: %#v", reviewed)
	}
	for _, action := range []string{
		"ModuleTemplateCreated", "ModuleTemplateRevisionReleased",
		"ModuleInstanceCreated", "ModuleInstanceInventoryConfirmed",
	} {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE action=?`, action).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("audit action %s count=%d", action, count)
		}
	}
}
```

Zusätzliche Tests erfassen exakt diese Fälle:

- Seite 1 mit `pageSize=2`, stabile Sortierung nach Name und ID, korrekte Gesamtzahlen;
- Suche über Vorlagenname und Exemplarkennung;
- Filter nach Anlage, Archivstatus, Spurweite, Eigentümer, Zustand, Verfügbarkeit und Prüfstatus;
- höchstens ein Entwurf je Vorlage;
- `expectedVersion` ergibt bei veraltetem Wert `ErrModuleVersionConflict`;
- Änderungen an `released` und `archived` ergeben `ErrModuleRevisionImmutable`;
- Bestandsbestätigung gegen eine Entwurfsrevision ergibt `ErrModuleLibraryValidation`;
- Wechsel eines Exemplars auf eine andere freigegebene Revision schreibt
  `ModuleInstanceRevisionChanged`;
- Entfernen einer Anlagenzuordnung mit referenzierender Konfiguration ergibt
  `ErrModuleTemplateInUse`.

- [ ] **Step 2: Repositorytests ausführen und fehlenden Konstruktor bestätigen**

Run: `cd backend; go test ./internal/infrastructure -run TestModuleLibraryRepository`

Expected: FAIL mit `undefined: infrastructure.NewModuleLibraryRepository`.

- [ ] **Step 3: Gemeinsame Seitenauswertung und stabile Filterabfrage implementieren**

```go
type ModuleLibraryRepository struct {
	db *sql.DB
}

func NewModuleLibraryRepository(db *sql.DB) *ModuleLibraryRepository {
	return &ModuleLibraryRepository{db: db}
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	return strings.ReplaceAll(value, `_`, `\_`)
}

func modulePageBounds(page, pageSize, total int) (offset, totalPages int) {
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	return (page - 1) * pageSize, totalPages
}

func templateFilterSQL(filter application.ModuleTemplateFilter) (string, []any) {
	clauses := []string{"1=1"}
	args := []any{}
	if filter.Query != "" {
		clauses = append(clauses, `(template.name LIKE ? ESCAPE '\' COLLATE NOCASE OR EXISTS (
			SELECT 1 FROM module_instances instance
			JOIN module_template_revisions instance_revision
			  ON instance_revision.id=instance.module_template_revision_id
			WHERE instance_revision.module_template_id=template.id
			  AND instance.inventory_number LIKE ? ESCAPE '\' COLLATE NOCASE
		))`)
		pattern := "%" + escapeLike(filter.Query) + "%"
		args = append(args, pattern, pattern)
	}
	if filter.LayoutID != "" {
		clauses = append(clauses, `EXISTS (
			SELECT 1 FROM layout_module_template_usages usage
			WHERE usage.module_template_id=template.id AND usage.layout_id=?
		)`)
		args = append(args, filter.LayoutID)
	}
	if filter.Archived != nil {
		clauses = append(clauses, "template.archived=?")
		args = append(args, boolToInt(*filter.Archived))
	}
	if filter.Gauge != "" {
		clauses = append(clauses, `EXISTS (
			SELECT 1 FROM module_template_revisions gauge_revision
			WHERE gauge_revision.module_template_id=template.id
			  AND gauge_revision.gauge=? COLLATE NOCASE
		)`)
		args = append(args, filter.Gauge)
	}
	return strings.Join(clauses, " AND "), args
}
```

`ListTemplates` verwendet dieselbe WHERE-Klausel für `COUNT(DISTINCT template.id)` und die
Datenseite. Die Datenseite sortiert mit
`ORDER BY template.name COLLATE NOCASE, template.id LIMIT ? OFFSET ?`. Die neueste Revision wird
pro Vorlage mit `ORDER BY (status='draft') DESC, revision_number DESC LIMIT 1` ermittelt. Leere
Ergebnisse werden als leere Arrays, nicht als `null`, zurückgegeben.

- [ ] **Step 4: Schreibvorgänge transaktional implementieren**

```go
func (r *ModuleLibraryRepository) ReviewInstance(
	ctx context.Context,
	id string,
	input application.ReviewModuleInstanceInput,
	actor string,
) (*application.ModuleInstance, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		var revisionStatus domain.ModuleTemplateRevisionStatus
		if err := tx.QueryRowContext(ctx, `
SELECT revision.status
FROM module_instances instance
JOIN module_template_revisions revision ON revision.id=instance.module_template_revision_id
WHERE instance.id=?`, id).Scan(&revisionStatus); errors.Is(err, sql.ErrNoRows) {
			return application.ErrModuleInstanceNotFound
		} else if err != nil {
			return fmt.Errorf("read module instance revision: %w", err)
		}
		if revisionStatus != domain.ModuleTemplateRevisionReleased {
			return application.ErrModuleLibraryValidation
		}
		result, err := tx.ExecContext(ctx, `
UPDATE module_instances
SET inventory_number=?, owner_label=?, condition_state=?, availability_status=?,
    review_status='confirmed', location_label=?, notes=?, version=version+1, updated_at=?
WHERE id=? AND version=?`, input.InventoryNumber, input.OwnerLabel, input.Condition,
			input.Availability, input.LocationLabel, input.Notes, now, id, input.ExpectedVersion)
		if err != nil {
			return mapModuleWriteError(err)
		}
		if err := requireModuleUpdated(ctx, tx, result, "module_instances", id,
			application.ErrModuleInstanceNotFound); err != nil {
			return err
		}
		return writeLayoutAudit(ctx, tx, "ModuleInstanceInventoryConfirmed",
			"module_instance", id, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return r.GetInstance(ctx, id)
}

func (r *ModuleLibraryRepository) withTx(ctx context.Context, work func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin module library transaction: %w", err)
	}
	if err := work(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit module library transaction: %w", err)
	}
	return nil
}

func mapModuleWriteError(err error) error {
	message := err.Error()
	if strings.Contains(message, "module_revision_immutable") {
		return application.ErrModuleRevisionImmutable
	}
	if strings.Contains(message, "module_revision_invalid_base") ||
		strings.Contains(message, "module_instance_revision_not_released") {
		return application.ErrModuleLibraryValidation
	}
	if strings.Contains(message, "ux_module_instances_inventory_number") ||
		strings.Contains(message, "UNIQUE constraint failed: module_instances.inventory_number") ||
		strings.Contains(message, "ux_module_template_revisions_draft") ||
		strings.Contains(message, "CHECK constraint failed: module_instances") {
		return application.ErrModuleLibraryValidation
	}
	return err
}

func requireModuleUpdated(
	ctx context.Context,
	tx *sql.Tx,
	result sql.Result,
	table string,
	id string,
	notFound error,
) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read module update result: %w", err)
	}
	if affected == 1 {
		return nil
	}
	queries := map[string]string{
		"module_templates":          `SELECT COUNT(*) FROM module_templates WHERE id=?`,
		"module_template_revisions": `SELECT COUNT(*) FROM module_template_revisions WHERE id=?`,
		"module_instances":          `SELECT COUNT(*) FROM module_instances WHERE id=?`,
	}
	var count int
	if err := tx.QueryRowContext(ctx, queries[table], id).Scan(&count); err != nil {
		return fmt.Errorf("check module resource: %w", err)
	}
	if count == 0 {
		return notFound
	}
	return application.ErrModuleVersionConflict
}
```

Die übrigen Mutationen verwenden dieselbe Transaktionsgrenze und diese Auditaktionen:
`ModuleTemplateCreated`, `ModuleTemplateUpdated`, `ModuleTemplateRevisionCreated`,
`ModuleTemplateRevisionUpdated`, `ModuleTemplateRevisionReleased`, `ModuleInstanceCreated`,
`ModuleInstanceUpdated`, `ModuleInstanceRevisionChanged`, `LayoutModuleTemplateLinked` und
`LayoutModuleTemplateUnlinked`. `ReleaseRevision` aktualisiert ausschließlich einen Datensatz mit
`status='draft' AND version=?`, setzt `status='released'`, `released_at`, erhöht die Version und
belässt ältere Revisionen unverändert.

Die Mutationen verwenden exakt diese SQL-Grundoperationen. IDs entstehen jeweils einmal über
`randomID()`, Zeitstempel einmal über `timestamp()` und Audit wird vor demselben Commit geschrieben:

```sql
-- CreateTemplate
INSERT INTO module_templates(
  id, name, description, version, archived, created_at, updated_at
) VALUES(?, ?, ?, 1, 0, ?, ?);
INSERT INTO module_template_revisions(
  id, module_template_id, revision_number, base_revision_id, status, kind, gauge,
  width_mm, height_mm, source_layout_unit_id, version, released_at, created_at, updated_at
) VALUES(?, ?, 1, NULL, 'draft', ?, ?, ?, ?, NULL, 1, NULL, ?, ?);

-- UpdateTemplate
UPDATE module_templates
SET name=?, description=?, archived=?, version=version+1, updated_at=?
WHERE id=? AND version=?;

-- CreateRevision, nachdem Vorlage und optionale Basisrevision geprüft wurden
INSERT INTO module_template_revisions(
  id, module_template_id, revision_number, base_revision_id, status, kind, gauge,
  width_mm, height_mm, source_layout_unit_id, version, released_at, created_at, updated_at
)
SELECT ?, template.id, COALESCE(MAX(revision.revision_number), 0)+1, NULLIF(?, ''),
       'draft', ?, ?, ?, ?, NULL, 1, NULL, ?, ?
FROM module_templates template
LEFT JOIN module_template_revisions revision ON revision.module_template_id=template.id
WHERE template.id=?
GROUP BY template.id;

-- UpdateRevision
UPDATE module_template_revisions
SET kind=?, gauge=?, width_mm=?, height_mm=?, version=version+1, updated_at=?
WHERE id=? AND status='draft' AND version=?;

-- ReleaseRevision
UPDATE module_template_revisions
SET status='released', released_at=?, version=version+1, updated_at=?
WHERE id=? AND status='draft' AND version=?;

-- CreateInstance, immer ungeprüft
INSERT INTO module_instances(
  id, module_template_revision_id, inventory_number, owner_label, condition_state,
  availability_status, review_status, location_label, notes, version, archived,
  created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, 'review_required', ?, ?, 1, ?, ?, ?);

-- UpdateInstance; Revisionswechsel setzt die Bestandsprüfung zurück
UPDATE module_instances
SET module_template_revision_id=?, inventory_number=?, owner_label=?, condition_state=?,
    availability_status=?,
    review_status=CASE WHEN module_template_revision_id<>? THEN 'review_required' ELSE review_status END,
    location_label=?, notes=?, archived=?, version=version+1, updated_at=?
WHERE id=? AND version=?;

-- LinkTemplateToLayout
INSERT INTO layout_module_template_usages(layout_id, module_template_id, origin, created_at)
SELECT layout.id, template.id, 'planned', ?
FROM layouts layout CROSS JOIN module_templates template
WHERE layout.id=? AND template.id=?
ON CONFLICT(layout_id, module_template_id) DO NOTHING;

-- UnlinkTemplateFromLayout nach der Belegungsprüfung
DELETE FROM layout_module_template_usages WHERE layout_id=? AND module_template_id=?;
```

Vor `CreateRevision` muss eine nicht leere `baseRevisionId` zu derselben Vorlage und zu einer
freigegebenen oder archivierten Revision gehören. Eine Entwurfsrevision ist nie selbst Basis eines
weiteren Entwurfs.
`RowsAffected()==0` wird je Operation in Nicht-gefunden, Versionskonflikt, unveränderliche Revision
oder bereits vorhandene Zuordnung aufgelöst. Vor Unlink prüft das Repository folgende Belegung:

```sql
SELECT EXISTS(
  SELECT 1
  FROM layout_configurations configuration
  JOIN layout_configuration_units placement
    ON placement.configuration_id=configuration.id
  JOIN module_template_revisions revision
    ON revision.id=placement.module_template_revision_id
       OR revision.source_layout_unit_id=placement.unit_id
  WHERE configuration.layout_id=? AND revision.module_template_id=?
);
```

- [ ] **Step 5: Repositorytests und Race-unabhängige Konflikte prüfen**

Run: `cd backend; go test ./internal/infrastructure -run TestModuleLibraryRepository -count=1`

Expected: PASS.

- [ ] **Step 6: Commit erstellen**

```powershell
git add backend/internal/infrastructure/module_library_repository.go `
  backend/internal/infrastructure/module_library_repository_test.go
git commit -m "feat(modules): persist module library records"
```

### Task 5: Bestehende Anlagen ohne Datenverlust an die Bibliothek anbinden

**Files:**
- Modify: `backend/internal/application/layouts.go:45-96,137-166`
- Modify: `backend/internal/infrastructure/layout_repository.go:118-333,397-573,586-704`
- Modify: `backend/internal/infrastructure/layout_outline_repository.go:1-105`
- Modify: `backend/internal/infrastructure/layout_position_repository.go:1-180`
- Modify: `backend/internal/infrastructure/layout_twin_repository.go:21-116`
- Modify: `backend/internal/infrastructure/track_planner_repository.go:55-70,698-714`
- Modify: `backend/internal/infrastructure/layout_repository_test.go:1-520`
- Modify: `backend/internal/infrastructure/layout_twin_repository_test.go:1-100`
- Modify: `backend/internal/infrastructure/track_planner_repository_test.go:1-270`
- Modify: `backend/internal/infrastructure/track_planner_schema_test.go:139-155`
- Modify: `backend/internal/infrastructure/accessory_product_deletion_test.go:122-137`
- Modify: `backend/internal/infrastructure/layout_accessory_schema_test.go:79-96,153-168`

**Interfaces:**
- Consumes: neue Tabellen aus Task 2.
- Produces: projektweit wiederverwendbare bestehende Module und kanonische Revisionsreferenzen in
  Anlagenkonfigurationen.

- [ ] **Step 1: Wiederverwendung derselben Revision in zwei Anlagen testen**

```go
func TestLayoutRepositoryListsLinkedTemplateAcrossLayouts(t *testing.T) {
	db := moduleLibraryTestDB(t)
	layouts := application.NewLayoutService(infrastructure.NewLayoutRepository(db))
	first, err := layouts.CreateLayout(t.Context(), application.CreateLayoutInput{
		Name: "Anlage A", Kind: domain.LayoutKindClub, Gauge: "TT", Scale: "1:120",
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	second, err := layouts.CreateLayout(t.Context(), application.CreateLayoutInput{
		Name: "Anlage B", Kind: domain.LayoutKindClub, Gauge: "TT", Scale: "1:120",
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := layouts.CreateUnit(t.Context(), first.ID, application.CreateLayoutUnitInput{
		Name: "Bahnhof", Kind: domain.LayoutUnitKindModule, WidthMM: 1200, HeightMM: 500,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	library := application.NewModuleLibraryService(infrastructure.NewModuleLibraryRepository(db))
	if err := library.LinkTemplateToLayout(t.Context(), second.ID, unit.ModuleTemplateID, "planner"); err != nil {
		t.Fatal(err)
	}
	secondUnits, err := layouts.ListUnits(t.Context(), second.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(secondUnits) != 1 || secondUnits[0].ID != unit.ID ||
		secondUnits[0].ModuleTemplateRevisionID != unit.ModuleTemplateRevisionID {
		t.Fatalf("linked unit missing: %#v", secondUnits)
	}
}
```

- [ ] **Step 2: Test ausführen und aktuelle Anlagenbindung bestätigen**

Run: `cd backend; go test ./internal/infrastructure -run TestLayoutRepositoryListsLinkedTemplateAcrossLayouts`

Expected: FAIL, weil `ListUnits` derzeit ausschließlich `layout_units.layout_id` abfragt.

- [ ] **Step 3: Legacy-Antworten um kanonische IDs erweitern**

```go
type LayoutUnit struct {
	ID                       string                              `json:"id"`
	LayoutID                 string                              `json:"layoutId"`
	ModuleTemplateID         string                              `json:"moduleTemplateId"`
	ModuleTemplateRevisionID string                              `json:"moduleTemplateRevisionId"`
	RevisionNumber           int                                 `json:"revisionNumber"`
	RevisionStatus           domain.ModuleTemplateRevisionStatus `json:"revisionStatus"`
	Name                     string                              `json:"name"`
	Kind                     domain.LayoutUnitKind               `json:"kind"`
	OwnerLabel               string                              `json:"ownerLabel,omitempty"`
	WidthMM                  float64                             `json:"widthMm"`
	HeightMM                 float64                             `json:"heightMm"`
	Version                  int                                 `json:"version"`
	Archived                 bool                                `json:"archived"`
	CreatedAt                string                              `json:"createdAt"`
	UpdatedAt                string                              `json:"updatedAt"`
}

type ConfigurationUnit struct {
	UnitID                   string  `json:"unitId"`
	ModuleTemplateRevisionID string  `json:"moduleTemplateRevisionId"`
	PlanRevisionID           string  `json:"planRevisionId,omitempty"`
	PositionXMM              float64 `json:"positionXMm"`
	PositionYMM              float64 `json:"positionYMm"`
	RotationDegrees          float64 `json:"rotationDegrees"`
	SortOrder                int     `json:"sortOrder"`
}
```

- [ ] **Step 4: Listen, Erstellen, Ändern und Konfigurationen synchronisieren**

`ListUnits(layoutID)` verwendet folgende kanonische Verbindung und setzt das ausgegebene
`layoutId` auf den angefragten Kontext:

```sql
SELECT unit.id, usage.layout_id, template.id, revision.id, revision.revision_number,
       revision.status, template.name, revision.kind, instance.owner_label,
       revision.width_mm, revision.height_mm, unit.version, template.archived,
       template.created_at, template.updated_at
FROM layout_module_template_usages usage
JOIN module_templates template ON template.id=usage.module_template_id
JOIN module_template_revisions revision ON revision.module_template_id=template.id
JOIN layout_units unit ON unit.id=revision.source_layout_unit_id
LEFT JOIN module_instances instance ON instance.id=unit.id
WHERE usage.layout_id=?
  AND revision.id=(
    SELECT candidate.id FROM module_template_revisions candidate
    WHERE candidate.module_template_id=template.id AND candidate.source_layout_unit_id IS NOT NULL
    ORDER BY (candidate.status='draft') DESC, candidate.revision_number DESC LIMIT 1
  )
ORDER BY template.name COLLATE NOCASE, template.id;
```

`CreateUnit` schreibt in einer Transaktion und in dieser Reihenfolge:

1. `module_templates`, ID gleich neuer `layout_units.id`;
2. `layout_units`, unveränderte Legacy-Felder;
3. `module_template_revisions`, ID gleich Unit-ID, Revision 1, Status `draft`;
4. `module_instances`, ID gleich Unit-ID, Status `review_required`;
5. `layout_module_template_usages`, Herkunft `planned`;
6. Audit `LayoutUnitCreated` und `ModuleTemplateCreated`.

`UpdateUnit` darf nur die über `source_layout_unit_id` verbundene Entwurfsrevision verändern. Es
synchronisiert Vorlagenname, Revisionsart und Maße sowie Eigentümer des migrierten Exemplars in
derselben Transaktion. Ein freigegebener Datensatz ergibt `ErrModuleRevisionImmutable`.

Jeder Schreibfehler aus `layout_units`, `layout_unit_ports`, `layout_unit_outline_points` und
`layout_technical_positions` läuft durch `mapModuleWriteError`. Damit ergeben die Trigger aus Task
2 einen fachlichen Fehler und nie einen unklassifizierten HTTP-500-Fehler.

Beim Speichern einer Konfiguration wird die Revisions-ID serverseitig aus `unit_id` ermittelt und
zusätzlich gespeichert:

```sql
INSERT INTO layout_configuration_units(
  configuration_id, unit_id, module_template_revision_id, plan_revision_id,
  position_x_mm, position_y_mm, rotation_degrees, sort_order
)
SELECT ?, ?, revision.id, NULLIF(?, ''), ?, ?, ?, ?
FROM module_template_revisions revision
JOIN layout_module_template_usages usage ON usage.module_template_id=revision.module_template_id
WHERE revision.source_layout_unit_id=? AND usage.layout_id=?;
```

Nach dem Insert muss `RowsAffected()==1` gelten. `listConfigurationUnits` liest beide IDs. Die
Validierung prüft die Anlagenzuordnung über `layout_module_template_usages` und nicht mehr über
`layout_units.layout_id`.

- [ ] **Step 5: Digitalen Zwilling und Gleisplanabfragen auf projektweite Verwendung umstellen**

In `layout_twin_repository.go` ersetzen die Existenz- und Auswahlabfragen die Bedingung
`unit.layout_id=?` durch den Join aus Step 4. In `track_planner_repository.go` wird die Spurweite so
gelesen:

```sql
JOIN module_template_revisions module_revision
  ON module_revision.source_layout_unit_id=unit.id
```

und `module_revision.gauge` ersetzt `layout.gauge`. Damit ändert eine zweite Anlagenverwendung
weder Spurweite noch technische Revision.

- [ ] **Step 6: Kompatibilitätsregressionen ausführen**

Alle Test-Fixtures, die `layout_units` direkt per SQL anlegen, ergänzen unmittelbar danach die
passende Vorlage, Revision, Instanz und Anlagenzuordnung. Dabei bleibt die jeweilige Unit-ID auch
Template-, Revisions- und Instanz-ID. Die Revision erhält `draft`, die Instanz
`review_required`, und `layout_module_template_usages.origin` erhält `planned`. Das betrifft
`track_planner_schema_test.go`, `accessory_product_deletion_test.go`, beide Fixtures in
`layout_accessory_schema_test.go` sowie die zwei direkten Layout-Unit-Fixtures in
`backup_test.go`. So erzeugen weder neue noch alte Tests einen Zustand, den die Anwendung selbst
nicht mehr erzeugen kann.

Run: `cd backend; go test ./internal/infrastructure -run 'TestLayoutRepository|TestLayoutTwin|TestTrackPlanner'`

Expected: PASS. Bestehende Konfigurationen, Port-Fanglogik, Konturen und Gleispläne behalten ihre
IDs und Werte.

- [ ] **Step 7: Commit erstellen**

```powershell
git add backend/internal/application/layouts.go `
  backend/internal/infrastructure/layout_repository.go `
  backend/internal/infrastructure/layout_outline_repository.go `
  backend/internal/infrastructure/layout_position_repository.go `
  backend/internal/infrastructure/layout_twin_repository.go `
  backend/internal/infrastructure/track_planner_repository.go `
  backend/internal/infrastructure/layout_repository_test.go `
  backend/internal/infrastructure/layout_twin_repository_test.go `
  backend/internal/infrastructure/track_planner_repository_test.go `
  backend/internal/infrastructure/track_planner_schema_test.go `
  backend/internal/infrastructure/accessory_product_deletion_test.go `
  backend/internal/infrastructure/layout_accessory_schema_test.go
git commit -m "feat(modules): reuse module revisions across layouts"
```

### Task 6: HTTP-Endpunkte, Rollen, CSRF und Fehlercodes

**Files:**
- Create: `backend/internal/api/module_library_handlers.go`
- Create: `backend/internal/api/module_library_handlers_test.go`
- Modify: `backend/internal/api/router.go:20-92`
- Modify: `backend/internal/api/layout_handlers.go:171-205`
- Modify: `backend/internal/api/routes.go:190-220`
- Modify: `backend/internal/api/routes_security_test.go:15-103`
- Modify: `backend/cmd/railkeeper/startup.go:220-268`

**Interfaces:**
- Consumes: `application.ModuleLibraryService`.
- Produces: JSON-Endpunkte unter `/api/v1/module-*` und Anlagenzuordnungen.

- [ ] **Step 1: End-to-End-Handlerverhalten testen**

```go
func TestModuleLibraryHandlersEnforcePaginationReviewAndRoles(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		application.CreateUserInput{Username: "planner", Password: "planner-password", Roles: []string{"Planner"}},
		application.CreateUserInput{Username: "viewer", Password: "viewer-password", Roles: []string{"Viewer"}},
		application.CreateUserInput{Username: "messe", Password: "messe-password", Roles: []string{"Messe"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	library := application.NewModuleLibraryService(infrastructure.NewModuleLibraryRepository(db))
	router := NewRouter(Config{AuthService: auth, ModuleLibraryService: library})
	planner := loginRouteTestUser(t, auth, "planner", "planner-password")
	viewer := loginRouteTestUser(t, auth, "viewer", "viewer-password")
	messe := loginRouteTestUser(t, auth, "messe", "messe-password")

	createdResponse := layoutRequest(t, router, planner, http.MethodPost,
		"/api/v1/module-templates", map[string]any{
			"name": "Abzweig Süd",
			"initialRevision": map[string]any{
				"kind": "module", "gauge": "TT", "widthMm": 2400, "heightMm": 850,
			},
		}, true)
	assertStatus(t, createdResponse, http.StatusCreated)
	var created application.ModuleTemplate
	decodeResponse(t, createdResponse, &created)
	pageResponse := layoutRequest(t, router, viewer, http.MethodGet,
		"/api/v1/module-templates?q=Abzweig&page=1&pageSize=25", nil, true)
	assertStatus(t, pageResponse, http.StatusOK)
	var page application.ModuleTemplatePage
	decodeResponse(t, pageResponse, &page)
	if page.TotalItems != 1 || page.Items[0].ID != created.ID {
		t.Fatalf("unexpected template page: %#v", page)
	}
	response := layoutRequest(t, router, messe, http.MethodGet, "/api/v1/module-templates", nil, true)
	if response.Code != http.StatusForbidden {
		t.Fatalf("Messe status=%d", response.Code)
	}
}
```

- [ ] **Step 2: Test ausführen und nicht registrierte Route bestätigen**

Run: `cd backend; go test ./internal/api -run TestModuleLibraryHandlersEnforcePaginationReviewAndRoles`

Expected: FAIL mit HTTP 404.

- [ ] **Step 3: Service verdrahten und Routen registrieren**

```go
{http.MethodGet, "/api/v1/module-templates", routeAccessViewer, (*App).listModuleTemplates, nil},
{http.MethodPost, "/api/v1/module-templates", routeAccessPlanner, (*App).createModuleTemplate, nil},
{http.MethodGet, "/api/v1/module-templates/{id}", routeAccessViewer, (*App).getModuleTemplate, nil},
{http.MethodPut, "/api/v1/module-templates/{id}", routeAccessPlanner, (*App).updateModuleTemplate, nil},
{http.MethodGet, "/api/v1/module-templates/{id}/revisions", routeAccessViewer, (*App).listModuleTemplateRevisions, nil},
{http.MethodPost, "/api/v1/module-templates/{id}/revisions", routeAccessPlanner, (*App).createModuleTemplateRevision, nil},
{http.MethodPut, "/api/v1/module-template-revisions/{id}", routeAccessPlanner, (*App).updateModuleTemplateRevision, nil},
{http.MethodPost, "/api/v1/module-template-revisions/{id}/release", routeAccessPlanner, (*App).releaseModuleTemplateRevision, nil},
{http.MethodGet, "/api/v1/module-instances", routeAccessViewer, (*App).listModuleInstances, nil},
{http.MethodPost, "/api/v1/module-instances", routeAccessPlanner, (*App).createModuleInstance, nil},
{http.MethodPut, "/api/v1/module-instances/{id}", routeAccessPlanner, (*App).updateModuleInstance, nil},
{http.MethodPost, "/api/v1/module-instances/{id}/inventory-review", routeAccessPlanner, (*App).reviewModuleInstance, nil},
{http.MethodPost, "/api/v1/layouts/{id}/module-templates/{templateID}", routeAccessPlanner, (*App).linkModuleTemplateToLayout, nil},
{http.MethodDelete, "/api/v1/layouts/{id}/module-templates/{templateID}", routeAccessPlanner, (*App).unlinkModuleTemplateFromLayout, nil},
```

`Config` und `App` erhalten `ModuleLibraryService *application.ModuleLibraryService`.
`startup.go` erstellt Repository und Service aus derselben Datenbank wie `LayoutService`.

- [ ] **Step 4: Query-Parsing und fachliche Fehlerabbildung implementieren**

```go
func modulePageQuery(r *http.Request) (int, int, error) {
	page, err := optionalPositiveQueryInt(r, "page")
	if err != nil {
		return 0, 0, application.ErrModuleLibraryValidation
	}
	pageSize, err := optionalPositiveQueryInt(r, "pageSize")
	if err != nil || pageSize > 100 {
		return 0, 0, application.ErrModuleLibraryValidation
	}
	return page, pageSize, nil
}

func (a *App) moduleLibraryError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, application.ErrModuleLibraryValidation):
		respondProblem(w, http.StatusBadRequest, "module_library_validation", "Module library data is invalid.")
	case errors.Is(err, application.ErrModuleTemplateNotFound),
		errors.Is(err, application.ErrModuleRevisionNotFound),
		errors.Is(err, application.ErrModuleInstanceNotFound):
		respondProblem(w, http.StatusNotFound, "module_library_not_found", "Module library resource not found.")
	case errors.Is(err, application.ErrModuleVersionConflict):
		respondProblem(w, http.StatusConflict, "module_version_conflict", "Module library data has changed.")
	case errors.Is(err, application.ErrModuleRevisionImmutable):
		respondProblem(w, http.StatusConflict, "module_revision_immutable", "Released module revisions are immutable.")
	case errors.Is(err, application.ErrModuleTemplateInUse):
		respondProblem(w, http.StatusConflict, "module_template_in_use", "The module template is used by this layout.")
	default:
		a.logger.Error("module library operation failed", "action", action, "error", err)
		respondProblem(w, http.StatusInternalServerError, "module_library_operation_failed", "Module library operation failed.")
	}
}
```

`layoutError` erhält für die weiterhin unterstützten Unit-, Port-, Kontur- und Positionsrouten
dieselbe Abbildung:

```go
case errors.Is(err, application.ErrModuleRevisionImmutable):
	respondProblem(w, http.StatusConflict, "module_revision_immutable", "Released module revisions are immutable.")
```

Ungültige `page`, `pageSize`, Bool- oder Enum-Querywerte ergeben 400. Gültige leere Seiten ergeben
200 mit `items: []`. POST liefert 201, PUT und Aktions-POST liefern 200, erfolgreiches Unlink 204.

- [ ] **Step 5: API- und Sicherheitsmatrix ausführen**

Run: `cd backend; go test ./internal/api -run 'TestModuleLibrary|TestProtectedRoutesRejectUnauthorizedAndInsufficientRoles|TestAPIRoutesDeclareAccess'`

Expected: PASS. Viewer-Schreibzugriffe und jeder Messe-Bibliothekszugriff ergeben 403; POST ohne
CSRF ergibt 403.

- [ ] **Step 6: Commit erstellen**

```powershell
git add backend/internal/api/module_library_handlers.go `
  backend/internal/api/module_library_handlers_test.go backend/internal/api/router.go `
  backend/internal/api/routes.go backend/internal/api/routes_security_test.go `
  backend/internal/api/layout_handlers.go `
  backend/cmd/railkeeper/startup.go
git commit -m "feat(modules): expose secured module library API"
```

### Task 7: OpenAPI als verbindlichen öffentlichen Vertrag ergänzen

**Files:**
- Modify: `openapi/railkeeper.yaml:3636-3768,8660-8890`
- Modify: `backend/internal/api/openapi_contract_test.go:170-205`

**Interfaces:**
- Consumes: JSON-Namen, Routen und Statuscodes aus Tasks 3 und 6.
- Produces: dokumentierten OpenAPI-Vertrag ohne Frontend- oder Backenddrift.

- [ ] **Step 1: Vertragsprüfung für alle neuen Schemas und Operationen ergänzen**

```go
func TestOpenAPIDocumentsProjectWideModuleLibrary(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	for _, schema := range []string{
		"ModuleTemplate", "ModuleTemplateInput", "UpdateModuleTemplateInput",
		"ModuleTemplatePage", "ModuleTemplateRevision", "ModuleTemplateRevisionInput",
		"UpdateModuleTemplateRevisionInput", "ModuleTemplateRevisionPage",
		"ReleaseModuleTemplateRevisionInput", "ModuleInstance", "ModuleInstanceInput",
		"UpdateModuleInstanceInput", "ModuleInstancePage", "ReviewModuleInstanceInput",
	} {
		if !strings.Contains(contract, "    "+schema+":\n") {
			t.Errorf("OpenAPI contract is missing schema %s", schema)
		}
	}
	for _, path := range []string{
		"  /module-templates:\n", "  /module-templates/{id}:\n",
		"  /module-templates/{id}/revisions:\n", "  /module-template-revisions/{id}:\n",
		"  /module-template-revisions/{id}/release:\n", "  /module-instances:\n",
		"  /module-instances/{id}:\n", "  /module-instances/{id}/inventory-review:\n",
		"  /layouts/{id}/module-templates/{templateID}:\n",
	} {
		if !strings.Contains(contract, path) {
			t.Errorf("OpenAPI contract is missing path %s", strings.TrimSpace(path))
		}
	}
}
```

- [ ] **Step 2: Vertragstest ausführen und fehlende Schemas bestätigen**

Run: `cd backend; go test ./internal/api -run TestOpenAPIDocumentsProjectWideModuleLibrary`

Expected: FAIL mit Meldungen für alle neuen Schemas.

- [ ] **Step 3: Schemas mit geschlossenen Enums und Pflichtfeldern dokumentieren**

```yaml
    ModuleTemplateRevision:
      type: object
      required:
        [id, moduleTemplateId, revisionNumber, status, kind, gauge, widthMm, heightMm,
         version, createdAt, updatedAt]
      properties:
        id: { type: string }
        moduleTemplateId: { type: string }
        revisionNumber: { type: integer, minimum: 1 }
        baseRevisionId: { type: string }
        status: { type: string, enum: [draft, released, archived] }
        kind: { type: string, enum: [baseboard, module, segment, area] }
        gauge: { type: string }
        widthMm: { type: number, minimum: 0 }
        heightMm: { type: number, minimum: 0 }
        sourceLayoutUnitId: { type: string }
        version: { type: integer, minimum: 1 }
        releasedAt: { type: string, format: date-time }
        createdAt: { type: string, format: date-time }
        updatedAt: { type: string, format: date-time }
    ModuleInstance:
      type: object
      required:
        [id, moduleTemplateRevisionId, moduleTemplateId, templateName, revisionNumber,
         condition, availability, reviewStatus, version, archived, createdAt, updatedAt]
      properties:
        id: { type: string }
        moduleTemplateRevisionId: { type: string }
        moduleTemplateId: { type: string }
        templateName: { type: string }
        revisionNumber: { type: integer, minimum: 1 }
        inventoryNumber: { type: string }
        ownerLabel: { type: string }
        condition:
          type: string
          enum: [unknown, ready, maintenance_due, defective]
        availability:
          type: string
          enum: [unchecked, available, unavailable]
        reviewStatus:
          type: string
          enum: [review_required, confirmed]
        locationLabel: { type: string }
        notes: { type: string }
        version: { type: integer, minimum: 1 }
        archived: { type: boolean }
        createdAt: { type: string, format: date-time }
        updatedAt: { type: string, format: date-time }
```

`ModuleTemplate`, die drei Seitentypen und sämtliche Inputs verwenden dieselben Feldnamen und
Pflichtregeln wie Task 3. Listenparameter dokumentieren `q`, `layoutId`, `templateId`, `gauge`,
`condition`, `availability`, `reviewStatus`, `archived`, `page` und `pageSize`; `pageSize` hat
`minimum: 1` und `maximum: 100`.

Die restlichen Komponenten verwenden diese exakte Pflichtfeldmatrix:

| Schema | Pflichtfelder |
|---|---|
| `ModuleTemplate` | `id`, `name`, `version`, `archived`, `instanceCount`, `reviewRequiredCount`, `createdAt`, `updatedAt` |
| `ModuleTemplateInput` | `name`, `initialRevision` |
| `UpdateModuleTemplateInput` | `name`, `archived`, `expectedVersion` |
| `ModuleTemplateRevisionInput` | `kind`, `gauge`, `widthMm`, `heightMm` |
| `UpdateModuleTemplateRevisionInput` | `kind`, `gauge`, `widthMm`, `heightMm`, `expectedVersion` |
| `ReleaseModuleTemplateRevisionInput` | `expectedVersion` |
| `ModuleInstanceInput` | `moduleTemplateRevisionId` |
| `UpdateModuleInstanceInput` | `moduleTemplateRevisionId`, `condition`, `availability`, `archived`, `expectedVersion` |
| `ReviewModuleInstanceInput` | `inventoryNumber`, `condition`, `availability`, `expectedVersion` |
| alle drei Seitentypen | `items`, `page`, `pageSize`, `totalItems`, `totalPages` |

`description`, `baseRevisionId`, `inventoryNumber`, `ownerLabel`, `locationLabel`, `notes`,
`sourceLayoutUnitId`, `releasedAt` und `latestRevision` bleiben optional. `archived` ist bei Create
optional und hat den Standardwert `false`.

- [ ] **Step 4: Jede Route mit Rollenbeschreibung und Fehlerantworten dokumentieren**

Jede GET-Operation nennt explizit „Viewer, Editor, Planner, or Admin role required. Messe is
excluded.“. Jede Mutation nennt „Planner or Admin role required.“. Mutationen mit
`expectedVersion` dokumentieren 409; Unlink dokumentiert zusätzlich den belegten Anlagenbezug.
Alle Request- und Response-Schemas referenzieren Komponenten, keine freien Inline-Objekte.

| Pfad und Methode | Erfolg | Antwortschema |
|---|---:|---|
| `GET /module-templates` | 200 | `ModuleTemplatePage` |
| `POST /module-templates` | 201 | `ModuleTemplate` |
| `GET /module-templates/{id}` | 200 | `ModuleTemplate` |
| `PUT /module-templates/{id}` | 200 | `ModuleTemplate` |
| `GET /module-templates/{id}/revisions` | 200 | `ModuleTemplateRevisionPage` |
| `POST /module-templates/{id}/revisions` | 201 | `ModuleTemplateRevision` |
| `PUT /module-template-revisions/{id}` | 200 | `ModuleTemplateRevision` |
| `POST /module-template-revisions/{id}/release` | 200 | `ModuleTemplateRevision` |
| `GET /module-instances` | 200 | `ModuleInstancePage` |
| `POST /module-instances` | 201 | `ModuleInstance` |
| `PUT /module-instances/{id}` | 200 | `ModuleInstance` |
| `POST /module-instances/{id}/inventory-review` | 200 | `ModuleInstance` |
| `POST /layouts/{id}/module-templates/{templateID}` | 204 | kein Body |
| `DELETE /layouts/{id}/module-templates/{templateID}` | 204 | kein Body |

- [ ] **Step 5: Gesamten OpenAPI-Abgleich ausführen**

Run: `cd backend; go test ./internal/api -run OpenAPI`

Expected: PASS, einschließlich der Prüfung, dass jede Frontendoperation dokumentiert ist.

- [ ] **Step 6: Commit erstellen**

```powershell
git add openapi/railkeeper.yaml backend/internal/api/openapi_contract_test.go
git commit -m "docs(api): document module library contracts"
```

### Task 8: Backup-Version 21 und Restore alter Bestände

**Files:**
- Create: `backend/internal/application/backup_module_library.go`
- Modify: `backend/internal/application/backup.go:23-205,326-410,678-875`
- Modify: `backend/internal/application/backup_test.go:591-657,1937-2128`

**Interfaces:**
- Consumes: Tabellen aus Task 2.
- Produces: Backup-Version 21, reproduzierbaren Roundtrip und deterministischen Backfill für
  Versionen 1 bis 20.

- [ ] **Step 1: Aktuellen Roundtrip und Version-20-Restore testen**

```go
func TestBackupVersionTwentyOnePreservesModuleLibraryAndLegacyTwentyBackfillsIt(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	service := application.NewBackupService(db, dataDir)
	layoutService := application.NewLayoutService(infrastructure.NewLayoutRepository(db))
	layout, err := layoutService.CreateLayout(t.Context(), application.CreateLayoutInput{
		Name: "Vereinsanlage", Kind: domain.LayoutKindClub, Gauge: "TT", Scale: "1:120",
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := layoutService.CreateUnit(t.Context(), layout.ID, application.CreateLayoutUnitInput{
		Name: "Bahnhof Süd", Kind: domain.LayoutUnitKindModule, OwnerLabel: "MEC",
		WidthMM: 1200, HeightMM: 500,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if doc.Version != 21 || len(doc.Tables["module_templates"]) != 1 ||
		len(doc.Tables["module_template_revisions"]) != 1 || len(doc.Tables["module_instances"]) != 1 {
		t.Fatalf("module library missing from version 21 backup: %#v", doc.Tables)
	}
	encoded, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	legacy := &application.BackupDocument{}
	if err := json.Unmarshal(encoded, legacy); err != nil {
		t.Fatal(err)
	}
	legacy.Version = 20
	delete(legacy.Tables, "module_templates")
	delete(legacy.Tables, "module_template_revisions")
	delete(legacy.Tables, "module_instances")
	delete(legacy.Tables, "layout_module_template_usages")
	for _, row := range legacy.Tables["layout_configuration_units"] {
		delete(row, "module_template_revision_id")
	}
	if _, err := service.Import(t.Context(), legacy); err != nil {
		t.Fatal(err)
	}
	var status, review string
	if err := db.QueryRow(`SELECT revision.status, instance.review_status
	 FROM module_template_revisions revision
	 JOIN module_instances instance ON instance.module_template_revision_id=revision.id
	 LIMIT 1`).Scan(&status, &review); err != nil {
		t.Fatal(err)
	}
	if status != "draft" || review != "review_required" {
		t.Fatalf("legacy module identity=%q/%q", status, review)
	}
	var restoredRevisionID string
	if err := db.QueryRow(`SELECT id
	 FROM module_template_revisions WHERE source_layout_unit_id=?`, unit.ID).Scan(&restoredRevisionID); err != nil {
		t.Fatal(err)
	}
	if restoredRevisionID != unit.ID {
		t.Fatalf("restored revision id=%q, want %q", restoredRevisionID, unit.ID)
	}
}
```

- [ ] **Step 2: Test ausführen und Versionsabweichung bestätigen**

Run: `cd backend; go test ./internal/application -run TestBackupVersionTwentyOnePreservesModuleLibraryAndLegacyTwentyBackfillsIt`

Expected: FAIL, weil Export noch Version 20 schreibt.

- [ ] **Step 3: Backup-Reihenfolge und Versionsrichtlinie erweitern**

```go
const (
	backupFormat  = "railkeeper-backup"
	backupVersion = 21
)
```

Direkt nach `layout_units` werden in `backupTableOrder` diese Tabellen eingefügt:

```go
"module_templates",
"module_template_revisions",
"module_instances",
"layout_module_template_usages",
```

Jede erhält in `backupTableVersions` den Wert
`backupTableVersionPolicy{introduced: 21, required: 21}`. Diese Reihenfolge stellt sicher, dass
Layouts und Legacy-Units vor den neuen Fremdschlüsseln importiert werden.

Alle bestehenden Assertions, die den jeweils aktuellen Export prüfen, wechseln in
`backup_test.go` von 20 auf 21. Tests, die bewusst ein historisches Dokument der Version 20
konstruieren, behalten den Wert 20. Das gilt konkret für die aktuellen Exportprüfungen an den
bisherigen Zeilen 76, 159, 380, 530, 608, 691, 769, 882, 938, 1001 und 1285.

- [ ] **Step 4: Legacy-Backfill nach dem Tabellenimport implementieren**

```go
func backfillRestoredModuleLibrary(ctx context.Context, tx *sql.Tx, version int) error {
	if version >= 21 {
		return nil
	}
	statements := []string{
		`INSERT INTO module_templates(id,name,description,version,archived,created_at,updated_at)
		 SELECT id,name,'',1,archived,created_at,updated_at FROM layout_units`,
		`INSERT INTO module_template_revisions(
		   id,module_template_id,revision_number,base_revision_id,status,kind,gauge,width_mm,height_mm,
		   source_layout_unit_id,version,released_at,created_at,updated_at)
		 SELECT unit.id,unit.id,1,NULL,'draft',unit.kind,layout.gauge,COALESCE(unit.width_mm,0),
		   COALESCE(unit.height_mm,0),unit.id,1,NULL,unit.created_at,unit.updated_at
		 FROM layout_units unit JOIN layouts layout ON layout.id=unit.layout_id`,
		`INSERT INTO module_instances(
		   id,module_template_revision_id,inventory_number,owner_label,condition_state,
		   availability_status,review_status,location_label,notes,version,archived,created_at,updated_at)
		 SELECT id,id,'',owner_label,'unknown','unchecked','review_required','','',1,archived,
		   created_at,updated_at FROM layout_units`,
		`INSERT INTO layout_module_template_usages(layout_id,module_template_id,origin,created_at)
		 SELECT layout_id,id,'migrated',created_at FROM layout_units`,
		`UPDATE layout_configuration_units SET module_template_revision_id=unit_id
		 WHERE module_template_revision_id IS NULL`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("backfill restored module library: %w", err)
		}
	}
	return nil
}
```

`Import` ruft diese Funktion nach den regulären Tabellenzeilen und vor `tx.Commit()` auf. `Validate`
ruft für Version 21 zusätzlich `validateBackupModuleLibrary` auf. Diese Prüfung meldet vor dem
Löschen lokaler Daten doppelte Vorlagenrevisionen, fehlende Referenzen, mehr als einen Entwurf je
Vorlage, doppelte nicht leere Kennungen und bestätigte Exemplare ohne freigegebene Revision oder
explizite Bestandswerte.

Die Vorprüfung arbeitet ausschließlich auf dem Backup-Dokument und verändert weder Datenbank noch
Dateisystem. Sie verwendet die bestehende Funktion `backupNonEmptyString` und wird vollständig in
`backup_module_library.go` implementiert:

```go
type backupModuleRevision struct {
	templateID string
	status     string
	number     float64
}

func backupUniqueIDs(table string, rows []map[string]any) (map[string]struct{}, []string) {
	ids := map[string]struct{}{}
	errorsFound := []string{}
	for index, row := range rows {
		id, valid := backupNonEmptyString(row["id"])
		if !valid {
			errorsFound = append(errorsFound,
				fmt.Sprintf("Modulbibliothek: %s Zeile %d hat keine gültige ID.", table, index+1))
			continue
		}
		if _, duplicate := ids[id]; duplicate {
			errorsFound = append(errorsFound,
				fmt.Sprintf("Modulbibliothek: %s enthält die ID %q mehrfach.", table, id))
		}
		ids[id] = struct{}{}
	}
	return ids, errorsFound
}

func validateBackupModuleLibrary(doc *BackupDocument) []string {
	if doc.Version < 21 {
		return nil
	}
	errorsFound := []string{}
	templateIDs, found := backupUniqueIDs("module_templates", doc.Tables["module_templates"])
	errorsFound = append(errorsFound, found...)
	layoutIDs, found := backupUniqueIDs("layouts", doc.Tables["layouts"])
	errorsFound = append(errorsFound, found...)
	unitIDs, found := backupUniqueIDs("layout_units", doc.Tables["layout_units"])
	errorsFound = append(errorsFound, found...)

	revisions := map[string]backupModuleRevision{}
	revisionNumbers := map[string]struct{}{}
	sourceUnits := map[string]struct{}{}
	drafts := map[string]int{}
	for index, row := range doc.Tables["module_template_revisions"] {
		prefix := fmt.Sprintf("Modulbibliothek: Vorlagenrevision Zeile %d", index+1)
		id, idValid := backupNonEmptyString(row["id"])
		templateID, templateValid := backupNonEmptyString(row["module_template_id"])
		status, statusValid := backupNonEmptyString(row["status"])
		number, numberValid := backupNumberValue(row["revision_number"])
		if !idValid || !templateValid || !statusValid || !numberValid || number < 1 {
			errorsFound = append(errorsFound, prefix+" ist unvollständig oder ungültig.")
			continue
		}
		if _, duplicate := revisions[id]; duplicate {
			errorsFound = append(errorsFound, prefix+" verwendet eine doppelte ID.")
		}
		if _, exists := templateIDs[templateID]; !exists {
			errorsFound = append(errorsFound, prefix+" verweist auf keine vorhandene Vorlage.")
		}
		if status != "draft" && status != "released" && status != "archived" {
			errorsFound = append(errorsFound, prefix+" hat einen ungültigen Status.")
		}
		_, hasReleasedAt := backupNonEmptyString(row["released_at"])
		if (status == "draft" && hasReleasedAt) || (status != "draft" && !hasReleasedAt) {
			errorsFound = append(errorsFound, prefix+" hat einen widersprüchlichen Freigabezeitpunkt.")
		}
		numberKey := fmt.Sprintf("%s:%v", templateID, number)
		if _, duplicate := revisionNumbers[numberKey]; duplicate {
			errorsFound = append(errorsFound, prefix+" verwendet eine doppelte Revisionsnummer.")
		}
		revisionNumbers[numberKey] = struct{}{}
		if status == "draft" {
			drafts[templateID]++
		}
		revisions[id] = backupModuleRevision{templateID: templateID, status: status, number: number}
	}
	for templateID, count := range drafts {
		if count > 1 {
			errorsFound = append(errorsFound,
				fmt.Sprintf("Modulbibliothek: Vorlage %q besitzt mehr als einen Entwurf.", templateID))
		}
	}
	for index, row := range doc.Tables["module_template_revisions"] {
		id, _ := backupNonEmptyString(row["id"])
		revision := revisions[id]
		if baseID, present := backupNonEmptyString(row["base_revision_id"]); present {
			base, exists := revisions[baseID]
			if !exists || base.templateID != revision.templateID || base.number >= revision.number ||
				(base.status != "released" && base.status != "archived") {
				errorsFound = append(errorsFound, fmt.Sprintf(
					"Modulbibliothek: Vorlagenrevision Zeile %d hat keine gültige Basisrevision.", index+1))
			}
		}
		if sourceID, present := backupNonEmptyString(row["source_layout_unit_id"]); present {
			if _, exists := unitIDs[sourceID]; !exists {
				errorsFound = append(errorsFound, fmt.Sprintf(
					"Modulbibliothek: Vorlagenrevision Zeile %d verweist auf kein vorhandenes Anlagenmodul.", index+1))
			}
			if _, duplicate := sourceUnits[sourceID]; duplicate {
				errorsFound = append(errorsFound, fmt.Sprintf(
					"Modulbibliothek: Anlagenmodul %q ist mehreren Revisionen zugeordnet.", sourceID))
			}
			sourceUnits[sourceID] = struct{}{}
		}
	}

	inventoryNumbers := map[string]struct{}{}
	_, found = backupUniqueIDs("module_instances", doc.Tables["module_instances"])
	errorsFound = append(errorsFound, found...)
	for index, row := range doc.Tables["module_instances"] {
		prefix := fmt.Sprintf("Modulbibliothek: Modulexemplar Zeile %d", index+1)
		revisionID, valid := backupNonEmptyString(row["module_template_revision_id"])
		revision, exists := revisions[revisionID]
		if !valid || !exists {
			errorsFound = append(errorsFound, prefix+" verweist auf keine vorhandene Revision.")
		}
		if inventoryNumber, present := backupNonEmptyString(row["inventory_number"]); present {
			key := strings.ToLower(inventoryNumber)
			if _, duplicate := inventoryNumbers[key]; duplicate {
				errorsFound = append(errorsFound, prefix+" verwendet eine doppelte Kennung.")
			}
			inventoryNumbers[key] = struct{}{}
		}
		reviewStatus, reviewValid := backupNonEmptyString(row["review_status"])
		condition, conditionValid := backupNonEmptyString(row["condition_state"])
		availability, availabilityValid := backupNonEmptyString(row["availability_status"])
		validCondition := condition == "unknown" || condition == "ready" ||
			condition == "maintenance_due" || condition == "defective"
		validAvailability := availability == "unchecked" || availability == "available" ||
			availability == "unavailable"
		if !reviewValid || (reviewStatus != "review_required" && reviewStatus != "confirmed") ||
			!conditionValid || !validCondition || !availabilityValid || !validAvailability {
			errorsFound = append(errorsFound, prefix+" hat ungültige Bestandswerte.")
			continue
		}
		_, hasInventoryNumber := backupNonEmptyString(row["inventory_number"])
		if reviewStatus == "confirmed" &&
			((revision.status != "released" && revision.status != "archived") || !hasInventoryNumber ||
			condition == "unknown" || availability == "unchecked") {
			errorsFound = append(errorsFound,
				prefix+" ist ohne freigegebene Revision und vollständige Bestandswerte bestätigt.")
		}
	}

	usages := map[string]struct{}{}
	for index, row := range doc.Tables["layout_module_template_usages"] {
		layoutID, layoutValid := backupNonEmptyString(row["layout_id"])
		templateID, templateValid := backupNonEmptyString(row["module_template_id"])
		origin, originValid := backupNonEmptyString(row["origin"])
		if !layoutValid || !templateValid || !originValid ||
			(origin != "migrated" && origin != "planned") {
			errorsFound = append(errorsFound, fmt.Sprintf(
				"Modulbibliothek: Anlagenzuordnung Zeile %d ist unvollständig.", index+1))
			continue
		}
		if _, exists := layoutIDs[layoutID]; !exists {
			errorsFound = append(errorsFound, fmt.Sprintf(
				"Modulbibliothek: Anlagenzuordnung Zeile %d verweist auf keine vorhandene Anlage.", index+1))
		}
		if _, exists := templateIDs[templateID]; !exists {
			errorsFound = append(errorsFound, fmt.Sprintf(
				"Modulbibliothek: Anlagenzuordnung Zeile %d verweist auf keine vorhandene Vorlage.", index+1))
		}
		key := layoutID + "\x00" + templateID
		if _, duplicate := usages[key]; duplicate {
			errorsFound = append(errorsFound, fmt.Sprintf(
				"Modulbibliothek: Anlagenzuordnung Zeile %d ist doppelt.", index+1))
		}
		usages[key] = struct{}{}
	}

	configurationLayouts := map[string]string{}
	for _, row := range doc.Tables["layout_configurations"] {
		id, idValid := backupNonEmptyString(row["id"])
		layoutID, layoutValid := backupNonEmptyString(row["layout_id"])
		if idValid && layoutValid {
			configurationLayouts[id] = layoutID
		}
	}
	for index, row := range doc.Tables["layout_configuration_units"] {
		configurationID, configurationValid := backupNonEmptyString(row["configuration_id"])
		revisionID, revisionValid := backupNonEmptyString(row["module_template_revision_id"])
		revision, revisionExists := revisions[revisionID]
		layoutID, configurationExists := configurationLayouts[configurationID]
		if !configurationValid || !revisionValid || !revisionExists || !configurationExists {
			errorsFound = append(errorsFound, fmt.Sprintf(
				"Modulbibliothek: Konfigurationsmodul Zeile %d hat keinen gültigen Revisionsbezug.", index+1))
			continue
		}
		if _, linked := usages[layoutID+"\x00"+revision.templateID]; !linked {
			errorsFound = append(errorsFound, fmt.Sprintf(
				"Modulbibliothek: Konfigurationsmodul Zeile %d verwendet eine nicht zugeordnete Vorlage.", index+1))
		}
	}
	return errorsFound
}
```

`Validate` hängt das Ergebnis genau einmal nach der Tabellenschleife an:

```go
result.Errors = append(result.Errors, validateBackupModuleLibrary(doc)...)
```

- [ ] **Step 5: Backup-Suite ausführen**

Run: `cd backend; go test ./internal/application -run Backup -count=1`

Expected: PASS für Version 21 sowie alle bisherigen Legacy-Versionen. Ein absichtlich beschädigter
Version-21-Vertrag wird als inkompatibel gemeldet, bevor Importdaten gelöscht werden.

- [ ] **Step 6: Commit erstellen**

```powershell
git add backend/internal/application/backup.go `
  backend/internal/application/backup_module_library.go `
  backend/internal/application/backup_test.go
git commit -m "feat(backup): preserve module library identities"
```

### Task 9: TypeScript-Vertrag und API-Client

**Files:**
- Modify: `frontend/src/shared/apiLayoutsAccessories.ts:1-240,1120-1185`
- Modify: `frontend/src/shared/apiLayoutsAccessories.test.ts:12-130`

**Interfaces:**
- Consumes: OpenAPI-Namen und HTTP-Routen aus Tasks 6 und 7.
- Produces: strikt typisierte API-Methoden für die Bibliotheksoberfläche.

- [ ] **Step 1: Exakte Requests im bestehenden Clienttest ergänzen**

```ts
await api.moduleTemplates({ query: "Bahnhof", layoutId: "layout/1", archived: false, page: 1, pageSize: 25 });
await api.createModuleTemplate({
  name: "Abzweig Süd",
  initialRevision: { kind: "module", gauge: "TT", widthMm: 2400, heightMm: 850 }
});
await api.moduleTemplateRevisions("template/1", { page: 1, pageSize: 25 });
await api.releaseModuleTemplateRevision("revision/1", { expectedVersion: 2 });
await api.moduleInstances({ templateId: "template/1", reviewStatus: "review_required", page: 1, pageSize: 25 });
await api.reviewModuleInstance("instance/1", {
  inventoryNumber: "M-001", ownerLabel: "MEC", condition: "ready",
  availability: "available", locationLabel: "Lager", notes: "Geprüft", expectedVersion: 1
});
await api.linkModuleTemplateToLayout("layout/1", "template/1");
await api.unlinkModuleTemplateFromLayout("layout/1", "template/1");
```

Der erwartete Requestvektor enthält exakt URL-kodierte IDs, Queryparameter in stabiler
Einfügereihenfolge, JSON-Body und CSRF-Header bei jeder Mutation.

- [ ] **Step 2: Clienttest ausführen und fehlende Methoden bestätigen**

Run: `cd frontend; npm.cmd run test:run -- src/shared/apiLayoutsAccessories.test.ts`

Expected: FAIL mit `Property 'moduleTemplates' does not exist`.

- [ ] **Step 3: Strikte Typen ergänzen**

```ts
export type ModuleTemplateRevisionStatus = "draft" | "released" | "archived";
export type ModuleInstanceCondition = "unknown" | "ready" | "maintenance_due" | "defective";
export type ModuleInstanceAvailability = "unchecked" | "available" | "unavailable";
export type ModuleInventoryReviewStatus = "review_required" | "confirmed";

export type ModuleTemplateRevision = {
  id: string;
  moduleTemplateId: string;
  revisionNumber: number;
  baseRevisionId?: string;
  status: ModuleTemplateRevisionStatus;
  kind: LayoutUnitKind;
  gauge: string;
  widthMm: number;
  heightMm: number;
  sourceLayoutUnitId?: string;
  version: number;
  releasedAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type ModuleTemplate = {
  id: string;
  name: string;
  description?: string;
  version: number;
  archived: boolean;
  latestRevision?: ModuleTemplateRevision;
  instanceCount: number;
  reviewRequiredCount: number;
  createdAt: string;
  updatedAt: string;
};

export type ModuleInstance = {
  id: string;
  moduleTemplateRevisionId: string;
  moduleTemplateId: string;
  templateName: string;
  revisionNumber: number;
  inventoryNumber?: string;
  ownerLabel?: string;
  condition: ModuleInstanceCondition;
  availability: ModuleInstanceAvailability;
  reviewStatus: ModuleInventoryReviewStatus;
  locationLabel?: string;
  notes?: string;
  version: number;
  archived: boolean;
  createdAt: string;
  updatedAt: string;
};

export type Page<T> = {
  items: T[];
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
};

export type CreateModuleTemplateRevisionInput = {
  baseRevisionId?: string;
  kind: LayoutUnitKind;
  gauge: string;
  widthMm: number;
  heightMm: number;
};

export type CreateModuleTemplateInput = {
  name: string;
  description?: string;
  initialRevision: CreateModuleTemplateRevisionInput;
};

export type UpdateModuleTemplateInput = {
  name: string;
  description?: string;
  archived: boolean;
  expectedVersion: number;
};

export type UpdateModuleTemplateRevisionInput = Omit<
  CreateModuleTemplateRevisionInput,
  "baseRevisionId"
> & { expectedVersion: number };

export type CreateModuleInstanceInput = {
  moduleTemplateRevisionId: string;
  inventoryNumber?: string;
  ownerLabel?: string;
  condition?: ModuleInstanceCondition;
  availability?: ModuleInstanceAvailability;
  locationLabel?: string;
  notes?: string;
  archived?: boolean;
};

export type UpdateModuleInstanceInput = Required<Pick<CreateModuleInstanceInput,
  "moduleTemplateRevisionId" | "condition" | "availability" | "archived"
>> & Omit<CreateModuleInstanceInput,
  "moduleTemplateRevisionId" | "condition" | "availability" | "archived"
> & { expectedVersion: number };

export type ReviewModuleInstanceInput = {
  inventoryNumber: string;
  ownerLabel?: string;
  condition: ModuleInstanceCondition;
  availability: ModuleInstanceAvailability;
  locationLabel?: string;
  notes?: string;
  expectedVersion: number;
};
```

Die Inputtypen spiegeln die Go-Inputs vollständig. `ReleaseModuleTemplateRevisionInput` ist
`{ expectedVersion: number }`. Sie verwenden keine `any`-Werte und machen `expectedVersion` bei
Änderungen verpflichtend.

- [ ] **Step 4: Querybuilder und API-Methoden ergänzen**

```ts
const modulePageQuery = (filter: {
  query?: string;
  layoutId?: string;
  templateId?: string;
  gauge?: string;
  condition?: ModuleInstanceCondition;
  availability?: ModuleInstanceAvailability;
  reviewStatus?: ModuleInventoryReviewStatus;
  archived?: boolean;
  page: number;
  pageSize: number;
}) => {
  const query = new URLSearchParams();
  if (filter.query?.trim()) query.set("q", filter.query.trim());
  if (filter.layoutId) query.set("layoutId", filter.layoutId);
  if (filter.templateId) query.set("templateId", filter.templateId);
  if (filter.gauge) query.set("gauge", filter.gauge);
  if (filter.condition) query.set("condition", filter.condition);
  if (filter.availability) query.set("availability", filter.availability);
  if (filter.reviewStatus) query.set("reviewStatus", filter.reviewStatus);
  if (filter.archived !== undefined) query.set("archived", String(filter.archived));
  query.set("page", String(filter.page));
  query.set("pageSize", String(filter.pageSize));
  return query.toString();
};
```

Methoden verwenden `request<Page<ModuleTemplate>>`, `request<Page<ModuleTemplateRevision>>` und
`request<Page<ModuleInstance>>`. Mutationen verwenden den bestehenden `json()`-Helper. Unlink
sendet `{ method: "DELETE" }` ohne Body.

Der Export ergänzt exakt diese Methoden: `moduleTemplates`, `moduleTemplate`,
`createModuleTemplate`, `updateModuleTemplate`, `moduleTemplateRevisions`,
`createModuleTemplateRevision`, `updateModuleTemplateRevision`,
`releaseModuleTemplateRevision`, `moduleInstances`, `createModuleInstance`,
`updateModuleInstance`, `reviewModuleInstance`, `linkModuleTemplateToLayout` und
`unlinkModuleTemplateFromLayout`. Namen, Rückgabetypen und Pfade entsprechen eins zu eins der
Matrix in Task 7.

- [ ] **Step 5: Clienttest und TypeScript-Build ausführen**

Run: `cd frontend; npm.cmd run test:run -- src/shared/apiLayoutsAccessories.test.ts; npm.cmd run build`

Expected: PASS.

- [ ] **Step 6: Commit erstellen**

```powershell
git add frontend/src/shared/apiLayoutsAccessories.ts frontend/src/shared/apiLayoutsAccessories.test.ts
git commit -m "feat(modules): add typed module library client"
```

### Task 10: Eigenständige und anlagenbezogene Modulbibliothek

**Files:**
- Create: `frontend/src/features/layouts/ModuleLibraryView.tsx`
- Create: `frontend/src/features/layouts/modules/ModuleLibraryPanel.tsx`
- Create: `frontend/src/features/layouts/modules/ModuleLibraryPanel.test.tsx`
- Create: `frontend/src/features/layouts/modules/ModuleTemplateDialog.tsx`
- Create: `frontend/src/features/layouts/modules/ModuleInstanceDialog.tsx`
- Create: `frontend/src/styles/module-library.css`
- Modify: `frontend/src/app/App.tsx:8-62,190-202`
- Modify: `frontend/src/app/navigationAvailability.ts:3-13`
- Modify: `frontend/src/features/layouts/LayoutsView.tsx:1-115`
- Modify: `frontend/src/features/layouts/LayoutWorkspace.tsx:1-170`
- Modify: `frontend/src/features/layouts/LayoutModulesPanel.tsx:1-89`
- Modify: `frontend/src/app/styles.css:20-23`

**Interfaces:**
- Consumes: TypeScript-API aus Task 9.
- Produces: Route `/modules` und dieselbe Bibliothekskomponente mit optionalem `layoutId`.

- [ ] **Step 1: Sichtbare Struktur und Kontextfilter testen**

```tsx
it("shows the same paginated library globally and in a layout context", async () => {
  const user = userEvent.setup();
  vi.spyOn(api, "moduleTemplates").mockResolvedValue({
    items: [templateFixture], page: 1, pageSize: 25, totalItems: 1, totalPages: 1
  });
  vi.spyOn(api, "moduleTemplateRevisions").mockResolvedValue({
    items: [revisionFixture], page: 1, pageSize: 25, totalItems: 1, totalPages: 1
  });
  vi.spyOn(api, "moduleInstances").mockResolvedValue({
    items: [instanceFixture], page: 1, pageSize: 25, totalItems: 1, totalPages: 1
  });
  render(<ModuleLibraryPanel layoutId="layout-1" canPlan />);
  expect(await screen.findByRole("heading", { name: "Modulbibliothek" })).toBeInTheDocument();
  expect(api.moduleTemplates).toHaveBeenCalledWith(expect.objectContaining({ layoutId: "layout-1" }));
  await user.click(screen.getByRole("button", { name: "Alle Module" }));
  await waitFor(() => expect(api.moduleTemplates).toHaveBeenLastCalledWith(
    expect.objectContaining({ layoutId: undefined })
  ));
  await user.click(screen.getByRole("button", { name: "Bahnhof Süd" }));
  expect(await screen.findByText("Revision 1")).toBeInTheDocument();
  expect(screen.getByText("Bestand prüfen")).toBeInTheDocument();
});
```

Weitere Tests prüfen leere, ladende und fehlerhafte Seiten, nächste/vorherige Seite, Suche mit
verzögerungsfreiem Absenden per Formular, Viewer ohne Schreibaktionen und das Verknüpfen einer
globalen Vorlage mit der aktiven Anlage.

- [ ] **Step 2: Referenzdateien vor der sichtbaren Änderung erneut prüfen**

Run:

```powershell
Get-FileHash -Algorithm SHA256 'C:\Users\droth\.codex\generated_images\01a07612-659f-7632-8706-cb2b937377fa\exec-ffa8ae91-5053-4c73-b629-83a701194a95.png'
Get-FileHash -Algorithm SHA256 'C:\Users\droth\.codex\generated_images\01a07612-659f-7632-8706-cb2b937377fa\exec-be12f866-ee67-43f9-820c-f7c5131e29ec.png'
```

Expected: exakt die beiden Werte aus den Global Constraints. Bei Abweichung endet die Ausführung
dieses Tasks ohne Frontendänderung und fordert die Originaldatei erneut an.

- [ ] **Step 3: Test ausführen und fehlende Komponente bestätigen**

Run: `cd frontend; npm.cmd run test:run -- src/features/layouts/modules/ModuleLibraryPanel.test.tsx`

Expected: FAIL mit `Cannot find module './ModuleLibraryPanel'`.

- [ ] **Step 4: Gemeinsame Bibliothekskomponente implementieren**

```tsx
export function ModuleLibraryPanel({ layoutId, canPlan }: { layoutId?: string; canPlan: boolean }) {
  const [scope, setScope] = useState<"layout" | "all">(layoutId ? "layout" : "all");
  const [query, setQuery] = useState("");
  const [page, setPage] = useState(1);
  const [selectedId, setSelectedId] = useState("");
  const [reloadToken, setReloadToken] = useState(0);
  const effectiveLayoutId = scope === "layout" ? layoutId : undefined;
  const { t } = useI18n();

  return <section className="module-library" aria-labelledby="module-library-title">
    <header className="module-library-head">
      <div>
        <p className="eyebrow">{t("modules.eyebrow")}</p>
        <h2 id="module-library-title">{t("modules.title")}</h2>
      </div>
      {canPlan ? <button type="button" className="primary-button">
        <Plus size={16} aria-hidden="true" />{t("modules.createTemplate")}
      </button> : null}
    </header>
    {layoutId ? <div className="module-library-scope" role="group" aria-label={t("modules.scope.label")}>
      <button type="button" aria-pressed={scope === "layout"} onClick={() => { setScope("layout"); setPage(1); }}>
        {t("modules.scope.layout")}
      </button>
      <button type="button" aria-pressed={scope === "all"} onClick={() => { setScope("all"); setPage(1); }}>
        {t("modules.scope.all")}
      </button>
    </div> : null}
    <ModuleTemplateTable
      layoutId={effectiveLayoutId}
      query={query}
      page={page}
      reloadToken={reloadToken}
      selectedId={selectedId}
      onQueryChange={setQuery}
      onPageChange={setPage}
      onSelect={setSelectedId}
    />
    <ModuleTemplateDetails templateId={selectedId} canPlan={canPlan} layoutId={layoutId}
      onChanged={() => setReloadToken((current) => current + 1)} />
  </section>;
}
```

`ModuleTemplateTable` und `ModuleTemplateDetails` bleiben lokale, fokussierte Komponenten in
derselben Datei. Die Tabelle zeigt Bezeichnung, Revision, Spurweite, Maße, Exemplare,
Prüfhinweise und Status. Der Detailbereich zeigt Revisionen und Exemplare in getrennten
Abschnitten. Status wird immer mit Text und Symbol dargestellt, nie nur durch Farbe.

Die lokalen Komponenten erhalten diese vollständigen Verträge, damit Laden und Auswählen nicht
mit Dialogzuständen vermischt werden:

```tsx
type ModuleTemplateTableProps = {
  layoutId?: string;
  query: string;
  page: number;
  reloadToken: number;
  selectedId: string;
  onQueryChange: (query: string) => void;
  onPageChange: (page: number) => void;
  onSelect: (templateId: string) => void;
};

type ModuleTemplateDetailsProps = {
  templateId: string;
  layoutId?: string;
  canPlan: boolean;
  onChanged: () => void;
};

type ModuleTemplateDialogProps = {
  mode: "create_template" | "create_revision" | "edit_template";
  template?: ModuleTemplate;
  baseRevision?: ModuleTemplateRevision;
  onSaved: (result: ModuleTemplate | ModuleTemplateRevision) => void;
  onClose: () => void;
};

type ModuleInstanceDialogProps = {
  template: ModuleTemplate;
  revisions: ModuleTemplateRevision[];
  instance?: ModuleInstance;
  onSaved: (instance: ModuleInstance) => void;
  onClose: () => void;
};
```

`ModuleTemplateDialog` zeigt im Modus `create_template` die Felder Bezeichnung, Beschreibung,
Art, Spurweite, Breite und Tiefe. Im Modus `create_revision` bleiben Bezeichnung und Beschreibung
nur lesbar; `baseRevisionId` wird aus der gewählten Basisrevision übernommen. Im Modus
`edit_template` sind nur Bezeichnung, Beschreibung und Archivstatus änderbar. Der Dialog sendet
je nach Modus exakt `createModuleTemplate`, `createModuleTemplateRevision` oder
`updateModuleTemplate` und behält Eingaben bei jedem API-Fehler.

`ModuleInstanceDialog` zeigt Revision, Kennung, Eigentümer, Zustand, Verfügbarkeit, Standort,
Notizen und Archivstatus. Ein neues oder auf eine andere Revision umgestelltes Exemplar bleibt
sichtbar **Bestand prüfen**; nur `ModuleInventoryReviewDialog` darf es bestätigen. Beide Dialoge
verwenden `LayoutConfirmDialog`, sobald geänderte Eingaben verworfen werden sollen.

Im Komponententest werden zusätzlich diese konkreten Aufrufe geprüft:

```tsx
await user.click(screen.getByRole("button", { name: "Modulvorlage anlegen" }));
await user.type(screen.getByRole("textbox", { name: "Bezeichnung" }), "  Abzweig Süd  ");
await user.selectOptions(screen.getByRole("combobox", { name: "Art" }), "module");
await user.type(screen.getByRole("spinbutton", { name: "Breite (mm)" }), "2400");
await user.type(screen.getByRole("spinbutton", { name: "Tiefe (mm)" }), "850");
await user.click(screen.getByRole("button", { name: "Anlegen" }));
expect(api.createModuleTemplate).toHaveBeenCalledWith({
  name: "Abzweig Süd",
  description: "",
  initialRevision: {
    kind: "module",
    gauge: "TT",
    widthMm: 2400,
    heightMm: 850
  }
});

await user.click(screen.getByRole("button", { name: "Neue Revision" }));
await user.click(screen.getByRole("button", { name: "Entwurf anlegen" }));
expect(api.createModuleTemplateRevision).toHaveBeenCalledWith(templateFixture.id,
  expect.objectContaining({ baseRevisionId: revisionFixture.id }));
```

- [ ] **Step 5: Eigenständige Route und Anlagenadapter anbinden**

`AppView` erhält `"modules"`; `/modules` lädt `ModuleLibraryView`. Diese Route verwendet dieselbe
Rollenlogik wie Anlagen und ist für Messe nicht erreichbar. `LayoutsView` erhält einen ruhigen Link
**Modulbibliothek**. `LayoutModulesPanel` wird zu diesem Adapter:

```tsx
export function LayoutModulesPanel({ layoutID, canPlan }: { layoutID: string; canPlan: boolean }) {
  return <ModuleLibraryPanel layoutId={layoutID} canPlan={canPlan} />;
}
```

Die übrigen Anlagen-Tabs behalten weiterhin `LayoutUnit[]`, bis Stufe 4 ihre Platzierungsverträge
umstellt.

- [ ] **Step 6: Fokussierte Styles mit vorhandenem Tokensystem ergänzen**

```css
.module-library {
  display: grid;
  gap: var(--space-lg);
  min-width: 0;
}

.module-library-head,
.module-library-toolbar,
.module-library-pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-md);
}

.module-library-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(18rem, 0.38fr);
  gap: var(--space-lg);
  min-width: 0;
}

@media (max-width: 900px) {
  .module-library-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .module-library-head,
  .module-library-toolbar {
    align-items: stretch;
    flex-direction: column;
  }
}
```

Nur vorhandene Farb-, Abstands-, Radius-, Schatten- und Typografietokens verwenden. Die
Bibliothek darf keine Hero-Fläche, Marketingkarte oder dekorative Animation einführen.

- [ ] **Step 7: Komponenten- und Buildprüfung ausführen**

Run: `cd frontend; npm.cmd run test:run -- src/features/layouts/modules/ModuleLibraryPanel.test.tsx src/app/App.test.tsx; npm.cmd run build`

Expected: PASS, ohne TypeScript-Casts auf `any` und ohne horizontales Seitenoverflow.

- [ ] **Step 8: Commit erstellen**

```powershell
git add frontend/src/features/layouts/ModuleLibraryView.tsx `
  frontend/src/features/layouts/modules/ModuleLibraryPanel.tsx `
  frontend/src/features/layouts/modules/ModuleLibraryPanel.test.tsx `
  frontend/src/features/layouts/modules/ModuleTemplateDialog.tsx `
  frontend/src/features/layouts/modules/ModuleInstanceDialog.tsx `
  frontend/src/styles/module-library.css frontend/src/app/App.tsx `
  frontend/src/app/navigationAvailability.ts frontend/src/features/layouts/LayoutsView.tsx `
  frontend/src/features/layouts/LayoutWorkspace.tsx `
  frontend/src/features/layouts/LayoutModulesPanel.tsx frontend/src/app/styles.css
git commit -m "feat(modules): add project-wide module library UI"
```

### Task 11: Bestandsprüfung, Konflikterhalt und zweisprachige Textqualität

**Files:**
- Create: `frontend/src/features/layouts/modules/ModuleInventoryReviewDialog.tsx`
- Create: `frontend/src/features/layouts/modules/ModuleInventoryReviewDialog.test.tsx`
- Modify: `frontend/src/features/layouts/modules/ModuleLibraryPanel.tsx`
- Modify: `frontend/src/shared/i18n/de.ts:2504-2605`
- Modify: `frontend/src/shared/i18n/en.ts:2504-2605`
- Modify: `frontend/src/shared/i18n.test.ts:41-103`

**Interfaces:**
- Consumes: `ModuleInstance`, `ReviewModuleInstanceInput`, `ApiError` und
  `api.reviewModuleInstance`.
- Produces: zugängliche explizite Prüfung migrierter Exemplare.

- [ ] **Step 1: Fachliche Sperren und Konfliktverhalten testen**

```tsx
it("keeps reviewed values after a version conflict and requires explicit stock facts", async () => {
  const user = userEvent.setup();
  const review = vi.spyOn(api, "reviewModuleInstance")
    .mockRejectedValue(new ApiError("changed", "module_version_conflict", 409));
  render(<ModuleInventoryReviewDialog instance={instanceFixture} revision={releasedRevisionFixture}
    onSaved={vi.fn()} onClose={vi.fn()} />);
  await user.type(screen.getByRole("textbox", { name: "Kennung" }), "M-001");
  await user.selectOptions(screen.getByRole("combobox", { name: "Zustand" }), "ready");
  await user.selectOptions(screen.getByRole("combobox", { name: "Verfügbarkeit" }), "available");
  await user.click(screen.getByRole("button", { name: "Bestand bestätigen" }));
  await screen.findByText("Der Datensatz wurde zwischenzeitlich geändert.");
  expect(screen.getByRole("textbox", { name: "Kennung" })).toHaveValue("M-001");
  expect(review).toHaveBeenCalledWith(instanceFixture.id, expect.objectContaining({
    inventoryNumber: "M-001", condition: "ready", availability: "available",
    expectedVersion: instanceFixture.version
  }));
});
```

Ein zweiter Test rendert eine Entwurfsrevision und erwartet den erklärenden Text
**Revision vor der Bestandsbestätigung freigeben** sowie einen deaktivierten Bestätigungsbutton.

- [ ] **Step 2: Dialogtest ausführen und fehlende Komponente bestätigen**

Run: `cd frontend; npm.cmd run test:run -- src/features/layouts/modules/ModuleInventoryReviewDialog.test.tsx`

Expected: FAIL mit fehlendem Modul.

- [ ] **Step 3: Dialog mit lokalem Entwurf und sichtbarer Konfliktbehandlung implementieren**

```tsx
const canConfirm = revision.status === "released" && form.inventoryNumber.trim() !== "" &&
  form.condition !== "unknown" && form.availability !== "unchecked";

const submit = async (event: FormEvent) => {
  event.preventDefault();
  setSaving(true);
  setMessage("");
  try {
    const saved = await api.reviewModuleInstance(instance.id, {
      inventoryNumber: form.inventoryNumber.trim(),
      ownerLabel: form.ownerLabel.trim(),
      condition: form.condition,
      availability: form.availability,
      locationLabel: form.locationLabel.trim(),
      notes: form.notes.trim(),
      expectedVersion: instance.version
    });
    onSaved(saved);
  } catch (reason) {
    setMessage(reason instanceof ApiError && reason.status === 409
      ? t("modules.review.conflict")
      : reason instanceof Error ? reason.message : t("modules.error.generic"));
  } finally {
    setSaving(false);
  }
};
```

Der Dialog verwendet native Labels, `AppSelect`, initialen Fokus auf Kennung, Rückgabe des Fokus
an den Auslöser und keine automatische Schließung bei Fehlern. Schließen mit geänderten Werten
fordert über `LayoutConfirmDialog` eine Bestätigung.

- [ ] **Step 4: Deutsche und englische Fachbegriffe ergänzen und absichern**

Verbindliche deutsche Begriffe: **Modulbibliothek**, **Modulvorlage**, **Vorlagenrevision**,
**Modulexemplar**, **Bestand prüfen**, **Bestand bestätigt**, **Verfügbarkeit**, **Entwurf**,
**Freigegeben**, **Archiviert**.

Verbindliche englische Begriffe: **Module library**, **Module template**, **Template revision**,
**Module instance**, **Review inventory**, **Inventory confirmed**, **Availability**, **Draft**,
**Released**, **Archived**.

```ts
it("uses the approved module-library terminology", async () => {
  const [{ deTranslations }, { enTranslations }] = await Promise.all([
    import("./i18n/de"), import("./i18n/en")
  ]);
  expect(deTranslations["modules.title"]).toBe("Modulbibliothek");
  expect(deTranslations["modules.review.required"]).toBe("Bestand prüfen");
  expect(enTranslations["modules.title"]).toBe("Module library");
  expect(enTranslations["modules.review.required"]).toBe("Review inventory");
});
```

- [ ] **Step 5: Dialog, i18n und gesamten Frontendtest ausführen**

Run: `cd frontend; npm.cmd run test:run -- src/features/layouts/modules/ModuleInventoryReviewDialog.test.tsx src/shared/i18n.test.ts; npm.cmd run test:run`

Expected: PASS. Deutsche und englische Schlüssel sind identisch; kein sichtbarer Schlüssel fällt
auf seinen Rohwert zurück.

- [ ] **Step 6: Commit erstellen**

```powershell
git add frontend/src/features/layouts/modules/ModuleInventoryReviewDialog.tsx `
  frontend/src/features/layouts/modules/ModuleInventoryReviewDialog.test.tsx `
  frontend/src/features/layouts/modules/ModuleLibraryPanel.tsx `
  frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts `
  frontend/src/shared/i18n.test.ts
git commit -m "feat(modules): add explicit inventory review"
```

### Task 12: Gesamtverifikation und visuelle Abnahme der Stufe 1

**Files:**
- Modify only if a verified defect is found: files already listed in Tasks 1 through 11.

**Interfaces:**
- Consumes: vollständige Stufe-1-Implementierung.
- Produces: nachweisbar grünen Build, migrationssicheren Datenbestand und geprüfte Desktop-,
  Mobil-, Hell- und Dunkelansichten.

- [ ] **Step 1: Formatierung und fokussierte statische Prüfung ausführen**

Run:

```powershell
gofmt -w backend/internal/domain/module_library.go `
  backend/internal/domain/module_library_test.go `
  backend/internal/application/module_library.go `
  backend/internal/application/module_library_test.go `
  backend/internal/application/backup_module_library.go `
  backend/internal/infrastructure/module_library_repository.go `
  backend/internal/infrastructure/module_library_repository_test.go `
  backend/internal/infrastructure/module_identity_migration_test.go `
  backend/internal/api/module_library_handlers.go `
  backend/internal/api/module_library_handlers_test.go
git diff --check
```

Expected: keine Ausgabe von `git diff --check`.

- [ ] **Step 2: Gesamte Backend- und Frontendprüfung ausführen**

Run:

```powershell
cd backend
go test ./...
cd ..\frontend
npm.cmd run test:run
npm.cmd run build
```

Expected: alle Go- und Vitest-Tests PASS; TypeScript- und Vite-Build erfolgreich.

- [ ] **Step 3: Migration und Restore mit Fremdschlüsselprüfung separat wiederholen**

Run:

```powershell
cd backend
go test ./internal/infrastructure -run TestProjectWideModuleIdentityMigrationPreservesLayoutUse -count=1
go test ./internal/application -run TestBackupVersionTwentyOnePreservesModuleLibraryAndLegacyTwentyBackfillsIt -count=1
```

Expected: beide Tests PASS und keine Fremdschlüsselverletzung.

- [ ] **Step 4: Browserprüfung mit laufendem lokalen Server durchführen**

Den Frontend-Build über den Go-Server öffnen und diese Matrix abarbeiten:

| Ansicht | Breite | Theme | Prüfpunkte |
|---|---:|---|---|
| `/modules` | 1440 px | dunkel | Dichte, Tabellenhierarchie, Detailpanel, lange Namen, Fokus |
| `/modules` | 1440 px | hell | Flächenhierarchie, Kontrast, Statussymbole, Fokus |
| `/modules` | 390 px | dunkel | Umbruch, 44-px-Aktionen, kein Seitenoverflow, Dialog |
| `/layouts` Modul-Tab | 1440 px | dunkel | Kontextfilter, Alle Module, Anlagenzuordnung |

Zusätzlich prüfen: leer, Laden, API-Fehler, eine Seite, mehrere Seiten, Viewer, Planner,
Versionskonflikt, Entwurfsrevision und freigegebene Revision. Die Stufe-1-Bibliothek bleibt eine
ruhige operative Liste. Sie ahmt die in Stufe 3 vorgesehene Zeichenfläche nicht vorzeitig nach.

- [ ] **Step 5: Sichtbare Texte manuell gegen die festgelegte Terminologie lesen**

Jeden neuen deutschen und englischen Text in der Oberfläche einmal vollständig anzeigen. Prüfen:
Umlaute, Großschreibung, Singular/Plural, Statusbegriffe, Platzhalter, abgeschnittene lange Namen
und verständliche Konfliktmeldungen. Gefundene Fehler direkt in beiden Sprachdateien und im
zugehörigen Test korrigieren.

- [ ] **Step 6: Letzten Implementierungscommit erstellen**

```powershell
git add backend frontend openapi/railkeeper.yaml
git commit -m "test(modules): verify project-wide module identity"
```

Expected: Der Commit enthält ausschließlich Korrekturen aus der Gesamtverifikation. Falls keine
Korrektur nötig war, wird kein leerer Commit erzeugt.

---

## Selbstprüfung des Plans

### Spezifikationsabdeckung

- Vorlage, Revision und Exemplar: Tasks 1 bis 4.
- Deterministische Migration bestehender `LayoutUnit`-Daten: Task 2.
- Erhalt der Anlagenverwendung und Konfigurationswerte: Task 5.
- Projektweite Wiederverwendung in mehreren Anlagen: Tasks 4, 5 und 10.
- Keine behauptete physische Nutzung in Bestandskonfigurationen: Tasks 2 und 5.
- Paginierte getrennte Verträge: Tasks 3, 4, 6, 7 und 9.
- Rollen, CSRF und Ausschluss von Messe: Task 6.
- Audit und Revisionswechsel: Task 4.
- Backup, Restore und alte Backups: Task 8.
- Explizite Bestandsprüfung: Tasks 4 und 11.
- Zweisprachigkeit, Rechtschreibung, responsive Zustände und Themes: Tasks 10 bis 12.
- Exakte visuelle Referenzen: Hash-Gate in Task 10. Der eigentliche Form- und Kopfeditor ist
  entsprechend der freigegebenen Stufenteilung nicht Teil dieses Plans.

### Bewusst außerhalb dieser Stufe

- stabile Kontursegmente, Geraden und Kreisbögen;
- Verbindungsseiten, Kopfansicht, Normfamilien und Profilrevisionen;
- gebundene Ports und Normprüfung;
- mehrfache Platzierung einer Vorlage innerhalb eines normalen Entwurfs;
- `Z`-Positionen, Ebenen und interaktive Ausstellungsaufbauten;
- Canvas-Massendarstellung für 1.000 Module und 100.000 Detailobjekte.

Diese Punkte besitzen bereits freigegebene Folgestufen und werden nicht als unvollständige
Stufe-1-Funktion dargestellt.

### Typkonsistenz

- Backend und Frontend verwenden identische Statuswerte und JSON-Feldnamen.
- `ModuleTemplateRevisionID` beziehungsweise `moduleTemplateRevisionId` bezeichnet überall die
  kanonische Revision.
- `sourceLayoutUnitId` ist ausschließlich die Stufe-1-Brücke zum bisherigen technischen Aggregat.
- `unitId` bleibt ausschließlich die Legacy-Referenz bestehender Anlagenfunktionen.
- Seiten verwenden durchgehend `items`, `page`, `pageSize`, `totalItems` und `totalPages`.
- Schreibkonflikte verwenden durchgehend `expectedVersion` und HTTP 409.

## Inline-Ausführung

Die Ausführung erfolgt gemäß Benutzerentscheidung vollständig inline mit
`superpowers:executing-plans`. Nach Tasks 4, 8 und 12 gibt es jeweils einen überprüfbaren
Checkpoint für Datenmodell, Restore-Sicherheit und Gesamtabnahme.
