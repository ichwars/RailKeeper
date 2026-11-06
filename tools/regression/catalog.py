"""Independent expected catalog relationships for a synthetic legacy reference."""

import copy
import json
from datetime import datetime, timedelta, timezone


FIELDS = {
    "name": "name", "manufacturer": "manufacturer", "articleNumber": "article_number",
    "articleSourceUrl": "article_source_url", "gauge": "gauge", "epoch": "epoch",
    "railwayCompany": "railway_company", "category": "category", "gattung": "gattung",
    "description": "description", "ean": "ean", "productionPeriod": "production_period", "listPrice": "list_price",
}
TABLES = ("catalog_articles", "catalog_article_revisions", "vehicle_article_links", "vehicle_set_article_links")
NOTE = "Übernommene Katalogdarstellung, Quellenzeitpunkt unbekannt. / Imported catalog representation, source time unknown."


def expected_catalog_backfill(document, actual):
    expected = copy.deepcopy(document)
    tables = expected["tables"]
    for table in TABLES:
        tables.setdefault(table, [])
    actual_rows = {table: {row["id" if table in TABLES[:2] else "vehicle_id" if table == TABLES[2] else "vehicle_set_id"]: row
                           for row in actual["tables"].get(table, [])} for table in TABLES}

    def append(table, identity, row, clock):
        actual_row = actual_rows[table].get(identity)
        if actual_row is None:
            raise AssertionError("Expected legacy catalog assignment is absent from " + table)
        value = actual_row[clock]
        timestamp = datetime.fromisoformat(value.replace("Z", "+00:00"))
        now = datetime.now(timezone.utc)
        if timestamp.tzinfo is None or not now - timedelta(minutes=15) <= timestamp <= now + timedelta(seconds=5):
            raise AssertionError("A new catalog adoption must have its own current timestamp")
        row[clock] = value
        for column in ("fields_json", "overrides_json"):
            if column in row:
                if json.loads(actual_row[column]) != row[column]:
                    raise AssertionError("Catalog values differ from the complete legacy representation")
                row[column] = actual_row[column]
        tables[table].append(row)

    vehicles = {row["id"]: row for row in tables["vehicles"]}
    sets = {row["id"]: row for row in tables["vehicle_sets"]}
    membership = {row["vehicle_id"]: row["vehicle_set_id"] for row in tables["vehicle_set_members"]}
    vehicle_links = {row["vehicle_id"]: row for row in tables[TABLES[2]]}
    set_links = {row["vehicle_set_id"]: row for row in tables[TABLES[3]]}
    articles = {(row["origin"], row["source_entity_id"]): row for row in tables[TABLES[0]] if row["origin"] != "manual"}
    fields = lambda row: {key: "" if row.get(column) is None else row[column] for key, column in FIELDS.items()}

    for kind, entities, links in (("set", sets, set_links), ("vehicle", vehicles, vehicle_links)):
        for identity in entities:
            if identity in links or (kind == "vehicle" and identity in membership):
                continue
            source = ("legacy_" + kind, identity)
            if source not in articles:
                row = {"id": "legacy-" + kind + ":" + identity, "origin": source[0], "source_entity_id": identity}
                append(TABLES[0], row["id"], row, "created_at")
                articles[source] = row
    revisions = {}
    for row in tables[TABLES[1]]:
        if row["article_id"] not in revisions or row["revision_number"] < revisions[row["article_id"]]["revision_number"]:
            revisions[row["article_id"]] = row
    for (origin, identity), article in articles.items():
        entities = sets if origin == "legacy_set" else vehicles if origin == "legacy_vehicle" else {}
        if identity not in entities or article["id"] in revisions:
            continue
        row = {"id": article["id"] + ":r1", "article_id": article["id"], "revision_number": 1,
               "fields_json": fields(entities[identity]), "note": NOTE}
        append(TABLES[1], row["id"], row, "created_at")
        revisions[article["id"]] = row
    revision_ids = {row["id"]: row for row in tables[TABLES[1]]}
    for kind, entities, links, table, column in (("set", sets, set_links, TABLES[3], "vehicle_set_id"),
                                                ("vehicle", vehicles, vehicle_links, TABLES[2], "vehicle_id")):
        for identity, entity in entities.items():
            if identity in links:
                continue
            if kind == "vehicle" and identity in membership:
                revision = revision_ids[set_links[membership[identity]]["revision_id"]]
            else:
                revision = revisions[articles[("legacy_" + kind, identity)]["id"]]
            base = json.loads(revision["fields_json"])
            overrides = {key: value for key, value in fields(entity).items() if base[key] != value}
            row = {column: identity, "revision_id": revision["id"], "overrides_json": overrides, "authority": "legacy"}
            append(table, identity, row, "updated_at")
            links[identity] = row
    return expected


def expected_catalog_authority(document, source_version):
    """Derive only the documented authority conversion, retaining every original value."""
    expected = copy.deepcopy(document)
    tables = expected["tables"]
    revisions = {row["id"]: json.loads(row["fields_json"]) for row in tables[TABLES[1]]}
    for links, entities, identity in ((TABLES[2], "vehicles", "vehicle_id"),
                                     (TABLES[3], "vehicle_sets", "vehicle_set_id")):
        legacy = {row["id"]: {key: row.get(column) or "" for key, column in FIELDS.items()}
                  for row in tables[entities]}
        for row in tables[links]:
            if source_version >= 26:
                row.setdefault("source_authority", None)
                continue
            row["source_authority"] = row["authority"]
            composed = {**revisions[row["revision_id"]], **json.loads(row["overrides_json"])}
            row["authority"] = "catalog" if composed == legacy[row[identity]] else "legacy"
    return expected
