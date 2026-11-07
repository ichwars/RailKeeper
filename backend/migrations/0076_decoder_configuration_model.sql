CREATE TABLE decoder_instances (
  id TEXT PRIMARY KEY,
  label TEXT NOT NULL,
  manufacturer TEXT NOT NULL DEFAULT '',
  model TEXT NOT NULL DEFAULT '',
  serial_number TEXT NOT NULL DEFAULT '',
  notes TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE decoder_installations (
  id TEXT PRIMARY KEY,
  decoder_id TEXT NOT NULL REFERENCES decoder_instances(id) ON DELETE RESTRICT,
  vehicle_id TEXT REFERENCES vehicles(id) ON DELETE SET NULL,
  vehicle_reference_id TEXT NOT NULL,
  vehicle_name TEXT NOT NULL,
  vehicle_inventory_number TEXT NOT NULL,
  state TEXT NOT NULL CHECK (state IN ('installed', 'removed', 'unresolved')),
  installed_at TEXT CHECK (installed_at IS NULL OR length(installed_at) = 10),
  removed_at TEXT CHECK (removed_at IS NULL OR length(removed_at) = 10),
  notes TEXT NOT NULL DEFAULT '',
  recorded_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE UNIQUE INDEX idx_decoder_current_installation ON decoder_installations(decoder_id) WHERE state='installed';
CREATE INDEX idx_decoder_installation_vehicle ON decoder_installations(vehicle_id);

-- Deleting inventory cannot establish when a physical decoder was removed.
CREATE TRIGGER decoder_installation_vehicle_removed
AFTER UPDATE OF vehicle_id ON decoder_installations
WHEN NEW.vehicle_id IS NULL AND OLD.vehicle_id IS NOT NULL AND NEW.state='installed'
BEGIN
  UPDATE decoder_installations SET state='unresolved', updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id=NEW.id;
END;

CREATE TABLE decoder_cv_workspaces (
  id TEXT PRIMARY KEY,
  vehicle_id TEXT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
  profile_text TEXT NOT NULL,
  origin TEXT NOT NULL CHECK (origin IN ('legacy', 'manual')),
  installation_id TEXT REFERENCES decoder_installations(id) ON DELETE RESTRICT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (vehicle_id, profile_text)
);

CREATE TABLE decoder_cv_snapshots (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL,
  vehicle_reference_id TEXT NOT NULL,
  vehicle_name TEXT NOT NULL,
  vehicle_inventory_number TEXT NOT NULL,
  profile_text TEXT NOT NULL,
  installation_id TEXT REFERENCES decoder_installations(id) ON DELETE RESTRICT,
  kind TEXT NOT NULL CHECK (kind IN ('legacy', 'manual', 'file_import', 'station_import', 'checkpoint')),
  scope TEXT NOT NULL CHECK (scope='partial_working_state'),
  payload_json TEXT NOT NULL CHECK (json_valid(payload_json) AND json_type(payload_json)='object'),
  hardware_json TEXT NOT NULL CHECK (json_valid(hardware_json) AND json_type(hardware_json)='object'),
  note TEXT NOT NULL DEFAULT '',
  recorded_at TEXT NOT NULL
);
CREATE INDEX idx_decoder_snapshot_workspace ON decoder_cv_snapshots(workspace_id, recorded_at);
CREATE INDEX idx_decoder_snapshot_vehicle ON decoder_cv_snapshots(vehicle_reference_id, recorded_at);

CREATE TABLE decoder_cv_snapshot_sources (
  snapshot_id TEXT NOT NULL REFERENCES decoder_cv_snapshots(id) ON DELETE RESTRICT,
  source_file_id TEXT NOT NULL,
  metadata_json TEXT NOT NULL CHECK (json_valid(metadata_json) AND json_type(metadata_json)='object'),
  content_state TEXT NOT NULL CHECK (content_state IN ('stored', 'unavailable')),
  blob_id TEXT REFERENCES file_blobs(id) ON DELETE RESTRICT,
  PRIMARY KEY (snapshot_id, source_file_id),
  CHECK ((content_state='stored' AND blob_id IS NOT NULL) OR (content_state='unavailable' AND blob_id IS NULL))
);
CREATE INDEX idx_decoder_snapshot_source_blob ON decoder_cv_snapshot_sources(blob_id);

CREATE TRIGGER decoder_cv_snapshot_immutable BEFORE UPDATE ON decoder_cv_snapshots
BEGIN SELECT RAISE(ABORT, 'decoder CV snapshots are immutable'); END;
CREATE TRIGGER decoder_cv_snapshot_source_immutable BEFORE UPDATE ON decoder_cv_snapshot_sources
BEGIN SELECT RAISE(ABORT, 'decoder CV snapshot sources are immutable'); END;
