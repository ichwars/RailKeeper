ALTER TABLE exhibition_entries ADD COLUMN gattung TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_entries ADD COLUMN series TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_entries ADD COLUMN manufacturer TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_entries ADD COLUMN epoch TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_entries ADD COLUMN railway_company TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_entries ADD COLUMN decoder_type TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_entries ADD COLUMN adapter TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_entries ADD COLUMN sx_address TEXT NOT NULL DEFAULT '';
ALTER TABLE exhibition_entries ADD COLUMN analog INTEGER NOT NULL DEFAULT 0;

ALTER TABLE vehicles ADD COLUMN decoder_type TEXT NOT NULL DEFAULT '';
