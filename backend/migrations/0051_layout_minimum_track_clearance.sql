ALTER TABLE layouts ADD COLUMN minimum_track_clearance_mm REAL
    CHECK(minimum_track_clearance_mm IS NULL OR minimum_track_clearance_mm > 0);
