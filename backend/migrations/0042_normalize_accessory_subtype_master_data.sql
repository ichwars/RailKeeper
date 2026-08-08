DELETE FROM master_data_entries AS obsolete
WHERE obsolete.type = 'article_subtype'
  AND EXISTS (
    SELECT 1
    FROM master_data_entries AS canonical
    WHERE canonical.type = 'accessory_subtype'
      AND canonical.key = obsolete.key
  );

UPDATE master_data_entries
SET type = 'accessory_subtype',
    updated_at = datetime('now')
WHERE type = 'article_subtype';
