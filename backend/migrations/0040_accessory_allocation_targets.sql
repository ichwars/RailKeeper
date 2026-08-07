DROP INDEX IF EXISTS ix_accessory_reservations_product;
DROP INDEX IF EXISTS ix_accessory_reservations_asset;
DROP INDEX IF EXISTS ix_accessory_reservations_location;
DROP INDEX IF EXISTS ix_accessory_reservations_layout;
DROP INDEX IF EXISTS ix_accessory_reservations_unit;

ALTER TABLE accessory_reservations RENAME TO accessory_reservations_legacy;

CREATE TABLE accessory_reservations (
  id TEXT PRIMARY KEY,
  product_id TEXT NOT NULL,
  asset_id TEXT,
  location_id TEXT NOT NULL,
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  vehicle_id TEXT,
  layout_id TEXT,
  layout_unit_id TEXT,
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'fulfilled', 'cancelled')),
  note TEXT NOT NULL DEFAULT '',
  created_by TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (product_id) REFERENCES accessory_products(id) ON DELETE RESTRICT,
  FOREIGN KEY (asset_id) REFERENCES accessory_assets(id) ON DELETE RESTRICT,
  FOREIGN KEY (location_id) REFERENCES storage_locations(id) ON DELETE RESTRICT,
  FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE RESTRICT,
  FOREIGN KEY (layout_id) REFERENCES layouts(id) ON DELETE RESTRICT,
  FOREIGN KEY (layout_unit_id) REFERENCES layout_units(id) ON DELETE RESTRICT,
  CHECK ((vehicle_id IS NOT NULL) + (layout_id IS NOT NULL) + (layout_unit_id IS NOT NULL) = 1),
  CHECK (asset_id IS NULL OR quantity = 1)
);

INSERT INTO accessory_reservations(
  id, product_id, asset_id, location_id, quantity, layout_id, layout_unit_id,
  status, note, created_by, created_at, updated_at
)
SELECT
  id, product_id, asset_id, location_id, quantity, layout_id, layout_unit_id,
  status, note, created_by, created_at, updated_at
FROM accessory_reservations_legacy;

DROP TABLE accessory_reservations_legacy;

CREATE INDEX ix_accessory_reservations_product ON accessory_reservations(product_id);
CREATE INDEX ix_accessory_reservations_asset ON accessory_reservations(asset_id);
CREATE UNIQUE INDEX ux_accessory_reservations_active_asset
  ON accessory_reservations(asset_id) WHERE asset_id IS NOT NULL AND status = 'active';
CREATE INDEX ix_accessory_reservations_location ON accessory_reservations(location_id);
CREATE INDEX ix_accessory_reservations_vehicle ON accessory_reservations(vehicle_id);
CREATE INDEX ix_accessory_reservations_layout ON accessory_reservations(layout_id);
CREATE INDEX ix_accessory_reservations_unit ON accessory_reservations(layout_unit_id);

ALTER TABLE accessory_installations ADD COLUMN removal_notes TEXT NOT NULL DEFAULT '';
