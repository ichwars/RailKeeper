ALTER TABLE exhibition_entries ADD COLUMN vehicle_id TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_exhibition_entries_vehicle ON exhibition_entries(vehicle_id, list_id);

UPDATE exhibition_entries
SET vehicle_id = (
  SELECT v.id
  FROM vehicles v
  WHERE exhibition_entries.vehicle_id = ''
    AND exhibition_entries.decoder_number <> ''
    AND v.digital_decoder_number = exhibition_entries.decoder_number
  LIMIT 1
)
WHERE vehicle_id = ''
  AND decoder_number <> ''
  AND EXISTS (
    SELECT 1
    FROM vehicles v
    WHERE v.digital_decoder_number = exhibition_entries.decoder_number
  );

UPDATE exhibition_entries
SET vehicle_id = (
  SELECT v.id
  FROM vehicles v
  WHERE exhibition_entries.vehicle_id = ''
    AND lower(trim(v.name)) = lower(trim(exhibition_entries.locomotive_name))
    AND lower(trim(v.manufacturer)) = lower(trim(exhibition_entries.manufacturer))
  LIMIT 1
)
WHERE vehicle_id = ''
  AND locomotive_name <> ''
  AND manufacturer <> ''
  AND EXISTS (
    SELECT 1
    FROM vehicles v
    WHERE lower(trim(v.name)) = lower(trim(exhibition_entries.locomotive_name))
      AND lower(trim(v.manufacturer)) = lower(trim(exhibition_entries.manufacturer))
  );
