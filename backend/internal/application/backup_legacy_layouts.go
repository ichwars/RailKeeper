package application

import (
	"context"
	"database/sql"
	"fmt"
	"railkeeper/backend/migrations"
)

// Retired planning tables are retained only for lossless legacy backup compatibility.
func backfillRestoredTrackGeometrySnapshots(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
UPDATE plan_track_objects
SET geometry_snapshot_json = (
  SELECT json_object(
    'id', geometry.id,
    'libraryId', geometry.library_id,
    'articleNumber', geometry.article_number,
    'name', geometry.name,
    'kind', geometry.kind,
    'lengthMm', geometry.length_mm,
    'minimumRadiusMm', geometry.minimum_radius_mm,
    'geometry', json(geometry.geometry_json),
    'sourceUrl', geometry.source_url,
    'status', geometry.status,
    'createdAt', geometry.created_at
  )
  FROM track_geometry_definitions geometry
  WHERE geometry.id = plan_track_objects.geometry_id
)
WHERE geometry_snapshot_json IS NULL`); err != nil {
		return fmt.Errorf("backfill restored track geometry snapshots: %w", err)
	}
	return nil
}

func retireRestoredLegacyAllocations(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, migrations.LegacyLayoutRetirementSQL); err != nil {
		return fmt.Errorf("retire legacy allocations: %w", err)
	}
	return nil
}
