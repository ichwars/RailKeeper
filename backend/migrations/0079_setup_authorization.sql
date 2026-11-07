CREATE TABLE setup_authorizations (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  token_hash TEXT NOT NULL,
  created_at TEXT NOT NULL,
  consumed_at TEXT
);
