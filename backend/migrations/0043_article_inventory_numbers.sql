ALTER TABLE accessory_products ADD COLUMN inventory_number TEXT NOT NULL DEFAULT '';

INSERT INTO inventory_number_schemes(
  id, category, prefix, next_number, padding, active, created_at, updated_at
)
VALUES(
  lower(hex(randomblob(16))), 'Artikel', 'RK-ART', 1, 6, 1, datetime('now'), datetime('now')
)
ON CONFLICT(category) DO NOTHING;

WITH ranked AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at, id) - 1 AS number_offset
  FROM accessory_products
)
UPDATE accessory_products
SET inventory_number = (
  SELECT printf('%s-%0*d', scheme.prefix, scheme.padding, scheme.next_number + ranked.number_offset)
  FROM ranked
  JOIN inventory_number_schemes scheme ON scheme.category = 'Artikel'
  WHERE ranked.id = accessory_products.id
)
WHERE inventory_number = '';

UPDATE inventory_number_schemes
SET next_number = next_number + (SELECT COUNT(*) FROM accessory_products),
    updated_at = datetime('now')
WHERE category = 'Artikel';

CREATE UNIQUE INDEX ux_accessory_products_inventory_number
  ON accessory_products(inventory_number);

CREATE TRIGGER accessory_products_inventory_number_required_insert
BEFORE INSERT ON accessory_products
WHEN TRIM(NEW.inventory_number) = ''
BEGIN
  SELECT RAISE(ABORT, 'accessory product inventory number required');
END;

CREATE TRIGGER accessory_products_inventory_number_required_update
BEFORE UPDATE OF inventory_number ON accessory_products
WHEN TRIM(NEW.inventory_number) = ''
BEGIN
  SELECT RAISE(ABORT, 'accessory product inventory number required');
END;
