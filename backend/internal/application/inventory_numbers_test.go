package application_test

import (
	"context"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestInventoryNumberSchemeListAndUpdate(t *testing.T) {
	db := testDB(t)
	service := application.NewInventoryNumberService(db)
	ctx := context.Background()

	schemes, err := service.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(schemes) == 0 {
		t.Fatal("expected seeded inventory number schemes")
	}
	for _, scheme := range schemes {
		if scheme.Category == "Zubehoer" {
			t.Fatal("accessory inventory scheme must not be seeded")
		}
	}

	updated, err := service.Update(ctx, "Lokomotive", application.InventoryNumberSchemeInput{
		Prefix:     "rk-engine",
		NextNumber: 42,
		Padding:    5,
		Active:     true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if updated.Prefix != "RK-ENGINE" || updated.NextNumber != 42 || updated.Padding != 5 || !updated.Active {
		t.Fatalf("unexpected scheme update: %#v", updated)
	}
	if updated.Preview != "RK-ENGINE-00042" {
		t.Fatalf("unexpected preview %q", updated.Preview)
	}
}

func TestInventoryNumberSchemeCreate(t *testing.T) {
	db := testDB(t)
	service := application.NewInventoryNumberService(db)
	ctx := context.Background()

	created, err := service.Create(ctx, application.InventoryNumberSchemeCreateInput{
		Category: "Güterwagen",
		InventoryNumberSchemeInput: application.InventoryNumberSchemeInput{
			Prefix:     "rk-gw",
			NextNumber: 7,
			Padding:    4,
			Active:     true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Category != "Güterwagen" || created.Prefix != "RK-GW" || created.Preview != "RK-GW-0007" {
		t.Fatalf("unexpected created scheme: %#v", created)
	}

	if _, err := service.Create(ctx, application.InventoryNumberSchemeCreateInput{
		Category: "Güterwagen",
		InventoryNumberSchemeInput: application.InventoryNumberSchemeInput{
			Prefix:     "RK-GW",
			NextNumber: 8,
			Padding:    4,
			Active:     true,
		},
	}); err != application.ErrInventoryNumberConflict {
		t.Fatalf("expected duplicate scheme conflict, got %v", err)
	}
}
