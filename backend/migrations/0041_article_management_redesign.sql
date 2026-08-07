ALTER TABLE accessory_products ADD COLUMN ean TEXT NOT NULL DEFAULT '';
ALTER TABLE accessory_products ADD COLUMN manufacturer_status TEXT NOT NULL DEFAULT 'unknown'
  CHECK (manufacturer_status IN ('announced', 'available', 'discontinued', 'unknown'));
ALTER TABLE accessory_products ADD COLUMN article_type TEXT NOT NULL DEFAULT 'other';
ALTER TABLE accessory_products ADD COLUMN subtype TEXT NOT NULL DEFAULT '';
ALTER TABLE accessory_products ADD COLUMN gauges_json TEXT NOT NULL DEFAULT '[]';
ALTER TABLE accessory_products ADD COLUMN scale TEXT NOT NULL DEFAULT '';
ALTER TABLE accessory_products ADD COLUMN package_quantity INTEGER NOT NULL DEFAULT 1
  CHECK (package_quantity > 0);
ALTER TABLE accessory_products ADD COLUMN stock_unit TEXT NOT NULL DEFAULT 'piece';
ALTER TABLE accessory_products ADD COLUMN minimum_stock INTEGER NOT NULL DEFAULT 0
  CHECK (minimum_stock >= 0);
ALTER TABLE accessory_products ADD COLUMN inventory_strategy TEXT NOT NULL DEFAULT 'quantity'
  CHECK (inventory_strategy IN ('quantity', 'individual', 'quantity_later_individual'));
ALTER TABLE accessory_products ADD COLUMN manufacturer_url TEXT NOT NULL DEFAULT '';
ALTER TABLE accessory_products ADD COLUMN product_url TEXT NOT NULL DEFAULT '';
ALTER TABLE accessory_products ADD COLUMN alternative_numbers_json TEXT NOT NULL DEFAULT '[]';
ALTER TABLE accessory_products ADD COLUMN keywords_json TEXT NOT NULL DEFAULT '[]';
ALTER TABLE accessory_products ADD COLUMN compatibility_notes TEXT NOT NULL DEFAULT '';
ALTER TABLE accessory_products ADD COLUMN internal_notes TEXT NOT NULL DEFAULT '';
ALTER TABLE accessory_products ADD COLUMN archived INTEGER NOT NULL DEFAULT 0 CHECK (archived IN (0, 1));

UPDATE accessory_products
SET
  article_type = 'other',
  subtype = category,
  inventory_strategy = CASE tracking_mode
    WHEN 'individual' THEN 'individual'
    ELSE 'quantity'
  END;

DROP INDEX ux_accessory_products_article;
CREATE INDEX ix_accessory_products_article_lookup
  ON accessory_products(manufacturer COLLATE NOCASE, article_number COLLATE NOCASE);
CREATE INDEX ix_accessory_products_article_type ON accessory_products(article_type);
CREATE INDEX ix_accessory_products_archived ON accessory_products(archived);
CREATE INDEX ix_accessory_products_ean ON accessory_products(ean);

CREATE TABLE accessory_product_attributes (
  id TEXT PRIMARY KEY,
  product_id TEXT NOT NULL,
  attribute_key TEXT NOT NULL,
  value_type TEXT NOT NULL CHECK (value_type IN (
    'text', 'number', 'boolean', 'date', 'single_select', 'multi_select'
  )),
  text_value TEXT,
  number_value REAL,
  boolean_value INTEGER CHECK (boolean_value IS NULL OR boolean_value IN (0, 1)),
  date_value TEXT,
  single_select_value TEXT,
  multi_select_value TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (product_id) REFERENCES accessory_products(id) ON DELETE CASCADE,
  UNIQUE (product_id, attribute_key),
  CHECK (
    (value_type = 'text' AND text_value IS NOT NULL AND number_value IS NULL AND boolean_value IS NULL
      AND date_value IS NULL AND single_select_value IS NULL AND multi_select_value IS NULL)
    OR (value_type = 'number' AND text_value IS NULL AND number_value IS NOT NULL AND boolean_value IS NULL
      AND date_value IS NULL AND single_select_value IS NULL AND multi_select_value IS NULL)
    OR (value_type = 'boolean' AND text_value IS NULL AND number_value IS NULL AND boolean_value IS NOT NULL
      AND date_value IS NULL AND single_select_value IS NULL AND multi_select_value IS NULL)
    OR (value_type = 'date' AND text_value IS NULL AND number_value IS NULL AND boolean_value IS NULL
      AND date_value IS NOT NULL AND single_select_value IS NULL AND multi_select_value IS NULL)
    OR (value_type = 'single_select' AND text_value IS NULL AND number_value IS NULL AND boolean_value IS NULL
      AND date_value IS NULL AND single_select_value IS NOT NULL AND multi_select_value IS NULL)
    OR (value_type = 'multi_select' AND text_value IS NULL AND number_value IS NULL AND boolean_value IS NULL
      AND date_value IS NULL AND single_select_value IS NULL AND multi_select_value IS NOT NULL)
  )
);

CREATE INDEX ix_accessory_product_attributes_product ON accessory_product_attributes(product_id);

CREATE TABLE accessory_stock_movements (
  id TEXT PRIMARY KEY,
  product_id TEXT NOT NULL,
  location_id TEXT NOT NULL,
  movement_type TEXT NOT NULL CHECK (movement_type IN (
    'purchase', 'adjustment', 'transfer_in', 'transfer_out', 'individualization', 'installation', 'removal'
  )),
  quantity INTEGER NOT NULL CHECK (quantity <> 0),
  source_type TEXT NOT NULL DEFAULT '',
  source_id TEXT NOT NULL DEFAULT '',
  actor TEXT NOT NULL DEFAULT '',
  note TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  FOREIGN KEY (product_id) REFERENCES accessory_products(id) ON DELETE RESTRICT,
  FOREIGN KEY (location_id) REFERENCES storage_locations(id) ON DELETE RESTRICT
);

CREATE INDEX ix_accessory_stock_movements_product ON accessory_stock_movements(product_id, created_at);
CREATE INDEX ix_accessory_stock_movements_location ON accessory_stock_movements(location_id, created_at);
CREATE INDEX ix_accessory_stock_movements_source ON accessory_stock_movements(source_type, source_id);

CREATE TABLE accessory_purchases (
  id TEXT PRIMARY KEY,
  product_id TEXT NOT NULL,
  destination_location_id TEXT,
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  purchased_at TEXT NOT NULL,
  supplier TEXT NOT NULL DEFAULT '',
  order_number TEXT NOT NULL DEFAULT '',
  unit_price TEXT NOT NULL DEFAULT '',
  note TEXT NOT NULL DEFAULT '',
  created_by TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (product_id) REFERENCES accessory_products(id) ON DELETE RESTRICT,
  FOREIGN KEY (destination_location_id) REFERENCES storage_locations(id) ON DELETE RESTRICT
);

CREATE INDEX ix_accessory_purchases_product ON accessory_purchases(product_id, purchased_at);
CREATE INDEX ix_accessory_purchases_destination ON accessory_purchases(destination_location_id);

CREATE TABLE accessory_documents (
  id TEXT PRIMARY KEY,
  product_id TEXT NOT NULL,
  file_blob_id TEXT NOT NULL,
  category TEXT NOT NULL DEFAULT 'other',
  name TEXT NOT NULL DEFAULT '',
  is_primary INTEGER NOT NULL DEFAULT 0 CHECK (is_primary IN (0, 1)),
  created_at TEXT NOT NULL,
  FOREIGN KEY (product_id) REFERENCES accessory_products(id) ON DELETE CASCADE,
  FOREIGN KEY (file_blob_id) REFERENCES file_blobs(id) ON DELETE RESTRICT
);

CREATE INDEX ix_accessory_documents_product ON accessory_documents(product_id);
CREATE INDEX ix_accessory_documents_blob ON accessory_documents(file_blob_id);

ALTER TABLE accessory_assets ADD COLUMN purchase_id TEXT REFERENCES accessory_purchases(id) ON DELETE SET NULL;
CREATE INDEX ix_accessory_assets_purchase ON accessory_assets(purchase_id);

ALTER TABLE accessory_reservations ADD COLUMN placement TEXT NOT NULL DEFAULT '';
ALTER TABLE accessory_reservations ADD COLUMN digital_address TEXT NOT NULL DEFAULT '';
ALTER TABLE accessory_reservations ADD COLUMN decoder_output TEXT NOT NULL DEFAULT '';
ALTER TABLE accessory_reservations ADD COLUMN connection TEXT NOT NULL DEFAULT '';
ALTER TABLE accessory_reservations ADD COLUMN wiring_notes TEXT NOT NULL DEFAULT '';

ALTER TABLE accessory_installations ADD COLUMN placement TEXT NOT NULL DEFAULT '';
ALTER TABLE accessory_installations ADD COLUMN digital_address TEXT NOT NULL DEFAULT '';
ALTER TABLE accessory_installations ADD COLUMN decoder_output TEXT NOT NULL DEFAULT '';
ALTER TABLE accessory_installations ADD COLUMN connection TEXT NOT NULL DEFAULT '';
ALTER TABLE accessory_installations ADD COLUMN wiring_notes TEXT NOT NULL DEFAULT '';

INSERT OR IGNORE INTO master_data_entries(
  id, type, key, label, active, sort_order, source_url, metadata_json, created_at, updated_at
) VALUES
  ('article-type-track', 'article_type', 'track', 'Track', 1, 10, '', '{}', datetime('now'), datetime('now')),
  ('article-type-signal', 'article_type', 'signal', 'Signal', 1, 20, '', '{}', datetime('now'), datetime('now')),
  ('article-type-decoder', 'article_type', 'decoder', 'Decoder', 1, 30, '', '{}', datetime('now'), datetime('now')),
  ('article-type-electrical-control', 'article_type', 'electrical_control', 'Electrical control', 1, 40, '', '{}', datetime('now'), datetime('now')),
  ('article-type-building-equipment', 'article_type', 'building_equipment', 'Building equipment', 1, 50, '', '{}', datetime('now'), datetime('now')),
  ('article-type-landscape-consumable', 'article_type', 'landscape_consumable', 'Landscape consumable', 1, 60, '', '{}', datetime('now'), datetime('now')),
  ('article-type-lighting', 'article_type', 'lighting', 'Lighting', 1, 70, '', '{}', datetime('now'), datetime('now')),
  ('article-type-other', 'article_type', 'other', 'Other', 1, 80, '', '{}', datetime('now'), datetime('now'));

INSERT OR IGNORE INTO master_data_entries(
  id, type, key, label, active, sort_order, source_url, metadata_json, created_at, updated_at
) VALUES
  ('article-subtype-track-straight', 'article_subtype', 'track:straight', 'Straight', 1, 10, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-track-curve', 'article_subtype', 'track:curve', 'Curve', 1, 20, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-track-flex', 'article_subtype', 'track:flex', 'Flex', 1, 30, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-track-turnout', 'article_subtype', 'track:turnout', 'Turnout', 1, 40, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-track-crossing', 'article_subtype', 'track:crossing', 'Crossing', 1, 50, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-track-double-slip', 'article_subtype', 'track:double_slip', 'Double slip', 1, 60, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-track-transition', 'article_subtype', 'track:transition', 'Transition', 1, 70, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-track-buffer-stop', 'article_subtype', 'track:buffer_stop', 'Buffer stop', 1, 80, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-signal-light', 'article_subtype', 'signal:light', 'Light', 1, 10, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-signal-semaphore', 'article_subtype', 'signal:semaphore', 'Semaphore', 1, 20, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-signal-main', 'article_subtype', 'signal:main', 'Main', 1, 30, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-signal-distant', 'article_subtype', 'signal:distant', 'Distant', 1, 40, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-signal-block', 'article_subtype', 'signal:block', 'Block', 1, 50, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-signal-entry', 'article_subtype', 'signal:entry', 'Entry', 1, 60, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-signal-exit', 'article_subtype', 'signal:exit', 'Exit', 1, 70, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-signal-shunting', 'article_subtype', 'signal:shunting', 'Shunting', 1, 80, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-decoder-locomotive', 'article_subtype', 'decoder:locomotive', 'Locomotive', 1, 10, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-decoder-function', 'article_subtype', 'decoder:function', 'Function', 1, 20, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-decoder-accessory', 'article_subtype', 'decoder:accessory', 'Accessory', 1, 30, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-decoder-switching', 'article_subtype', 'decoder:switching', 'Switching', 1, 40, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-decoder-servo', 'article_subtype', 'decoder:servo', 'Servo', 1, 50, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-decoder-feedback', 'article_subtype', 'decoder:feedback', 'Feedback', 1, 60, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-electrical-control-turnout-drive', 'article_subtype', 'electrical_control:turnout_drive', 'Turnout drive', 1, 10, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-electrical-control-feedback', 'article_subtype', 'electrical_control:feedback', 'Feedback', 1, 20, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-electrical-control-booster', 'article_subtype', 'electrical_control:booster', 'Booster', 1, 30, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-electrical-control-power-supply', 'article_subtype', 'electrical_control:power_supply', 'Power supply', 1, 40, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-electrical-control-sensor', 'article_subtype', 'electrical_control:sensor', 'Sensor', 1, 50, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-electrical-control-relay', 'article_subtype', 'electrical_control:relay', 'Relay', 1, 60, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-electrical-control-distribution', 'article_subtype', 'electrical_control:distribution', 'Distribution', 1, 70, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-electrical-control-control-element', 'article_subtype', 'electrical_control:control_element', 'Control element', 1, 80, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-building-equipment-building', 'article_subtype', 'building_equipment:building', 'Building', 1, 10, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-building-equipment-platform', 'article_subtype', 'building_equipment:platform', 'Platform', 1, 20, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-building-equipment-bridge', 'article_subtype', 'building_equipment:bridge', 'Bridge', 1, 30, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-building-equipment-tunnel-portal', 'article_subtype', 'building_equipment:tunnel_portal', 'Tunnel portal', 1, 40, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-building-equipment-road-vehicle', 'article_subtype', 'building_equipment:road_vehicle', 'Road vehicle', 1, 50, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-building-equipment-figure', 'article_subtype', 'building_equipment:figure', 'Figure', 1, 60, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-building-equipment-street-equipment', 'article_subtype', 'building_equipment:street_equipment', 'Street equipment', 1, 70, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-building-equipment-interior-equipment', 'article_subtype', 'building_equipment:interior_equipment', 'Interior equipment', 1, 80, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-landscape-consumable-grass', 'article_subtype', 'landscape_consumable:grass', 'Grass', 1, 10, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-landscape-consumable-scatter', 'article_subtype', 'landscape_consumable:scatter', 'Scatter', 1, 20, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-landscape-consumable-tree', 'article_subtype', 'landscape_consumable:tree', 'Tree', 1, 30, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-landscape-consumable-water', 'article_subtype', 'landscape_consumable:water', 'Water', 1, 40, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-landscape-consumable-paint', 'article_subtype', 'landscape_consumable:paint', 'Paint', 1, 50, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-landscape-consumable-adhesive', 'article_subtype', 'landscape_consumable:adhesive', 'Adhesive', 1, 60, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-landscape-consumable-ballast', 'article_subtype', 'landscape_consumable:ballast', 'Ballast', 1, 70, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-landscape-consumable-wire', 'article_subtype', 'landscape_consumable:wire', 'Wire', 1, 80, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-landscape-consumable-cable', 'article_subtype', 'landscape_consumable:cable', 'Cable', 1, 90, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-landscape-consumable-fastener', 'article_subtype', 'landscape_consumable:fastener', 'Fastener', 1, 100, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-lighting-lamp', 'article_subtype', 'lighting:lamp', 'Lamp', 1, 10, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-lighting-led', 'article_subtype', 'lighting:led', 'LED', 1, 20, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-lighting-light-strip', 'article_subtype', 'lighting:light_strip', 'Light strip', 1, 30, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-lighting-building-lighting', 'article_subtype', 'lighting:building_lighting', 'Building lighting', 1, 40, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-lighting-effect-lighting', 'article_subtype', 'lighting:effect_lighting', 'Effect lighting', 1, 50, '', '{}', datetime('now'), datetime('now')),
  ('article-subtype-other-other', 'article_subtype', 'other:other', 'Other', 1, 10, '', '{}', datetime('now'), datetime('now'));

INSERT OR IGNORE INTO master_data_entries(
  id, type, key, label, active, sort_order, source_url, metadata_json, created_at, updated_at
) VALUES
  ('stock-unit-piece', 'stock_unit', 'piece', 'Piece', 1, 10, '', '{}', datetime('now'), datetime('now')),
  ('stock-unit-pack', 'stock_unit', 'pack', 'Pack', 1, 20, '', '{}', datetime('now'), datetime('now')),
  ('stock-unit-meter', 'stock_unit', 'meter', 'Meter', 1, 30, '', '{}', datetime('now'), datetime('now')),
  ('stock-unit-gram', 'stock_unit', 'gram', 'Gram', 1, 40, '', '{}', datetime('now'), datetime('now')),
  ('stock-unit-milliliter', 'stock_unit', 'milliliter', 'Milliliter', 1, 50, '', '{}', datetime('now'), datetime('now')),
  ('controlled-field-kind-text', 'controlled_field_kind', 'text', 'Text', 1, 10, '', '{}', datetime('now'), datetime('now')),
  ('controlled-field-kind-number', 'controlled_field_kind', 'number', 'Number', 1, 20, '', '{}', datetime('now'), datetime('now')),
  ('controlled-field-kind-boolean', 'controlled_field_kind', 'boolean', 'Boolean', 1, 30, '', '{}', datetime('now'), datetime('now')),
  ('controlled-field-kind-date', 'controlled_field_kind', 'date', 'Date', 1, 40, '', '{}', datetime('now'), datetime('now')),
  ('controlled-field-kind-single-select', 'controlled_field_kind', 'single_select', 'Single select', 1, 50, '', '{}', datetime('now'), datetime('now')),
  ('controlled-field-kind-multi-select', 'controlled_field_kind', 'multi_select', 'Multi select', 1, 60, '', '{}', datetime('now'), datetime('now'));
