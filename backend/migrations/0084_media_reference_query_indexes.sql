-- Reverse references used when deleting an unreferenced blob.
CREATE INDEX idx_vehicle_images_blob ON vehicle_images(blob_id);
CREATE INDEX idx_vehicle_images_thumbnail_blob ON vehicle_images(thumbnail_blob_id);
CREATE INDEX idx_vehicle_attachments_blob ON vehicle_attachments(blob_id);
CREATE INDEX idx_vehicle_cv_files_blob ON vehicle_cv_files(blob_id);
CREATE INDEX idx_vehicle_sets_image_blob ON vehicle_sets(set_image_blob_id) WHERE set_image_blob_id IS NOT NULL;
CREATE INDEX idx_vehicle_sets_thumbnail_blob
  ON vehicle_sets(set_image_thumbnail_blob_id) WHERE set_image_thumbnail_blob_id IS NOT NULL;
