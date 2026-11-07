-- Preserve exact legacy working labels. They do not identify physical hardware.
INSERT OR IGNORE INTO decoder_cv_workspaces
  (id, vehicle_id, profile_text, origin, installation_id, created_at, updated_at)
SELECT lower(hex(randomblob(16))), vehicle_id, profile_text, 'legacy', NULL,
       strftime('%Y-%m-%dT%H:%M:%fZ','now'), strftime('%Y-%m-%dT%H:%M:%fZ','now')
FROM (
  SELECT vehicle_id, COALESCE(decoder_profile,'') AS profile_text FROM vehicle_cv_values
  UNION
  SELECT vehicle_id, COALESCE(decoder_profile,'') AS profile_text FROM vehicle_cv_files
);
