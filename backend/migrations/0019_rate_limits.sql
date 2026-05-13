CREATE TABLE IF NOT EXISTS rate_limit_attempts (
  id TEXT PRIMARY KEY,
  scope TEXT NOT NULL,
  key TEXT NOT NULL,
  attempted_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_rate_limit_attempts_scope_key_time
  ON rate_limit_attempts(scope, key, attempted_at);
