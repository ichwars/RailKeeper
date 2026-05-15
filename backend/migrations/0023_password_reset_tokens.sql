ALTER TABLE password_reset_requests ADD COLUMN token_hash TEXT;
ALTER TABLE password_reset_requests ADD COLUMN expires_at TEXT;
ALTER TABLE password_reset_requests ADD COLUMN used_at TEXT;

CREATE INDEX IF NOT EXISTS idx_password_reset_requests_token_hash
  ON password_reset_requests(token_hash);

CREATE INDEX IF NOT EXISTS idx_password_reset_requests_expires_at
  ON password_reset_requests(expires_at);
