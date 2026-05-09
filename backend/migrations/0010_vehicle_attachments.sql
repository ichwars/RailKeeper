CREATE TABLE IF NOT EXISTS vehicle_attachments (
  id TEXT PRIMARY KEY,
  vehicle_id TEXT NOT NULL,
  file_name TEXT NOT NULL,
  original_name TEXT NOT NULL,
  description TEXT,
  category TEXT,
  mime_type TEXT,
  size_bytes INTEGER NOT NULL DEFAULT 0,
  storage_path TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_vehicle_attachments_vehicle_id ON vehicle_attachments(vehicle_id, created_at);
