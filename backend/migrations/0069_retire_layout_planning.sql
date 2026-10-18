-- Planning data remains a dormant archive for older backups. No application API writes it.
-- This SQL also runs inside the restore transaction, so it must remain idempotent.
CREATE TABLE IF NOT EXISTS retired_allocation_records (
  source_table TEXT NOT NULL,
  source_id TEXT NOT NULL,
  row_json TEXT NOT NULL CHECK (json_valid(row_json)),
  retired_at TEXT NOT NULL,
  PRIMARY KEY (source_table, source_id)
);

INSERT OR IGNORE INTO retired_allocation_records(source_table, source_id, row_json, retired_at)
SELECT 'accessory_reservations', id, json_object(
  'id', id,
  'product_id', product_id,
  'asset_id', asset_id,
  'location_id', location_id,
  'quantity', quantity,
  'vehicle_id', vehicle_id,
  'layout_id', layout_id,
  'layout_unit_id', layout_unit_id,
  'status', status,
  'note', note,
  'created_by', created_by,
  'created_at', created_at,
  'updated_at', updated_at,
  'placement', placement,
  'digital_address', digital_address,
  'decoder_output', decoder_output,
  'connection', connection,
  'wiring_notes', wiring_notes
), strftime('%Y-%m-%dT%H:%M:%SZ','now')
FROM accessory_reservations WHERE vehicle_id IS NULL;

INSERT OR IGNORE INTO retired_allocation_records(source_table, source_id, row_json, retired_at)
SELECT 'accessory_installations', id, json_object(
  'id', id,
  'product_id', product_id,
  'asset_id', asset_id,
  'source_location_id', source_location_id,
  'quantity', quantity,
  'vehicle_id', vehicle_id,
  'layout_id', layout_id,
  'layout_unit_id', layout_unit_id,
  'condition_state', condition_state,
  'installed_by', installed_by,
  'installed_at', installed_at,
  'removed_by', removed_by,
  'removed_at', removed_at,
  'removal_disposition', removal_disposition,
  'notes', notes,
  'removal_notes', removal_notes,
  'placement', placement,
  'digital_address', digital_address,
  'decoder_output', decoder_output,
  'connection', connection,
  'wiring_notes', wiring_notes
), strftime('%Y-%m-%dT%H:%M:%SZ','now')
FROM accessory_installations WHERE vehicle_id IS NULL;


-- Reservations have never removed counted stock. Release only individually reserved items.
UPDATE accessory_assets SET lifecycle_state='stored',
 storage_location_id=(SELECT location_id FROM accessory_reservations r WHERE r.asset_id=accessory_assets.id AND r.status='active' AND r.vehicle_id IS NULL),
 updated_at=strftime('%Y-%m-%dT%H:%M:%SZ','now')
WHERE lifecycle_state='reserved' AND id IN (SELECT asset_id FROM accessory_reservations WHERE vehicle_id IS NULL AND status='active');
UPDATE accessory_reservations SET status='cancelled', updated_at=strftime('%Y-%m-%dT%H:%M:%SZ','now')
WHERE vehicle_id IS NULL AND status='active';

-- Counted installed stock is returned exactly once; individual items remain individually tracked.
INSERT INTO accessory_stock(product_id,location_id,quantity,updated_at)
SELECT product_id,source_location_id,SUM(quantity),strftime('%Y-%m-%dT%H:%M:%SZ','now')
FROM accessory_installations WHERE vehicle_id IS NULL AND removed_at IS NULL AND asset_id IS NULL
GROUP BY product_id,source_location_id
ON CONFLICT(product_id,location_id) DO UPDATE SET quantity=quantity+excluded.quantity,updated_at=excluded.updated_at;
INSERT INTO accessory_stock_movements(id,product_id,location_id,movement_type,quantity,source_type,source_id,actor,note,created_at)
SELECT lower(hex(randomblob(16))),product_id,source_location_id,'removal',quantity,'retired_layout',id,'system',
 'Anlagenplanung entfernt / Layout planning retired',strftime('%Y-%m-%dT%H:%M:%SZ','now')
FROM accessory_installations WHERE vehicle_id IS NULL AND removed_at IS NULL AND asset_id IS NULL;
UPDATE accessory_assets SET lifecycle_state=CASE WHEN condition_state IN ('maintenance_due','defective') THEN 'maintenance' ELSE 'stored' END,
 storage_location_id=(SELECT source_location_id FROM accessory_installations i WHERE i.asset_id=accessory_assets.id AND i.vehicle_id IS NULL AND i.removed_at IS NULL),
 updated_at=strftime('%Y-%m-%dT%H:%M:%SZ','now')
WHERE id IN (SELECT asset_id FROM accessory_installations WHERE vehicle_id IS NULL AND removed_at IS NULL);
UPDATE accessory_installations SET removed_at=strftime('%Y-%m-%dT%H:%M:%SZ','now'),removed_by='system',removal_disposition='stored',
 removal_notes='Anlagenplanung entfernt / Layout planning retired'
WHERE vehicle_id IS NULL AND removed_at IS NULL;
