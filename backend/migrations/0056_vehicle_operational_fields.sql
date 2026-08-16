ALTER TABLE vehicles ADD COLUMN maximum_speed_kmh INTEGER
  CHECK (maximum_speed_kmh IS NULL OR maximum_speed_kmh BETWEEN 1 AND 1000);
ALTER TABLE vehicles ADD COLUMN home_base TEXT NOT NULL DEFAULT '';
