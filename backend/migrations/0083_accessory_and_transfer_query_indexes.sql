-- Stock calculations and lookups by normalized article identity or attribute key.
CREATE INDEX idx_accessory_reservations_active_stock
  ON accessory_reservations(product_id, location_id, asset_id, quantity) WHERE status='active';
CREATE INDEX idx_accessory_installations_active_product
  ON accessory_installations(product_id, asset_id, quantity) WHERE removed_at IS NULL;
CREATE INDEX idx_accessory_products_normalized_article
  ON accessory_products(lower(trim(manufacturer)), lower(trim(article_number)), name COLLATE NOCASE, id);
CREATE INDEX idx_accessory_products_inventory_nocase ON accessory_products(inventory_number COLLATE NOCASE);
CREATE INDEX idx_accessory_attributes_key_product ON accessory_product_attributes(attribute_key, product_id);

-- Keep creation order when timestamps tie, including history for a selected profile.
CREATE INDEX idx_data_transfer_jobs_created_order ON data_transfer_jobs(created_at DESC, id DESC);
CREATE INDEX idx_data_transfer_jobs_profile_order ON data_transfer_jobs(profile_id, created_at DESC, id DESC);
CREATE INDEX idx_data_transfer_issues_job_order ON data_transfer_job_issues(job_id, created_at, id);
DROP INDEX idx_data_transfer_jobs_created_at;
DROP INDEX idx_data_transfer_issues_job_id;
