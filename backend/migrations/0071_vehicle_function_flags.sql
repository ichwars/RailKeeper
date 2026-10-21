-- The old checkbox was labelled "Inverted" but stored direction dependency.
-- Preserve the ambiguous value without assigning either meaning to it.
ALTER TABLE vehicle_functions RENAME TO vehicle_functions_legacy_flags;

CREATE TABLE vehicle_functions (
  id TEXT PRIMARY KEY,
  vehicle_id TEXT NOT NULL,
  function_key TEXT NOT NULL,
  name TEXT NOT NULL DEFAULT '',
  symbol_key TEXT NOT NULL DEFAULT '',
  function_type TEXT NOT NULL DEFAULT 'standard',
  mode TEXT NOT NULL DEFAULT 'dauer',
  direction_dependent INTEGER CHECK (direction_dependent IN (0, 1)),
  inverted INTEGER CHECK (inverted IN (0, 1)),
  legacy_direction_dependent INTEGER CHECK (legacy_direction_dependent IN (0, 1)),
  notes TEXT NOT NULL DEFAULT '',
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE,
  UNIQUE (vehicle_id, function_key)
);

INSERT INTO vehicle_functions
  (id, vehicle_id, function_key, name, symbol_key, function_type, mode,
   direction_dependent, inverted, legacy_direction_dependent, notes, sort_order, created_at, updated_at)
SELECT id, vehicle_id, function_key, name, symbol_key, function_type, mode,
       NULL, NULL, direction_dependent, notes, sort_order, created_at, updated_at
FROM vehicle_functions_legacy_flags;

DROP TABLE vehicle_functions_legacy_flags;
CREATE INDEX idx_vehicle_functions_vehicle_id ON vehicle_functions(vehicle_id, sort_order, function_key);
