-- Case-insensitive import matching and exhibition expiry.
CREATE INDEX idx_vehicles_inventory_nocase ON vehicles(inventory_number COLLATE NOCASE);
CREATE INDEX idx_vehicle_sets_inventory_nocase ON vehicle_sets(inventory_number COLLATE NOCASE);
CREATE INDEX idx_vehicles_manufacturer_nocase ON vehicles(manufacturer COLLATE NOCASE);
CREATE INDEX idx_vehicles_exhibition_active ON vehicles(id) WHERE exhibition=1;

-- Vehicle CV lists and the reverse lookup from a source file to its CVs.
CREATE INDEX idx_vehicle_cv_values_order
  ON vehicle_cv_values(vehicle_id, decoder_profile, index_cv31, index_cv32, cv_number);
-- The CV identity index already covers the previous vehicle/profile/number prefix.
DROP INDEX idx_vehicle_cv_values_vehicle_id;
CREATE INDEX idx_vehicle_cv_values_source ON vehicle_cv_values(source_file_id, id);
CREATE INDEX idx_vehicle_cv_history_vehicle_order
  ON vehicle_cv_value_history(vehicle_id, changed_at DESC, id DESC);

-- Keep the existing image index for promoting by sort_order without is_primary.
CREATE INDEX idx_vehicle_images_primary_order ON vehicle_images(vehicle_id, is_primary DESC, sort_order, created_at);

CREATE INDEX idx_exhibition_entries_list_order
  ON exhibition_entries(list_id, sort_order, locomotive_name COLLATE NOCASE, owner COLLATE NOCASE);
DROP INDEX idx_exhibition_entries_list;
