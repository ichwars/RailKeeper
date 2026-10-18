ALTER TABLE vehicle_cv_values ADD COLUMN source_line INTEGER
  CHECK (source_line IS NULL OR source_line BETWEEN 1 AND 10000000);
ALTER TABLE vehicle_cv_values ADD COLUMN source_exported_at TEXT NOT NULL DEFAULT ''
  CHECK (source_exported_at = '' OR length(source_exported_at) = 19);
