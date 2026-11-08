CREATE TABLE login_origins (
  token_hash TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  budget_key TEXT NOT NULL UNIQUE,
  state_hash TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE INDEX idx_login_origins_user ON login_origins(user_id, created_at);

CREATE TRIGGER revoke_login_origins_credentials
AFTER UPDATE OF username,password_hash,two_factor_secret,two_factor_enabled ON users
WHEN OLD.username IS NOT NEW.username OR OLD.password_hash IS NOT NEW.password_hash
  OR OLD.two_factor_secret IS NOT NEW.two_factor_secret OR OLD.two_factor_enabled IS NOT NEW.two_factor_enabled
BEGIN
  DELETE FROM login_origins WHERE user_id=NEW.id;
END;

CREATE TRIGGER revoke_login_origins_role_added AFTER INSERT ON user_roles
BEGIN
  DELETE FROM login_origins WHERE user_id=NEW.user_id;
END;

CREATE TRIGGER revoke_login_origins_role_removed AFTER DELETE ON user_roles
BEGIN
  DELETE FROM login_origins WHERE user_id=OLD.user_id;
END;
