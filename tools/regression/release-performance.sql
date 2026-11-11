-- Load into a fresh, migrated test database with no collection data.
-- These rows model query cardinality. Image records use synthetic metadata.
PRAGMA foreign_keys = ON;
BEGIN;

CREATE TEMP TABLE perf_sequence (n INTEGER PRIMARY KEY);
WITH RECURSIVE numbers(n) AS (
  VALUES (1) UNION ALL SELECT n + 1 FROM numbers WHERE n < 144000
)
INSERT INTO perf_sequence SELECT n FROM numbers;

INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, description, exhibition, created_at, updated_at)
SELECT 'v'||n, printf('PERF-%06d', n), 'Maker '||(n%300), 'Vehicle '||n, 'TT', printf('%0200d', n),
       n%100=0, '2026-01-01', printf('2026-01-%02dT%02d:00:00Z', 1+(n%28), n%24)
FROM perf_sequence WHERE n <= 6000;

INSERT INTO vehicle_sets(id, inventory_number, name, manufacturer, gauge, category, gattung, created_at, updated_at)
SELECT 'set'||n, printf('PERF-SET-%06d', n), 'Set '||n, 'Maker '||(n%150), 'TT', 'Wagen', 'Güterwagen',
       '2026-01-01', '2026-01-01'
FROM perf_sequence WHERE n <= 1000;

INSERT INTO vehicle_set_members(vehicle_set_id, vehicle_id, position)
SELECT 'set'||(1+(n-1)/3), 'v'||n, 1+(n-1)%3 FROM perf_sequence WHERE n <= 3000;

INSERT INTO vehicle_images(id, vehicle_id, url, is_primary, sort_order, blob_id, thumbnail_blob_id, created_at)
SELECT 'image'||n, 'v'||(1+(n-1)%6000), 'https://example.test/image.png', n<=6000, (n-1)/6000,
       'blob-image'||n, CASE WHEN n%5<>0 THEN 'blob-thumb'||n ELSE '' END, '2026-01-01'
FROM perf_sequence WHERE n <= 24000;

INSERT INTO vehicle_cv_files(id, vehicle_id, file_name, original_name, storage_path, blob_id, created_at, updated_at)
SELECT 'source'||n, 'v'||n, 'cv.xml', 'cv.xml', '', 'blob-source'||n, '2026-01-01', '2026-01-01'
FROM perf_sequence WHERE n <= 4000;

INSERT INTO vehicle_cv_values(id, vehicle_id, cv_number, value, decoder_profile, source_file_id, created_at, updated_at)
SELECT 'cv'||n, 'v'||(1+(n-1)%4000), 1+(n-1)/4000, n%256, 'DCC',
       CASE WHEN n%3<>0 THEN 'source'||(1+(n-1)%4000) ELSE '' END, '2026-01-01', '2026-01-01'
FROM perf_sequence WHERE n <= 96000;

INSERT INTO vehicle_cv_value_history(id, cv_value_id, vehicle_id, old_value, new_value, changed_at)
SELECT 'h'||n, 'cv'||(1+(n-1)%96000), 'v'||(1+(n-1)%4000), 0, 1,
       printf('2026-01-%02dT%02d:00:00Z', 1+(n%28), n%24)
FROM perf_sequence;

INSERT INTO users(id, username, password_hash, created_at)
SELECT 'perf-user-'||n, 'perf-user-'||n, 'synthetic-disabled', '2026-01-01'
FROM perf_sequence WHERE n <= 1500;

INSERT INTO audit_logs(id, actor_user_id, action, created_at)
SELECT 'audit'||n, 'perf-user-'||(1+(n-1)%1500), 'test',
       printf('2026-01-%02dT%02d:00:00Z', 1+(n%28), n%24)
FROM perf_sequence WHERE n <= 80000;

DROP TABLE perf_sequence;
ANALYZE;
COMMIT;
