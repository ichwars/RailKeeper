CREATE TABLE catalog_articles (
  id TEXT PRIMARY KEY,
  origin TEXT NOT NULL CHECK (origin IN ('legacy_vehicle', 'legacy_set', 'manual')),
  source_entity_id TEXT,
  created_at TEXT NOT NULL,
  CHECK ((origin = 'manual' AND source_entity_id IS NULL)
    OR (origin <> 'manual' AND source_entity_id IS NOT NULL AND length(source_entity_id) > 0))
);

CREATE UNIQUE INDEX idx_catalog_article_legacy_source
  ON catalog_articles(origin, source_entity_id) WHERE origin <> 'manual';

CREATE TABLE catalog_article_revisions (
  id TEXT PRIMARY KEY,
  article_id TEXT NOT NULL REFERENCES catalog_articles(id) ON DELETE RESTRICT,
  revision_number INTEGER NOT NULL CHECK (typeof(revision_number) = 'integer' AND revision_number BETWEEN 1 AND 2147483647),
  fields_json TEXT NOT NULL CHECK (json_valid(fields_json) AND json_type(fields_json) = 'object'),
  note TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  UNIQUE (article_id, revision_number)
);

CREATE TRIGGER catalog_article_revision_immutable
BEFORE UPDATE ON catalog_article_revisions
BEGIN
  SELECT RAISE(ABORT, 'catalog article revisions are immutable');
END;

CREATE TABLE vehicle_article_links (
  vehicle_id TEXT PRIMARY KEY REFERENCES vehicles(id) ON DELETE CASCADE,
  revision_id TEXT NOT NULL REFERENCES catalog_article_revisions(id) ON DELETE RESTRICT,
  overrides_json TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(overrides_json) AND json_type(overrides_json) = 'object'),
  authority TEXT NOT NULL DEFAULT 'legacy' CHECK (authority IN ('legacy', 'catalog')),
  updated_at TEXT NOT NULL
);

CREATE INDEX idx_vehicle_article_revision ON vehicle_article_links(revision_id);

CREATE TABLE vehicle_set_article_links (
  vehicle_set_id TEXT PRIMARY KEY REFERENCES vehicle_sets(id) ON DELETE CASCADE,
  revision_id TEXT NOT NULL REFERENCES catalog_article_revisions(id) ON DELETE RESTRICT,
  overrides_json TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(overrides_json) AND json_type(overrides_json) = 'object'),
  authority TEXT NOT NULL DEFAULT 'legacy' CHECK (authority IN ('legacy', 'catalog')),
  updated_at TEXT NOT NULL
);

CREATE INDEX idx_vehicle_set_article_revision ON vehicle_set_article_links(revision_id);
