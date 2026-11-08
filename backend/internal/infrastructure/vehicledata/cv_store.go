package vehicledata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/decoder/workflow"
)

type cvQuery interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}
type cvQueries struct{ query cvQuery }
type cvStore struct {
	*cvQueries
	db *sql.DB
}
type cvMutation struct {
	*cvQueries
	*workspaceWriter
	tx *sql.Tx
}

func NewCVStore(db *sql.DB) workflow.CVRepository {
	return &cvStore{cvQueries: &cvQueries{query: db}, db: db}
}
func (s *cvStore) begin(ctx context.Context) (*cvMutation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &cvMutation{cvQueries: &cvQueries{query: tx}, workspaceWriter: &workspaceWriter{tx: tx}, tx: tx}, nil
}
func (s *cvStore) BeginCVTransaction(ctx context.Context) (workflow.CVMutation, error) {
	return s.begin(ctx)
}
func (s *cvStore) BeginVehicleCVWrite(ctx context.Context, vehicleID string) (workflow.CVMutation, error) {
	m, err := s.begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin CV write: %w", err)
	}
	var exists int
	err = m.tx.QueryRowContext(ctx, `SELECT 1 FROM vehicles WHERE id=?`, vehicleID).Scan(&exists)
	if err != nil {
		_ = m.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return nil, workflow.ErrCVNotFound
		}
		return nil, fmt.Errorf("check CV vehicle: %w", err)
	}
	return m, nil
}
func (m *cvMutation) Commit() error   { return m.tx.Commit() }
func (m *cvMutation) Rollback() error { return m.tx.Rollback() }
func (m *cvMutation) CVBinding(ctx context.Context, vehicleID, workspaceID string) (decoder.CVBinding, error) {
	binding, err := ReadCVWorkspaceBinding(ctx, m.tx, vehicleID, workspaceID)
	if errors.Is(err, sql.ErrNoRows) {
		return binding, decoder.ErrCVValidation
	}
	if err != nil {
		return binding, fmt.Errorf("read CV working group binding: %w", err)
	}
	return binding, nil
}
func (m *cvMutation) CVSourceVehicle(ctx context.Context, id string) (string, bool, error) {
	var vehicleID string
	err := m.tx.QueryRowContext(ctx, `SELECT vehicle_id FROM vehicle_cv_files WHERE id=?`, id).Scan(&vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return vehicleID, err == nil, err
}
func (m *cvMutation) FindCV(ctx context.Context, item decoder.CVRecord) (decoder.CVRecord, bool, error) {
	value, err := FindCV(ctx, m.tx, item)
	if errors.Is(err, sql.ErrNoRows) {
		return value, false, nil
	}
	return value, err == nil, err
}
func (m *cvMutation) InsertCV(ctx context.Context, item decoder.CVRecord) error {
	return InsertCV(ctx, m.tx, item)
}
func (m *cvMutation) UpdateCV(ctx context.Context, item decoder.CVRecord) error {
	return UpdateCV(ctx, m.tx, item)
}
func (m *cvMutation) DeleteCV(ctx context.Context, vehicleID, id string) error {
	return DeleteCV(ctx, m.tx, vehicleID, id)
}
func (m *cvMutation) InsertCVHistory(ctx context.Context, h decoder.VehicleCVValueHistory) error {
	return InsertCVHistory(ctx, m.tx, h.ID, h.VehicleID, h.CVValueID, h.OldValue, h.NewValue, h.ChangedAt)
}
func (m *cvMutation) InsertCVFile(ctx context.Context, file decoder.VehicleCVFile) error {
	if _, err := m.tx.ExecContext(ctx, `
INSERT INTO vehicle_cv_files(id, vehicle_id, file_name, original_name, description, decoder_profile, mime_type, size_bytes, storage_path, blob_id, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, file.ID, file.VehicleID, file.FileName, file.OriginalName, file.Description, file.DecoderProfile, file.MimeType, file.SizeBytes, file.StoragePath, file.BlobID, file.CreatedAt, file.UpdatedAt); err != nil {
		return fmt.Errorf("create vehicle cv file: %w", err)
	}
	return nil
}
func (m *cvMutation) DeleteCVFile(ctx context.Context, vehicleID, id string) error {
	if _, err := m.tx.ExecContext(ctx, `DELETE FROM vehicle_cv_files WHERE id=? AND vehicle_id=?`, id, vehicleID); err != nil {
		return fmt.Errorf("delete vehicle CV file: %w", err)
	}
	return nil
}
func (m *cvMutation) ECoSMappingCount(ctx context.Context, vehicleID, objectID string) (int, error) {
	var matches int
	err := m.tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM vehicle_external_mappings
WHERE vehicle_id=? AND provider='ecos' AND external_id=?`, vehicleID, objectID).Scan(&matches)
	return matches, err
}
func (m *cvMutation) CVAddressExists(ctx context.Context, vehicleID string, number int, cv31, cv32 *int) (bool, error) {
	var exists int
	err := m.tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM vehicle_cv_values
WHERE vehicle_id=? AND cv_number=? AND COALESCE(index_cv31,-1)=COALESCE(?,-1)
AND COALESCE(index_cv32,-1)=COALESCE(?,-1)`, vehicleID, number, cv31, cv32).Scan(&exists)
	return exists > 0, err
}
