package application_test

import (
	"testing"

	"railkeeper/backend/internal/application"
)

func TestImageCreationThenMetadataSavePreservesZeroOrder(t *testing.T) {
	s := application.NewVehicleService(testDB(t))
	input := application.CreateVehicleInput{Manufacturer: "Reference", Name: "Order", Gauge: "TT",
		Category: "Lokomotive", Gattung: "Diesellok", Images: []application.VehicleImageInput{
			{URL: "https://example.test/zero.png"},
			{URL: "https://example.test/primary.png", SortOrder: 1, IsPrimary: true},
			{URL: "https://example.test/other.png", SortOrder: 2},
		}}
	before, err := s.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	input.Images = nil
	for _, image := range before.Images {
		input.Images = append(input.Images, application.VehicleImageInput{ID: image.ID, URL: image.URL,
			IsPrimary: image.IsPrimary, SortOrder: image.SortOrder, Title: "Changed title"})
	}
	input.ExpectedRevision, input.ExpectedImagesRevision = &before.Revision, &before.ImagesRevision
	after, err := s.Update(t.Context(), before.ID, input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	for index, image := range after.Images {
		if image.ID != before.Images[index].ID || image.SortOrder != before.Images[index].SortOrder ||
			image.IsPrimary != before.Images[index].IsPrimary || image.Title != "Changed title" {
			t.Fatalf("metadata save changed stored ordering: before=%+v after=%+v", before.Images, after.Images)
		}
	}
	// Explicit reordering still changes order, and new zero-valued inputs keep the legacy index default.
	input.ExpectedRevision, input.ExpectedImagesRevision = &after.Revision, &after.ImagesRevision
	input.Images = []application.VehicleImageInput{
		{ID: after.Images[2].ID, URL: after.Images[2].URL, IsPrimary: true},
		{ID: after.Images[0].ID, URL: after.Images[0].URL, SortOrder: 1},
		{URL: "https://example.test/new.png"},
	}
	reordered, err := s.Update(t.Context(), before.ID, input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if reordered.Images[0].ID != after.Images[2].ID || reordered.Images[0].SortOrder != 0 ||
		reordered.Images[1].SortOrder != 1 || reordered.Images[2].SortOrder != 2 {
		t.Fatalf("explicit reorder or new-image default changed: %+v", reordered.Images)
	}
}
