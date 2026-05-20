UPDATE master_data_entries
SET label = TRIM(SUBSTR(label, INSTR(label, ' - ') + 3)),
    source_url = '',
    updated_at = datetime('now')
WHERE type = 'cv8_manufacturer'
  AND label GLOB '[0-9][0-9][0-9] - *';

UPDATE master_data_entries
SET source_url = '',
    updated_at = datetime('now')
WHERE type = 'cv8_manufacturer'
  AND COALESCE(source_url, '') <> '';
