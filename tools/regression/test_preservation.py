import copy
import sqlite3
import tempfile
import unittest
from contextlib import closing
from pathlib import Path

from fixture import record_comparison_projection

from preservation import (
    assert_same, backup_manifest, database_manifest, expected_backup_upgrade, expected_fresh_restore, file_manifest, vehicle_edit_manifest,
)


class PreservationTests(unittest.TestCase):
    def test_decoder_schema_only_adds_five_empty_tables_and_retains_every_old_row(self):
        from decoder import TABLES
        original = {"format": "railkeeper-backup", "version": 26,
                    "tables": {"vehicle_cv_values": [{"id": "cv", "decoder_profile": " Raw ", "index_cv31": None}]}}
        expected = expected_backup_upgrade(original, 27, decoder_model=True)
        self.assertEqual(expected["version"], 27)
        self.assertEqual(expected["tables"]["vehicle_cv_values"], original["tables"]["vehicle_cv_values"])
        self.assertEqual(set(expected["tables"]) - set(original["tables"]), set(TABLES))
        self.assertTrue(all(expected["tables"][table] == [] for table in TABLES))
        self.assertEqual(original["version"], 26)
        with self.assertRaises(AssertionError):
            expected_backup_upgrade(original, 27)
        with self.assertRaises(AssertionError):
            expected_backup_upgrade(original, 28, decoder_model=True)

    def test_decoder_upgrade_does_not_accept_unknown_or_accidentally_prefilled_data(self):
        original = {"version": 26, "tables": {"decoder_instances": [{"id": "unexpected"}]}}
        with self.assertRaises(AssertionError):
            expected_backup_upgrade(original, 27, decoder_model=True)

    def test_catalog_schema_upgrade_only_adds_known_empty_tables(self):
        original = {"format": "railkeeper-backup", "version": 24, "tables": {"vehicles": [{"id": "same", "name": "Keep"}]}}
        expected = expected_backup_upgrade(original, 25)
        self.assertEqual(expected["tables"]["vehicles"], original["tables"]["vehicles"])
        self.assertEqual(expected["tables"]["catalog_article_revisions"], [])
        self.assertEqual(original["version"], 24)
        changed = copy.deepcopy(expected)
        changed["tables"]["vehicle_article_links"] = [{"vehicle_id": "unexpected"}]
        with self.assertRaises(AssertionError):
            assert_same(expected, changed, "unexpected backfill")
        with self.assertRaises(AssertionError):
            expected_backup_upgrade(original, 26)

    def test_cv_api_projection_retains_indices_sources_and_other_revisions(self):
        cv = {"id": "cv", "vehicleId": "v", "cvNumber": 257, "value": 0, "indexCV31": 0,
              "indexCV32": 0, "sourceLine": 2, "sourceExportedAt": "unknown", "revision": "computed"}
        original = {"revision": "vehicle", "cvValues": [cv], "functions": [{"revision": "function"}]}
        projected = record_comparison_projection(original)
        self.assertEqual(projected["revision"], "vehicle")
        self.assertEqual(projected["functions"], original["functions"])
        self.assertEqual(projected["cvValues"], [{key: value for key, value in cv.items() if key != "revision"}])
        self.assertEqual(original["cvValues"][0]["revision"], "computed")

    def test_new_revision_projection_only_omits_computed_tokens(self):
        vehicle = {"id": "v", "inventoryNumber": "1", "manufacturer": "Test", "name": "Test",
                   "gauge": "TT", "qrCodeEnabled": True, "revision": "core", "imagesRevision": "media"}
        image = {"id": "i", "vehicleId": "v", "url": "private", "isPrimary": True, "sortOrder": 0,
                 "revision": "image", "title": "Keep"}
        attachment = {"id": "a", "vehicleId": "v", "fileName": "a.txt", "originalName": "a.txt",
                      "sizeBytes": 7, "revision": "attachment", "description": "Keep"}
        source = {"vehicles": [vehicle], "images": [image], "attachments": [attachment],
                  "other": {"revision": 3, "imagesRevision": "keep"}}
        result = record_comparison_projection(source)
        self.assertEqual(result["vehicles"], [{k: v for k, v in vehicle.items() if k not in {"revision", "imagesRevision"}}])
        self.assertEqual(result["images"], [{k: v for k, v in image.items() if k != "revision"}])
        creation = {**image, "imagesRevision": "created collection"}
        self.assertEqual(record_comparison_projection(creation), result["images"][0])
        self.assertEqual(result["attachments"], [{k: v for k, v in attachment.items() if k != "revision"}])
        self.assertEqual(result["other"], source["other"])
        self.assertEqual(source["vehicles"][0]["revision"], "core")
        cv_file = {**attachment, "decoderProfile": "Legacy", "sourceReferenceCount": 3}
        projected = record_comparison_projection(cv_file)
        self.assertEqual(projected, {key: value for key, value in cv_file.items() if key not in {"revision", "sourceReferenceCount"}})


    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.db_path = self.root / "railkeeper.db"
        with closing(sqlite3.connect(self.db_path)) as db, db:
            db.execute("CREATE TABLE values_fixture(id TEXT PRIMARY KEY, value, page, payload BLOB)")
            db.executemany("INSERT INTO values_fixture VALUES(?,?,?,?)", [
                ("one", 7, None, b"image"), ("two", 7, 0, b"manual"),
            ])

    def test_same_count_changed_value_id_null_or_blob_is_detected(self):
        before = database_manifest(self.db_path)
        for column, value in [("id", "different"), ("value", "7"), ("page", 0), ("payload", b"changed")]:
            with self.subTest(column=column):
                with closing(sqlite3.connect(self.db_path)) as db, db:
                    db.execute("BEGIN")
                    db.execute(f"UPDATE values_fixture SET {column}=? WHERE id='one'", (value,))
                    db.commit()
                after = database_manifest(self.db_path)
                self.assertEqual(before["values_fixture"]["count"], after["values_fixture"]["count"])
                with self.assertRaises(AssertionError):
                    assert_same(before, after, "regression")
                with closing(sqlite3.connect(self.db_path)) as db, db:
                    db.execute("DELETE FROM values_fixture")
                    db.executemany("INSERT INTO values_fixture VALUES(?,?,?,?)", [
                        ("one", 7, None, b"image"), ("two", 7, 0, b"manual"),
                    ])

    def test_row_order_is_irrelevant_and_manifest_is_read_only(self):
        before = database_manifest(self.db_path)
        original = self.db_path.read_bytes()
        database_manifest(self.db_path)
        self.assertEqual(original, self.db_path.read_bytes())
        with closing(sqlite3.connect(self.db_path)) as db, db:
            db.execute("DELETE FROM values_fixture WHERE id='one'")
            db.execute("INSERT INTO values_fixture VALUES('one',7,NULL,?)", (b"image",))
        assert_same(before, database_manifest(self.db_path), "row order")

    def test_broken_foreign_key_is_rejected(self):
        with closing(sqlite3.connect(self.db_path)) as db, db:
            db.execute("CREATE TABLE children(parent TEXT REFERENCES values_fixture(id))")
            db.execute("INSERT INTO children VALUES('missing')")
        with self.assertRaisesRegex(AssertionError, "foreign-key"):
            database_manifest(self.db_path)

    def test_file_bytes_changed_at_same_size_are_detected(self):
        path = self.root / "manual.txt"
        path.write_bytes(b"old")
        before = file_manifest(self.root)
        path.write_bytes(b"new")
        with self.assertRaises(AssertionError):
            assert_same(before, file_manifest(self.root), "file change")

    def test_export_ignores_only_envelope_clock_and_order(self):
        doc = {"format": "railkeeper-backup", "version": 24, "createdAt": "first",
               "tables": {"vehicles": [{"id": "a", "name": "A"}, {"id": "b", "name": "B"}]},
               "profiles": [], "files": []}
        changed = copy.deepcopy(doc)
        changed["createdAt"] = "later"
        changed["tables"]["vehicles"].reverse()
        self.assertEqual(backup_manifest(doc), backup_manifest(changed))
        changed["tables"]["vehicles"][0]["name"] = "DIFFERENT"
        self.assertNotEqual(backup_manifest(doc), backup_manifest(changed))

    def test_fresh_restore_policy_does_not_mask_unrelated_fields(self):
        symbol = {"id": "symbol", "type": "symbols", "key": "light", "origin": "bundled",
                  "metadata_json": '{"library":"railkeeper-workshop-line"}', "updated_at": "source"}
        custom = {"id": "custom", "type": "manufacturer", "key": "custom", "origin": "custom",
                  "metadata_json": "{}", "updated_at": "original"}
        source = {"version": 25, "tables": {"master_data_entries": [symbol, custom]},
                  "profiles": [{"id": "profile", "created_by_user_id": "source-user"}]}
        local = {"version": 25, "tables": {"master_data_entries": [{**symbol, "updated_at": "target"}]}}
        expected = expected_fresh_restore(source, local, {"target-user"})
        self.assertEqual(expected["tables"]["master_data_entries"][0], custom)
        self.assertEqual(expected["tables"]["master_data_entries"][1]["updated_at"], "target")
        self.assertIsNone(expected["profiles"][0]["created_by_user_id"])
        self.assertEqual(source["profiles"][0]["created_by_user_id"], "source-user")
        self.assertEqual(source["tables"]["master_data_entries"][0]["updated_at"], "source")

    def test_vehicle_edit_allows_only_its_own_edit_clocks(self):
        doc = {"format": "railkeeper-backup", "version": 24, "files": [], "profiles": [],
               "tables": {"vehicles": [{"id": "a", "updated_at": "old"}, {"id": "b", "updated_at": "old"}],
                          "vehicle_images": [{"id": "image", "vehicle_id": "a", "source_url": "", "updated_at": "old"}]}}
        changed = copy.deepcopy(doc)
        changed["tables"]["vehicles"][0]["updated_at"] = "new"
        changed["tables"]["vehicle_images"][0]["updated_at"] = "new"
        self.assertEqual(vehicle_edit_manifest(doc, "a"), vehicle_edit_manifest(changed, "a"))
        changed["tables"]["vehicle_images"][0]["source_url"] = "invented"
        self.assertNotEqual(vehicle_edit_manifest(doc, "a"), vehicle_edit_manifest(changed, "a"))
        changed = copy.deepcopy(doc)
        changed["tables"]["vehicles"][1]["updated_at"] = "new"
        self.assertNotEqual(vehicle_edit_manifest(doc, "a"), vehicle_edit_manifest(changed, "a"))


if __name__ == "__main__":
    unittest.main()
