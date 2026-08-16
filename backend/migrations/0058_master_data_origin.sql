ALTER TABLE master_data_entries
ADD COLUMN origin TEXT NOT NULL DEFAULT 'custom'
  CHECK (origin IN ('bundled', 'custom'));

CREATE INDEX idx_master_data_entries_origin
  ON master_data_entries(origin, type, key);
