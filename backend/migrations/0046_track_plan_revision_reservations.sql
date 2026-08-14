ALTER TABLE plan_track_objects ADD COLUMN lineage_id TEXT NOT NULL DEFAULT '';

UPDATE plan_track_objects SET lineage_id=id WHERE lineage_id='';

CREATE TRIGGER tr_plan_track_objects_default_lineage
AFTER INSERT ON plan_track_objects
WHEN NEW.lineage_id=''
BEGIN
  UPDATE plan_track_objects SET lineage_id=NEW.id WHERE id=NEW.id;
END;

CREATE UNIQUE INDEX ux_plan_track_objects_revision_lineage
  ON plan_track_objects(revision_id, lineage_id);

CREATE TABLE plan_track_object_reservations (
  reservation_id TEXT PRIMARY KEY,
  track_object_id TEXT NOT NULL,
  active INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (reservation_id) REFERENCES accessory_reservations(id) ON DELETE CASCADE,
  FOREIGN KEY (track_object_id) REFERENCES plan_track_objects(id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX ux_plan_track_object_reservations_active_object
  ON plan_track_object_reservations(track_object_id) WHERE active=1;
CREATE INDEX ix_plan_track_object_reservations_object
  ON plan_track_object_reservations(track_object_id, active);
