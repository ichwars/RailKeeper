CREATE TABLE IF NOT EXISTS file_blobs (
  id TEXT PRIMARY KEY,
  original_size INTEGER NOT NULL DEFAULT 0,
  compressed_size INTEGER NOT NULL DEFAULT 0,
  compression TEXT NOT NULL DEFAULT 'zlib',
  sha256 TEXT NOT NULL DEFAULT '',
  data BLOB NOT NULL,
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_file_blobs_sha256 ON file_blobs(sha256);

ALTER TABLE vehicle_images ADD COLUMN blob_id TEXT NOT NULL DEFAULT '';
ALTER TABLE vehicle_images ADD COLUMN thumbnail_blob_id TEXT NOT NULL DEFAULT '';
ALTER TABLE vehicle_attachments ADD COLUMN blob_id TEXT NOT NULL DEFAULT '';
ALTER TABLE vehicle_cv_files ADD COLUMN blob_id TEXT NOT NULL DEFAULT '';
