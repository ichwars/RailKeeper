package workflow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/decoder"
)

func (s *CVService) CreateCVValue(ctx context.Context, vehicleID string, input decoder.VehicleCVValueInput) (*decoder.VehicleCVValue, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	input = decoder.CleanCVValueInput(input)
	if vehicleID == "" || !decoder.ValidCVValueInput(input) {
		return nil, decoder.ErrCVValidation
	}
	tx, err := s.repository.BeginVehicleCVWrite(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	item, err := writeCV(ctx, tx, vehicleID, nil, input, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit CV create: %w", err)
	}
	value := decoder.CVRecordValue(item)
	return &value, nil
}

func (s *CVService) UpdateCVValue(ctx context.Context, vehicleID, cvValueID string, input decoder.VehicleCVValueInput) (*decoder.VehicleCVValue, error) {
	vehicleID, cvValueID = strings.TrimSpace(vehicleID), strings.TrimSpace(cvValueID)
	input = decoder.CleanCVValueInput(input)
	if vehicleID == "" || cvValueID == "" || !decoder.ValidCVValueInput(input) {
		return nil, decoder.ErrCVValidation
	}
	tx, err := s.repository.BeginVehicleCVWrite(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	previous, found, err := tx.CVValue(ctx, vehicleID, cvValueID)
	if err == nil && !found {
		return nil, ErrCVNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read CV before update: %w", err)
	}
	if _, err := writeCV(ctx, tx, vehicleID, &previous, input, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit CV update: %w", err)
	}
	return s.GetCVValue(ctx, vehicleID, cvValueID)
}

func (s *CVService) DeleteCVValueCompared(ctx context.Context, vehicleID, cvValueID string, input decoder.VehicleCVDeletionInput) (*decoder.VehicleCVValue, error) {
	vehicleID, cvValueID = strings.TrimSpace(vehicleID), strings.TrimSpace(cvValueID)
	tx, err := s.repository.BeginVehicleCVWrite(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	record, found, err := tx.CVValue(ctx, vehicleID, cvValueID)
	if err == nil && !found {
		return nil, ErrCVNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read CV before delete: %w", err)
	}
	if err := checkCVDeletionScope(ctx, tx, vehicleID, record.DecoderProfile, input); err != nil {
		return nil, err
	}
	if input.ExpectedRevision != nil {
		if err := decoder.CheckCVRevision(&record, input.ExpectedRevision); err != nil {
			return nil, err
		}
	}
	item := decoder.CVRecordValue(record)
	item.History, err = tx.CVHistory(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	if err := tx.DeleteCV(ctx, vehicleID, cvValueID); err != nil {
		return nil, fmt.Errorf("delete CV: %w", err)
	}
	if err := TouchWorkspace(ctx, tx, vehicleID, item.DecoderProfile, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit CV delete: %w", err)
	}
	return &item, nil
}

func (s *CVService) GetCVValue(ctx context.Context, vehicleID, cvValueID string) (*decoder.VehicleCVValue, error) {
	record, found, err := s.repository.CVValue(ctx, strings.TrimSpace(vehicleID), strings.TrimSpace(cvValueID))
	if err == nil && !found {
		return nil, ErrCVNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get CV value: %w", err)
	}
	item := decoder.CVRecordValue(record)
	item.History, err = s.repository.CVHistory(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	return &item, nil
}
