ALTER TABLE track_geometry_libraries
  ADD COLUMN verification_note TEXT NOT NULL DEFAULT '';
ALTER TABLE track_geometry_libraries
  ADD COLUMN verified_at TEXT;
ALTER TABLE track_geometry_libraries
  ADD COLUMN verified_by TEXT;

ALTER TABLE plan_track_objects
  ADD COLUMN geometry_snapshot_json TEXT;

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
WHERE geometry_snapshot_json IS NULL;
