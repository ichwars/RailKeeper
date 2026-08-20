package application

import (
	"context"
	"errors"
	"testing"
)

func TestDataTransferProfileRejectsInvalidCSVCombination(t *testing.T) {
	service := NewDataTransferService(profileRepositoryStub{}, t.TempDir())
	for _, areas := range [][]TransferArea{
		{TransferVehicles, TransferAccessories},
		{TransferArea("masterData")},
	} {
		_, err := service.CreateProfile(t.Context(), CreateDataTransferProfileInput{
			Name: "Invalid", Direction: TransferExport, Format: TransferCSV, Areas: areas,
		}, "admin-1")
		if !errors.Is(err, ErrDataTransferValidation) {
			t.Fatalf("expected validation error for %#v, got %v", areas, err)
		}
	}
}

func TestDataTransferProfileServiceUpdatesDisablesAndListsProfiles(t *testing.T) {
	repository := &dataTransferProfileRepositoryStub{profiles: []DataTransferProfile{
		{ID: "profile-1", Name: "Vehicle export", Direction: TransferExport, Format: TransferCSV,
			Areas: []TransferArea{TransferVehicles}, Enabled: true, CreatedByUserID: "admin-1"},
	}}
	service := NewDataTransferService(repository, t.TempDir())

	profiles, err := service.ListProfiles(t.Context())
	if err != nil || len(profiles) != 1 || profiles[0].ID != "profile-1" {
		t.Fatalf("unexpected listed profiles: %#v, %v", profiles, err)
	}

	updated, err := service.UpdateProfile(t.Context(), " profile-1 ", UpdateDataTransferProfileInput{
		Name: " Vehicle JSON export ", Direction: TransferExport, Format: TransferJSON,
		Areas: []TransferArea{TransferVehicles, TransferAccessories},
	})
	if err != nil || updated.Name != "Vehicle JSON export" || len(updated.Areas) != 2 || !updated.Enabled {
		t.Fatalf("unexpected updated profile: %#v, %v", updated, err)
	}

	disabled, err := service.DisableProfile(t.Context(), "profile-1")
	if err != nil || disabled.Enabled {
		t.Fatalf("unexpected disabled profile: %#v, %v", disabled, err)
	}
}

type profileRepositoryStub struct {
	DataTransferRepository
}

func (profileRepositoryStub) CreateProfile(
	context.Context,
	DataTransferProfile,
) (DataTransferProfile, error) {
	return DataTransferProfile{}, nil
}

type dataTransferProfileRepositoryStub struct {
	DataTransferRepository
	profiles []DataTransferProfile
}

func (s *dataTransferProfileRepositoryStub) GetProfile(_ context.Context, id string) (DataTransferProfile, error) {
	for _, profile := range s.profiles {
		if profile.ID == id {
			return profile, nil
		}
	}
	return DataTransferProfile{}, ErrDataTransferNotFound
}

func (s *dataTransferProfileRepositoryStub) ListProfiles(context.Context) ([]DataTransferProfile, error) {
	return s.profiles, nil
}

func (s *dataTransferProfileRepositoryStub) UpdateProfile(
	_ context.Context,
	profile DataTransferProfile,
) (DataTransferProfile, error) {
	for index := range s.profiles {
		if s.profiles[index].ID == profile.ID {
			s.profiles[index] = profile
			return profile, nil
		}
	}
	return DataTransferProfile{}, ErrDataTransferNotFound
}
