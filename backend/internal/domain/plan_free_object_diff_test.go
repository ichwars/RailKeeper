package domain

import "testing"

func TestCompareFreePlanObjectRevisionsReportsSortedChanges(t *testing.T) {
	rectangle := FreePlanObjectShape{
		SchemaVersion: 1, Kind: FreePlanRectangle, WidthMM: freeFloat(300), HeightMM: freeFloat(80),
	}
	changed := rectangle
	changed.WidthMM = freeFloat(320)
	base := []PlanFreeObject{
		{ID: "remove", LineageID: "b-remove", Name: "Entfernt", Category: FreePlanStructure, Shape: rectangle},
		{ID: "old", LineageID: "c-change", Name: "Bahnsteig", Category: FreePlanPlatform, Shape: rectangle},
		{ID: "same-old", LineageID: "d-same", Name: "Gleich", Category: FreePlanScenery, Shape: rectangle},
	}
	current := []PlanFreeObject{
		{ID: "add", LineageID: "a-add", Name: "Neu", Category: FreePlanAnnotation, Shape: rectangle},
		{ID: "new", LineageID: "c-change", Name: "Bahnsteig", Category: FreePlanPlatform, Shape: changed},
		{ID: "same-new", LineageID: "d-same", Name: "Gleich", Category: FreePlanScenery, Shape: rectangle},
	}

	diff := CompareFreePlanObjectRevisions(base, current)
	if len(diff) != 3 {
		t.Fatalf("expected three changes, got %#v", diff)
	}
	if diff[0].LineageID != "a-add" || diff[0].Type != TrackPlanObjectAdded || diff[0].After == nil {
		t.Fatalf("unexpected added change: %#v", diff[0])
	}
	if diff[1].LineageID != "b-remove" || diff[1].Type != TrackPlanObjectRemoved || diff[1].Before == nil {
		t.Fatalf("unexpected removed change: %#v", diff[1])
	}
	if diff[2].LineageID != "c-change" || diff[2].Type != TrackPlanObjectChanged ||
		diff[2].Before == nil || diff[2].After == nil {
		t.Fatalf("unexpected changed entry: %#v", diff[2])
	}
}

func TestCompareFreePlanObjectRevisionsDetectsEveryMeaningfulField(t *testing.T) {
	base := PlanFreeObject{
		ID: "old", LineageID: "lineage", Name: "Objekt", Category: FreePlanStructure,
		PositionXMM: 10, PositionYMM: 20, RotationDegrees: 30,
		Shape: FreePlanObjectShape{
			SchemaVersion: 1, Kind: FreePlanRectangle, WidthMM: freeFloat(100), HeightMM: freeFloat(50),
		},
	}
	mutations := []func(*PlanFreeObject){
		func(object *PlanFreeObject) { object.Name = "Anders" },
		func(object *PlanFreeObject) { object.Category = FreePlanPlatform },
		func(object *PlanFreeObject) { object.PositionXMM++ },
		func(object *PlanFreeObject) { object.PositionYMM++ },
		func(object *PlanFreeObject) { object.RotationDegrees++ },
		func(object *PlanFreeObject) { object.Shape.HeightMM = freeFloat(55) },
	}
	for _, mutate := range mutations {
		current := base
		mutate(&current)
		if diff := CompareFreePlanObjectRevisions([]PlanFreeObject{base}, []PlanFreeObject{current}); len(diff) != 1 {
			t.Fatalf("mutation was not detected: %#v", current)
		}
	}
}
