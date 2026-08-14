ALTER TABLE plan_track_objects
  ADD COLUMN elevation_start_mm REAL NOT NULL DEFAULT 0;

ALTER TABLE plan_track_objects
  ADD COLUMN elevation_end_mm REAL NOT NULL DEFAULT 0;

