# Zubehörartikel sicher löschen Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Admins können vollständig unbenutzte Zubehörartikel endgültig löschen, während jeder vorhandene Bestand oder Nutzungsbezug die Löschung atomar und nachvollziehbar verhindert.

**Architecture:** Die Löschregel liegt vollständig im Go-Backend und wird innerhalb derselben SQLite-Transaktion geprüft und ausgeführt. Die Repository-Methode liefert die Blob-IDs gelöschter Metadaten zurück, der Application Service bereinigt anschließend nicht mehr referenzierte Dateien. Das React-Frontend bietet eine einzige geteilte Admin-Aktion für Tabelle, Kacheln und Mobilansicht an und führt sie über den vorhandenen Bestätigungsdialog aus.

**Tech Stack:** Go 1.x, `net/http`, SQLite, React, TypeScript, Vite, Vitest, Testing Library, OpenAPI 3

## Global Constraints

- Endgültiges Löschen ist nur für Admins erlaubt; Editor, Viewer, Planner und Messe bleiben ausgeschlossen.
- Ein Artikel ist nur löschbar, wenn sein Gesamtbestand 0 ist und weder Käufe, Bestandsbewegungen, Reservierungen, Einbauten, Nutzungshistorie, Einzelobjekte noch technische Anlagenpositionen auf ihn verweisen.
- Die Prüfung und die Datenbanklöschung müssen in einer einzigen SQLite-Schreibtransaktion erfolgen.
- Bilder, Dokumente und benutzerdefinierte Attribute gelten als Artikelmetadaten und dürfen bei einer ansonsten zulässigen Löschung entfernt werden.
- Datei-Blobs dürfen erst nach erfolgreichem Commit und nur über `DeleteIfUnreferenced` physisch entfernt werden.
- Archivieren und Wiederherstellen bleiben unverändert für Admin und Editor verfügbar; vorheriges Archivieren ist keine Löschvoraussetzung.
- Die schreibende Route bleibt CSRF-geschützt und erzeugt einen Audit-Eintrag.
- Backend, Frontend-API und `openapi/railkeeper.yaml` müssen dieselbe Route und dieselben Fehlerzustände abbilden.
- Deutsche und englische UI-Texte müssen vollständig gepflegt werden.
- Keine neue Migration und keine neue UI-Bibliothek einführen.
- Bestehende Nutzeränderungen im Worktree nicht überschreiben; `frontend/dist`, `data`, Caches und lokale Zugangsdaten nicht committen.

---

## File map

- `backend/internal/application/accessories.go`: neuer domänenspezifischer Fehler, Repository-Vertrag, Service-Methode und optionaler Blob-Cleaner.
- `backend/internal/application/accessories_test.go`: Service-Vertrag, Normalisierung, Fehlerweitergabe und Blob-Bereinigung.
- `backend/internal/infrastructure/accessory_product_deletion.go`: fokussierte atomare Löschoperation und Sperrprüfung.
- `backend/internal/infrastructure/accessory_product_deletion_test.go`: persistente Erfolgs-, Sperr-, Rollback- und Audit-Tests.
- `backend/cmd/railkeeper/main.go`: vorhandenen File-Blob-Service in den Zubehör-Service injizieren.
- `backend/internal/api/accessory_handlers.go`: DELETE-Handler und stabile Problem-Antwort `accessory_delete_blocked`.
- `backend/internal/api/routes.go`: Admin- und CSRF-geschützte DELETE-Route.
- `backend/internal/api/accessory_handlers_test.go`: Rollen-, CSRF-, Erfolgs-, 404- und 409-Verhalten.
- `openapi/railkeeper.yaml`: öffentlicher DELETE-Vertrag.
- `frontend/src/shared/apiLayoutsAccessories.ts`: typisierte DELETE-Anfrage.
- `frontend/src/shared/apiLayoutsAccessories.test.ts`: URL-Encoding und HTTP-Methode.
- `frontend/src/features/accessories/useArticleOverview.ts`: Delete-Command mit Reload und weitergereichtem Fehler.
- `frontend/src/features/accessories/useArticleOverview.test.tsx`: Reload- und Fehlerverhalten des Hooks.
- `frontend/src/features/accessories/ArticleActions.tsx`: Admin-Löschaktion und korrekte Mehrpunkt-Menü-Tastatursteuerung.
- `frontend/src/features/accessories/ArticleActions.test.tsx`: Sichtbarkeit, Callback und Fokusnavigation.
- `frontend/src/features/accessories/ArticleTable.tsx`: Delete-Props an die gemeinsame Aktion weiterreichen.
- `frontend/src/features/accessories/ArticleCardGrid.tsx`: Delete-Props an die gemeinsame Aktion weiterreichen.
- `frontend/src/features/accessories/ArticleCompactList.tsx`: Delete-Props an die gemeinsame Aktion weiterreichen.
- `frontend/src/features/accessories/AccessoriesView.tsx`: Admin-Erkennung, Dialogzustand und erfolgreicher Reload.
- `frontend/src/features/accessories/AccessoriesView.test.tsx`: Rollen, Bestätigung, Erfolg und Konflikt in allen Ansichten.
- `frontend/src/shared/i18n/de.ts`: deutsche Löschtexte.
- `frontend/src/shared/i18n/en.ts`: englische Löschtexte.
- `frontend/src/styles/accessories.css`: Gefahrfarbe für den Menüeintrag.

### Task 1: Application-Vertrag für die sichere Produktlöschung

**Files:**
- Modify: `backend/internal/application/accessories.go`
- Modify: `backend/internal/application/accessories_test.go`

**Interfaces:**
- Consumes: `FileBlobReferenceCleaner.DeleteIfUnreferenced(context.Context, string) error` aus `accessory_documents.go`.
- Produces: `ErrAccessoryDeleteBlocked`, `AccessoryCatalogRepository.DeleteProduct(context.Context, string, string) ([]string, error)` und `AccessoryService.DeleteProduct(context.Context, string, string) error`.
- Produces: rückwärtskompatibler Konstruktor `NewAccessoryService(AccessoryRepository, ...FileBlobReferenceCleaner) *AccessoryService`.

- [ ] **Step 1: Den fehlenden Service-Vertrag mit fehlschlagenden Tests festschreiben**

Erweitere `accessoryRepositorySpy` um:

```go
deleteProductID      string
deleteActor          string
deleteBlobIDs        []string
deleteProductErr     error
```

und implementiere im Spy:

```go
func (spy *accessoryRepositorySpy) DeleteProduct(
	_ context.Context,
	id string,
	actor string,
) ([]string, error) {
	spy.deleteProductID = id
	spy.deleteActor = actor
	return append([]string(nil), spy.deleteBlobIDs...), spy.deleteProductErr
}
```

Füge außerdem diesen lokalen Cleaner und die Tests hinzu:

```go
type accessoryBlobCleanerSpy struct {
	deleted []string
	err     error
}

func (spy *accessoryBlobCleanerSpy) DeleteIfUnreferenced(_ context.Context, id string) error {
	spy.deleted = append(spy.deleted, id)
	return spy.err
}

func TestAccessoryServiceDeletesProductAndCleansUnreferencedBlobs(t *testing.T) {
	repository := &accessoryRepositorySpy{deleteBlobIDs: []string{"blob-1", "blob-2"}}
	blobs := &accessoryBlobCleanerSpy{}
	service := NewAccessoryService(repository, blobs)

	if err := service.DeleteProduct(t.Context(), " product-1 ", "admin-1"); err != nil {
		t.Fatal(err)
	}
	if repository.deleteProductID != "product-1" || repository.deleteActor != "admin-1" {
		t.Fatalf("unexpected repository call: id=%q actor=%q",
			repository.deleteProductID, repository.deleteActor)
	}
	if !slices.Equal(blobs.deleted, []string{"blob-1", "blob-2"}) {
		t.Fatalf("deleted blob IDs = %#v", blobs.deleted)
	}
}

func TestAccessoryServiceRejectsEmptyDeleteIDAndPreservesErrors(t *testing.T) {
	repository := &accessoryRepositorySpy{deleteProductErr: ErrAccessoryDeleteBlocked}
	service := NewAccessoryService(repository)

	if err := service.DeleteProduct(t.Context(), "  ", "admin-1"); !errors.Is(err, ErrAccessoryValidation) {
		t.Fatalf("empty product ID error = %v", err)
	}
	if err := service.DeleteProduct(t.Context(), "product-1", "admin-1");
		!errors.Is(err, ErrAccessoryDeleteBlocked) {
		t.Fatalf("repository delete error = %v", err)
	}
}

func TestAccessoryServiceReportsBlobCleanupFailure(t *testing.T) {
	want := errors.New("cleanup failed")
	repository := &accessoryRepositorySpy{deleteBlobIDs: []string{"blob-1"}}
	service := NewAccessoryService(repository, &accessoryBlobCleanerSpy{err: want})

	if err := service.DeleteProduct(t.Context(), "product-1", "admin-1"); !errors.Is(err, want) {
		t.Fatalf("blob cleanup error = %v", err)
	}
}
```

- [ ] **Step 2: Den RED-Zustand prüfen**

Run:

```powershell
cd backend
go test ./internal/application -run 'TestAccessoryService(DeletesProduct|RejectsEmptyDeleteID|ReportsBlobCleanupFailure)$'
```

Expected: FAIL, weil `DeleteProduct`, `ErrAccessoryDeleteBlocked` und der zweite Konstruktorparameter noch fehlen.

- [ ] **Step 3: Den minimalen Application-Vertrag implementieren**

Ergänze den Fehlerblock und das Repository-Interface:

```go
ErrAccessoryDeleteBlocked = errors.New("accessory deletion blocked")
```

```go
DeleteProduct(context.Context, string, string) ([]string, error)
```

Ersetze Service und Konstruktor durch:

```go
type AccessoryService struct {
	repository AccessoryRepository
	blobs      FileBlobReferenceCleaner
}

func NewAccessoryService(
	repository AccessoryRepository,
	blobCleaners ...FileBlobReferenceCleaner,
) *AccessoryService {
	var blobs FileBlobReferenceCleaner
	if len(blobCleaners) > 0 {
		blobs = blobCleaners[0]
	}
	return &AccessoryService{repository: repository, blobs: blobs}
}
```

Füge die Service-Methode direkt bei den Produktmutationen hinzu:

```go
func (s *AccessoryService) DeleteProduct(ctx context.Context, id, actor string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrAccessoryValidation
	}
	blobIDs, err := s.repository.DeleteProduct(ctx, id, actor)
	if err != nil {
		return err
	}
	if s.blobs == nil {
		return nil
	}
	for _, blobID := range blobIDs {
		if err := s.blobs.DeleteIfUnreferenced(ctx, blobID); err != nil {
			return fmt.Errorf("delete accessory product blob %q: %w", blobID, err)
		}
	}
	return nil
}
```

- [ ] **Step 4: Formatieren und GREEN prüfen**

Run:

```powershell
gofmt -w internal/application/accessories.go internal/application/accessories_test.go
go test ./internal/application
```

Expected: PASS.

- [ ] **Step 5: Den Application-Vertrag committen**

```powershell
git add backend/internal/application/accessories.go backend/internal/application/accessories_test.go
git commit -m "feat(accessories): define safe product deletion"
```

### Task 2: Atomare SQLite-Löschung und Datei-Bereinigung

**Files:**
- Create: `backend/internal/infrastructure/accessory_product_deletion.go`
- Create: `backend/internal/infrastructure/accessory_product_deletion_test.go`
- Modify: `backend/cmd/railkeeper/main.go`

**Interfaces:**
- Consumes: `AccessoryCatalogRepository.DeleteProduct(context.Context, string, string) ([]string, error)`.
- Consumes: `reserveAccessoryWriteTransaction`, `writeAccessoryAudit` und `AccessoryRepository.withTx`.
- Produces: transaktionale Implementierung mit `AccessoryProductDeleted`-Audit und Blob-ID-Rückgabe.

- [ ] **Step 1: Erfolgs- und Rollback-Verhalten als RED-Tests anlegen**

Erstelle die neue Testdatei im Paket `infrastructure_test`. Verwende `testAccessoryService(t)` aus `accessory_repository_test.go` und lege Produkte über `CreateProduct` an. Der Erfolgstest muss vor der Löschung einen Nullbestand, ein `file_blobs`-Objekt, ein `accessory_documents`-Objekt und einen benutzerdefinierten Attributwert anlegen. Danach gilt:

```go
func TestAccessoryProductDeleteRemovesUnusedProductMetadataAndAudits(t *testing.T) {
	_, db := testAccessoryService(t)
	blobs := application.NewFileBlobService(db, t.TempDir())
	blobID, err := blobs.Store(t.Context(), []byte("accessory image"))
	if err != nil {
		t.Fatal(err)
	}
	service := application.NewAccessoryService(infrastructure.NewAccessoryRepository(db), blobs)
	product := createDeletableAccessoryProduct(t, service)
	insertDeletableProductMetadata(t, db, product.ID, blobID)

	if err := service.DeleteProduct(t.Context(), product.ID, "admin-1"); err != nil {
		t.Fatal(err)
	}
	assertRowCount(t, db, "accessory_products", "id", product.ID, 0)
	assertRowCount(t, db, "accessory_stock", "product_id", product.ID, 0)
	assertRowCount(t, db, "accessory_product_attributes", "product_id", product.ID, 0)
	assertRowCount(t, db, "accessory_documents", "product_id", product.ID, 0)
	assertRowCount(t, db, "file_blobs", "id", blobID, 0)
	assertAccessoryAuditCount(t, db, "AccessoryProductDeleted", 1)
}
```

`insertDeletableProductMetadata` führt die folgenden Inserts aus. Verwende für Testdaten die feste
Zeit `2026-08-14T10:00:00Z` und den zuvor mit dem echten `FileBlobService` gespeicherten Blob:

```go
func insertDeletableProductMetadata(t *testing.T, db *sql.DB, productID, blobID string) {
	t.Helper()
	const timestamp = "2026-08-14T10:00:00Z"
	statements := []struct {
	query string
	args  []any
}{
	{`INSERT INTO storage_locations(id, name, description, archived, created_at, updated_at)
	  VALUES('location-delete', 'Löschlager', '', 0, ?, ?)`, []any{timestamp, timestamp}},
	{`INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
	  VALUES(?, 'location-delete', 0, ?)`, []any{productID, timestamp}},
	{`INSERT INTO accessory_product_attributes(
	    id, product_id, attribute_key, value_type, text_value, created_at, updated_at
	  ) VALUES('attribute-delete', ?, 'test-note', 'text', 'Metadatum', ?, ?)`,
	  []any{productID, timestamp, timestamp}},
	{`INSERT INTO accessory_documents(
	    id, product_id, file_blob_id, file_name, original_name, description, category,
	    mime_type, size_bytes, is_primary, created_by, created_at, updated_at
	  ) VALUES('document-delete', ?, ?, 'bild.png', 'bild.png', '', 'image',
	    'image/png', 15, 1, 'admin-1', ?, ?)`, []any{productID, blobID, timestamp, timestamp}},
}
for _, statement := range statements {
	if _, err := db.ExecContext(t.Context(), statement.query, statement.args...); err != nil {
		t.Fatal(err)
		}
	}
}
```

Definiere die zwei übrigen Hilfen vollständig:

```go
func createDeletableAccessoryProduct(
	t *testing.T,
	service *application.AccessoryService,
) *application.AccessoryProduct {
	t.Helper()
	product, err := service.CreateProduct(t.Context(), application.CreateAccessoryProductInput{
		Manufacturer: "Tillig", ArticleNumber: "DELETE-1", Name: "Löschtest",
		Category: "Gleismaterial", TrackingMode: domain.AccessoryTrackingModeQuantity,
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	return product
}

func assertRowCount(
	t *testing.T,
	db *sql.DB,
	table string,
	column string,
	value string,
	want int,
) {
	t.Helper()
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ?", table, column)
	var got int
	if err := db.QueryRowContext(t.Context(), query, value).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("%s row count = %d, want %d", table, got, want)
	}
}
```

Die neue Testdatei importiert `database/sql`, `errors`, `fmt`, `testing` sowie
die Pakete `application`, `domain` und `infrastructure`.

Lege anschließend einen tabellengesteuerten Sperrtest an. Jede Fixture erhält eine frische Datenbank und genau einen dieser Bezüge:

```go
tests := []struct {
	name   string
	insert func(*testing.T, *sql.DB, string)
}{
	{name: "positive stock", insert: insertPositiveAccessoryStock},
	{name: "asset", insert: insertAccessoryAssetReference},
	{name: "stock movement", insert: insertAccessoryMovementReference},
	{name: "purchase", insert: insertAccessoryPurchaseReference},
	{name: "reservation", insert: insertAccessoryReservationReference},
	{name: "installation", insert: insertAccessoryInstallationReference},
	{name: "layout technical position", insert: insertLayoutTechnicalPositionReference},
}
```

Jeder Subtest legt vor seinem spezifischen Bezug diese gemeinsamen Fremdschlüsselziele an:

```sql
INSERT INTO storage_locations(id, name, description, archived, created_at, updated_at)
VALUES('location-block', 'Sperrlager', '', 0, '2026-08-14T10:00:00Z', '2026-08-14T10:00:00Z');
INSERT INTO layouts(id, name, kind, gauge, scale, description, version, archived, created_at, updated_at)
VALUES('layout-block', 'Sperranlage', 'club', 'TT', '1:120', '', 1, 0,
  '2026-08-14T10:00:00Z', '2026-08-14T10:00:00Z');
INSERT INTO layout_units(
  id, layout_id, name, kind, owner_label, width_mm, height_mm, version, archived, created_at, updated_at
) VALUES('unit-block', 'layout-block', 'Sperrmodul', 'module', '', 1000, 500, 1, 0,
  '2026-08-14T10:00:00Z', '2026-08-14T10:00:00Z');
```

Die sieben Fixture-Funktionen führen jeweils genau eines dieser Inserts aus, wobei `?` die
Produkt-ID erhält:

```sql
INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
VALUES(?, 'location-block', 1, '2026-08-14T10:00:00Z');

INSERT INTO accessory_assets(
  id, product_id, condition_state, lifecycle_state, created_at, updated_at
) VALUES('asset-block', ?, 'ready', 'stored', '2026-08-14T10:00:00Z', '2026-08-14T10:00:00Z');

INSERT INTO accessory_stock_movements(
  id, product_id, location_id, movement_type, quantity, created_at
) VALUES('movement-block', ?, 'location-block', 'adjustment', 1, '2026-08-14T10:00:00Z');

INSERT INTO accessory_purchases(
  id, product_id, quantity, purchased_at, created_at, updated_at
) VALUES('purchase-block', ?, 1, '2026-08-14', '2026-08-14T10:00:00Z', '2026-08-14T10:00:00Z');

INSERT INTO accessory_reservations(
  id, product_id, location_id, quantity, layout_id, status, created_by, created_at, updated_at
) VALUES('reservation-block', ?, 'location-block', 1, 'layout-block', 'cancelled', 'admin-1',
  '2026-08-14T10:00:00Z', '2026-08-14T10:00:00Z');

INSERT INTO accessory_installations(
  id, product_id, source_location_id, quantity, layout_id, condition_state,
  installed_by, installed_at, removed_by, removed_at, removal_disposition, notes, removal_notes
) VALUES('installation-block', ?, 'location-block', 1, 'layout-block', 'ready',
  'admin-1', '2026-08-14T10:00:00Z', 'admin-1', '2026-08-14T11:00:00Z', 'stored', '', '');

INSERT INTO layout_technical_positions(
  id, layout_unit_id, label, kind, position_x_mm, position_y_mm, rotation_degrees,
  product_id, description, version, archived, created_at, updated_at
) VALUES('position-block', 'unit-block', 'Sperrposition', 'turnout', 10, 20, 0,
  ?, '', 1, 0, '2026-08-14T10:00:00Z', '2026-08-14T10:00:00Z');
```

Der Subtest erstellt seine isolierte Datenbank und den Service mit Cleaner so:

```go
_, db := testAccessoryService(t)
blobs := application.NewFileBlobService(db, t.TempDir())
blobID, err := blobs.Store(t.Context(), []byte("blocked accessory image"))
if err != nil {
	t.Fatal(err)
}
service := application.NewAccessoryService(infrastructure.NewAccessoryRepository(db), blobs)
product := createDeletableAccessoryProduct(t, service)
insertDeletableProductMetadata(t, db, product.ID, blobID)
insertAccessoryDeleteReferencePrerequisites(t, db)
test.insert(t, db, product.ID)
```

`insertAccessoryDeleteReferencePrerequisites` führt die drei gemeinsamen SQL-Inserts für Lager,
Anlage und Modul aus dem vorherigen Block aus. Danach muss der Testkörper den Fehler, die
unveränderten Metadaten und das Ausbleiben jeder Blob-Bereinigung prüfen:

```go
if err := service.DeleteProduct(t.Context(), product.ID, "admin-1");
	!errors.Is(err, application.ErrAccessoryDeleteBlocked) {
	t.Fatalf("delete error = %v", err)
}
assertRowCount(t, db, "accessory_products", "id", product.ID, 1)
assertRowCount(t, db, "accessory_documents", "product_id", product.ID, 1)
assertRowCount(t, db, "file_blobs", "id", blobID, 1)
assertAccessoryAuditCount(t, db, "AccessoryProductDeleted", 0)
```

Ergänze außerdem den Not-found-Test:

```go
func TestAccessoryProductDeleteReturnsNotFound(t *testing.T) {
	service, _ := testAccessoryService(t)
	if err := service.DeleteProduct(t.Context(), "missing", "admin-1");
		!errors.Is(err, application.ErrAccessoryNotFound) {
		t.Fatalf("delete missing product error = %v", err)
	}
}
```

- [ ] **Step 2: Die Infrastrukturtests im RED-Zustand ausführen**

Run:

```powershell
cd backend
go test ./internal/infrastructure -run 'TestAccessoryProductDelete'
```

Expected: FAIL, weil `AccessoryRepository.DeleteProduct` noch nicht implementiert ist.

- [ ] **Step 3: Die atomare Repository-Methode implementieren**

Erstelle `accessory_product_deletion.go` mit der Methode:

```go
func (r *AccessoryRepository) DeleteProduct(
	ctx context.Context,
	id string,
	actor string,
) ([]string, error) {
	var blobIDs []string
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		if err := reserveAccessoryWriteTransaction(ctx, tx); err != nil {
			return err
		}
		exists, err := accessoryRecordExists(ctx, tx,
			`SELECT COUNT(*) FROM accessory_products WHERE id = ?`, id)
		if err != nil {
			return fmt.Errorf("check accessory product: %w", err)
		}
		if !exists {
			return application.ErrAccessoryNotFound
		}

		var blocked bool
		err = tx.QueryRowContext(ctx, `
			SELECT
				EXISTS(SELECT 1 FROM accessory_stock WHERE product_id = ? AND quantity <> 0)
				OR EXISTS(SELECT 1 FROM accessory_assets WHERE product_id = ?)
				OR EXISTS(SELECT 1 FROM accessory_stock_movements WHERE product_id = ?)
				OR EXISTS(SELECT 1 FROM accessory_purchases WHERE product_id = ?)
				OR EXISTS(SELECT 1 FROM accessory_reservations WHERE product_id = ?)
				OR EXISTS(SELECT 1 FROM accessory_installations WHERE product_id = ?)
				OR EXISTS(SELECT 1 FROM layout_technical_positions WHERE product_id = ?)
		`, id, id, id, id, id, id, id).Scan(&blocked)
		if err != nil {
			return fmt.Errorf("check accessory product deletion references: %w", err)
		}
		if blocked {
			return application.ErrAccessoryDeleteBlocked
		}

		rows, err := tx.QueryContext(ctx,
			`SELECT file_blob_id FROM accessory_documents WHERE product_id = ? ORDER BY id`, id)
		if err != nil {
			return fmt.Errorf("list accessory product blobs: %w", err)
		}
		for rows.Next() {
			var blobID string
			if err := rows.Scan(&blobID); err != nil {
				return fmt.Errorf("scan accessory product blob: %w", err)
			}
			blobIDs = append(blobIDs, blobID)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate accessory product blobs: %w", err)
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("close accessory product blobs: %w", err)
		}

		for _, statement := range []string{
			`DELETE FROM accessory_documents WHERE product_id = ?`,
			`DELETE FROM accessory_stock WHERE product_id = ?`,
			`DELETE FROM accessory_products WHERE id = ?`,
		} {
			if _, err := tx.ExecContext(ctx, statement, id); err != nil {
				return fmt.Errorf("delete accessory product data: %w", err)
			}
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryProductDeleted",
			"accessory_product", id, actor, time.Now().UTC().Format(time.RFC3339Nano), "{}")
	})
	if err != nil {
		return nil, err
	}
	return blobIDs, nil
}
```

Imports: `context`, `database/sql`, `fmt`, `time`, `railkeeper/backend/internal/application`.

- [ ] **Step 4: Den Produktions-Service mit dem bestehenden Blob-Cleaner verdrahten**

Ändere in `backend/cmd/railkeeper/main.go` ausschließlich die Konstruktion:

```go
accessoryService := application.NewAccessoryService(accessoryRepository, fileBlobService)
```

Alle bestehenden Ein-Argument-Aufrufe bleiben durch den variadischen Konstruktor gültig.

- [ ] **Step 5: Formatieren und die Persistenzschicht GREEN prüfen**

Run:

```powershell
gofmt -w internal/infrastructure/accessory_product_deletion.go `
  internal/infrastructure/accessory_product_deletion_test.go cmd/railkeeper/main.go
go test ./internal/application ./internal/infrastructure
```

Expected: PASS, einschließlich aller sieben Sperrgründe, Metadatenlöschung und Audit.

- [ ] **Step 6: Die Persistenzänderung committen**

```powershell
git add backend/internal/infrastructure/accessory_product_deletion.go `
  backend/internal/infrastructure/accessory_product_deletion_test.go backend/cmd/railkeeper/main.go
git commit -m "feat(accessories): safely delete unused products"
```

### Task 3: Admin-DELETE-Route und OpenAPI-Vertrag

**Files:**
- Modify: `backend/internal/api/accessory_handlers.go`
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/internal/api/accessory_handlers_test.go`
- Modify: `openapi/railkeeper.yaml`

**Interfaces:**
- Consumes: `AccessoryService.DeleteProduct(context.Context, string, string) error`.
- Produces: `DELETE /api/v1/accessory-products/{id}` mit 204, 403, 404 und 409.
- Produces: Problem-Code `accessory_delete_blocked` für einen fachlich nicht löschbaren Artikel.

- [ ] **Step 1: Rollen-, CSRF- und Statusverhalten als RED-Test ergänzen**

Füge einen fokussierten Test hinzu, der Admin- und Editor-Sitzung mit den vorhandenen Hilfen erstellt, zwei unbenutzte Produkte per `POST` anlegt und diese Assertions ausführt:

```go
func TestAccessoryProductDeleteRequiresAdminAndCSRF(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "admin-delete", Password: "admin-password", Roles: []string{"Admin"}},
		{Username: "editor-delete", Password: "editor-password", Roles: []string{"Editor"}},
		{Username: "viewer-delete", Password: "viewer-password", Roles: []string{"Viewer"}},
		{Username: "planner-delete", Password: "planner-password", Roles: []string{"Planner"}},
		{Username: "messe-delete", Password: "messe-password", Roles: []string{"Messe"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	router := NewRouter(Config{
		AuthService: auth,
		AccessoryService: application.NewAccessoryService(
			infrastructure.NewAccessoryRepository(db),
		),
	})
	admin := loginRouteTestUser(t, auth, "admin-delete", "admin-password")
	nonAdmins := []*application.LoginResult{
		loginRouteTestUser(t, auth, "editor-delete", "editor-password"),
		loginRouteTestUser(t, auth, "viewer-delete", "viewer-password"),
		loginRouteTestUser(t, auth, "planner-delete", "planner-password"),
		loginRouteTestUser(t, auth, "messe-delete", "messe-password"),
	}
	create := func(articleNumber string) application.AccessoryProduct {
		response := layoutRequest(t, router, admin, http.MethodPost,
			"/api/v1/accessory-products", map[string]any{
				"manufacturer": "Tillig", "articleNumber": articleNumber,
				"name": "Löschtest", "category": "Gleismaterial", "trackingMode": "quantity",
			}, true)
		assertStatus(t, response, http.StatusCreated)
		var product application.AccessoryProduct
		decodeResponse(t, response, &product)
		return product
	}

	protected := create("DELETE-CSRF")
	path := "/api/v1/accessory-products/" + protected.ID
	for _, session := range nonAdmins {
		assertProblem(t, layoutRequest(t, router, session, http.MethodDelete, path, nil, true),
			http.StatusForbidden, "forbidden")
	}
	assertProblem(t, layoutRequest(t, router, admin, http.MethodDelete, path, nil, false),
		http.StatusForbidden, "csrf_required")
	assertStatus(t, layoutRequest(t, router, admin, http.MethodDelete, path, nil, true),
		http.StatusNoContent)
	assertProblem(t, layoutRequest(t, router, admin, http.MethodGet, path, nil, true),
		http.StatusNotFound, "accessory_not_found")
	assertProblem(t, layoutRequest(t, router, admin, http.MethodDelete,
		"/api/v1/accessory-products/missing", nil, true),
		http.StatusNotFound, "accessory_not_found")
}
```

Füge den vollständigen Konflikttest hinzu:

```go
func TestAccessoryProductDeleteRejectsUsedProduct(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "admin-blocked-delete", Password: "admin-password", Roles: []string{"Admin"},
	}); err != nil {
		t.Fatal(err)
	}
	admin := loginRouteTestUser(t, auth, "admin-blocked-delete", "admin-password")
	router := NewRouter(Config{
		AuthService: auth,
		AccessoryService: application.NewAccessoryService(
			infrastructure.NewAccessoryRepository(db),
		),
	})
	locationResponse := layoutRequest(t, router, admin, http.MethodPost,
		"/api/v1/storage-locations", map[string]any{"name": "Sperrlager"}, true)
	assertStatus(t, locationResponse, http.StatusCreated)
	var location application.StorageLocation
	decodeResponse(t, locationResponse, &location)

	productResponse := layoutRequest(t, router, admin, http.MethodPost,
		"/api/v1/accessory-products", map[string]any{
			"manufacturer": "Tillig", "articleNumber": "DELETE-BLOCKED", "name": "Benutzter Artikel",
			"category": "Gleismaterial", "trackingMode": "quantity",
		}, true)
	assertStatus(t, productResponse, http.StatusCreated)
	var product application.AccessoryProduct
	decodeResponse(t, productResponse, &product)
	path := "/api/v1/accessory-products/" + product.ID
	assertStatus(t, layoutRequest(t, router, admin, http.MethodPost,
		path+"/stock-adjustments", map[string]any{
			"locationId": location.ID, "delta": 1,
		}, true), http.StatusOK)

	assertProblem(t, layoutRequest(t, router, admin, http.MethodDelete, path, nil, true),
		http.StatusConflict, "accessory_delete_blocked")
	assertStatus(t, layoutRequest(t, router, admin, http.MethodGet, path, nil, true), http.StatusOK)
}
```

- [ ] **Step 2: Die API-Tests im RED-Zustand ausführen**

Run:

```powershell
cd backend
go test ./internal/api -run 'TestAccessoryProductDelete'
```

Expected: FAIL mit 404/405, weil Route und Handler noch fehlen.

- [ ] **Step 3: Handler, Fehlerabbildung und Admin-Route implementieren**

Füge in `accessory_handlers.go` ein:

```go
func (a *App) deleteAccessoryProduct(w http.ResponseWriter, r *http.Request) {
	if err := a.accessoryService.DeleteProduct(
		r.Context(), r.PathValue("id"), actorUserID(r),
	); err != nil {
		a.accessoryError(w, err, "delete accessory product")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

Ergänze in `accessoryError` direkt nach `ErrAccessoryNotFound`:

```go
case errors.Is(err, application.ErrAccessoryDeleteBlocked):
	respondProblem(w, http.StatusConflict, "accessory_delete_blocked",
		"Accessory product has stock or usage history and cannot be deleted.")
```

Füge in `routes.go` direkt nach der PUT-Route hinzu:

```go
{http.MethodDelete, "/api/v1/accessory-products/{id}", routeAccessAdmin,
	(*App).deleteAccessoryProduct, nil},
```

Die vorhandene zentrale Router-Logik übernimmt Session, Adminrolle und CSRF für diese schreibende Route.

- [ ] **Step 4: Den OpenAPI-Pfad vervollständigen**

Ergänze unter `/accessory-products/{id}` neben `get` und `put`:

```yaml
    delete:
      tags: [Accessories]
      summary: Permanently delete an unused accessory product
      description: >-
        Admin role required. Deletion is rejected while stock, assets, purchases,
        stock movements, reservations, installations, usage history, or layout
        technical positions reference the product. Product metadata is deleted.
      security:
        - sessionCookie: []
          csrfHeader: []
      responses:
        "204":
          description: Accessory product deleted
        "403":
          description: Admin role and valid CSRF token required
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Problem"
        "404":
          description: Accessory product not found
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Problem"
        "409":
          description: Accessory product has stock or usage references
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Problem"
```

- [ ] **Step 5: Formatieren und GREEN prüfen**

Run:

```powershell
gofmt -w internal/api/accessory_handlers.go internal/api/routes.go `
  internal/api/accessory_handlers_test.go
go test ./internal/api -run 'TestAccessoryProductDelete'
go test ./internal/api -run 'TestOpenAPIDocumentsRegisteredAPIRoutes|TestFrontendAPIAdapterUsesDocumentedRoutes'
go test ./internal/api
```

Expected: alle drei Läufe PASS; der Contract-Test bestätigt die registrierte DELETE-Operation.

- [ ] **Step 6: Route und Vertrag committen**

```powershell
git add backend/internal/api/accessory_handlers.go backend/internal/api/routes.go `
  backend/internal/api/accessory_handlers_test.go openapi/railkeeper.yaml
git commit -m "feat(api): expose admin accessory deletion"
```

### Task 4: Frontend-Command und Reload-Verhalten

**Files:**
- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/shared/apiLayoutsAccessories.test.ts`
- Modify: `frontend/src/features/accessories/useArticleOverview.ts`
- Modify: `frontend/src/features/accessories/useArticleOverview.test.tsx`

**Interfaces:**
- Consumes: `DELETE /api/v1/accessory-products/{id}`.
- Produces: `api.deleteAccessoryProduct(id: string): Promise<void>`.
- Produces: `useArticleOverview().deleteArticle(id: string): Promise<void>`; Erfolg triggert `reload`, Fehler werden unverändert an den Bestätigungsdialog weitergereicht.

- [ ] **Step 1: API-Aufruf und Hook-Semantik als RED-Tests ergänzen**

Erweitere den bestehenden API-Routentest zwischen Restore und Usage History:

```ts
await api.deleteAccessoryProduct("product/1");
```

und die erwartete Sequenz um:

```ts
["DELETE", "/api/v1/accessory-products/product%2F1"],
```

Ergänze im Hook-Test-Setup:

```ts
vi.spyOn(api, "deleteAccessoryProduct").mockResolvedValue(undefined);
```

und füge zwei Tests hinzu:

```ts
it("deletes an article and reloads the overview", async () => {
  const { result } = renderHook(() => useArticleOverview());
  await waitFor(() => expect(result.current.loading).toBe(false));
  vi.mocked(api.accessoryArticles).mockClear();

  await act(async () => result.current.deleteArticle("product/1"));

  expect(api.deleteAccessoryProduct).toHaveBeenCalledWith("product/1");
  await waitFor(() => expect(api.accessoryArticles).toHaveBeenCalledTimes(1));
});

it("rejects a delete failure without reloading", async () => {
  vi.mocked(api.deleteAccessoryProduct).mockRejectedValueOnce(new Error("Artikel wird verwendet"));
  const { result } = renderHook(() => useArticleOverview());
  await waitFor(() => expect(result.current.loading).toBe(false));

  await expect(act(async () => result.current.deleteArticle("product/1")))
    .rejects.toThrow("Artikel wird verwendet");
  expect(api.accessoryArticles).toHaveBeenCalledTimes(1);
});
```

- [ ] **Step 2: Frontend-Tests im RED-Zustand ausführen**

Run:

```powershell
cd frontend
npm.cmd test -- --run src/shared/apiLayoutsAccessories.test.ts `
  src/features/accessories/useArticleOverview.test.tsx
```

Expected: FAIL, weil API-Methode und Hook-Command fehlen.

- [ ] **Step 3: API und Hook minimal implementieren**

Füge neben Archive und Restore ein:

```ts
deleteAccessoryProduct: (id: string) =>
  request<void>(`/accessory-products/${encodeURIComponent(id)}`, { method: "DELETE" }),
```

Füge im Hook ein:

```ts
const deleteArticle = useCallback(async (id: string) => {
  setError("");
  await api.deleteAccessoryProduct(id);
  reload();
}, [reload]);
```

und gib `deleteArticle` im Return-Objekt zurück.

- [ ] **Step 4: GREEN prüfen**

Run:

```powershell
npm.cmd test -- --run src/shared/apiLayoutsAccessories.test.ts `
  src/features/accessories/useArticleOverview.test.tsx
```

Expected: PASS.

- [ ] **Step 5: Frontend-Command committen**

```powershell
git add frontend/src/shared/apiLayoutsAccessories.ts `
  frontend/src/shared/apiLayoutsAccessories.test.ts `
  frontend/src/features/accessories/useArticleOverview.ts `
  frontend/src/features/accessories/useArticleOverview.test.tsx
git commit -m "feat(accessories): wire article deletion command"
```

### Task 5: Gemeinsame Admin-Löschaktion und Bestätigungsdialog

**Files:**
- Modify: `frontend/src/features/accessories/ArticleActions.tsx`
- Modify: `frontend/src/features/accessories/ArticleActions.test.tsx`
- Modify: `frontend/src/features/accessories/ArticleTable.tsx`
- Modify: `frontend/src/features/accessories/ArticleCardGrid.tsx`
- Modify: `frontend/src/features/accessories/ArticleCompactList.tsx`
- Modify: `frontend/src/features/accessories/AccessoriesView.tsx`
- Modify: `frontend/src/features/accessories/AccessoriesView.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/accessories.css`

**Interfaces:**
- Consumes: `useArticleOverview().deleteArticle(id: string): Promise<void>`.
- Produces: optionale Präsentations-Props `canDelete?: boolean` und `onDelete?: (article: AccessoryArticleListItem) => void`.
- Produces: gefährlicher `AccessoryConfirmDialog`, der bei Erfolg schließt und bei 409 offen bleibt.

- [ ] **Step 1: Gemeinsame Aktion, Rollen und Fokusnavigation als RED-Tests festschreiben**

Ergänze in `ArticleActions.test.tsx`:

```tsx
it("shows delete only when explicitly allowed and routes the selected article", async () => {
  const user = userEvent.setup();
  const onDelete = vi.fn();
  render(<ArticleActions article={article} canEdit canDelete onView={vi.fn()} onEdit={vi.fn()}
    onArchive={vi.fn()} onRestore={vi.fn()} onDelete={onDelete} />);

  await user.click(screen.getByRole("button", { name: /Weitere Aktionen/ }));
  await user.click(screen.getByRole("menuitem", { name: "Artikel löschen" }));

  expect(onDelete).toHaveBeenCalledWith(article);
});

it("does not expose delete to editors", async () => {
  const user = userEvent.setup();
  render(<ArticleActions article={article} canEdit canDelete={false} onView={vi.fn()}
    onEdit={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()} onDelete={vi.fn()} />);

  await user.click(screen.getByRole("button", { name: /Weitere Aktionen/ }));
  expect(screen.queryByRole("menuitem", { name: "Artikel löschen" })).not.toBeInTheDocument();
});
```

Erweitere den bestehenden Fokus-Test mit `canDelete` und `onDelete`. Prüfe danach zwei Menüelemente:

```tsx
const [archiveItem, deleteItem] = screen.getAllByRole("menuitem");
expect(archiveItem).toHaveFocus();
await user.keyboard("{ArrowDown}");
expect(deleteItem).toHaveFocus();
await user.keyboard("{ArrowDown}");
expect(archiveItem).toHaveFocus();
await user.keyboard("{End}");
expect(deleteItem).toHaveFocus();
await user.keyboard("{Home}");
expect(archiveItem).toHaveFocus();
```

Passe bestehende Renders entweder mit `canDelete={false}` an oder lasse den neuen Prop optional.

- [ ] **Step 2: View-Dialog und Fehlerzustand als RED-Tests ergänzen**

Ergänze im `beforeEach` von `AccessoriesView.test.tsx`:

```ts
vi.spyOn(api, "deleteAccessoryProduct").mockResolvedValue(undefined);
```

Füge folgende Tests mit dem bestehenden Artikel-Fixture und den vorhandenen API-Mocks hinzu:

```tsx
it("keeps permanent deletion hidden from editors", async () => {
  const user = userEvent.setup();
  render(<AccessoriesView roles={["Editor"]} />);
  await screen.findByText("Gerades Modellgleis");

  await user.click(screen.getByRole("button", { name: /Weitere Aktionen/ }));
  expect(screen.queryByRole("menuitem", { name: "Artikel löschen" })).not.toBeInTheDocument();
});

it("confirms admin deletion with article identity and reloads", async () => {
  const user = userEvent.setup();
  render(<AccessoriesView roles={["Admin"]} />);
  await screen.findByText("Gerades Modellgleis");

  await user.click(screen.getByRole("button", { name: /Weitere Aktionen/ }));
  await user.click(screen.getByRole("menuitem", { name: "Artikel löschen" }));
  const dialog = screen.getByRole("dialog", { name: "Artikel endgültig löschen" });
  expect(within(dialog).getByText(/RK-ART-000001/)).toBeInTheDocument();
  expect(within(dialog).getByText(/Gerades Modellgleis/)).toBeInTheDocument();
  await user.click(within(dialog).getByRole("button", { name: "Endgültig löschen" }));

  expect(api.deleteAccessoryProduct).toHaveBeenCalledWith("article-1");
  await waitFor(() => expect(api.accessoryArticles).toHaveBeenCalledTimes(2));
  expect(screen.queryByRole("dialog", { name: "Artikel endgültig löschen" }))
    .not.toBeInTheDocument();
});

it("keeps the dialog and article visible when deletion is blocked", async () => {
  vi.mocked(api.deleteAccessoryProduct).mockRejectedValueOnce(
    new Error("Accessory product has stock or usage history and cannot be deleted."),
  );
  const user = userEvent.setup();
  render(<AccessoriesView roles={["Admin"]} />);
  await screen.findByText("Gerades Modellgleis");

  await user.click(screen.getByRole("button", { name: /Weitere Aktionen/ }));
  await user.click(screen.getByRole("menuitem", { name: "Artikel löschen" }));
  await user.click(screen.getByRole("button", { name: "Endgültig löschen" }));

  expect(await screen.findByRole("alert")).toHaveTextContent("stock or usage history");
  expect(screen.getByRole("dialog", { name: "Artikel endgültig löschen" })).toBeInTheDocument();
  expect(screen.getByText("Gerades Modellgleis")).toBeInTheDocument();
});

it("localizes the admin delete action and confirmation in English", async () => {
  setLanguage("en");
  const user = userEvent.setup();
  render(<AccessoriesView roles={["Admin"]} />);
  await screen.findByText("Gerades Modellgleis");

  await user.click(screen.getByRole("button", { name: /More actions/ }));
  await user.click(screen.getByRole("menuitem", { name: "Delete article" }));

  const dialog = screen.getByRole("dialog", { name: "Permanently delete article" });
  expect(within(dialog).getByRole("button", { name: "Delete permanently" })).toBeInTheDocument();
  expect(within(dialog).getByText(/RK-ART-000001: Gerades Modellgleis/)).toBeInTheDocument();
});
```

Importiere `within` aus Testing Library. Der bestehende Fixture-Vertrag verwendet die ID
`article-1` und die Bestandsnummer `RK-ART-000001`; die Assertions bleiben exakt darauf festgelegt.

- [ ] **Step 3: RED für Aktions- und View-Tests prüfen**

Run:

```powershell
cd frontend
npm.cmd test -- --run src/features/accessories/ArticleActions.test.tsx `
  src/features/accessories/AccessoriesView.test.tsx
```

Expected: FAIL, weil Lösch-Props, Texte und Dialogzustand fehlen.

- [ ] **Step 4: Die gemeinsame Löschaktion und korrekte Menü-Navigation implementieren**

Erweitere `ArticleActionsProps`:

```ts
canDelete?: boolean;
onDelete?: (article: AccessoryArticleListItem) => void;
```

Destrukturiere beide Props mit `canDelete = false`. Zeige den Overflow-Bereich bei `canEdit || canDelete`. Rendere Archiv/Wiederherstellen nur bei `canEdit` und direkt danach:

```tsx
{canDelete && onDelete ? (
  <button
    type="button"
    role="menuitem"
    tabIndex={-1}
    className="danger-menu-item"
    onClick={() => {
      setOpen(false);
      onDelete(article);
    }}
  >
    {t("accessories.actions.delete")}
  </button>
) : null}
```

Setze bei allen Menüelementen `tabIndex={-1}`; der Öffnungseffekt fokussiert weiterhin das erste. Ersetze `handleMenuKeyDown` vollständig durch:

```ts
const handleMenuKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
  if (event.key === "Tab") {
    setOpen(false);
    return;
  }
  if (!["ArrowDown", "ArrowUp", "Home", "End"].includes(event.key)) return;
  event.preventDefault();
  const items = Array.from(
    menuRef.current?.querySelectorAll<HTMLButtonElement>("[role='menuitem']") || [],
  );
  if (items.length === 0) return;
  const current = Math.max(0, items.indexOf(document.activeElement as HTMLButtonElement));
  const next = event.key === "Home" ? 0
    : event.key === "End" ? items.length - 1
      : event.key === "ArrowDown" ? (current + 1) % items.length
        : (current - 1 + items.length) % items.length;
  items[next]?.focus();
};
```

- [ ] **Step 5: Delete-Props durch alle drei Präsentationen führen**

Ergänze in `ArticleTableProps`, `ArticleCardGridProps` und `ArticleCompactListProps` jeweils:

```ts
canDelete?: boolean;
onDelete?: (article: AccessoryArticleListItem) => void;
```

Destrukturiere `canDelete = false` und `onDelete`. Übergib in jedem vorhandenen `<ArticleActions>`:

```tsx
canDelete={canDelete}
onDelete={onDelete}
```

Die drei Komponenten enthalten keine eigene Rollenentscheidung.

- [ ] **Step 6: Adminzustand und Bestätigungsdialog in der View implementieren**

Importiere `AccessoryConfirmDialog` und ergänze:

```ts
const canDelete = roles.includes("Admin");
const [pendingDeleteArticle, setPendingDeleteArticle] =
  useState<AccessoryArticleListItem | null>(null);
```

Übergib an Tabelle, Kacheln und Kompaktliste:

```tsx
canDelete={canDelete}
onDelete={setPendingDeleteArticle}
```

Rendere einmal am Ende der View, außerhalb des Listen-Panels:

```tsx
<AccessoryConfirmDialog
  action={pendingDeleteArticle ? {
    title: t("accessories.delete.title"),
    body: t("accessories.delete.body", {
      inventoryNumber: pendingDeleteArticle.inventoryNumber,
      name: pendingDeleteArticle.name,
    }),
    confirmLabel: t("accessories.delete.confirm"),
    dangerous: true,
    run: () => overview.deleteArticle(pendingDeleteArticle.id),
  } : null}
  onClose={() => setPendingDeleteArticle(null)}
/>
```

Der vorhandene `AccessoryConfirmDialog` schließt nur nach erfolgreichem Promise. Bei 409 zeigt er den Fehler als `role="alert"` und bleibt offen.

- [ ] **Step 7: Deutsche und englische Texte sowie Gefahrfarbe ergänzen**

Ergänze in beiden Wörterbüchern neben Archive/Restore:

```ts
// de.ts
"accessories.actions.delete": "Artikel löschen",
"accessories.delete.title": "Artikel endgültig löschen",
"accessories.delete.body":
  "Soll {inventoryNumber}: {name} endgültig gelöscht werden? Dies ist nur bei vollständig unbenutzten Artikeln möglich.",
"accessories.delete.confirm": "Endgültig löschen",

// en.ts
"accessories.actions.delete": "Delete article",
"accessories.delete.title": "Permanently delete article",
"accessories.delete.body":
  "Permanently delete {inventoryNumber}: {name}? This is only possible for completely unused articles.",
"accessories.delete.confirm": "Delete permanently",
```

Ergänze in `accessories.css` direkt bei den Menübutton-Regeln:

```css
.article-action-menu .danger-menu-item {
  color: var(--danger);
}
```

- [ ] **Step 8: UI-Tests GREEN prüfen**

Run:

```powershell
npm.cmd test -- --run src/features/accessories/ArticleActions.test.tsx `
  src/features/accessories/AccessoriesView.test.tsx `
  src/features/accessories/ArticleTable.test.tsx `
  src/features/accessories/ArticleCardGrid.test.tsx `
  src/features/accessories/ArticleCompactList.test.tsx
```

Expected: PASS. Wenn Präsentationstests streng typisierte Pflichtprops verwenden, `canDelete={false}` und `onDelete={vi.fn()}` explizit ergänzen.

- [ ] **Step 9: Die UI-Funktion committen**

```powershell
git add frontend/src/features/accessories/ArticleActions.tsx `
  frontend/src/features/accessories/ArticleActions.test.tsx `
  frontend/src/features/accessories/ArticleTable.tsx `
  frontend/src/features/accessories/ArticleCardGrid.tsx `
  frontend/src/features/accessories/ArticleCompactList.tsx `
  frontend/src/features/accessories/AccessoriesView.tsx `
  frontend/src/features/accessories/AccessoriesView.test.tsx `
  frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts `
  frontend/src/styles/accessories.css
git commit -m "feat(accessories): add admin delete action"
```

### Task 6: Gesamtverifikation und Browser-Abnahme

**Files:**
- Verify: alle in Tasks 1 bis 5 geänderten Dateien
- Do not modify: `frontend/dist`, `data`, lokale Datenbank- oder Upload-Dateien

**Interfaces:**
- Consumes: vollständige Backend-, API- und Frontend-Implementierung.
- Produces: reproduzierbare Testnachweise und visuelle Abnahme ohne Löschung realer Nutzerdaten.

- [ ] **Step 1: Backend vollständig prüfen**

Run:

```powershell
cd backend
gofmt -w internal/application/accessories.go internal/application/accessories_test.go `
  internal/infrastructure/accessory_product_deletion.go `
  internal/infrastructure/accessory_product_deletion_test.go `
  internal/api/accessory_handlers.go internal/api/accessory_handlers_test.go `
  internal/api/routes.go cmd/railkeeper/main.go
go test ./...
```

Expected: PASS für alle Backend-Pakete.

- [ ] **Step 2: Frontend vollständig prüfen und Produktionsbundle bauen**

Run:

```powershell
cd ..\frontend
npm.cmd test -- --run
npm.cmd run build
```

Expected: alle Vitest-Tests PASS; TypeScript/Vite-Build erfolgreich.

- [ ] **Step 3: Vertrag und Diff statisch prüfen**

Run:

```powershell
cd ..
rg -n "accessory_delete_blocked|deleteAccessoryProduct|AccessoryProductDeleted|accessories.actions.delete" `
  backend frontend/src openapi/railkeeper.yaml
git diff --check
git status --short
```

Expected: Route, Fehlercode, Audit und UI-Text sind in ihren jeweiligen Schichten vorhanden; `git diff --check` meldet nichts; nur beabsichtigte Dateien sind geändert.

- [ ] **Step 4: Browser-Verhalten gegen den lokalen Server prüfen**

Prüfe zuerst den bereits laufenden lokalen Server:

```powershell
Invoke-WebRequest http://127.0.0.1:18084/health -UseBasicParsing
```

Expected: HTTP 200. Ist auf Port 18084 kein Server aktiv, starte ihn in einem normalen PowerShell-
Fenster mit diesen exakten Einstellungen:

```powershell
$railKeeperRepo = 'C:\Users\droth\Documents\GitHub\RailKeeper'
$env:RAILKEEPER_ADDR = ':18084'
$env:RAILKEEPER_DATA_DIR = "$railKeeperRepo\data"
$env:RAILKEEPER_MIGRATIONS_DIR = "$railKeeperRepo\backend\migrations"
$env:RAILKEEPER_SEEDS_DIR = "$railKeeperRepo\backend\seeds"
$env:RAILKEEPER_STATIC_DIR = "$railKeeperRepo\frontend\dist"
$env:GOCACHE = "$railKeeperRepo\.cache\go-build"
Set-Location "$railKeeperRepo\backend"
go run ./cmd/railkeeper
```

Öffne beziehungsweise aktualisiere anschließend `http://127.0.0.1:18084/accessories` und prüfe:

1. Admin sieht im Drei-Punkte-Menü in Tabellen-, Kachel- und Mobilansicht `Artikel löschen`.
2. Editor sieht Archive/Wiederherstellen, aber keine Löschaktion.
3. Der Dialog nennt Bestandsnummer und Artikelname, nutzt den gefährlichen Button und lässt sich mit Abbrechen oder Escape schließen.
4. Fokus wandert mit Pfeiltasten, Home und End korrekt zwischen Archive/Wiederherstellen und Löschen.
5. Light/Dark und schmale Breite behalten lesbare Menü- und Dialogdarstellung.

Keine Löschung eines bestehenden Nutzerartikels bestätigen. Die tatsächliche erfolgreiche Löschung und alle Sperrfälle werden ausschließlich gegen die temporären Testdatenbanken aus Task 2 und 3 verifiziert.

- [ ] **Step 5: Abschlussstatus kontrollieren**

Run:

```powershell
git log -7 --oneline
git status --short
```

Expected: fünf fokussierte Feature-Commits aus Tasks 1 bis 5 sowie Design- und Plan-Commit; der
Worktree ist sauber oder enthält ausschließlich vorbestehende, eindeutig fremde Nutzeränderungen.
