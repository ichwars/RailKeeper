CREATE TABLE layout_unit_outline_points (
  layout_unit_id TEXT NOT NULL,
  point_index INTEGER NOT NULL CHECK (point_index >= 0),
  position_x_mm REAL NOT NULL,
  position_y_mm REAL NOT NULL,
  PRIMARY KEY (layout_unit_id, point_index),
  FOREIGN KEY (layout_unit_id) REFERENCES layout_units(id) ON DELETE CASCADE
);

CREATE TABLE layout_technical_positions (
  id TEXT PRIMARY KEY,
  layout_unit_id TEXT NOT NULL,
  label TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN (
    'turnout', 'signal', 'feedback', 'decoder', 'lighting', 'power', 'sensor', 'other'
  )),
  position_x_mm REAL NOT NULL,
  position_y_mm REAL NOT NULL,
  rotation_degrees REAL NOT NULL DEFAULT 0
    CHECK (rotation_degrees >= 0 AND rotation_degrees < 360),
  product_id TEXT,
  description TEXT NOT NULL DEFAULT '',
  version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
  archived INTEGER NOT NULL DEFAULT 0 CHECK (archived IN (0, 1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (layout_unit_id) REFERENCES layout_units(id) ON DELETE RESTRICT,
  FOREIGN KEY (product_id) REFERENCES accessory_products(id) ON DELETE RESTRICT
);

CREATE INDEX ix_layout_technical_positions_unit
  ON layout_technical_positions(layout_unit_id, archived, label COLLATE NOCASE);
CREATE INDEX ix_layout_technical_positions_product
  ON layout_technical_positions(product_id);

CREATE TABLE accessory_reservation_positions (
  reservation_id TEXT PRIMARY KEY,
  position_id TEXT NOT NULL,
  FOREIGN KEY (reservation_id) REFERENCES accessory_reservations(id) ON DELETE CASCADE,
  FOREIGN KEY (position_id) REFERENCES layout_technical_positions(id) ON DELETE RESTRICT
);

CREATE INDEX ix_accessory_reservation_positions_position
  ON accessory_reservation_positions(position_id);

CREATE TABLE accessory_installation_positions (
  installation_id TEXT PRIMARY KEY,
  position_id TEXT NOT NULL,
  FOREIGN KEY (installation_id) REFERENCES accessory_installations(id) ON DELETE CASCADE,
  FOREIGN KEY (position_id) REFERENCES layout_technical_positions(id) ON DELETE RESTRICT
);

CREATE INDEX ix_accessory_installation_positions_position
  ON accessory_installation_positions(position_id);
