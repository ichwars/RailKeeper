package application

import (
	"math"
	"strings"
	"testing"

	"railkeeper/backend/internal/domain"
)

func TestFreePlanObjectServiceNormalizesCreateAndUpdate(t *testing.T) {
	repository := &trackPlannerRepositorySpy{}
	service := NewTrackPlannerService(repository)
	width, height := 500.0, 70.0
	created, err := service.CreateFreeObject(t.Context(), " revision-1 ", CreateFreePlanObjectInput{
		Name: " Bahnsteig 1 ", Category: domain.FreePlanPlatform,
		PositionXMM: 200, PositionYMM: 100, RotationDegrees: -15,
		Shape: domain.FreePlanObjectShape{
			SchemaVersion: 1, Kind: domain.FreePlanRectangle, WidthMM: &width, HeightMM: &height,
		},
	}, "planner")
	if err != nil || created.Name != "Bahnsteig 1" || created.RotationDegrees != 345 {
		t.Fatalf("free object was not normalized: %#v, %v", created, err)
	}
	fontSize := 8.0
	updated, err := service.UpdateFreeObject(t.Context(), " free-1 ", UpdateFreePlanObjectInput{
		CreateFreePlanObjectInput: CreateFreePlanObjectInput{
			Name: " Gleistext ", Category: domain.FreePlanAnnotation, RotationDegrees: 725,
			Shape: domain.FreePlanObjectShape{
				SchemaVersion: 1, Kind: domain.FreePlanLabel, Text: " Gleis 1 ", FontSizeMM: &fontSize,
			},
		},
		ExpectedVersion: 2,
	}, "planner")
	if err != nil || updated.Name != "Gleistext" || updated.Shape.Text != "Gleis 1" ||
		updated.RotationDegrees != 5 || repository.freeUpdated.ExpectedVersion != 2 {
		t.Fatalf("free object update was not normalized: %#v, %#v, %v", updated, repository.freeUpdated, err)
	}
	if err := service.DeleteFreeObject(t.Context(), " free-1 ", 3, "planner"); err != nil {
		t.Fatal(err)
	}
	if repository.freeDeletedID != "free-1" || repository.freeDeleted != 3 {
		t.Fatalf("free object delete was not normalized: %#v", repository)
	}
}

func TestFreePlanObjectServiceRejectsInvalidInputs(t *testing.T) {
	service := NewTrackPlannerService(&trackPlannerRepositorySpy{})
	width, height := 100.0, 50.0
	valid := CreateFreePlanObjectInput{
		Name: "Gebäude", Category: domain.FreePlanStructure,
		Shape: domain.FreePlanObjectShape{
			SchemaVersion: 1, Kind: domain.FreePlanRectangle, WidthMM: &width, HeightMM: &height,
		},
	}
	invalid := []CreateFreePlanObjectInput{
		{},
		{Name: strings.Repeat("x", 81), Category: valid.Category, Shape: valid.Shape},
		{Name: valid.Name, Category: "invalid", Shape: valid.Shape},
		{Name: valid.Name, Category: valid.Category, PositionXMM: math.NaN(), Shape: valid.Shape},
		{Name: valid.Name, Category: valid.Category, PositionYMM: math.Inf(1), Shape: valid.Shape},
		{Name: valid.Name, Category: valid.Category, RotationDegrees: math.NaN(), Shape: valid.Shape},
		{Name: valid.Name, Category: valid.Category, Shape: domain.FreePlanObjectShape{}},
	}
	for _, input := range invalid {
		if _, err := service.CreateFreeObject(t.Context(), "revision-1", input, "planner"); err != ErrTrackPlanValidation {
			t.Fatalf("expected invalid free create input %#v, got %v", input, err)
		}
	}
	if _, err := service.CreateFreeObject(t.Context(), " ", valid, "planner"); err != ErrTrackPlanValidation {
		t.Fatalf("expected invalid revision, got %v", err)
	}
	if _, err := service.UpdateFreeObject(t.Context(), "free-1", UpdateFreePlanObjectInput{
		CreateFreePlanObjectInput: valid,
	}, "planner"); err != ErrTrackPlanValidation {
		t.Fatalf("expected invalid update version, got %v", err)
	}
	if err := service.DeleteFreeObject(t.Context(), " ", 1, "planner"); err != ErrTrackPlanValidation {
		t.Fatalf("expected invalid delete id, got %v", err)
	}
	if err := service.DeleteFreeObject(t.Context(), "free-1", 0, "planner"); err != ErrTrackPlanValidation {
		t.Fatalf("expected invalid delete version, got %v", err)
	}
}
