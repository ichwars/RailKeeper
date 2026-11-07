CREATE VIEW IF NOT EXISTS vehicle_catalog_legacy_fields AS
SELECT id AS entity_id, json_object(
  'name', COALESCE(name, ''), 'manufacturer', COALESCE(manufacturer, ''),
  'articleNumber', COALESCE(article_number, ''), 'articleSourceUrl', COALESCE(article_source_url, ''),
  'gauge', COALESCE(gauge, ''), 'epoch', COALESCE(epoch, ''),
  'railwayCompany', COALESCE(railway_company, ''), 'category', COALESCE(category, ''),
  'gattung', COALESCE(gattung, ''), 'description', COALESCE(description, ''),
  'ean', COALESCE(ean, ''), 'productionPeriod', COALESCE(production_period, ''),
  'listPrice', COALESCE(list_price, '')
) AS fields_json FROM vehicles;

CREATE VIEW IF NOT EXISTS vehicle_set_catalog_legacy_fields AS
SELECT id AS entity_id, json_object(
  'name', COALESCE(name, ''), 'manufacturer', COALESCE(manufacturer, ''),
  'articleNumber', COALESCE(article_number, ''), 'articleSourceUrl', COALESCE(article_source_url, ''),
  'gauge', COALESCE(gauge, ''), 'epoch', COALESCE(epoch, ''),
  'railwayCompany', COALESCE(railway_company, ''), 'category', COALESCE(category, ''),
  'gattung', COALESCE(gattung, ''), 'description', COALESCE(description, ''),
  'ean', COALESCE(ean, ''), 'productionPeriod', COALESCE(production_period, ''),
  'listPrice', COALESCE(list_price, '')
) AS fields_json FROM vehicle_sets;

INSERT INTO catalog_articles(id, origin, source_entity_id, created_at)
SELECT 'legacy-set:' || s.entity_id, 'legacy_set', s.entity_id, strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
FROM vehicle_set_catalog_legacy_fields s
WHERE NOT EXISTS (SELECT 1 FROM vehicle_set_article_links l WHERE l.vehicle_set_id=s.entity_id)
  AND NOT EXISTS (SELECT 1 FROM catalog_articles a WHERE a.origin='legacy_set' AND a.source_entity_id=s.entity_id)
UNION ALL
SELECT 'legacy-vehicle:' || v.entity_id, 'legacy_vehicle', v.entity_id, strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
FROM vehicle_catalog_legacy_fields v
WHERE NOT EXISTS (SELECT 1 FROM vehicle_article_links l WHERE l.vehicle_id=v.entity_id)
  AND NOT EXISTS (SELECT 1 FROM vehicle_set_members m WHERE m.vehicle_id=v.entity_id)
  AND NOT EXISTS (SELECT 1 FROM catalog_articles a WHERE a.origin='legacy_vehicle' AND a.source_entity_id=v.entity_id);

INSERT INTO catalog_article_revisions(id, article_id, revision_number, fields_json, note, created_at)
SELECT a.id || ':r1', a.id, 1, COALESCE(v.fields_json, s.fields_json),
  'Übernommene Katalogdarstellung, Quellenzeitpunkt unbekannt. / Imported catalog representation, source time unknown.',
  strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
FROM catalog_articles a
LEFT JOIN vehicle_catalog_legacy_fields v ON a.origin='legacy_vehicle' AND v.entity_id=a.source_entity_id
LEFT JOIN vehicle_set_catalog_legacy_fields s ON a.origin='legacy_set' AND s.entity_id=a.source_entity_id
WHERE (v.entity_id IS NOT NULL OR s.entity_id IS NOT NULL)
  AND NOT EXISTS (SELECT 1 FROM catalog_article_revisions r WHERE r.article_id=a.id);

INSERT INTO vehicle_set_article_links(vehicle_set_id, revision_id, overrides_json, authority, updated_at)
SELECT s.entity_id, r.id,
  (SELECT json_group_object(e.key, e.value) FROM json_each(s.fields_json) e
   WHERE e.value IS NOT json_extract(r.fields_json, '$.' || e.key)),
  'legacy', strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
FROM vehicle_set_catalog_legacy_fields s
JOIN catalog_articles a ON a.origin='legacy_set' AND a.source_entity_id=s.entity_id
JOIN catalog_article_revisions r ON r.article_id=a.id
  AND r.revision_number=(SELECT MIN(r2.revision_number) FROM catalog_article_revisions r2 WHERE r2.article_id=a.id)
WHERE NOT EXISTS (SELECT 1 FROM vehicle_set_article_links l WHERE l.vehicle_set_id=s.entity_id);

INSERT INTO vehicle_article_links(vehicle_id, revision_id, overrides_json, authority, updated_at)
SELECT v.entity_id, r.id,
  (SELECT json_group_object(e.key, e.value) FROM json_each(v.fields_json) e
   WHERE e.value IS NOT json_extract(r.fields_json, '$.' || e.key)),
  'legacy', strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
FROM vehicle_catalog_legacy_fields v
LEFT JOIN vehicle_set_members m ON m.vehicle_id=v.entity_id
LEFT JOIN vehicle_set_article_links sl ON sl.vehicle_set_id=m.vehicle_set_id
JOIN catalog_article_revisions r ON r.id=COALESCE(sl.revision_id,
  (SELECT rv.id FROM catalog_articles av JOIN catalog_article_revisions rv ON rv.article_id=av.id
   WHERE av.origin='legacy_vehicle' AND av.source_entity_id=v.entity_id ORDER BY rv.revision_number LIMIT 1))
WHERE NOT EXISTS (SELECT 1 FROM vehicle_article_links l WHERE l.vehicle_id=v.entity_id);
