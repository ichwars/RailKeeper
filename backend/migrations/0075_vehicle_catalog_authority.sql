ALTER TABLE vehicle_article_links ADD COLUMN source_authority TEXT
  CHECK (source_authority IS NULL OR source_authority IN ('legacy', 'catalog'));

CREATE VIEW vehicle_catalog_composed_fields AS
SELECT l.vehicle_id AS entity_id, json_patch(r.fields_json, l.overrides_json) AS fields_json
FROM vehicle_article_links l JOIN catalog_article_revisions r ON r.id=l.revision_id;

CREATE VIEW vehicle_catalog_consistency AS
SELECT v.entity_id, NOT EXISTS (
  SELECT 1 FROM json_each(v.fields_json) e
  WHERE e.value IS NOT json_extract(c.fields_json, '$.' || e.key)
) AS consistent
FROM vehicle_catalog_legacy_fields v JOIN vehicle_catalog_composed_fields c ON c.entity_id=v.entity_id;

UPDATE vehicle_article_links SET source_authority=authority,
  authority=CASE WHEN (SELECT consistent FROM vehicle_catalog_consistency c WHERE c.entity_id=vehicle_article_links.vehicle_id)
    THEN 'catalog' ELSE 'legacy' END;

ALTER TABLE vehicle_set_article_links ADD COLUMN source_authority TEXT
  CHECK (source_authority IS NULL OR source_authority IN ('legacy', 'catalog'));

CREATE VIEW set_catalog_composed_fields AS
SELECT l.vehicle_set_id AS entity_id, json_patch(r.fields_json, l.overrides_json) AS fields_json
FROM vehicle_set_article_links l JOIN catalog_article_revisions r ON r.id=l.revision_id;

CREATE VIEW set_catalog_consistency AS
SELECT v.entity_id, NOT EXISTS (
  SELECT 1 FROM json_each(v.fields_json) e
  WHERE e.value IS NOT json_extract(c.fields_json, '$.' || e.key)
) AS consistent
FROM vehicle_set_catalog_legacy_fields v JOIN set_catalog_composed_fields c ON c.entity_id=v.entity_id;

UPDATE vehicle_set_article_links SET source_authority=authority,
  authority=CASE WHEN (SELECT consistent FROM set_catalog_consistency c WHERE c.entity_id=vehicle_set_article_links.vehicle_set_id)
    THEN 'catalog' ELSE 'legacy' END;

-- BEGIN catalog projection triggers

CREATE TRIGGER IF NOT EXISTS catalog_vehicle_create AFTER INSERT ON vehicles
WHEN NOT EXISTS (SELECT 1 FROM vehicle_article_links l WHERE l.vehicle_id=NEW.id)
BEGIN
  INSERT INTO catalog_articles(id, origin, source_entity_id, created_at)
  VALUES('owned-vehicle:' || NEW.id, 'manual', NULL, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
  ON CONFLICT(id) DO NOTHING;
  INSERT INTO catalog_article_revisions(id, article_id, revision_number, fields_json, note, created_at)
  SELECT 'owned-vehicle:' || NEW.id || ':r1', 'owned-vehicle:' || NEW.id, 1, v.fields_json,
    'Bestätigte Bestandsangaben. / Confirmed inventory fields.', strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
  FROM vehicle_catalog_legacy_fields v WHERE v.entity_id=NEW.id
    AND NOT EXISTS (SELECT 1 FROM catalog_article_revisions r WHERE r.article_id='owned-vehicle:' || NEW.id);
  INSERT INTO vehicle_article_links(vehicle_id, revision_id, overrides_json, authority, updated_at)
  SELECT NEW.id, r.id,
    (SELECT json_group_object(e.key, e.value) FROM json_each(v.fields_json) e
     WHERE e.value IS NOT json_extract(r.fields_json, '$.' || e.key)),
    'catalog', strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
  FROM vehicle_catalog_legacy_fields v JOIN catalog_article_revisions r ON r.article_id='owned-vehicle:' || NEW.id
  WHERE v.entity_id=NEW.id
    AND r.revision_number=(SELECT MIN(r2.revision_number) FROM catalog_article_revisions r2 WHERE r2.article_id='owned-vehicle:' || NEW.id);
END;

CREATE TRIGGER IF NOT EXISTS catalog_vehicle_update AFTER UPDATE OF name, manufacturer, article_number, article_source_url, gauge, epoch, railway_company, category, gattung, description, ean, production_period, list_price ON vehicles
WHEN COALESCE(NEW.name, '') IS NOT COALESCE(OLD.name, '')
  OR COALESCE(NEW.manufacturer, '') IS NOT COALESCE(OLD.manufacturer, '')
  OR COALESCE(NEW.article_number, '') IS NOT COALESCE(OLD.article_number, '')
  OR COALESCE(NEW.article_source_url, '') IS NOT COALESCE(OLD.article_source_url, '')
  OR COALESCE(NEW.gauge, '') IS NOT COALESCE(OLD.gauge, '')
  OR COALESCE(NEW.epoch, '') IS NOT COALESCE(OLD.epoch, '')
  OR COALESCE(NEW.railway_company, '') IS NOT COALESCE(OLD.railway_company, '')
  OR COALESCE(NEW.category, '') IS NOT COALESCE(OLD.category, '')
  OR COALESCE(NEW.gattung, '') IS NOT COALESCE(OLD.gattung, '')
  OR COALESCE(NEW.description, '') IS NOT COALESCE(OLD.description, '')
  OR COALESCE(NEW.ean, '') IS NOT COALESCE(OLD.ean, '')
  OR COALESCE(NEW.production_period, '') IS NOT COALESCE(OLD.production_period, '')
  OR COALESCE(NEW.list_price, '') IS NOT COALESCE(OLD.list_price, '')
BEGIN
  UPDATE vehicle_article_links SET overrides_json=(
    SELECT json_group_object(e.key, e.value) FROM vehicle_catalog_legacy_fields v, json_each(v.fields_json) e
    JOIN catalog_article_revisions r ON r.id=vehicle_article_links.revision_id
    WHERE v.entity_id=NEW.id AND e.value IS NOT json_extract(r.fields_json, '$.' || e.key)
  ), updated_at=strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
  WHERE vehicle_id=NEW.id AND authority='catalog';
  UPDATE vehicle_article_links SET source_authority=COALESCE(source_authority, authority), authority='catalog'
  WHERE vehicle_id=NEW.id AND authority='legacy'
    AND (SELECT consistent FROM vehicle_catalog_consistency c WHERE c.entity_id=NEW.id);
END;

CREATE TRIGGER IF NOT EXISTS catalog_vehicle_link AFTER INSERT ON vehicle_article_links
WHEN NEW.authority='legacy'
  AND (SELECT consistent FROM vehicle_catalog_consistency c WHERE c.entity_id=NEW.vehicle_id)
BEGIN
  UPDATE vehicle_article_links SET source_authority=COALESCE(source_authority, authority), authority='catalog'
  WHERE vehicle_id=NEW.vehicle_id;
END;

CREATE TRIGGER IF NOT EXISTS catalog_set_create AFTER INSERT ON vehicle_sets
WHEN NOT EXISTS (SELECT 1 FROM vehicle_set_article_links l WHERE l.vehicle_set_id=NEW.id)
BEGIN
  INSERT INTO catalog_articles(id, origin, source_entity_id, created_at)
  VALUES('owned-set:' || NEW.id, 'manual', NULL, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
  ON CONFLICT(id) DO NOTHING;
  INSERT INTO catalog_article_revisions(id, article_id, revision_number, fields_json, note, created_at)
  SELECT 'owned-set:' || NEW.id || ':r1', 'owned-set:' || NEW.id, 1, v.fields_json,
    'Bestätigte Bestandsangaben. / Confirmed inventory fields.', strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
  FROM vehicle_set_catalog_legacy_fields v WHERE v.entity_id=NEW.id
    AND NOT EXISTS (SELECT 1 FROM catalog_article_revisions r WHERE r.article_id='owned-set:' || NEW.id);
  INSERT INTO vehicle_set_article_links(vehicle_set_id, revision_id, overrides_json, authority, updated_at)
  SELECT NEW.id, r.id,
    (SELECT json_group_object(e.key, e.value) FROM json_each(v.fields_json) e
     WHERE e.value IS NOT json_extract(r.fields_json, '$.' || e.key)),
    'catalog', strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
  FROM vehicle_set_catalog_legacy_fields v JOIN catalog_article_revisions r ON r.article_id='owned-set:' || NEW.id
  WHERE v.entity_id=NEW.id
    AND r.revision_number=(SELECT MIN(r2.revision_number) FROM catalog_article_revisions r2 WHERE r2.article_id='owned-set:' || NEW.id);
END;

CREATE TRIGGER IF NOT EXISTS catalog_set_update AFTER UPDATE OF name, manufacturer, article_number, article_source_url, gauge, epoch, railway_company, category, gattung, description, ean, production_period, list_price ON vehicle_sets
WHEN COALESCE(NEW.name, '') IS NOT COALESCE(OLD.name, '')
  OR COALESCE(NEW.manufacturer, '') IS NOT COALESCE(OLD.manufacturer, '')
  OR COALESCE(NEW.article_number, '') IS NOT COALESCE(OLD.article_number, '')
  OR COALESCE(NEW.article_source_url, '') IS NOT COALESCE(OLD.article_source_url, '')
  OR COALESCE(NEW.gauge, '') IS NOT COALESCE(OLD.gauge, '')
  OR COALESCE(NEW.epoch, '') IS NOT COALESCE(OLD.epoch, '')
  OR COALESCE(NEW.railway_company, '') IS NOT COALESCE(OLD.railway_company, '')
  OR COALESCE(NEW.category, '') IS NOT COALESCE(OLD.category, '')
  OR COALESCE(NEW.gattung, '') IS NOT COALESCE(OLD.gattung, '')
  OR COALESCE(NEW.description, '') IS NOT COALESCE(OLD.description, '')
  OR COALESCE(NEW.ean, '') IS NOT COALESCE(OLD.ean, '')
  OR COALESCE(NEW.production_period, '') IS NOT COALESCE(OLD.production_period, '')
  OR COALESCE(NEW.list_price, '') IS NOT COALESCE(OLD.list_price, '')
BEGIN
  UPDATE vehicle_set_article_links SET overrides_json=(
    SELECT json_group_object(e.key, e.value) FROM vehicle_set_catalog_legacy_fields v, json_each(v.fields_json) e
    JOIN catalog_article_revisions r ON r.id=vehicle_set_article_links.revision_id
    WHERE v.entity_id=NEW.id AND e.value IS NOT json_extract(r.fields_json, '$.' || e.key)
  ), updated_at=strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
  WHERE vehicle_set_id=NEW.id AND authority='catalog';
  UPDATE vehicle_set_article_links SET source_authority=COALESCE(source_authority, authority), authority='catalog'
  WHERE vehicle_set_id=NEW.id AND authority='legacy'
    AND (SELECT consistent FROM set_catalog_consistency c WHERE c.entity_id=NEW.id);
END;

CREATE TRIGGER IF NOT EXISTS catalog_set_link AFTER INSERT ON vehicle_set_article_links
WHEN NEW.authority='legacy'
  AND (SELECT consistent FROM set_catalog_consistency c WHERE c.entity_id=NEW.vehicle_set_id)
BEGIN
  UPDATE vehicle_set_article_links SET source_authority=COALESCE(source_authority, authority), authority='catalog'
  WHERE vehicle_set_id=NEW.vehicle_set_id;
END;
