PRAGMA defer_foreign_keys=ON;

DROP TRIGGER tr_plan_track_reservation_deactivate;

ALTER TABLE track_geometry_definitions RENAME TO track_geometry_definitions_legacy;
ALTER TABLE plan_track_objects RENAME TO plan_track_objects_legacy;
ALTER TABLE plan_track_object_reservations RENAME TO plan_track_object_reservations_legacy;

CREATE TABLE track_geometry_definitions (
  id TEXT PRIMARY KEY,
  library_id TEXT NOT NULL,
  article_number TEXT NOT NULL,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('straight', 'curve', 'turnout', 'crossing', 'flex')),
  length_mm REAL NOT NULL CHECK (length_mm > 0),
  minimum_radius_mm REAL CHECK (minimum_radius_mm IS NULL OR minimum_radius_mm > 0),
  geometry_json TEXT NOT NULL,
  source_url TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('draft', 'verified', 'retired')),
  created_at TEXT NOT NULL,
  FOREIGN KEY (library_id) REFERENCES track_geometry_libraries(id) ON DELETE RESTRICT,
  UNIQUE (library_id, article_number)
);

CREATE TABLE plan_track_objects (
  id TEXT PRIMARY KEY,
  revision_id TEXT NOT NULL,
  geometry_id TEXT NOT NULL,
  position_x_mm REAL NOT NULL,
  position_y_mm REAL NOT NULL,
  rotation_degrees REAL NOT NULL CHECK (rotation_degrees >= 0 AND rotation_degrees < 360),
  version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  lineage_id TEXT NOT NULL DEFAULT '',
  elevation_start_mm REAL NOT NULL DEFAULT 0,
  elevation_end_mm REAL NOT NULL DEFAULT 0,
  flex_path_json TEXT,
  FOREIGN KEY (revision_id) REFERENCES plan_revisions(id) ON DELETE CASCADE,
  FOREIGN KEY (geometry_id) REFERENCES track_geometry_definitions(id) ON DELETE RESTRICT
);

CREATE TABLE plan_track_object_reservations (
  reservation_id TEXT PRIMARY KEY,
  track_object_id TEXT NOT NULL,
  active INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (reservation_id) REFERENCES accessory_reservations(id) ON DELETE CASCADE,
  FOREIGN KEY (track_object_id) REFERENCES plan_track_objects(id) ON DELETE RESTRICT
);

INSERT INTO track_geometry_definitions(
  id, library_id, article_number, name, kind, length_mm, minimum_radius_mm,
  geometry_json, source_url, status, created_at
)
SELECT id, library_id, article_number, name, kind, length_mm, NULL,
       geometry_json, source_url, status, created_at
FROM track_geometry_definitions_legacy;

INSERT INTO plan_track_objects(
  id, revision_id, geometry_id, position_x_mm, position_y_mm, rotation_degrees,
  version, created_at, updated_at, lineage_id, elevation_start_mm, elevation_end_mm,
  flex_path_json
)
SELECT id, revision_id, geometry_id, position_x_mm, position_y_mm, rotation_degrees,
       version, created_at, updated_at, lineage_id, elevation_start_mm, elevation_end_mm,
       NULL
FROM plan_track_objects_legacy;

INSERT INTO plan_track_object_reservations(
  reservation_id, track_object_id, active, created_at, updated_at
)
SELECT reservation_id, track_object_id, active, created_at, updated_at
FROM plan_track_object_reservations_legacy;

DROP TABLE plan_track_object_reservations_legacy;
DROP TABLE plan_track_objects_legacy;
DROP TABLE track_geometry_definitions_legacy;

CREATE INDEX ix_track_geometry_definitions_library
  ON track_geometry_definitions(library_id, status, article_number COLLATE NOCASE);
CREATE INDEX ix_plan_track_objects_revision
  ON plan_track_objects(revision_id, created_at, id);
CREATE INDEX ix_plan_track_objects_geometry
  ON plan_track_objects(geometry_id);
CREATE UNIQUE INDEX ux_plan_track_objects_revision_lineage
  ON plan_track_objects(revision_id, lineage_id);
CREATE UNIQUE INDEX ux_plan_track_object_reservations_active_object
  ON plan_track_object_reservations(track_object_id) WHERE active=1;
CREATE INDEX ix_plan_track_object_reservations_object
  ON plan_track_object_reservations(track_object_id, active);

CREATE TRIGGER tr_plan_track_objects_default_lineage
AFTER INSERT ON plan_track_objects
WHEN NEW.lineage_id=''
BEGIN
  UPDATE plan_track_objects SET lineage_id=NEW.id WHERE id=NEW.id;
END;

CREATE TRIGGER tr_plan_track_reservation_deactivate
AFTER UPDATE OF status ON accessory_reservations
WHEN OLD.status='active' AND NEW.status<>'active'
BEGIN
  UPDATE plan_track_object_reservations
  SET active=0, updated_at=NEW.updated_at
  WHERE reservation_id=NEW.id AND active=1;
END;

ALTER TABLE layouts ADD COLUMN minimum_flex_radius_mm REAL
  CHECK(minimum_flex_radius_mm IS NULL OR minimum_flex_radius_mm > 0);

INSERT INTO track_geometry_definitions(
  id, library_id, article_number, name, kind, length_mm, minimum_radius_mm,
  geometry_json, source_url, status, created_at
) VALUES(
  'tillig-tt-modellgleis-83125-v1',
  'tillig-tt-modellgleis-v1',
  '83125',
  'Flexgleis Holzschwelle',
  'flex',
  664,
  543,
  '{"schemaVersion":1,"ports":[{"id":"a","xMm":0,"yMm":0,"directionDegrees":180},{"id":"b","xMm":664,"yMm":0,"directionDegrees":0}],"routes":[{"id":"main","points":[{"xMm":0,"yMm":0},{"xMm":664,"yMm":0}]}]}',
  'https://www.tillig.com/Produkte/produktinfo-83125.html',
  'verified',
  '2026-08-10T00:00:00Z'
);
