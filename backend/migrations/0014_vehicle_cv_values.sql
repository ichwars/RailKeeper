CREATE TABLE IF NOT EXISTS vehicle_cv_values (
  id TEXT PRIMARY KEY,
  vehicle_id TEXT NOT NULL,
  cv_number INTEGER NOT NULL,
  value INTEGER NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL DEFAULT '',
  decoder_profile TEXT NOT NULL DEFAULT '',
  source_file_id TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE,
  UNIQUE (vehicle_id, cv_number, decoder_profile)
);

CREATE INDEX IF NOT EXISTS idx_vehicle_cv_values_vehicle_id
  ON vehicle_cv_values(vehicle_id, decoder_profile, cv_number);

CREATE TABLE IF NOT EXISTS vehicle_cv_value_history (
  id TEXT PRIMARY KEY,
  cv_value_id TEXT NOT NULL,
  vehicle_id TEXT NOT NULL,
  old_value INTEGER NOT NULL,
  new_value INTEGER NOT NULL,
  changed_at TEXT NOT NULL,
  FOREIGN KEY (cv_value_id) REFERENCES vehicle_cv_values(id) ON DELETE CASCADE,
  FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_vehicle_cv_value_history_value_id
  ON vehicle_cv_value_history(cv_value_id, changed_at);

CREATE TABLE IF NOT EXISTS vehicle_cv_files (
  id TEXT PRIMARY KEY,
  vehicle_id TEXT NOT NULL,
  file_name TEXT NOT NULL,
  original_name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  decoder_profile TEXT NOT NULL DEFAULT '',
  mime_type TEXT NOT NULL DEFAULT '',
  size_bytes INTEGER NOT NULL DEFAULT 0,
  storage_path TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_vehicle_cv_files_vehicle_id
  ON vehicle_cv_files(vehicle_id, created_at);
