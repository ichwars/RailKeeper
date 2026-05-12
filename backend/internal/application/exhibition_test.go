package application_test

import (
	"context"
	"errors"
	"testing"

	"railkeeper2/backend/internal/application"
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
