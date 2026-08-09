package infrastructure_test

import (
	"math"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestLayoutTwinTransformsConfigurationAndAggregatesLiveStatuses(t *testing.T) {
	fixture := newAllocationFixture(t)
	ctx := t.Context()
	layouts := application.NewLayoutService(infrastructure.NewLayoutRepository(fixture.db))
	if _, err := layouts.UpdateUnit(ctx, fixture.unit.ID, application.UpdateLayoutUnitInput{
		CreateLayoutUnitInput: application.CreateLayoutUnitInput{
			Name: fixture.unit.Name, Kind: fixture.unit.Kind, WidthMM: 100, HeightMM: 50,
		},
		ExpectedVersion: fixture.unit.Version,
	}, "planner-1"); err != nil {
		t.Fatal(err)
	}
	positionInputs := []application.CreateLayoutTechnicalPositionInput{
		{Label: "Signal reserviert", Kind: domain.LayoutPositionSignal,
			PositionXMM: 10, PositionYMM: 20, ProductID: fixture.quantityProduct.ID},
		{Label: "Signal defekt", Kind: domain.LayoutPositionSignal,
			PositionXMM: 20, PositionYMM: 20, ProductID: fixture.quantityProduct.ID},
		{Label: "Signal geplant", Kind: domain.LayoutPositionSignal,
			PositionXMM: 30, PositionYMM: 20, ProductID: fixture.quantityProduct.ID},
	}
	positions := make([]*application.LayoutTechnicalPosition, 0, len(positionInputs))
	for _, input := range positionInputs {
		position, err := layouts.CreateTechnicalPosition(ctx, fixture.unit.ID, input, "planner-1")
		if err != nil {
			t.Fatal(err)
		}
		positions = append(positions, position)
	}
	target := application.AllocationTargetInput{LayoutUnitID: fixture.unit.ID}
	if _, err := fixture.allocations.CreateReservation(ctx, application.CreateAccessoryReservationInput{
		ProductID: fixture.quantityProduct.ID, LocationID: fixture.location.ID, Quantity: 1,
		AllocationTargetInput: target, TechnicalPositionID: positions[0].ID,
	}, "planner-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.allocations.Install(ctx, application.CreateAccessoryInstallationInput{
		ProductID: fixture.quantityProduct.ID, SourceLocationID: fixture.location.ID, Quantity: 1,
		AllocationTargetInput: target, TechnicalPositionID: positions[1].ID,
		Condition: domain.AccessoryConditionDefective,
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	configuration, err := layouts.SaveConfiguration(ctx, fixture.layout.ID,
		application.SaveLayoutConfigurationInput{
			Name: "Ausstellung", Units: []application.ConfigurationUnitInput{{
				UnitID: fixture.unit.ID, PositionXMM: 100, PositionYMM: 200, RotationDegrees: 90,
			}},
		}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}

	twin, err := layouts.GetTwin(ctx, fixture.layout.ID,
		application.LayoutTwinSelection{ConfigurationID: configuration.ID})
	if err != nil {
		t.Fatal(err)
	}
	if twin.ConfigurationName != "Ausstellung" || len(twin.Units) != 1 || !twin.HasGeometry {
		t.Fatalf("unexpected twin header: %#v", twin)
	}
	if !near(twin.Bounds.MinXMM, 50) || !near(twin.Bounds.MinYMM, 200) ||
		!near(twin.Bounds.WidthMM, 50) || !near(twin.Bounds.HeightMM, 100) {
		t.Fatalf("unexpected transformed bounds: %#v", twin.Bounds)
	}
	got := twin.Units[0].Positions
	if len(got) != 3 {
		t.Fatalf("unexpected positions: %#v", got)
	}
	byLabel := map[string]application.LayoutTwinPosition{}
	for _, position := range got {
		byLabel[position.Label] = position
	}
	reserved := byLabel["Signal reserviert"]
	if !near(reserved.GlobalXMM, 80) || !near(reserved.GlobalYMM, 210) ||
		len(reserved.Statuses) != 1 || reserved.Statuses[0] != application.LayoutTwinReserved ||
		len(reserved.Reservations) != 1 {
		t.Fatalf("unexpected reserved position: %#v", reserved)
	}
	defective := byLabel["Signal defekt"]
	if len(defective.Statuses) != 2 || defective.Statuses[0] != application.LayoutTwinInstalled ||
		defective.Statuses[1] != application.LayoutTwinDefective || len(defective.Installations) != 1 {
		t.Fatalf("unexpected defective position: %#v", defective)
	}
	planned := byLabel["Signal geplant"]
	if len(planned.Statuses) != 1 || planned.Statuses[0] != application.LayoutTwinPlanned {
		t.Fatalf("unexpected planned position: %#v", planned)
	}
}

func near(got float64, want float64) bool {
	return math.Abs(got-want) < 0.000001
}
