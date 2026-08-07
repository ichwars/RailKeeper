package application

import (
	"context"
	"errors"
	"testing"

	"railkeeper/backend/internal/domain"
)

type accessoryAllocationRepositorySpy struct {
	createdReservation   CreateAccessoryReservationInput
	cancelledReservation string
	installed            CreateAccessoryInstallationInput
	removedID            string
	removed              RemoveAccessoryInstallationInput
	conditionID          string
	condition            UpdateAccessoryInstallationConditionInput
	historyProductID     string
}

func (s *accessoryAllocationRepositorySpy) GetUsageHistory(
	_ context.Context,
	productID string,
) (*AccessoryUsageHistory, error) {
	s.historyProductID = productID
	return &AccessoryUsageHistory{ProductID: productID, Events: []AccessoryUsageEvent{}}, nil
}

func (s *accessoryAllocationRepositorySpy) ListReservations(
	context.Context,
	string,
) ([]AccessoryReservation, error) {
	return []AccessoryReservation{}, nil
}

func (s *accessoryAllocationRepositorySpy) CreateReservation(
	_ context.Context,
	input CreateAccessoryReservationInput,
	_ string,
) (*AccessoryReservation, error) {
	s.createdReservation = input
	return &AccessoryReservation{ProductID: input.ProductID, Quantity: input.Quantity}, nil
}

func (s *accessoryAllocationRepositorySpy) CancelReservation(
	_ context.Context,
	id string,
	_ string,
) (*AccessoryReservation, error) {
	s.cancelledReservation = id
	return &AccessoryReservation{ID: id}, nil
}

func (s *accessoryAllocationRepositorySpy) ListInstallations(
	context.Context,
	string,
) ([]AccessoryInstallation, error) {
	return []AccessoryInstallation{}, nil
}

func (s *accessoryAllocationRepositorySpy) Install(
	_ context.Context,
	input CreateAccessoryInstallationInput,
	_ string,
) (*AccessoryInstallation, error) {
	s.installed = input
	return &AccessoryInstallation{ProductID: input.ProductID, Condition: input.Condition}, nil
}

func (s *accessoryAllocationRepositorySpy) RemoveInstallation(
	_ context.Context,
	id string,
	input RemoveAccessoryInstallationInput,
	_ string,
) (*AccessoryInstallation, error) {
	s.removedID = id
	s.removed = input
	return &AccessoryInstallation{ID: id, RemovalDisposition: input.Disposition}, nil
}

func (s *accessoryAllocationRepositorySpy) UpdateInstallationCondition(
	_ context.Context,
	id string,
	input UpdateAccessoryInstallationConditionInput,
	_ string,
) (*AccessoryInstallation, error) {
	s.conditionID = id
	s.condition = input
	return &AccessoryInstallation{ID: id, Condition: input.Condition}, nil
}

func (s *accessoryAllocationRepositorySpy) GetAllocationSummary(
	context.Context,
	string,
) (*AccessoryAllocationSummary, error) {
	return &AccessoryAllocationSummary{}, nil
}

func TestAccessoryAllocationServiceRequiresExactlyOneTarget(t *testing.T) {
	service := NewAccessoryAllocationService(&accessoryAllocationRepositorySpy{})
	validTargets := []AllocationTargetInput{
		{VehicleID: "vehicle-1"},
		{LayoutID: "layout-1"},
		{LayoutUnitID: "unit-1"},
	}
	for _, target := range validTargets {
		if _, err := service.CreateReservation(t.Context(), CreateAccessoryReservationInput{
			ProductID: "product-1", LocationID: "location-1", Quantity: 1, AllocationTargetInput: target,
		}, "planner-1"); err != nil {
			t.Fatalf("valid target %#v rejected: %v", target, err)
		}
	}
	invalidTargets := []AllocationTargetInput{
		{},
		{VehicleID: "vehicle-1", LayoutID: "layout-1"},
		{LayoutID: "layout-1", LayoutUnitID: "unit-1"},
	}
	for _, target := range invalidTargets {
		if _, err := service.CreateReservation(t.Context(), CreateAccessoryReservationInput{
			ProductID: "product-1", LocationID: "location-1", Quantity: 1, AllocationTargetInput: target,
		}, "planner-1"); !errors.Is(err, ErrAccessoryValidation) {
			t.Fatalf("target %#v: expected validation error, got %v", target, err)
		}
	}
}

func TestAccessoryAllocationServiceNormalizesReservationAndInstallation(t *testing.T) {
	repository := &accessoryAllocationRepositorySpy{}
	service := NewAccessoryAllocationService(repository)

	if _, err := service.CreateReservation(t.Context(), CreateAccessoryReservationInput{
		ProductID: " product-1 ", AssetID: " asset-1 ", LocationID: " location-1 ", Quantity: 1,
		AllocationTargetInput: AllocationTargetInput{LayoutUnitID: " unit-1 "}, Note: " planned ",
		Placement: " signal bridge ", DigitalAddress: " 42 ", DecoderOutput: " A1 ",
		Connection: " terminal 3 ", WiringNotes: " blue wire ",
	}, "planner-1"); err != nil {
		t.Fatal(err)
	}
	reservation := repository.createdReservation
	if reservation.ProductID != "product-1" || reservation.AssetID != "asset-1" ||
		reservation.LocationID != "location-1" || reservation.LayoutUnitID != "unit-1" ||
		reservation.Note != "planned" || reservation.Placement != "signal bridge" ||
		reservation.DigitalAddress != "42" || reservation.DecoderOutput != "A1" ||
		reservation.Connection != "terminal 3" || reservation.WiringNotes != "blue wire" {
		t.Fatalf("reservation was not normalized: %#v", reservation)
	}

	installation, err := service.Install(t.Context(), CreateAccessoryInstallationInput{
		ReservationID: " reservation-1 ", ProductID: " product-1 ", SourceLocationID: " location-1 ",
		Quantity: 2, AllocationTargetInput: AllocationTargetInput{LayoutID: " layout-1 "}, Notes: " installed ",
		Placement: " platform 1 ", DigitalAddress: " 17 ", DecoderOutput: " B2 ",
		Connection: " bus 1 ", WiringNotes: " yellow wire ",
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if installation.Condition != domain.AccessoryConditionUnknown ||
		repository.installed.ReservationID != "reservation-1" || repository.installed.LayoutID != "layout-1" ||
		repository.installed.Notes != "installed" || repository.installed.Placement != "platform 1" ||
		repository.installed.DigitalAddress != "17" || repository.installed.DecoderOutput != "B2" ||
		repository.installed.Connection != "bus 1" || repository.installed.WiringNotes != "yellow wire" {
		t.Fatalf("installation was not normalized and defaulted: %#v", repository.installed)
	}
}

func TestAccessoryAllocationServiceGetsNormalizedUsageHistory(t *testing.T) {
	repository := &accessoryAllocationRepositorySpy{}
	service := NewAccessoryAllocationService(repository)
	history, err := service.GetUsageHistory(t.Context(), " product-1 ")
	if err != nil {
		t.Fatal(err)
	}
	if repository.historyProductID != "product-1" || history.ProductID != "product-1" {
		t.Fatalf("unexpected usage history request: %#v", history)
	}
	if _, err := service.GetUsageHistory(t.Context(), " "); !errors.Is(err, ErrAccessoryValidation) {
		t.Fatalf("expected empty product id rejection, got %v", err)
	}
}

func TestAccessoryAllocationServiceRejectsInvalidQuantitiesAndIdentifiers(t *testing.T) {
	service := NewAccessoryAllocationService(&accessoryAllocationRepositorySpy{})
	validTarget := AllocationTargetInput{LayoutID: "layout-1"}
	reservations := []CreateAccessoryReservationInput{
		{LocationID: "location-1", Quantity: 1, AllocationTargetInput: validTarget},
		{ProductID: "product-1", Quantity: 1, AllocationTargetInput: validTarget},
		{ProductID: "product-1", LocationID: "location-1", Quantity: 0, AllocationTargetInput: validTarget},
		{ProductID: "product-1", AssetID: "asset-1", LocationID: "location-1", Quantity: 2,
			AllocationTargetInput: validTarget},
	}
	for _, input := range reservations {
		if _, err := service.CreateReservation(t.Context(), input, "planner-1"); !errors.Is(err, ErrAccessoryValidation) {
			t.Fatalf("reservation %#v: expected validation error, got %v", input, err)
		}
	}
	if _, err := service.CancelReservation(t.Context(), " ", "planner-1"); !errors.Is(err, ErrAccessoryValidation) {
		t.Fatalf("expected empty cancellation id to fail, got %v", err)
	}
	if _, err := service.Install(t.Context(), CreateAccessoryInstallationInput{
		ProductID: "product-1", AssetID: "asset-1", SourceLocationID: "location-1", Quantity: 2,
		AllocationTargetInput: validTarget,
	}, "editor-1"); !errors.Is(err, ErrAccessoryValidation) {
		t.Fatalf("expected individual quantity to fail, got %v", err)
	}
	if _, err := service.GetAllocationSummary(t.Context(), " "); !errors.Is(err, ErrAccessoryValidation) {
		t.Fatalf("expected empty summary product id to fail, got %v", err)
	}
}

func TestAccessoryAllocationServiceValidatesRemovalAndConditionChanges(t *testing.T) {
	repository := &accessoryAllocationRepositorySpy{}
	service := NewAccessoryAllocationService(repository)

	if _, err := service.RemoveInstallation(t.Context(), " installation-1 ", RemoveAccessoryInstallationInput{
		Disposition: domain.AccessoryRemovalStored, StorageLocationID: " location-1 ", Notes: " returned ",
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	if repository.removedID != "installation-1" || repository.removed.StorageLocationID != "location-1" ||
		repository.removed.Notes != "returned" {
		t.Fatalf("removal was not normalized: %#v", repository.removed)
	}
	invalidRemovals := []RemoveAccessoryInstallationInput{
		{},
		{Disposition: "sold"},
		{Disposition: domain.AccessoryRemovalStored},
		{Disposition: domain.AccessoryRemovalRetired, StorageLocationID: "location-1"},
	}
	for _, input := range invalidRemovals {
		if _, err := service.RemoveInstallation(t.Context(), "installation-1", input, "editor-1"); !errors.Is(err, ErrAccessoryValidation) {
			t.Fatalf("removal %#v: expected validation error, got %v", input, err)
		}
	}
	if _, err := service.UpdateInstallationCondition(t.Context(), "installation-1",
		UpdateAccessoryInstallationConditionInput{Condition: domain.AccessoryConditionDefective}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	if repository.conditionID != "installation-1" ||
		repository.condition.Condition != domain.AccessoryConditionDefective {
		t.Fatalf("condition was not passed through: %#v", repository.condition)
	}
	if _, err := service.UpdateInstallationCondition(t.Context(), "installation-1",
		UpdateAccessoryInstallationConditionInput{Condition: "broken"}, "editor-1"); !errors.Is(err, ErrAccessoryValidation) {
		t.Fatalf("expected invalid condition to fail, got %v", err)
	}
}
