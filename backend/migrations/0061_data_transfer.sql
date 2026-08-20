CREATE TABLE data_transfer_profiles (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  direction TEXT NOT NULL CHECK(direction IN ('import', 'export')),
  format TEXT NOT NULL CHECK(format IN ('csv', 'railkeeper-json')),
  areas_json TEXT NOT NULL,
  options_json TEXT NOT NULL DEFAULT '{}',
  enabled INTEGER NOT NULL DEFAULT 1 CHECK(enabled IN (0, 1)),
  created_by_user_id TEXT,
  last_used_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE data_transfer_jobs (
  id TEXT PRIMARY KEY,
  profile_id TEXT REFERENCES data_transfer_profiles(id) ON DELETE SET NULL,
  profile_name TEXT NOT NULL DEFAULT '',
  direction TEXT NOT NULL CHECK(direction IN ('import', 'export')),
  format TEXT NOT NULL CHECK(format IN ('csv', 'railkeeper-json')),
  areas_json TEXT NOT NULL,
  options_json TEXT NOT NULL DEFAULT '{}',
  state TEXT NOT NULL,
  stage TEXT NOT NULL,
  source_name TEXT NOT NULL DEFAULT '',
  source_sha256 TEXT NOT NULL DEFAULT '',
  package_version INTEGER NOT NULL DEFAULT 0,
  total_records INTEGER NOT NULL DEFAULT 0,
  ready_records INTEGER NOT NULL DEFAULT 0,
  warning_records INTEGER NOT NULL DEFAULT 0,
  error_records INTEGER NOT NULL DEFAULT 0,
  preview_json TEXT NOT NULL DEFAULT '{}',
  created_by_user_id TEXT,
  confirmed_by_user_id TEXT,
  confirmed_at TEXT,
  completed_at TEXT,
  result_message TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE data_transfer_job_issues (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL REFERENCES data_transfer_jobs(id) ON DELETE CASCADE,
  area TEXT NOT NULL,
  record_key TEXT NOT NULL DEFAULT '',
  row_number INTEGER,
  field TEXT NOT NULL DEFAULT '',
  severity TEXT NOT NULL CHECK(severity IN ('warning', 'error')),
  code TEXT NOT NULL,
  message TEXT NOT NULL,
  proposed_resolution TEXT NOT NULL DEFAULT '',
  selected_resolution TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE data_transfer_artifacts (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL REFERENCES data_transfer_jobs(id) ON DELETE CASCADE,
  relative_path TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL,
  mime_type TEXT NOT NULL,
  size_bytes INTEGER NOT NULL,
  sha256 TEXT NOT NULL,
  deleted_at TEXT,
  created_at TEXT NOT NULL
);

CREATE INDEX idx_data_transfer_jobs_created_at ON data_transfer_jobs(created_at DESC);
CREATE INDEX idx_data_transfer_jobs_state ON data_transfer_jobs(state);
CREATE INDEX idx_data_transfer_issues_job_id ON data_transfer_job_issues(job_id);
CREATE INDEX idx_data_transfer_artifacts_job_id ON data_transfer_artifacts(job_id);
