package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/media"
)

func captureSnapshotSources(ctx context.Context, repository SnapshotCapture, workspace SnapshotWorkspace, values []decoder.VehicleCVValue) ([]SnapshotSource, error) {
	ids := map[string]bool{}
	for _, value := range values {
		if value.SourceFileID != "" {
			ids[value.SourceFileID] = true
		}
	}
	storedIDs, err := repository.SourceIDs(ctx, workspace.VehicleID, workspace.Profile)
	if err != nil {
		return nil, err
	}
	for _, id := range storedIDs {
		ids[id] = true
	}
	keys := make([]string, 0, len(ids))
	for id := range ids {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	sources := make([]SnapshotSource, 0, len(keys))
	for _, id := range keys {
		source, err := freezeSnapshotSource(ctx, repository, workspace.VehicleID, id)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, nil
}
func freezeSnapshotSource(ctx context.Context, repository SnapshotCapture, vehicleID, sourceID string) (SnapshotSource, error) {
	source := SnapshotSource{ID: sourceID, State: "unavailable"}
	file, found, err := repository.SourceFile(ctx, sourceID)
	if err != nil {
		return source, fmt.Errorf("read original decoder snapshot source: %w", err)
	}
	if !found {
		file = &decoder.VehicleCVFile{ID: sourceID}
	}
	file.Revision, file.SourceReferenceCount = "", 0
	metadata, err := json.Marshal(file)
	if err != nil {
		return source, err
	}
	source.Metadata = string(metadata)
	// Foreign legacy references retain their evidence and are never silently reassigned.
	if file.VehicleID != vehicleID {
		if file.VehicleID == "" {
			return resolveArchivedSource(ctx, repository, vehicleID, source)
		}
		return source, nil
	}
	if file.BlobID != "" {
		source.BlobID = OptionalText{Text: file.BlobID, Valid: true}
	} else if file.StoragePath != "" {
		data, err := repository.ReadSourceBytes(file.StoragePath)
		if errors.Is(err, media.ErrSourceMissing) {
			return resolveArchivedSource(ctx, repository, vehicleID, source)
		}
		if err != nil {
			return source, fmt.Errorf("read decoder source inside snapshot data directory: %w", err)
		}
		blob, err := repository.StoreSourceBytes(ctx, data)
		if err != nil {
			return source, err
		}
		source.BlobID = OptionalText{Text: blob, Valid: true}
	}
	if source.BlobID.Valid {
		source.State = "stored"
		return source, nil
	}
	return resolveArchivedSource(ctx, repository, vehicleID, source)
}
func resolveArchivedSource(ctx context.Context, repository SnapshotCapture, vehicleID string, source SnapshotSource) (SnapshotSource, error) {
	archived, err := readArchivedSource(ctx, repository, vehicleID, source.ID)
	if err != nil {
		return source, err
	}
	if archived != nil {
		return *archived, nil
	}
	return source, nil
}
func readArchivedSource(ctx context.Context, repository SnapshotCapture, vehicleID, sourceID string) (*SnapshotSource, error) {
	variants, err := repository.ArchivedVariants(ctx, vehicleID, sourceID)
	if err != nil {
		return nil, err
	}
	if variants == 0 {
		return nil, nil
	}
	if variants > 1 {
		originals, err := repository.ArchivedOriginals(ctx, vehicleID, sourceID)
		if err != nil {
			return nil, err
		}
		defer func() { _ = originals.Close() }()
		var original []byte
		seen := false
		for originals.Next() {
			stored, err := originals.Blob()
			if err != nil {
				return nil, err
			}
			content, err := media.Decode(stored.Data, stored.Compression)
			if err != nil || (seen && !bytes.Equal(original, content)) {
				return nil, nil
			}
			original, seen = content, true
		}
		if err := originals.Err(); err != nil {
			return nil, err
		}
		if err := originals.Close(); err != nil {
			return nil, err
		}
	}
	return repository.LatestArchivedSource(ctx, vehicleID, sourceID)
}
