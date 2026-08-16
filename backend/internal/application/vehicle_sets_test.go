package application_test

import (
	"context"
	"errors"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestCreateVehicleSetValidatesAllMembersBeforeLocalizingImages(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	localized := 0
	service.SetImageLocalizer(func(
		_ context.Context,
		_ string,
		images []application.VehicleImageInput,
	) ([]application.VehicleImageInput, error) {
		localized++
		return images, nil
	})

	_, err := service.CreateSet(context.Background(), application.CreateVehicleSetInput{
		Set: application.VehicleSetInput{
			Name: "Bildtest", Manufacturer: "Roco", Gauge: "H0", Category: "Wagen", Gattung: "Reisezugwagen",
		},
		Members: []application.CreateVehicleInput{
			{Name: "Wagen 1", Images: []application.VehicleImageInput{{URL: "https://example.test/image.jpg"}}},
			{Name: "Wagen 2", MaximumSpeedKmh: intPointer(0)},
		},
	}, "actor-1")
	if !errors.Is(err, application.ErrVehicleOperationalValidation) {
		t.Fatalf("expected operational validation error, got %v", err)
	}
	if localized != 0 {
		t.Fatalf("localized %d members before the complete request was valid", localized)
	}
}

func TestCreateVehicleSetCreatesOrderedMembersAndListMetadata(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	ctx := context.Background()

	created, err := service.CreateSet(ctx, application.CreateVehicleSetInput{
		Set: application.VehicleSetInput{
			Name: "TEE Roland", Manufacturer: "Märklin", ArticleNumber: "37605", Gauge: "H0",
			Category: "Triebzug", Gattung: "Dieseltriebzug", Epoch: "IV",
		},
		Members: []application.CreateVehicleInput{
			{InventoryNumber: "RK-SET-000001", Name: "Motorwagen", VehicleNumber: "VT 11.5"},
			{InventoryNumber: "RK-SET-000002", Name: "Steuerwagen", VehicleNumber: "VS 11.5"},
		},
	}, "actor-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(created.Members) != 2 {
		t.Fatalf("expected two set members, got %d", len(created.Members))
	}
	if created.InventoryNumber != "RK-SET-000001" {
		t.Fatalf("unexpected set inventory number %q", created.InventoryNumber)
	}
	for index, member := range created.Members {
		if member.VehicleSetID != created.ID || member.VehicleSetName != "TEE Roland" {
			t.Fatalf("member %d has incomplete set metadata: %#v", index, member)
		}
		if member.VehicleSetPosition != index+1 || member.VehicleSetSize != 2 {
			t.Fatalf("member %d has invalid position metadata: %#v", index, member)
		}
		if member.Manufacturer != "Märklin" || member.ArticleNumber != "37605" {
			t.Fatalf("member %d did not inherit shared article data: %#v", index, member)
		}
	}

	listed, err := service.List(ctx, "Roland")
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 2 || listed[0].VehicleSetID != created.ID || listed[1].VehicleSetID != created.ID {
		t.Fatalf("set search/list metadata mismatch: %#v", listed)
	}
}

func TestCreateVehicleSetRollsBackAllMembersOnConflict(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	ctx := context.Background()

	if _, err := service.Create(ctx, application.CreateVehicleInput{
		InventoryNumber: "RK-DUP-000001", Manufacturer: "Piko", Name: "Bestand", Gauge: "H0",
		Category: "Lokomotive", Gattung: "Diesellok",
	}, "actor-1"); err != nil {
		t.Fatal(err)
	}

	_, err := service.CreateSet(ctx, application.CreateVehicleSetInput{
		Set: application.VehicleSetInput{
			Name: "Testset", Manufacturer: "Roco", Gauge: "H0", Category: "Wagen", Gattung: "Reisezugwagen",
		},
		Members: []application.CreateVehicleInput{
			{InventoryNumber: "RK-NEW-000001", Name: "Wagen 1"},
			{InventoryNumber: "RK-DUP-000001", Name: "Wagen 2"},
		},
	}, "actor-1")
	if !errors.Is(err, application.ErrInventoryNumberConflict) {
		t.Fatalf("expected inventory number conflict, got %v", err)
	}

	var vehicleCount, setCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicles`).Scan(&vehicleCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_sets`).Scan(&setCount); err != nil {
		t.Fatal(err)
	}
	if vehicleCount != 1 || setCount != 0 {
		t.Fatalf("set creation was not atomic: vehicles=%d sets=%d", vehicleCount, setCount)
	}

	created, err := service.CreateSet(ctx, application.CreateVehicleSetInput{
		Set: application.VehicleSetInput{
			Name: "Folgeset", Manufacturer: "Roco", Gauge: "H0", Category: "Wagen", Gattung: "Reisezugwagen",
		},
		Members: []application.CreateVehicleInput{{Name: "Wagen 1"}, {Name: "Wagen 2"}},
	}, "actor-1")
	if err != nil {
		t.Fatal(err)
	}
	if created.InventoryNumber != "RK-SET-000001" {
		t.Fatalf("rolled-back set consumed an inventory number: %q", created.InventoryNumber)
	}
}

func TestDeleteFinalVehicleSetMemberRemovesEmptySet(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	ctx := context.Background()

	created, err := service.CreateSet(ctx, application.CreateVehicleSetInput{
		Set: application.VehicleSetInput{
			Name: "Löschtest", Manufacturer: "Fleischmann", Gauge: "N", Category: "Wagen",
			Gattung: "Güterwagen",
		},
		Members: []application.CreateVehicleInput{{Name: "Wagen 1"}, {Name: "Wagen 2"}},
	}, "actor-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, member := range created.Members {
		if err := service.Delete(ctx, member.ID, "actor-1"); err != nil {
			t.Fatal(err)
		}
	}

	var setCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_sets WHERE id=?`, created.ID).Scan(&setCount); err != nil {
		t.Fatal(err)
	}
	if setCount != 0 {
		t.Fatalf("expected empty set to be removed, got %d rows", setCount)
	}
}

func TestUpdateVehicleSetMemberPreservesSharedSetData(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	ctx := context.Background()

	created, err := service.CreateSet(ctx, application.CreateVehicleSetInput{
		Set: application.VehicleSetInput{
			Name: "TEE Roland", Manufacturer: "Märklin", ArticleNumber: "37605", Gauge: "H0",
			Category: "Triebzug", Gattung: "Dieseltriebzug", PurchasePrice: "299.90",
		},
		Members: []application.CreateVehicleInput{{Name: "Motorwagen"}, {Name: "Steuerwagen"}},
	}, "actor-1")
	if err != nil {
		t.Fatal(err)
	}

	member := created.Members[0]
	input := application.CreateVehicleInput{
		InventoryNumber: member.InventoryNumber,
		Manufacturer:    "Roco",
		ArticleNumber:   "changed",
		Name:            "Motorwagen neu",
		Gauge:           "N",
		Category:        "Lokomotive",
		Gattung:         "Diesellok",
		PurchasePrice:   "1.00",
	}
	updated, err := service.Update(ctx, member.ID, input, "actor-1")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Motorwagen neu" {
		t.Fatalf("member-specific name was not updated: %q", updated.Name)
	}
	if updated.Manufacturer != "Märklin" || updated.ArticleNumber != "37605" || updated.Gauge != "H0" ||
		updated.Category != "Triebzug" || updated.Gattung != "Dieseltriebzug" || updated.PurchasePrice != "299.90" {
		t.Fatalf("shared set data changed through member update: %#v", updated)
	}
}

func intPointer(value int) *int {
	return &value
}
