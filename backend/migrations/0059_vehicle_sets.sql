CREATE TABLE vehicle_sets (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  manufacturer TEXT NOT NULL,
  article_number TEXT NOT NULL DEFAULT '',
  article_source_url TEXT NOT NULL DEFAULT '',
  gauge TEXT NOT NULL,
  epoch TEXT NOT NULL DEFAULT '',
  railway_company TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL,
  gattung TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  ean TEXT NOT NULL DEFAULT '',
  production_period TEXT NOT NULL DEFAULT '',
  list_price TEXT NOT NULL DEFAULT '',
  acquisition_type TEXT NOT NULL DEFAULT '',
  acquired_from TEXT NOT NULL DEFAULT '',
  purchase_price TEXT NOT NULL DEFAULT '',
  purchase_date TEXT NOT NULL DEFAULT '',
  storage_location TEXT NOT NULL DEFAULT '',
  storage_details TEXT NOT NULL DEFAULT '',
  condition TEXT NOT NULL DEFAULT '',
  condition_details TEXT NOT NULL DEFAULT '',
  packaging TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE vehicle_set_members (
  vehicle_set_id TEXT NOT NULL,
  vehicle_id TEXT NOT NULL UNIQUE,
  position INTEGER NOT NULL CHECK (position >= 1),
  label TEXT NOT NULL DEFAULT '',
  PRIMARY KEY (vehicle_set_id, position),
  FOREIGN KEY (vehicle_set_id) REFERENCES vehicle_sets(id) ON DELETE CASCADE,
  FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX idx_vehicle_set_members_set
  ON vehicle_set_members(vehicle_set_id, position);

CREATE INDEX idx_vehicle_sets_article
  ON vehicle_sets(manufacturer, article_number);
