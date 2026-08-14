CREATE TABLE plan_free_objects (
  id TEXT PRIMARY KEY,
  lineage_id TEXT NOT NULL,
  revision_id TEXT NOT NULL REFERENCES plan_revisions(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  category TEXT NOT NULL CHECK(category IN ('structure', 'platform', 'scenery', 'annotation')),
  position_x_mm REAL NOT NULL,
  position_y_mm REAL NOT NULL,
  rotation_degrees REAL NOT NULL,
  shape_json TEXT NOT NULL,
  version INTEGER NOT NULL DEFAULT 1 CHECK(version >= 1),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(revision_id, lineage_id)
);

CREATE INDEX idx_plan_free_objects_revision
ON plan_free_objects(revision_id, created_at, id);
