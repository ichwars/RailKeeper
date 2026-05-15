UPDATE master_data_entries
SET label='Piko', updated_at=datetime('now')
WHERE type='manufacturer' AND key='piko-spielwaren';

UPDATE master_data_entries
SET label='Tillig', updated_at=datetime('now')
WHERE type='manufacturer' AND key='tillig-modellbahnen';

UPDATE vehicles
SET manufacturer='Piko', updated_at=datetime('now')
WHERE manufacturer='Piko Spielwaren';

UPDATE vehicles
SET manufacturer='Tillig', updated_at=datetime('now')
WHERE manufacturer='Tillig Modellbahnen';
