package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

func (repository *TrackPlannerRepository) ListTrackLibraries(
	ctx context.Context,
) ([]domain.TrackGeometryLibrary, error) {
	rows, err := repository.db.QueryContext(ctx, trackLibraryListSelect+`
ORDER BY library.manufacturer COLLATE NOCASE, library.track_system COLLATE NOCASE,
         library.version COLLATE NOCASE, library.id`)
	if err != nil {
		return nil, fmt.Errorf("list track libraries: %w", err)
	}
	defer func() { _ = rows.Close() }()
	libraries := []domain.TrackGeometryLibrary{}
	for rows.Next() {
		library, err := scanTrackLibrary(rows)
		if err != nil {
			return nil, fmt.Errorf("scan track library: %w", err)
		}
		libraries = append(libraries, *library)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate track libraries: %w", err)
	}
	return libraries, nil
}

func (repository *TrackPlannerRepository) TrackLibraryVersionExists(
	ctx context.Context,
	metadata domain.TrackLibraryPackageMetadata,
) (bool, error) {
	var count int
	if err := repository.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM track_geometry_libraries
WHERE manufacturer=? COLLATE NOCASE AND track_system=? COLLATE NOCASE
  AND gauge=? COLLATE NOCASE AND version=? COLLATE NOCASE`, metadata.Manufacturer,
		metadata.TrackSystem, metadata.Gauge, metadata.Version).Scan(&count); err != nil {
		return false, fmt.Errorf("check track library version: %w", err)
	}
	return count > 0, nil
}

func (repository *TrackPlannerRepository) ImportTrackLibrary(
	ctx context.Context,
	doc domain.TrackLibraryPackage,
	actor string,
) (*domain.TrackGeometryLibrary, error) {
	now := timestamp()
	libraryID := randomID()
	err := repository.withTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO track_geometry_libraries(
  id, manufacturer, track_system, gauge, scale, version, source_url, status,
  verification_note, verified_at, verified_by, created_at
) VALUES(?, ?, ?, ?, ?, ?, ?, 'draft', '', NULL, NULL, ?)`, libraryID,
			doc.Library.Manufacturer, doc.Library.TrackSystem, doc.Library.Gauge, doc.Library.Scale,
			doc.Library.Version, doc.Library.SourceURL, now); err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrTrackLibraryConflict
			}
			return fmt.Errorf("insert track library: %w", err)
		}
		for _, definition := range doc.Definitions {
			geometryJSON, err := json.Marshal(definition.Geometry)
			if err != nil {
				return application.ErrTrackLibraryValidation
			}
			if _, err := tx.ExecContext(ctx, `
INSERT INTO track_geometry_definitions(
  id, library_id, article_number, name, kind, length_mm, minimum_radius_mm,
  geometry_json, source_url, status, created_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, 'draft', ?)`, randomID(), libraryID,
				definition.ArticleNumber, definition.Name, definition.Kind, definition.LengthMM,
				definition.MinimumRadiusMM, string(geometryJSON), definition.SourceURL, now); err != nil {
				if isSQLiteConstraint(err) {
					return application.ErrTrackLibraryConflict
				}
				return fmt.Errorf("insert track geometry definition: %w", err)
			}
		}
		return writeLayoutAudit(
			ctx, tx, "TrackGeometryLibraryImported", "track_geometry_library", libraryID, actor, now,
		)
	})
	if err != nil {
		return nil, err
	}
	return repository.getTrackLibrary(ctx, libraryID)
}

func (repository *TrackPlannerRepository) ExportTrackLibrary(
	ctx context.Context,
	id string,
) (*domain.TrackLibraryPackage, error) {
	library, err := repository.getTrackLibrary(ctx, id)
	if err != nil {
		return nil, err
	}
	doc := &domain.TrackLibraryPackage{
		Format: domain.TrackLibraryPackageFormat, SchemaVersion: domain.TrackLibraryPackageSchemaVersion,
		ExportedAt: timestamp(),
		Library: domain.TrackLibraryPackageMetadata{
			Manufacturer: library.Manufacturer, TrackSystem: library.TrackSystem, Gauge: library.Gauge,
			Scale: library.Scale, Version: library.Version, SourceURL: library.SourceURL,
			Status: library.Status,
		},
		Definitions: []domain.TrackLibraryPackageDefinition{},
	}
	rows, err := repository.db.QueryContext(ctx, trackGeometrySelect+`
WHERE geometry.library_id=?
ORDER BY geometry.article_number COLLATE NOCASE, geometry.id`, id)
	if err != nil {
		return nil, fmt.Errorf("export track library definitions: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		definition, err := scanTrackGeometry(rows)
		if err != nil {
			return nil, fmt.Errorf("scan track library export: %w", err)
		}
		doc.Definitions = append(doc.Definitions, domain.TrackLibraryPackageDefinition{
			ArticleNumber: definition.ArticleNumber, Name: definition.Name, Kind: definition.Kind,
			LengthMM: definition.LengthMM, MinimumRadiusMM: definition.MinimumRadiusMM,
			Geometry: definition.Geometry, SourceURL: definition.SourceURL, Status: definition.Status,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate track library export: %w", err)
	}
	return doc, nil
}

func (repository *TrackPlannerRepository) UpdateTrackLibraryStatus(
	ctx context.Context,
	id string,
	status domain.TrackGeometryStatus,
	note string,
	actor string,
) (*domain.TrackGeometryLibrary, error) {
	now := timestamp()
	err := repository.withTx(ctx, func(tx *sql.Tx) error {
		var current domain.TrackGeometryStatus
		if err := tx.QueryRowContext(ctx, `
SELECT status FROM track_geometry_libraries WHERE id=?`, id).Scan(&current); errors.Is(err, sql.ErrNoRows) {
			return application.ErrTrackLibraryNotFound
		} else if err != nil {
			return fmt.Errorf("read track library status: %w", err)
		}
		action := "TrackGeometryLibraryRetired"
		if status == domain.TrackGeometryVerified {
			action = "TrackGeometryLibraryVerified"
			if _, err := tx.ExecContext(ctx, `
UPDATE track_geometry_definitions SET status='verified' WHERE library_id=?`, id); err != nil {
				return fmt.Errorf("verify track geometry definitions: %w", err)
			}
			if _, err := tx.ExecContext(ctx, `
UPDATE track_geometry_libraries
SET status='verified', verification_note=?, verified_at=?, verified_by=?
WHERE id=?`, note, now, actor, id); err != nil {
				return fmt.Errorf("verify track library: %w", err)
			}
		} else if _, err := tx.ExecContext(ctx, `
UPDATE track_geometry_libraries SET status='retired' WHERE id=?`, id); err != nil {
			return fmt.Errorf("retire track library: %w", err)
		}
		return writeLayoutAudit(ctx, tx, action, "track_geometry_library", id, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return repository.getTrackLibrary(ctx, id)
}

func (repository *TrackPlannerRepository) getTrackLibrary(
	ctx context.Context,
	id string,
) (*domain.TrackGeometryLibrary, error) {
	library, err := scanTrackLibrary(repository.db.QueryRowContext(ctx, trackLibraryListSelect+`
HAVING library.id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrTrackLibraryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get track library: %w", err)
	}
	return library, nil
}

const trackLibraryListSelect = `
SELECT library.id, library.manufacturer, library.track_system, library.gauge, library.scale,
       library.version, library.source_url, library.status, library.verification_note,
       COALESCE(library.verified_at, ''), COALESCE(library.verified_by, ''),
       COUNT(geometry.id), library.created_at
FROM track_geometry_libraries library
LEFT JOIN track_geometry_definitions geometry ON geometry.library_id=library.id
GROUP BY library.id, library.manufacturer, library.track_system, library.gauge, library.scale,
         library.version, library.source_url, library.status, library.verification_note,
         library.verified_at, library.verified_by, library.created_at
`

func scanTrackLibrary(scanner trackScanner) (*domain.TrackGeometryLibrary, error) {
	library := &domain.TrackGeometryLibrary{}
	if err := scanner.Scan(&library.ID, &library.Manufacturer, &library.TrackSystem, &library.Gauge,
		&library.Scale, &library.Version, &library.SourceURL, &library.Status,
		&library.VerificationNote, &library.VerifiedAt, &library.VerifiedBy,
		&library.DefinitionCount, &library.CreatedAt); err != nil {
		return nil, err
	}
	return library, nil
}
