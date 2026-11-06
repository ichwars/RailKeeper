import copy
import json
import unittest
from datetime import datetime, timezone

from catalog import NOTE, TABLES, expected_catalog_backfill
from preservation import assert_same, backup_manifest


def fixture():
    now = datetime.now(timezone.utc).isoformat(timespec="milliseconds").replace("+00:00", "Z")
    fields = {"name": "Same", "manufacturer": "Piko", "articleNumber": "123", "articleSourceUrl": "", "gauge": "H0",
              "epoch": "", "railwayCompany": "", "category": "Lokomotive", "gattung": "Diesellok", "description": "",
              "ean": "", "productionPeriod": "", "listPrice": "100"}
    legacy = {"manufacturer": "Piko", "name": "Same", "article_number": "123", "gauge": "H0", "category": "Lokomotive",
              "gattung": "Diesellok", "list_price": "100", "purchase_price": "75", "condition": "Individual"}
    original = {"format": "railkeeper-backup", "version": 25, "tables": {"vehicles": [{**legacy, "id": "a"}, {**legacy, "id": "b"}],
                "vehicle_sets": [], "vehicle_set_members": [], **{table: [] for table in TABLES}}, "files": [], "profiles": []}
    actual = copy.deepcopy(original)
    for identity in ("a", "b"):
        article = "legacy-vehicle:" + identity
        actual["tables"][TABLES[0]].append({"id": article, "origin": "legacy_vehicle", "source_entity_id": identity, "created_at": now})
        actual["tables"][TABLES[1]].append({"id": article + ":r1", "article_id": article, "revision_number": 1,
                                          "fields_json": json.dumps(fields), "note": NOTE, "created_at": now})
        actual["tables"][TABLES[2]].append({"vehicle_id": identity, "revision_id": article + ":r1", "overrides_json": "{}",
                                          "authority": "legacy", "updated_at": now})
    return original, actual


class CatalogPreservationTests(unittest.TestCase):
    def test_existing_manual_articles_and_unresolved_overrides_remain_exact(self):
        _, original = fixture()
        for article in original["tables"][TABLES[0]]:
            article.update(origin="manual", source_entity_id=None, created_at="2020-01-01T00:00:00Z")
        original["tables"][TABLES[2]][0]["overrides_json"] = '{"name":"Pending catalog choice"}'
        expected = expected_catalog_backfill(original, copy.deepcopy(original))
        self.assertEqual(backup_manifest(expected), backup_manifest(original))
        self.assertEqual(len(expected["tables"][TABLES[0]]), 2)

    def test_identical_owned_vehicles_keep_separate_articles_and_all_individual_data(self):
        original, actual = fixture()
        expected = expected_catalog_backfill(original, actual)
        self.assertEqual(backup_manifest(expected), backup_manifest(actual))
        self.assertEqual(expected["tables"]["vehicles"], original["tables"]["vehicles"])
        self.assertEqual(original["tables"][TABLES[0]], [])
        self.assertNotEqual(expected["tables"][TABLES[2]][0]["revision_id"], expected["tables"][TABLES[2]][1]["revision_id"])

    def test_wrong_fields_overrides_missing_links_and_invented_source_time_are_detected(self):
        for change in ("field", "override", "missing", "time"):
            with self.subTest(change=change):
                original, actual = fixture()
                if change == "field":
                    values = json.loads(actual["tables"][TABLES[1]][0]["fields_json"]); values["name"] = "Wrong"
                    actual["tables"][TABLES[1]][0]["fields_json"] = json.dumps(values)
                if change == "override": actual["tables"][TABLES[2]][0]["overrides_json"] = '{"purchasePrice":"75"}'
                if change == "missing": actual["tables"][TABLES[2]].pop()
                if change == "time": actual["tables"][TABLES[0]][0]["created_at"] = "2020-01-01T00:00:00Z"
                with self.assertRaises(AssertionError): expected_catalog_backfill(original, actual)

    def test_extra_rows_and_changed_authority_are_not_masked_by_new_timestamps(self):
        for change in ("extra", "authority"):
            with self.subTest(change=change):
                original, actual = fixture()
                if change == "extra": actual["tables"][TABLES[0]].append({**actual["tables"][TABLES[0]][0], "id": "unexpected"})
                else: actual["tables"][TABLES[2]][0]["authority"] = "catalog"
                expected = expected_catalog_backfill(original, actual)
                with self.assertRaises(AssertionError): assert_same(backup_manifest(expected), backup_manifest(actual), "Catalog content")


if __name__ == "__main__":
    unittest.main()
