ALTER TABLE data_transfer_profiles
ADD COLUMN name_key TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX idx_data_transfer_profiles_active_name
ON data_transfer_profiles(direction, name_key)
WHERE enabled=1 AND name_key<>'';
