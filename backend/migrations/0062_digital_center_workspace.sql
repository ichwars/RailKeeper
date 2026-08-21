CREATE TABLE digital_center_read_sessions (
  id TEXT PRIMARY KEY,
  provider TEXT NOT NULL,
  state TEXT NOT NULL,
  host TEXT NOT NULL DEFAULT '',
  port INTEGER NOT NULL DEFAULT 0,
  capability_json TEXT NOT NULL DEFAULT '{}',
  read_started_at TEXT,
  read_completed_at TEXT,
  created_by_user_id TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE digital_center_work_items (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL REFERENCES digital_center_read_sessions(id) ON DELETE CASCADE,
  center_object_id TEXT NOT NULL,
  vehicle_id TEXT,
  name TEXT NOT NULL DEFAULT '',
  decoder_address INTEGER NOT NULL DEFAULT 0,
  protocol TEXT NOT NULL DEFAULT '',
  compare_status TEXT NOT NULL,
  station_status TEXT NOT NULL DEFAULT '',
  center_json TEXT NOT NULL DEFAULT '{}',
  railkeeper_json TEXT NOT NULL DEFAULT '{}',
  proposed_json TEXT NOT NULL DEFAULT '{}',
  conflict_json TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(session_id, center_object_id)
);

CREATE TABLE digital_center_session_messages (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL REFERENCES digital_center_read_sessions(id) ON DELETE CASCADE,
  severity TEXT NOT NULL CHECK(severity IN ('info', 'warning', 'error')),
  code TEXT NOT NULL,
  message TEXT NOT NULL,
  next_action TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

CREATE TABLE digital_center_write_grants (
  id TEXT PRIMARY KEY,
  token_hash TEXT NOT NULL UNIQUE,
  session_id TEXT NOT NULL REFERENCES digital_center_read_sessions(id) ON DELETE CASCADE,
  work_item_id TEXT NOT NULL REFERENCES digital_center_work_items(id) ON DELETE CASCADE,
  preview_hash TEXT NOT NULL,
  actor_user_id TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  consumed_at TEXT,
  created_at TEXT NOT NULL
);

CREATE INDEX idx_digital_center_sessions_created_at ON digital_center_read_sessions(created_at DESC);
CREATE INDEX idx_digital_center_work_items_session ON digital_center_work_items(session_id);
CREATE INDEX idx_digital_center_messages_session ON digital_center_session_messages(session_id, created_at DESC);
CREATE INDEX idx_digital_center_grants_expiry ON digital_center_write_grants(expires_at);
