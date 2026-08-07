INSERT INTO roles(id, name)
SELECT lower(hex(randomblob(16))), 'Planner'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'Planner');

CREATE TABLE storage_locations (
  id TEXT PRIMARY KEY,
  parent_id TEXT,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  archived INTEGER NOT NULL DEFAULT 0 CHECK (archived IN (0, 1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (parent_id) REFERENCES storage_locations(id) ON DELETE RESTRICT,
  CHECK (parent_id IS NULL OR parent_id <> id)
);

CREATE UNIQUE INDEX ux_storage_locations_root_name
  ON storage_locations(name COLLATE NOCASE) WHERE parent_id IS NULL;
CREATE UNIQUE INDEX ux_storage_locations_parent_name
  ON storage_locations(parent_id, name COLLATE NOCASE) WHERE parent_id IS NOT NULL;
CREATE INDEX ix_storage_locations_parent ON storage_locations(parent_id);

CREATE TABLE accessory_products (
  id TEXT PRIMARY KEY,
  manufacturer TEXT NOT NULL,
  article_number TEXT NOT NULL DEFAULT '',
  name TEXT NOT NULL,
  category TEXT NOT NULL,
  tracking_mode TEXT NOT NULL CHECK (tracking_mode IN ('quantity', 'individual')),
  description TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE UNIQUE INDEX ux_accessory_products_article
  ON accessory_products(manufacturer COLLATE NOCASE, article_number COLLATE NOCASE)
  WHERE article_number <> '';
CREATE INDEX ix_accessory_products_name ON accessory_products(name COLLATE NOCASE);

CREATE TABLE accessory_stock (
  product_id TEXT NOT NULL,
  location_id TEXT NOT NULL,
  quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
  updated_at TEXT NOT NULL,
  PRIMARY KEY (product_id, location_id),
  FOREIGN KEY (product_id) REFERENCES accessory_products(id) ON DELETE RESTRICT,
  FOREIGN KEY (location_id) REFERENCES storage_locations(id) ON DELETE RESTRICT
);

CREATE INDEX ix_accessory_stock_location ON accessory_stock(location_id);

CREATE TABLE accessory_assets (
  id TEXT PRIMARY KEY,
  product_id TEXT NOT NULL,
  inventory_number TEXT,
  serial_number TEXT NOT NULL DEFAULT '',
  condition_state TEXT NOT NULL DEFAULT 'unknown'
    CHECK (condition_state IN ('ready', 'maintenance_due', 'defective', 'unknown')),
  lifecycle_state TEXT NOT NULL DEFAULT 'stored'
    CHECK (lifecycle_state IN ('stored', 'reserved', 'installed', 'maintenance', 'retired')),
  storage_location_id TEXT,
  purchase_date TEXT NOT NULL DEFAULT '',
  purchase_price TEXT NOT NULL DEFAULT '',
  warranty_until TEXT NOT NULL DEFAULT '',
  notes TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (product_id) REFERENCES accessory_products(id) ON DELETE RESTRICT,
  FOREIGN KEY (storage_location_id) REFERENCES storage_locations(id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX ux_accessory_assets_inventory_number
  ON accessory_assets(inventory_number COLLATE NOCASE)
  WHERE inventory_number IS NOT NULL AND inventory_number <> '';
CREATE INDEX ix_accessory_assets_product ON accessory_assets(product_id);
CREATE INDEX ix_accessory_assets_location ON accessory_assets(storage_location_id);

CREATE TABLE layouts (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('private', 'club')),
  gauge TEXT NOT NULL,
  scale TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
  archived INTEGER NOT NULL DEFAULT 0 CHECK (archived IN (0, 1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX ix_layouts_name ON layouts(name COLLATE NOCASE);

CREATE TABLE layout_units (
  id TEXT PRIMARY KEY,
  layout_id TEXT NOT NULL,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('baseboard', 'module', 'segment', 'area')),
  owner_label TEXT NOT NULL DEFAULT '',
  width_mm REAL CHECK (width_mm IS NULL OR width_mm >= 0),
  height_mm REAL CHECK (height_mm IS NULL OR height_mm >= 0),
  version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
  archived INTEGER NOT NULL DEFAULT 0 CHECK (archived IN (0, 1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (layout_id) REFERENCES layouts(id) ON DELETE RESTRICT,
  UNIQUE (layout_id, name)
);

CREATE INDEX ix_layout_units_layout ON layout_units(layout_id);

CREATE TABLE plan_variants (
  id TEXT PRIMARY KEY,
  layout_unit_id TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  archived INTEGER NOT NULL DEFAULT 0 CHECK (archived IN (0, 1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (layout_unit_id) REFERENCES layout_units(id) ON DELETE RESTRICT,
  UNIQUE (layout_unit_id, name)
);

CREATE INDEX ix_plan_variants_unit ON plan_variants(layout_unit_id);

CREATE TABLE plan_revisions (
  id TEXT PRIMARY KEY,
  variant_id TEXT NOT NULL,
  revision_number INTEGER NOT NULL CHECK (revision_number > 0),
  status TEXT NOT NULL CHECK (status IN ('draft', 'review', 'published', 'archived')),
  base_revision_id TEXT,
  version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
  created_by TEXT NOT NULL,
  published_by TEXT,
  published_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (variant_id) REFERENCES plan_variants(id) ON DELETE RESTRICT,
  FOREIGN KEY (base_revision_id) REFERENCES plan_revisions(id) ON DELETE RESTRICT,
  UNIQUE (variant_id, revision_number)
);

CREATE UNIQUE INDEX ux_plan_revisions_published
  ON plan_revisions(variant_id) WHERE status = 'published';
CREATE INDEX ix_plan_revisions_variant ON plan_revisions(variant_id);
CREATE INDEX ix_plan_revisions_base ON plan_revisions(base_revision_id);

CREATE TABLE layout_configurations (
  id TEXT PRIMARY KEY,
  layout_id TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
  archived INTEGER NOT NULL DEFAULT 0 CHECK (archived IN (0, 1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (layout_id) REFERENCES layouts(id) ON DELETE RESTRICT,
  UNIQUE (layout_id, name)
);

CREATE INDEX ix_layout_configurations_layout ON layout_configurations(layout_id);

CREATE TABLE layout_configuration_units (
  configuration_id TEXT NOT NULL,
  unit_id TEXT NOT NULL,
  plan_revision_id TEXT,
  position_x_mm REAL NOT NULL DEFAULT 0,
  position_y_mm REAL NOT NULL DEFAULT 0,
  rotation_degrees REAL NOT NULL DEFAULT 0 CHECK (rotation_degrees >= 0 AND rotation_degrees < 360),
  sort_order INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (configuration_id, unit_id),
  FOREIGN KEY (configuration_id) REFERENCES layout_configurations(id) ON DELETE CASCADE,
  FOREIGN KEY (unit_id) REFERENCES layout_units(id) ON DELETE RESTRICT,
  FOREIGN KEY (plan_revision_id) REFERENCES plan_revisions(id) ON DELETE RESTRICT
);

CREATE INDEX ix_layout_configuration_units_unit ON layout_configuration_units(unit_id);
CREATE INDEX ix_layout_configuration_units_revision ON layout_configuration_units(plan_revision_id);

CREATE TABLE accessory_reservations (
  id TEXT PRIMARY KEY,
  product_id TEXT NOT NULL,
  asset_id TEXT,
  location_id TEXT NOT NULL,
  quantity INTEGER NOT NULL CHECK (quantity > 0),
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
  FOREIGN KEY (layout_id) REFERENCES layouts(id) ON DELETE RESTRICT,
  FOREIGN KEY (layout_unit_id) REFERENCES layout_units(id) ON DELETE RESTRICT,
  CHECK ((layout_id IS NOT NULL) <> (layout_unit_id IS NOT NULL)),
  CHECK (asset_id IS NULL OR quantity = 1)
);

CREATE INDEX ix_accessory_reservations_product ON accessory_reservations(product_id);
CREATE INDEX ix_accessory_reservations_asset ON accessory_reservations(asset_id);
CREATE INDEX ix_accessory_reservations_location ON accessory_reservations(location_id);
CREATE INDEX ix_accessory_reservations_layout ON accessory_reservations(layout_id);
CREATE INDEX ix_accessory_reservations_unit ON accessory_reservations(layout_unit_id);

CREATE TABLE accessory_installations (
  id TEXT PRIMARY KEY,
  product_id TEXT NOT NULL,
  asset_id TEXT,
  source_location_id TEXT NOT NULL,
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  vehicle_id TEXT,
  layout_id TEXT,
  layout_unit_id TEXT,
  condition_state TEXT NOT NULL DEFAULT 'unknown'
    CHECK (condition_state IN ('ready', 'maintenance_due', 'defective', 'unknown')),
  installed_by TEXT NOT NULL,
  installed_at TEXT NOT NULL,
  removed_by TEXT,
  removed_at TEXT,
  removal_disposition TEXT
    CHECK (removal_disposition IS NULL OR removal_disposition IN ('stored', 'maintenance', 'defective', 'retired')),
  notes TEXT NOT NULL DEFAULT '',
  FOREIGN KEY (product_id) REFERENCES accessory_products(id) ON DELETE RESTRICT,
  FOREIGN KEY (asset_id) REFERENCES accessory_assets(id) ON DELETE RESTRICT,
  FOREIGN KEY (source_location_id) REFERENCES storage_locations(id) ON DELETE RESTRICT,
  FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE RESTRICT,
  FOREIGN KEY (layout_id) REFERENCES layouts(id) ON DELETE RESTRICT,
  FOREIGN KEY (layout_unit_id) REFERENCES layout_units(id) ON DELETE RESTRICT,
  CHECK ((vehicle_id IS NOT NULL) + (layout_id IS NOT NULL) + (layout_unit_id IS NOT NULL) = 1),
  CHECK (asset_id IS NULL OR quantity = 1),
  CHECK ((removed_at IS NULL AND removed_by IS NULL AND removal_disposition IS NULL) OR removed_at IS NOT NULL)
);

CREATE UNIQUE INDEX ux_accessory_installations_active_asset
  ON accessory_installations(asset_id) WHERE asset_id IS NOT NULL AND removed_at IS NULL;
CREATE INDEX ix_accessory_installations_product ON accessory_installations(product_id);
CREATE INDEX ix_accessory_installations_vehicle ON accessory_installations(vehicle_id);
CREATE INDEX ix_accessory_installations_layout ON accessory_installations(layout_id);
CREATE INDEX ix_accessory_installations_unit ON accessory_installations(layout_unit_id);
