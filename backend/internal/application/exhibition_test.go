package application_test

import (
	"context"
	"errors"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestExhibitionListLocksEntries(t *testing.T) {
	db := testDB(t)
	service := application.NewExhibitionService(db)
	ctx := context.Background()

	list, err := service.Create(ctx, application.ExhibitionListInput{
		Designation: "Messe Dortmund",
		Date:        "2026-05-12",
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.CreateEntry(ctx, list.ID, application.ExhibitionEntryInput{
		Owner:          "Test Besitzer",
		LocomotiveName: "BR 218",
		DTDecoder:      true,
		DecoderNumber:  "D-218",
		FunctionKeys:   "F0 Licht",
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := service.SetLocked(ctx, list.ID, true); err != nil {
		t.Fatal(err)
	}

	_, err = service.CreateEntry(ctx, list.ID, application.ExhibitionEntryInput{
		Owner:          "Zweiter Besitzer",
		LocomotiveName: "BR 103",
	})
	if !errors.Is(err, application.ErrExhibitionLocked) {
		t.Fatalf("expected locked list error, got %v", err)
	}
}

func TestExhibitionEntryAcceptsMultipleDayScopes(t *testing.T) {
	db := testDB(t)
	service := application.NewExhibitionService(db)
	ctx := context.Background()

	list, err := service.Create(ctx, application.ExhibitionListInput{
		Designation: "Messe Leipzig",
		Date:        "2026-05-13",
	})
	if err != nil {
		t.Fatal(err)
	}

	entry, err := service.CreateEntry(ctx, list.ID, application.ExhibitionEntryInput{
		Owner:          "Test Besitzer",
		LocomotiveName: "BR 38",
		DayScope:       "day4,day1,day2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if entry.DayScope != "day1,day2,day4" {
		t.Fatalf("expected normalized multi-day scope, got %q", entry.DayScope)
	}
}
