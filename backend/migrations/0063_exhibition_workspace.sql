ALTER TABLE exhibition_lists ADD COLUMN end_date TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_lists ADD COLUMN location TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_lists ADD COLUMN description TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_lists ADD COLUMN organization_notes TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_lists ADD COLUMN status TEXT NOT NULL DEFAULT 'open';
ALTER TABLE exhibition_lists ADD COLUMN revision INTEGER NOT NULL DEFAULT 1;
ALTER TABLE exhibition_lists ADD COLUMN lock_reason TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_lists ADD COLUMN locked_at TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_lists ADD COLUMN completed_at TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_lists ADD COLUMN archived_at TEXT NOT NULL DEFAULT '';

UPDATE exhibition_lists
SET end_date = list_date,
    status = CASE WHEN locked = 1 THEN 'locked' ELSE 'open' END
WHERE end_date = '';

ALTER TABLE exhibition_entries ADD COLUMN interface_name TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_entries ADD COLUMN availability TEXT NOT NULL DEFAULT 'available';
ALTER TABLE exhibition_entries ADD COLUMN revision INTEGER NOT NULL DEFAULT 1;

UPDATE exhibition_entries
SET interface_name = adapter
WHERE interface_name = '' AND adapter <> '';

CREATE TABLE exhibition_conflict_exceptions (
  id TEXT PRIMARY KEY,
  list_id TEXT NOT NULL,
  conflict_key TEXT NOT NULL,
  reason TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (list_id) REFERENCES exhibition_lists(id) ON DELETE CASCADE,
  UNIQUE (list_id, conflict_key)
);

CREATE INDEX idx_exhibition_conflict_exceptions_list
ON exhibition_conflict_exceptions(list_id, conflict_key);
