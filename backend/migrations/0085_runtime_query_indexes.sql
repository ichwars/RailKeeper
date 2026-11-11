-- Security log and expiration queries with independent filter and order prefixes.
CREATE INDEX idx_audit_logs_created_order ON audit_logs(created_at DESC, id DESC);
CREATE INDEX idx_rate_limit_attempts_scope_expiry ON rate_limit_attempts(scope, attempted_at);
CREATE INDEX idx_login_origins_expiry ON login_origins(expires_at);
CREATE INDEX idx_password_reset_pending_user ON password_reset_requests(user_id) WHERE handled_at IS NULL;

-- Saved reads by station. Ascending time keys support reverse scans
-- in the same direction as the rowid tie-breaker.
CREATE INDEX idx_digital_center_sessions_ready_station
  ON digital_center_read_sessions(provider, host, port, read_completed_at) WHERE state='ready';
CREATE INDEX idx_digital_center_work_items_session_order
  ON digital_center_work_items(session_id, name COLLATE NOCASE, center_object_id, id);
DROP INDEX idx_digital_center_work_items_session;
