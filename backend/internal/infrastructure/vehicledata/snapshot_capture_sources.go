package vehicledata

import (
	"context"
	"database/sql"
	"errors"
	"os"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/decoder/workflow"
	"railkeeper/backend/internal/infrastructure/blobdata"
	"railkeeper/backend/internal/media"
)

func (m *snapshotCaptureMutation) SourceIDs(ctx context.Context, vehicleID, profile string) ([]string, error) {
	ids := []string{}
	rows, err := m.tx.QueryContext(ctx, `SELECT id FROM vehicle_cv_files WHERE vehicle_id=? AND COALESCE(decoder_profile,'')=?`,
		vehicleID, profile)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return ids, nil
}
func (m *snapshotCaptureMutation) SourceFile(ctx context.Context, id string) (*decoder.VehicleCVFile, bool, error) {
	file, err := ScanCVFile(m.tx.QueryRowContext(ctx, CVFileSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	return file, err == nil, err
}
func (m *snapshotCaptureMutation) ReadSourceBytes(path string) ([]byte, error) {
	if m.sourceRoot == "" {
		return nil, media.ErrSourceMissing
	}
	data, err := blobdata.ReadConfinedDataFile(m.sourceRoot, path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, media.ErrSourceMissing
	}
	return data, err
}
func (m *snapshotCaptureMutation) StoreSourceBytes(ctx context.Context, data []byte) (string, error) {
	blob, err := media.NewBlob(data)
	if err != nil {
		return "", err
	}
	if err := blobdata.Insert(ctx, m.tx, blob); err != nil {
		return "", err
	}
	return blob.ID, nil
}

const archivedSourceMatches = ` FROM decoder_cv_snapshot_sources s JOIN decoder_cv_snapshots p ON p.id=s.snapshot_id
WHERE p.vehicle_reference_id=? AND s.source_file_id=? AND s.content_state='stored'`

func (m *snapshotCaptureMutation) ArchivedVariants(ctx context.Context, vehicleID, sourceID string) (int, error) {
	var variants int
	if err := m.tx.QueryRowContext(ctx, `SELECT COUNT(DISTINCT s.blob_id)`+archivedSourceMatches, vehicleID, sourceID).Scan(&variants); err != nil {
		return 0, err
	}
	return variants, nil
}

type originalCursor struct{ rows *sql.Rows }

func (c *originalCursor) Next() bool   { return c.rows.Next() }
func (c *originalCursor) Err() error   { return c.rows.Err() }
func (c *originalCursor) Close() error { return c.rows.Close() }
func (c *originalCursor) Blob() (media.StoredBlob, error) {
	var blob media.StoredBlob
	err := c.rows.Scan(&blob.Compression, &blob.Data)
	return blob, err
}
func (m *snapshotCaptureMutation) ArchivedOriginals(ctx context.Context, vehicleID, sourceID string) (workflow.OriginalCursor, error) {
	rows, err := m.tx.QueryContext(ctx, `SELECT compression, data FROM file_blobs WHERE id IN
(SELECT DISTINCT s.blob_id`+archivedSourceMatches+`)`, vehicleID, sourceID)
	if err != nil {
		return nil, err
	}
	return &originalCursor{rows: rows}, nil
}
func (m *snapshotCaptureMutation) LatestArchivedSource(ctx context.Context, vehicleID, sourceID string) (*workflow.SnapshotSource, error) {
	item := workflow.SnapshotSource{ID: sourceID, State: "stored"}
	var blob sql.NullString
	if err := m.tx.QueryRowContext(ctx, `SELECT s.metadata_json, s.blob_id`+archivedSourceMatches+` ORDER BY p.recorded_at DESC, p.id DESC LIMIT 1`,
		vehicleID, sourceID).Scan(&item.Metadata, &blob); err != nil {
		return nil, err
	}
	item.BlobID = workflow.OptionalText{Text: blob.String, Valid: blob.Valid}
	return &item, nil
}
