package application

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

type decoderSnapshotSource struct {
	ID, Metadata, State string
	BlobID              sql.NullString
}

func loadDecoderSnapshotSourcesTx(ctx context.Context, tx *sql.Tx, workspace decoderWorkspaceRecord,
	values []VehicleCVValue, sourceRoot string) ([]decoderSnapshotSource, error) {
	ids := map[string]bool{}
	for _, value := range values {
		if value.SourceFileID != "" {
			ids[value.SourceFileID] = true
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT id FROM vehicle_cv_files WHERE vehicle_id=? AND COALESCE(decoder_profile,'')=?`,
		workspace.VehicleID, workspace.Profile)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(ids))
	for id := range ids {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	sources := make([]decoderSnapshotSource, 0, len(keys))
	for _, id := range keys {
		source, err := freezeDecoderSnapshotSourceTx(ctx, tx, workspace.VehicleID, id, sourceRoot)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, nil
}

func freezeDecoderSnapshotSourceTx(ctx context.Context, tx *sql.Tx, vehicleID, sourceID, sourceRoot string) (decoderSnapshotSource, error) {
	source := decoderSnapshotSource{ID: sourceID, State: "unavailable"}
	file, err := scanVehicleCVFile(tx.QueryRowContext(ctx, vehicleCVFileSelect+` WHERE id=?`, sourceID))
	if errors.Is(err, sql.ErrNoRows) {
		file = &VehicleCVFile{ID: sourceID}
	} else if err != nil {
		return source, fmt.Errorf("read original decoder snapshot source: %w", err)
	}
	file.Revision, file.SourceReferenceCount = "", 0
	metadata, err := json.Marshal(file)
	if err != nil {
		return source, err
	}
	source.Metadata = string(metadata)
	// A legacy reference to another vehicle is preserved, not silently reassigned.
	if file.VehicleID != vehicleID {
		if file.VehicleID == "" {
			return resolveArchivedDecoderSourceTx(ctx, tx, vehicleID, source)
		}
		return source, nil
	}
	if file.BlobID != "" {
		source.BlobID = sql.NullString{String: file.BlobID, Valid: true}
	} else if file.StoragePath != "" && sourceRoot != "" {
		data, err := NewFileBlobService(nil, sourceRoot).readConfinedDataFile(file.StoragePath)
		if errors.Is(err, os.ErrNotExist) {
			return resolveArchivedDecoderSourceTx(ctx, tx, vehicleID, source)
		}
		if err != nil {
			return source, fmt.Errorf("read decoder source inside snapshot data directory: %w", err)
		}
		blob, err := storeFileBlob(ctx, tx, data)
		if err != nil {
			return source, err
		}
		source.BlobID = sql.NullString{String: blob, Valid: true}
	}
	if source.BlobID.Valid {
		source.State = "stored"
		return source, nil
	}
	return resolveArchivedDecoderSourceTx(ctx, tx, vehicleID, source)
}

func resolveArchivedDecoderSourceTx(ctx context.Context, tx *sql.Tx, vehicleID string, source decoderSnapshotSource) (decoderSnapshotSource, error) {
	archived, err := readArchivedDecoderSourceTx(ctx, tx, vehicleID, source.ID)
	if err != nil {
		return source, err
	}
	if archived != nil {
		return *archived, nil
	}
	return source, nil
}

func readArchivedDecoderSourceTx(ctx context.Context, tx *sql.Tx, vehicleID, sourceID string) (*decoderSnapshotSource, error) {
	const matches = ` FROM decoder_cv_snapshot_sources s JOIN decoder_cv_snapshots p ON p.id=s.snapshot_id
WHERE p.vehicle_reference_id=? AND s.source_file_id=? AND s.content_state='stored'`
	var variants int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(DISTINCT s.blob_id)`+matches, vehicleID, sourceID).Scan(&variants); err != nil {
		return nil, err
	}
	if variants == 0 {
		return nil, nil
	}
	if variants > 1 {
		// Different storage identities can still contain exactly the same original bytes.
		rows, err := tx.QueryContext(ctx, `SELECT compression, data FROM file_blobs WHERE id IN
(SELECT DISTINCT s.blob_id`+matches+`)`, vehicleID, sourceID)
		if err != nil {
			return nil, err
		}
		defer func() { _ = rows.Close() }()
		var original []byte
		seen := false
		for rows.Next() {
			var compression string
			var stored []byte
			if err := rows.Scan(&compression, &stored); err != nil {
				return nil, err
			}
			content, err := decompressBlobData(stored, compression)
			if err != nil || (seen && !bytes.Equal(original, content)) {
				return nil, nil // Conflicting or unreadable originals stay unresolved.
			}
			original, seen = content, true
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		if err := rows.Close(); err != nil {
			return nil, err
		}
	}
	item := decoderSnapshotSource{ID: sourceID, State: "stored"}
	if err := tx.QueryRowContext(ctx, `SELECT s.metadata_json, s.blob_id`+matches+` ORDER BY p.recorded_at DESC, p.id DESC LIMIT 1`,
		vehicleID, sourceID).Scan(&item.Metadata, &item.BlobID); err != nil {
		return nil, err
	}
	return &item, nil
}
