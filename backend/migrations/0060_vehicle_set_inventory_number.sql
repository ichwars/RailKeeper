ALTER TABLE vehicle_sets ADD COLUMN inventory_number TEXT NOT NULL DEFAULT '';

INSERT INTO inventory_number_schemes(
  id, category, prefix, next_number, padding, active, created_at, updated_at
)
VALUES(
  lower(hex(randomblob(16))), 'Set', 'RK-SET', 1, 6, 1, datetime('now'), datetime('now')
)
ON CONFLICT(category) DO NOTHING;

WITH ranked AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at, id) - 1 AS number_offset
  FROM vehicle_sets
)
UPDATE vehicle_sets
SET inventory_number = (
  SELECT printf('%s-%0*d', scheme.prefix, scheme.padding, scheme.next_number + ranked.number_offset)
  FROM ranked
  JOIN inventory_number_schemes scheme ON scheme.category = 'Set'
  WHERE ranked.id = vehicle_sets.id
)
WHERE inventory_number = '';

UPDATE inventory_number_schemes
SET next_number = next_number + (SELECT COUNT(*) FROM vehicle_sets),
    updated_at = datetime('now')
WHERE category = 'Set';

CREATE UNIQUE INDEX ux_vehicle_sets_inventory_number
  ON vehicle_sets(inventory_number);

CREATE TRIGGER vehicle_sets_inventory_number_required_insert
BEFORE INSERT ON vehicle_sets
WHEN TRIM(NEW.inventory_number) = ''
BEGIN
  SELECT RAISE(ABORT, 'vehicle set inventory number required');
END;

CREATE TRIGGER vehicle_sets_inventory_number_required_update
BEFORE UPDATE OF inventory_number ON vehicle_sets
WHEN TRIM(NEW.inventory_number) = ''
BEGIN
  SELECT RAISE(ABORT, 'vehicle set inventory number required');
END;
