CREATE TABLE IF NOT EXISTS vehicle_functions (
  id TEXT PRIMARY KEY,
  vehicle_id TEXT NOT NULL,
  function_key TEXT NOT NULL,
  name TEXT NOT NULL DEFAULT '',
  symbol_key TEXT NOT NULL DEFAULT '',
  function_type TEXT NOT NULL DEFAULT 'standard',
  mode TEXT NOT NULL DEFAULT 'dauer',
  direction_dependent INTEGER NOT NULL DEFAULT 0,
  notes TEXT NOT NULL DEFAULT '',
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE,
  UNIQUE (vehicle_id, function_key)
);

CREATE INDEX IF NOT EXISTS idx_vehicle_functions_vehicle_id
  ON vehicle_functions(vehicle_id, sort_order, function_key);

INSERT OR IGNORE INTO master_data_entries(id, type, key, label, active, sort_order, source_url, metadata_json, created_at, updated_at)
VALUES
  ('symbols:light', 'symbols', 'light', 'Licht', 1, 10, '', '{"category":"Licht","icon":"lightbulb"}', datetime('now'), datetime('now')),
  ('symbols:sound', 'symbols', 'sound', 'Sound', 1, 20, '', '{"category":"Sound","icon":"volume-2"}', datetime('now'), datetime('now')),
  ('symbols:horn', 'symbols', 'horn', 'Lokpfeife / Horn', 1, 30, '', '{"category":"Sound","icon":"megaphone"}', datetime('now'), datetime('now')),
  ('symbols:coupling', 'symbols', 'coupling', 'Kupplung', 1, 40, '', '{"category":"Kupplung","icon":"link"}', datetime('now'), datetime('now')),
  ('symbols:smoke', 'symbols', 'smoke', 'Rauchgenerator', 1, 50, '', '{"category":"Dampf","icon":"cloud"}', datetime('now'), datetime('now')),
  ('symbols:drive', 'symbols', 'drive', 'Antrieb / Rangiergang', 1, 60, '', '{"category":"Fahren","icon":"gauge"}', datetime('now'), datetime('now')),
  ('symbols:warning', 'symbols', 'warning', 'Warnung', 1, 70, '', '{"category":"Warnung","icon":"triangle-alert"}', datetime('now'), datetime('now')),
  ('symbols:standard', 'symbols', 'standard', 'Standard', 1, 999, '', '{"category":"Standard","icon":"circle"}', datetime('now'), datetime('now'));
