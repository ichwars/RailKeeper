ALTER TABLE vehicle_sets
  ADD COLUMN main_image_mode TEXT NOT NULL DEFAULT 'automatic'
  CHECK (main_image_mode IN ('automatic', 'member', 'dedicated'));

ALTER TABLE vehicle_sets
  ADD COLUMN main_member_image_id TEXT;

ALTER TABLE vehicle_sets ADD COLUMN set_image_file_name TEXT NOT NULL DEFAULT '';
ALTER TABLE vehicle_sets ADD COLUMN set_image_mime_type TEXT NOT NULL DEFAULT '';
ALTER TABLE vehicle_sets ADD COLUMN set_image_blob_id TEXT REFERENCES file_blobs(id);
ALTER TABLE vehicle_sets ADD COLUMN set_image_thumbnail_blob_id TEXT REFERENCES file_blobs(id);
ALTER TABLE vehicle_sets ADD COLUMN set_image_created_at TEXT NOT NULL DEFAULT '';
ALTER TABLE vehicle_sets ADD COLUMN set_image_updated_at TEXT NOT NULL DEFAULT '';

CREATE TRIGGER vehicle_sets_cleanup_owned_blobs
AFTER DELETE ON vehicle_sets
BEGIN
  DELETE FROM file_blobs
  WHERE id IN (OLD.set_image_blob_id, OLD.set_image_thumbnail_blob_id)
    AND NOT EXISTS (SELECT 1 FROM vehicle_images WHERE blob_id=file_blobs.id OR thumbnail_blob_id=file_blobs.id)
    AND NOT EXISTS (SELECT 1 FROM vehicle_attachments WHERE blob_id=file_blobs.id)
    AND NOT EXISTS (SELECT 1 FROM vehicle_cv_files WHERE blob_id=file_blobs.id)
    AND NOT EXISTS (SELECT 1 FROM accessory_documents WHERE file_blob_id=file_blobs.id)
    AND NOT EXISTS (
      SELECT 1 FROM vehicle_sets
      WHERE set_image_blob_id=file_blobs.id OR set_image_thumbnail_blob_id=file_blobs.id
    );
END;
