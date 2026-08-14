CREATE TABLE track_geometry_libraries (
  id TEXT PRIMARY KEY,
  manufacturer TEXT NOT NULL,
  track_system TEXT NOT NULL,
  gauge TEXT NOT NULL,
  scale TEXT NOT NULL,
  version TEXT NOT NULL,
  source_url TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('draft', 'verified', 'retired')),
  created_at TEXT NOT NULL,
  UNIQUE (manufacturer, track_system, gauge, version)
);

CREATE TABLE track_geometry_definitions (
  id TEXT PRIMARY KEY,
  library_id TEXT NOT NULL,
  article_number TEXT NOT NULL,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('straight', 'curve', 'turnout', 'crossing')),
  length_mm REAL NOT NULL CHECK (length_mm > 0),
  geometry_json TEXT NOT NULL,
  source_url TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('draft', 'verified', 'retired')),
  created_at TEXT NOT NULL,
  FOREIGN KEY (library_id) REFERENCES track_geometry_libraries(id) ON DELETE RESTRICT,
  UNIQUE (library_id, article_number)
);

CREATE INDEX ix_track_geometry_definitions_library
  ON track_geometry_definitions(library_id, status, article_number COLLATE NOCASE);

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
  FOREIGN KEY (revision_id) REFERENCES plan_revisions(id) ON DELETE CASCADE,
  FOREIGN KEY (geometry_id) REFERENCES track_geometry_definitions(id) ON DELETE RESTRICT
);

CREATE INDEX ix_plan_track_objects_revision
  ON plan_track_objects(revision_id, created_at, id);
CREATE INDEX ix_plan_track_objects_geometry
  ON plan_track_objects(geometry_id);

INSERT INTO track_geometry_libraries(
  id, manufacturer, track_system, gauge, scale, version, source_url, status, created_at
) VALUES(
  'tillig-tt-modellgleis-v1',
  'Tillig',
  'TT Modellgleis',
  'TT',
  '1:120',
  '1',
  'https://www.tillig.com/Produkte/produktinfo-83101.html',
  'verified',
  '2026-08-09T00:00:00Z'
);

INSERT INTO track_geometry_definitions(
  id, library_id, article_number, name, kind, length_mm, geometry_json,
  source_url, status, created_at
) VALUES(
  'tillig-tt-modellgleis-83101-v1',
  'tillig-tt-modellgleis-v1',
  '83101',
  'Gleisstück G1',
  'straight',
  166,
  '{"schemaVersion":1,"ports":[{"id":"a","xMm":0,"yMm":0,"directionDegrees":180},{"id":"b","xMm":166,"yMm":0,"directionDegrees":0}],"routes":[{"id":"main","points":[{"xMm":0,"yMm":0},{"xMm":166,"yMm":0}]}]}',
  'https://www.tillig.com/Produkte/produktinfo-83101.html',
  'verified',
  '2026-08-09T00:00:00Z'
);
