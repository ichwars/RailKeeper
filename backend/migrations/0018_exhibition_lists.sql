INSERT INTO roles(id, name)
SELECT lower(hex(randomblob(16))), 'Messe'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name='Messe');

CREATE TABLE IF NOT EXISTS exhibition_lists (
  id TEXT PRIMARY KEY,
  designation TEXT NOT NULL,
  list_date TEXT NOT NULL,
  locked INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_exhibition_lists_date ON exhibition_lists(list_date DESC, designation);

CREATE TABLE IF NOT EXISTS exhibition_entries (
  id TEXT PRIMARY KEY,
  list_id TEXT NOT NULL,
  owner TEXT NOT NULL,
  image_url TEXT,
  locomotive_name TEXT NOT NULL,
  dt_decoder INTEGER NOT NULL DEFAULT 0,
  decoder_number TEXT,
  function_keys TEXT,
  notes TEXT,
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (list_id) REFERENCES exhibition_lists(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_exhibition_entries_list ON exhibition_entries(list_id, sort_order, locomotive_name);
