package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrDataTransferValidation = errors.New("data transfer validation failed")
	ErrDataTransferNotFound   = errors.New("data transfer profile not found")
)

type CreateDataTransferProfileInput struct {
	Name      string            `json:"name"`
	Direction TransferDirection `json:"direction"`
	Format    TransferFormat    `json:"format"`
	Areas     []TransferArea    `json:"areas"`
	Options   map[string]any    `json:"options"`
}

type UpdateDataTransferProfileInput struct {
	Name      string            `json:"name"`
	Direction TransferDirection `json:"direction"`
	Format    TransferFormat    `json:"format"`
	Areas     []TransferArea    `json:"areas"`
	Options   map[string]any    `json:"options"`
}

type DataTransferService struct {
	repository DataTransferRepository
	dataDir    string
}

func NewDataTransferService(repository DataTransferRepository, dataDir string) *DataTransferService {
	return &DataTransferService{repository: repository, dataDir: strings.TrimSpace(dataDir)}
}

func (s *DataTransferService) CreateProfile(
	ctx context.Context,
	input CreateDataTransferProfileInput,
	actorUserID string,
) (DataTransferProfile, error) {
	input.Name = strings.TrimSpace(input.Name)
	actorUserID = strings.TrimSpace(actorUserID)
	if input.Name == "" || actorUserID == "" || !validTransferDirection(input.Direction) ||
		!validTransferFormat(input.Format) {
		return DataTransferProfile{}, ErrDataTransferValidation
	}
	if err := validateTransferSelection(input.Format, input.Areas); err != nil {
		return DataTransferProfile{}, err
	}
	return s.repository.CreateProfile(ctx, DataTransferProfile{
		Name: input.Name, Direction: input.Direction, Format: input.Format, Areas: input.Areas, Options: input.Options,
		Enabled: true, CreatedByUserID: actorUserID,
	})
}

func (s *DataTransferService) ListProfiles(ctx context.Context) ([]DataTransferProfile, error) {
	return s.repository.ListProfiles(ctx)
}

func (s *DataTransferService) UpdateProfile(
	ctx context.Context,
	id string,
	input UpdateDataTransferProfileInput,
) (DataTransferProfile, error) {
	id = strings.TrimSpace(id)
	input.Name = strings.TrimSpace(input.Name)
	if id == "" || input.Name == "" || !validTransferDirection(input.Direction) || !validTransferFormat(input.Format) {
		return DataTransferProfile{}, ErrDataTransferValidation
	}
	if err := validateTransferSelection(input.Format, input.Areas); err != nil {
		return DataTransferProfile{}, err
	}
	profile, err := s.profile(ctx, id)
	if err != nil {
		return DataTransferProfile{}, err
	}
	profile.Name = input.Name
	profile.Direction = input.Direction
	profile.Format = input.Format
	profile.Areas = input.Areas
	profile.Options = input.Options
	return s.repository.UpdateProfile(ctx, profile)
}

func (s *DataTransferService) DisableProfile(ctx context.Context, id string) (DataTransferProfile, error) {
	profile, err := s.profile(ctx, strings.TrimSpace(id))
	if err != nil {
		return DataTransferProfile{}, err
	}
	profile.Enabled = false
	return s.repository.UpdateProfile(ctx, profile)
}

func (s *DataTransferService) profile(ctx context.Context, id string) (DataTransferProfile, error) {
	if id == "" {
		return DataTransferProfile{}, ErrDataTransferValidation
	}
	profile, err := s.repository.GetProfile(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return DataTransferProfile{}, ErrDataTransferNotFound
	}
	return profile, err
}

func validateTransferSelection(format TransferFormat, areas []TransferArea) error {
	allowed := map[TransferArea]bool{
		TransferVehicles: true, TransferAccessories: true, TransferExhibitionLists: true,
	}
	if len(areas) == 0 {
		return fmt.Errorf("%w: select at least one area", ErrDataTransferValidation)
	}
	seen := map[TransferArea]bool{}
	for _, area := range areas {
		if !allowed[area] || seen[area] {
			return fmt.Errorf("%w: unsupported or repeated area %q", ErrDataTransferValidation, area)
		}
		seen[area] = true
	}
	if format == TransferCSV && len(areas) != 1 {
		return fmt.Errorf("%w: csv requires exactly one area", ErrDataTransferValidation)
	}
	if format == TransferCSV && areas[0] == TransferExhibitionLists {
		return fmt.Errorf("%w: exhibition lists require railkeeper-json", ErrDataTransferValidation)
	}
	return nil
}

func validTransferDirection(direction TransferDirection) bool {
	return direction == TransferImport || direction == TransferExport
}

func validTransferFormat(format TransferFormat) bool {
	return format == TransferCSV || format == TransferJSON
}
