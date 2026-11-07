CREATE TABLE vehicle_valuations (
  id TEXT PRIMARY KEY,
  vehicle_id TEXT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
  amount TEXT NOT NULL CHECK (length(amount) BETWEEN 4 AND 19),
  currency TEXT NOT NULL CHECK (length(currency)=3 AND currency NOT GLOB '*[^A-Z]*'),
  valued_on TEXT NOT NULL CHECK (length(valued_on)=10),
  source TEXT NOT NULL CHECK (length(trim(source)) BETWEEN 1 AND 1000),
  notes TEXT NOT NULL DEFAULT '',
  recorded_at TEXT NOT NULL CHECK (length(recorded_at)=30),
  replaces_id TEXT UNIQUE,
  withdrawn INTEGER NOT NULL DEFAULT 0 CHECK (withdrawn IN (0,1)),
  UNIQUE (vehicle_id, id),
  FOREIGN KEY (vehicle_id,replaces_id) REFERENCES vehicle_valuations(vehicle_id,id)
    DEFERRABLE INITIALLY DEFERRED,
  CHECK (replaces_id IS NULL OR replaces_id<>id),
  CHECK (withdrawn=0 OR replaces_id IS NOT NULL)
);
CREATE INDEX idx_vehicle_valuation_history
  ON vehicle_valuations(vehicle_id,valued_on DESC,recorded_at DESC,id DESC);
