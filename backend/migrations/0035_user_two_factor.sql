ALTER TABLE users ADD COLUMN two_factor_secret TEXT;
ALTER TABLE users ADD COLUMN two_factor_enabled INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN two_factor_enabled_at TEXT;

CREATE INDEX IF NOT EXISTS idx_users_two_factor_enabled
  ON users(two_factor_enabled);
