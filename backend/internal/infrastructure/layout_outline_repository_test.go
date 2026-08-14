package infrastructure_test

import (
	"errors"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

func TestLayoutUnitOutlineRepositoryReplacesPointsAndVersionsUnit(t *testing.T) {
	db, service := testLayoutPositionService(t)
	layout, err := service.CreateLayout(t.Context(), application.CreateLayoutInput{
		Name: "Anlage", Kind: domain.LayoutKindPrivate, Gauge: "TT", Scale: "1:120",
	}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := service.CreateUnit(t.Context(), layout.ID, application.CreateLayoutUnitInput{
		Name: "Modul", Kind: domain.LayoutUnitKindModule, WidthMM: 100, HeightMM: 50,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	points := []application.LayoutTwinPoint{
		{XMM: 0, YMM: 0}, {XMM: 100, YMM: 0}, {XMM: 80, YMM: 60}, {XMM: 0, YMM: 50},
	}
	outline, err := service.UpdateUnitOutline(t.Context(), unit.ID,
		application.UpdateLayoutUnitOutlineInput{Points: points, ExpectedVersion: unit.Version}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if outline.Version != 2 {
		t.Fatalf("unexpected outline version: %#v", outline)
	}
	var count, auditCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM layout_unit_outline_points WHERE layout_unit_id=?`, unit.ID).
		Scan(&count); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs
WHERE action='LayoutUnitOutlineUpdated' AND target_id=?`, unit.ID).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if count != 4 || auditCount != 1 {
		t.Fatalf("unexpected persistence: points=%d audit=%d", count, auditCount)
	}
	_, err = service.UpdateUnitOutline(t.Context(), unit.ID,
		application.UpdateLayoutUnitOutlineInput{Points: points, ExpectedVersion: unit.Version}, "planner")
	if !errors.Is(err, application.ErrLayoutVersionConflict) {
		t.Fatalf("expected version conflict, got %v", err)
	}
	twin, err := service.GetTwin(t.Context(), layout.ID, application.LayoutTwinSelection{UnitID: unit.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(twin.Units) != 1 || twin.Units[0].Version != 2 || len(twin.Units[0].LocalOutline) != 4 ||
		twin.Units[0].LocalOutline[2].YMM != 60 {
		t.Fatalf("twin lost custom outline: %#v", twin)
	}
}
