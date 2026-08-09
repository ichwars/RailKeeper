CREATE TABLE layout_unit_ports (
  id TEXT PRIMARY KEY,
  layout_unit_id TEXT NOT NULL,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('track', 'power', 'digital', 'feedback', 'accessory', 'other')),
  interface_key TEXT NOT NULL,
  x_mm REAL NOT NULL CHECK (x_mm >= 0),
  y_mm REAL NOT NULL CHECK (y_mm >= 0),
  direction_degrees REAL NOT NULL CHECK (direction_degrees >= 0 AND direction_degrees < 360),
  notes TEXT NOT NULL DEFAULT '',
  version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
  archived INTEGER NOT NULL DEFAULT 0 CHECK (archived IN (0, 1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (layout_unit_id) REFERENCES layout_units(id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX ux_layout_unit_ports_name
  ON layout_unit_ports(layout_unit_id, name COLLATE NOCASE);
CREATE INDEX ix_layout_unit_ports_unit ON layout_unit_ports(layout_unit_id);
CREATE INDEX ix_layout_unit_ports_interface
  ON layout_unit_ports(kind, interface_key COLLATE NOCASE);
