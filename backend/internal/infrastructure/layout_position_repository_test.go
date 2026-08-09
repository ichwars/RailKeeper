package infrastructure_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestLayoutTechnicalPositionRepositoryPersistsAndVersionsPositions(t *testing.T) {
	db, service := testLayoutPositionService(t)
	ctx := t.Context()

	layout, err := service.CreateLayout(ctx, application.CreateLayoutInput{
		Name: "Clubanlage", Kind: domain.LayoutKindClub, Gauge: "TT", Scale: "1:120",
	}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := service.CreateUnit(ctx, layout.ID, application.CreateLayoutUnitInput{
		Name: "Bahnhof", Kind: domain.LayoutUnitKindModule, WidthMM: 1200, HeightMM: 500,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO accessory_products(
      id, inventory_number, manufacturer, name, category, tracking_mode, created_at, updated_at
    ) VALUES('product-1', 'RK-ART-POS-1', 'Tillig', 'Signal', 'signal', 'quantity', 'now', 'now')`); err != nil {
		t.Fatal(err)
	}

	position, err := service.CreateTechnicalPosition(ctx, unit.ID, application.CreateLayoutTechnicalPositionInput{
		Label: "Signal B", Kind: domain.LayoutPositionSignal, PositionXMM: 100, PositionYMM: 50,
		RotationDegrees: -30, ProductID: "product-1", Description: "Ausfahrt",
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if position.Version != 1 || position.RotationDegrees != 330 || position.ProductID != "product-1" {
		t.Fatalf("unexpected created position: %#v", position)
	}
	if _, err := service.CreateTechnicalPosition(ctx, unit.ID, application.CreateLayoutTechnicalPositionInput{
		Label: "Signal A", Kind: domain.LayoutPositionSignal,
	}, "planner"); err != nil {
		t.Fatal(err)
	}

	positions, err := service.ListTechnicalPositions(ctx, unit.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(positions) != 2 || positions[0].Label != "Signal A" || positions[1].Label != "Signal B" {
		t.Fatalf("unexpected sorted positions: %#v", positions)
	}

	updated, err := service.UpdateTechnicalPosition(ctx, position.ID, application.UpdateLayoutTechnicalPositionInput{
		CreateLayoutTechnicalPositionInput: application.CreateLayoutTechnicalPositionInput{
			Label: "Signal B2", Kind: domain.LayoutPositionSignal, PositionXMM: 110, PositionYMM: 55,
			RotationDegrees: 15, ProductID: "product-1", Description: "Ausfahrt", Archived: true,
		},
		ExpectedVersion: position.Version,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != 2 || !updated.Archived || updated.Label != "Signal B2" {
		t.Fatalf("unexpected updated position: %#v", updated)
	}
	_, err = service.UpdateTechnicalPosition(ctx, position.ID, application.UpdateLayoutTechnicalPositionInput{
		CreateLayoutTechnicalPositionInput: application.CreateLayoutTechnicalPositionInput{
			Label: "Stale", Kind: domain.LayoutPositionSignal,
		},
		ExpectedVersion: position.Version,
	}, "planner")
	if !errors.Is(err, application.ErrLayoutPositionVersionConflict) {
		t.Fatalf("expected version conflict, got %v", err)
	}

	var createdAudit, updatedAudit int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs
      WHERE action='LayoutTechnicalPositionCreated' AND target_id=?`, position.ID).Scan(&createdAudit); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs
      WHERE action='LayoutTechnicalPositionUpdated' AND target_id=?`, position.ID).Scan(&updatedAudit); err != nil {
		t.Fatal(err)
	}
	if createdAudit != 1 || updatedAudit != 1 {
		t.Fatalf("unexpected audit counts: created=%d updated=%d", createdAudit, updatedAudit)
	}
}

func TestLayoutTechnicalPositionRepositoryRejectsMissingReferences(t *testing.T) {
	_, service := testLayoutPositionService(t)
	input := application.CreateLayoutTechnicalPositionInput{
		Label: "Signal", Kind: domain.LayoutPositionSignal,
	}
	if _, err := service.CreateTechnicalPosition(t.Context(), "missing", input, "planner"); !errors.Is(err, application.ErrLayoutNotFound) {
		t.Fatalf("expected missing unit error, got %v", err)
	}

	layout, err := service.CreateLayout(t.Context(), application.CreateLayoutInput{
		Name: "Anlage", Kind: domain.LayoutKindPrivate, Gauge: "TT", Scale: "1:120",
	}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := service.CreateUnit(t.Context(), layout.ID, application.CreateLayoutUnitInput{
		Name: "Platte", Kind: domain.LayoutUnitKindBaseboard,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	input.ProductID = "missing"
	if _, err := service.CreateTechnicalPosition(t.Context(), unit.ID, input, "planner"); !errors.Is(err, application.ErrLayoutPositionProductNotFound) {
		t.Fatalf("expected missing product error, got %v", err)
	}
}

func testLayoutPositionService(t *testing.T) (*sql.DB, *application.LayoutService) {
	t.Helper()
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	return db, application.NewLayoutService(infrastructure.NewLayoutRepository(db))
}
