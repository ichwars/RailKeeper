CREATE TABLE vehicle_cv_values_indexed (
  id TEXT PRIMARY KEY,
  vehicle_id TEXT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
  cv_number INTEGER NOT NULL,
  value INTEGER NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL DEFAULT '',
  decoder_profile TEXT NOT NULL DEFAULT '',
  source_file_id TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  protocol TEXT NOT NULL DEFAULT '',
  index_cv31 INTEGER,
  index_cv32 INTEGER,
  CHECK ((index_cv31 IS NULL AND index_cv32 IS NULL) OR
    (index_cv31 IS NOT NULL AND index_cv32 IS NOT NULL AND
     index_cv31 BETWEEN 0 AND 255 AND index_cv32 BETWEEN 0 AND 255 AND cv_number BETWEEN 257 AND 512))
);

INSERT INTO vehicle_cv_values_indexed
  (id, vehicle_id, cv_number, value, description, category, decoder_profile, source_file_id,
   created_at, updated_at, protocol)
SELECT id, vehicle_id, cv_number, value, description, category, decoder_profile, source_file_id,
       created_at, updated_at, protocol FROM vehicle_cv_values;

CREATE TABLE vehicle_cv_value_history_indexed (
  id TEXT PRIMARY KEY,
  cv_value_id TEXT NOT NULL REFERENCES vehicle_cv_values_indexed(id) ON DELETE CASCADE,
  vehicle_id TEXT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
  old_value INTEGER NOT NULL,
  new_value INTEGER NOT NULL,
  changed_at TEXT NOT NULL
);
INSERT INTO vehicle_cv_value_history_indexed SELECT * FROM vehicle_cv_value_history;
DROP TABLE vehicle_cv_value_history;
DROP TABLE vehicle_cv_values;
ALTER TABLE vehicle_cv_values_indexed RENAME TO vehicle_cv_values;
ALTER TABLE vehicle_cv_value_history_indexed RENAME TO vehicle_cv_value_history;

CREATE UNIQUE INDEX idx_vehicle_cv_values_identity ON vehicle_cv_values
  (vehicle_id, decoder_profile, cv_number, COALESCE(index_cv31, -1), COALESCE(index_cv32, -1));
CREATE INDEX idx_vehicle_cv_values_vehicle_id ON vehicle_cv_values(vehicle_id, decoder_profile, cv_number);
CREATE INDEX idx_vehicle_cv_value_history_value_id ON vehicle_cv_value_history(cv_value_id, changed_at);
