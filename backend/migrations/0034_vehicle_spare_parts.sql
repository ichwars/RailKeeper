CREATE TABLE IF NOT EXISTS vehicle_spare_parts (
  id TEXT PRIMARY KEY,
  vehicle_id TEXT NOT NULL,
  article_number TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  price TEXT NOT NULL DEFAULT '',
  url TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_vehicle_spare_parts_vehicle_id
  ON vehicle_spare_parts(vehicle_id, article_number, created_at);
