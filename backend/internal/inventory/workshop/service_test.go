package workshop

import (
	"context"
	"errors"
	"testing"
	"time"
)

type workshopRepository struct {
	Repository
	maintenance VehicleMaintenance
	spare       VehicleSparePart
	failure     error
	creates     int
	updates     int
	vehicleID   string
}

func (r *workshopRepository) CreateMaintenance(_ context.Context, value VehicleMaintenance) error {
	r.maintenance = value
	r.creates++
	return r.failure
}
func (r *workshopRepository) ListSpareParts(context.Context, string) ([]VehicleSparePart, error) {
	return []VehicleSparePart{r.spare}, r.failure
}
func (r *workshopRepository) UpdateSparePart(_ context.Context, vehicleID, id string, input VehicleSparePartInput, now string) error {
	r.updates++
	r.vehicleID = vehicleID
	if id != r.spare.ID {
		return errors.New("duplicate identity changed")
	}
	r.spare.ArticleNumber, r.spare.Description = input.ArticleNumber, input.Description
	r.spare.Price, r.spare.URL, r.spare.UpdatedAt = input.Price, input.URL, now
	return r.failure
}
func (r *workshopRepository) GetSparePart(_ context.Context, vehicleID, id string) (*VehicleSparePart, error) {
	if vehicleID != r.vehicleID || id != r.spare.ID {
		return nil, errors.New("spare part escaped its vehicle scope")
	}
	return &r.spare, r.failure
}

func TestMaintenanceRetainsUnknownDatesAndRecordedCostFormat(t *testing.T) {
	repository := &workshopRepository{}
	parent := ""
	service := New(repository, func(_ context.Context, id string) error { parent = id; return nil })
	value, err := service.CreateMaintenance(t.Context(), " vehicle ", VehicleMaintenanceInput{
		Status: " fällig ", Cost: " 129,90 € ", Notes: " Recorded maintenance ",
	})
	if err != nil || value == nil || value.ID == "" || parent != "vehicle" || repository.creates != 1 ||
		value.Kind != "Wartung" || value.Status != "faellig" || value.Cost != "129,90" ||
		value.DueDate != "" || value.CompletedAt != "" || value.Notes != "Recorded maintenance" {
		t.Fatal("workshop defaults inferred dates or changed recorded values", value, err)
	}
	if _, err := time.Parse(time.RFC3339, value.CreatedAt); err != nil || value.CreatedAt != value.UpdatedAt {
		t.Fatal("creation time was not an actual shared UTC timestamp", value, err)
	}
}

func TestMaintenanceDoesNotAcknowledgeParentOrStorageFailure(t *testing.T) {
	for _, phase := range []string{"parent", "write"} {
		t.Run(phase, func(t *testing.T) {
			sentinel := errors.New("synthetic workshop failure")
			repository := &workshopRepository{}
			parent := func(context.Context, string) error { return nil }
			if phase == "parent" {
				parent = func(context.Context, string) error { return sentinel }
			} else {
				repository.failure = sentinel
			}
			value, err := New(repository, parent).CreateMaintenance(t.Context(), "vehicle", VehicleMaintenanceInput{})
			if !errors.Is(err, sentinel) || value != nil || phase == "parent" && repository.creates != 0 {
				t.Fatal("failed workshop creation changed or acknowledged a record", value, err)
			}
		})
	}
}

func TestInvalidMaintenanceIsRejectedBeforeParentOrStorageAccess(t *testing.T) {
	for _, input := range []VehicleMaintenanceInput{
		{DueDate: "2026-02-30"}, {CompletedAt: "unknown"}, {Cost: "-1"},
		{Kind: "unsupported"}, {Status: "unsupported"}, {ConditionRating: "unsupported"},
	} {
		if value, err := New(nil, nil).CreateMaintenance(t.Context(), "vehicle", input); !errors.Is(err, ErrVehicleValidation) || value != nil {
			t.Fatal("invalid input reached a persistence boundary", value, err)
		}
	}
}

func TestDuplicateSparePartKeepsExistingIdentityAndRecordedFields(t *testing.T) {
	repository := &workshopRepository{spare: VehicleSparePart{ID: "recorded-id", VehicleID: "vehicle",
		ArticleNumber: "ET-12 34", Description: "Recorded description", Price: "12,90", URL: "https://example.invalid/original", CreatedAt: "2020-01-02T03:04:05Z"}}
	service := New(repository, func(context.Context, string) error { return nil })
	value, err := service.CreateSparePart(t.Context(), "vehicle", VehicleSparePartInput{
		ArticleNumber: "1234", Description: "Incoming description", Price: "99.99", URL: "https://example.invalid/new",
	})
	if err != nil || value.ID != "recorded-id" || repository.creates != 0 || repository.updates != 1 ||
		value.Description != "Recorded description" || value.Price != "12,90" || value.URL != "https://example.invalid/original" ||
		value.CreatedAt != "2020-01-02T03:04:05Z" || value.ArticleNumber != "ET-12 34" {
		t.Fatal("duplicate adoption changed identity, price, source or creation history", value, err)
	}
}
