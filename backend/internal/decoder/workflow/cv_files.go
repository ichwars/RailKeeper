package workflow

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/decoder"
)

func (s *CVService) CreateCVFile(ctx context.Context, vehicleID string, input decoder.VehicleCVFileInput) (*decoder.VehicleCVFile, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	input = decoder.CleanCVFileInput(input)
	if vehicleID == "" || input.FileName == "" || input.OriginalName == "" || (input.StoragePath == "" && input.BlobID == "") {
		return nil, decoder.ErrCVValidation
	}
	tx, err := s.repository.BeginVehicleCVWrite(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	profile, err := resolveCVWorkspaceProfile(ctx, tx, vehicleID, input.WorkspaceID, input.ExpectedWorkspaceRevision, input.DecoderProfile)
	if err != nil {
		return nil, err
	}
	input.DecoderProfile = profile
	now := time.Now().UTC().Format(time.RFC3339)
	file := decoder.VehicleCVFile{
		ID:             credentials.RandomID(),
		VehicleID:      vehicleID,
		FileName:       input.FileName,
		OriginalName:   input.OriginalName,
		Description:    input.Description,
		DecoderProfile: input.DecoderProfile,
		MimeType:       input.MimeType,
		SizeBytes:      input.SizeBytes,
		StoragePath:    input.StoragePath,
		BlobID:         input.BlobID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := tx.InsertCVFile(ctx, file); err != nil {
		return nil, err
	}
	if err := TouchWorkspace(ctx, tx, vehicleID, file.DecoderProfile, now); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit CV file creation: %w", err)
	}
	file.Revision = decoder.CVFileRevision(file, "[]")
	return &file, nil
}

func (s *CVService) DeleteCVFileCompared(ctx context.Context, vehicleID, fileID string, input decoder.VehicleCVDeletionInput) (*decoder.VehicleCVFile, error) {
	tx, err := s.repository.BeginCVTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin CV file delete: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	file, err := tx.CVFile(ctx, vehicleID, fileID)
	if err != nil {
		return nil, err
	}
	if err := checkCVDeletionScope(ctx, tx, file.VehicleID, file.DecoderProfile, input); err != nil {
		if errors.Is(err, decoder.ErrCVConflict) {
			return nil, ErrCVFileConflict
		}
		return nil, err
	}
	if err := checkCVFileRevision(file.Revision, input.ExpectedRevision); err != nil {
		return nil, err
	}
	if err := tx.DeleteCVFile(ctx, file.VehicleID, file.ID); err != nil {
		return nil, err
	}
	if err := TouchWorkspace(ctx, tx, file.VehicleID, file.DecoderProfile, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit CV file delete: %w", err)
	}
	return file, nil
}

func checkCVFileRevision(actual string, expected *string) error {
	if expected == nil {
		return nil
	}
	if !decoder.ValidCVRevision(*expected) {
		return decoder.ErrCVValidation
	}
	if actual != *expected {
		return ErrCVFileConflict
	}
	return nil
}
func (s *CVService) GetCVFile(ctx context.Context, vehicleID, fileID string) (*decoder.VehicleCVFile, error) {
	return s.repository.CVFile(ctx, vehicleID, fileID)
}
func (s *CVService) Files(ctx context.Context, vehicleID string) ([]decoder.VehicleCVFile, error) {
	return s.repository.CVFiles(ctx, vehicleID)
}
