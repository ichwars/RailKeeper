import copy
import json
import unittest

from decoder import NOTE, TABLES, expected_decoder_backfill
from preservation import assert_same, backup_manifest


def fixture():
    old = "2026-10-20T08:30:00Z"
    original = {"format": "railkeeper-backup", "version": 27, "createdAt": "2026-11-07T00:00:00Z",
                "tables": {table: [] for table in TABLES}, "files": [], "profiles": []}
    tables = original["tables"]
    tables["vehicles"] = [{"id": "v", "name": "Reference", "inventory_number": "INV-1"}]
    tables["vehicle_cv_values"] = [
        {"id": "plain", "vehicle_id": "v", "cv_number": 257, "value": -1, "decoder_profile": " Raw ",
         "index_cv31": None, "index_cv32": None, "source_file_id": "file", "source_line": 2,
         "source_exported_at": "2026-10-20T08:30:00", "created_at": old, "updated_at": old},
        {"id": "indexed", "vehicle_id": "v", "cv_number": 257, "value": 11, "decoder_profile": " Raw ",
         "index_cv31": 0, "index_cv32": 0, "source_file_id": "", "created_at": old, "updated_at": old}]
    history = {"id": "h", "cv_value_id": "plain", "vehicle_id": "v", "old_value": 3, "new_value": 5, "changed_at": old}
    tables["vehicle_cv_value_history"] = [history]
    tables["vehicle_cv_files"] = [{"id": "file", "vehicle_id": "v", "file_name": "original.csv",
        "original_name": "original.csv", "decoder_profile": " Raw ", "mime_type": "text/csv", "size_bytes": 13,
        "blob_id": "blob", "storage_path": "", "created_at": old, "updated_at": old}]
    tables["file_blobs"] = [{"id": "blob", "compression": "none", "data": "Q1Y7VmFsdWUKMTszCg=="}]
    observed = copy.deepcopy(original)
    observed["createdAt"] = "2026-11-07T00:00:05Z"
    now = "2026-11-07T00:00:01.123Z"
    workspace, snapshot = "a" * 32, "b" * 32
    observed["tables"]["decoder_cv_workspaces"] = [{"id": workspace, "vehicle_id": "v", "profile_text": " Raw ",
        "origin": "legacy", "installation_id": None, "created_at": now, "updated_at": now}]
    payload = {"values": [
        {"id": "plain", "vehicleId": "v", "cvNumber": 257, "value": -1, "decoderProfile": " Raw ",
         "sourceFileId": "file", "sourceLine": 2, "sourceExportedAt": "2026-10-20T08:30:00",
         "createdAt": old, "updatedAt": old, "history": [
             {"id": "h", "cvValueId": "plain", "vehicleId": "v", "oldValue": 3, "newValue": 5, "changedAt": old}]},
        {"id": "indexed", "vehicleId": "v", "cvNumber": 257, "value": 11, "decoderProfile": " Raw ",
         "indexCV31": 0, "indexCV32": 0, "createdAt": old, "updatedAt": old}]}
    observed["tables"]["decoder_cv_snapshots"] = [{"id": snapshot, "workspace_id": workspace,
        "vehicle_reference_id": "v", "vehicle_name": "Reference", "vehicle_inventory_number": "INV-1",
        "profile_text": " Raw ", "installation_id": None, "kind": "legacy", "scope": "partial_working_state",
        "payload_json": json.dumps(payload), "hardware_json": "{}", "note": NOTE, "recorded_at": now}]
    metadata = {"id": "file", "vehicleId": "v", "fileName": "original.csv", "originalName": "original.csv",
                "decoderProfile": " Raw ", "mimeType": "text/csv", "sizeBytes": 13, "createdAt": old, "updatedAt": old}
    observed["tables"]["decoder_cv_snapshot_sources"] = [{"snapshot_id": snapshot, "source_file_id": "file",
        "metadata_json": json.dumps(metadata), "content_state": "stored", "blob_id": "blob"}]
    return original, observed


class DecoderAdoptionTests(unittest.TestCase):
    def check(self, original, observed):
        expected = expected_decoder_backfill(original, observed)
        assert_same(backup_manifest(expected), backup_manifest(observed), "Independent complete decoder adoption")
        return expected

    def test_complete_old_values_indices_sources_history_and_raw_profiles(self):
        original, observed = fixture()
        self.check(original, observed)
        self.assertEqual(original["tables"]["decoder_cv_snapshots"], [])

    def test_existing_recorded_snapshots_are_not_reclassified_or_refreshed(self):
        _, observed = fixture()
        original = copy.deepcopy(observed)
        self.check(original, copy.deepcopy(observed))
        mutated = copy.deepcopy(observed)
        mutated["tables"]["decoder_cv_snapshots"][0]["recorded_at"] = "2026-11-07T00:00:04Z"
        with self.assertRaises(AssertionError):
            self.check(original, mutated)

    def test_every_semantic_mutation_is_detected(self):
        def payload(observed, change):
            row = observed["tables"]["decoder_cv_snapshots"][0]
            value = json.loads(row["payload_json"])
            change(value)
            row["payload_json"] = json.dumps(value)
        mutations = {
            "CV identity": lambda o: payload(o, lambda p: p["values"][0].update(id="other")),
            "unknown value": lambda o: payload(o, lambda p: p["values"][0].update(value=0)),
            "NULL versus zero page": lambda o: payload(o, lambda p: p["values"][0].update(indexCV31=0, indexCV32=0)),
            "source line": lambda o: payload(o, lambda p: p["values"][0].update(sourceLine=3)),
            "source clock": lambda o: payload(o, lambda p: p["values"][0].update(sourceExportedAt="2026-11-07T00:00:01")),
            "history": lambda o: payload(o, lambda p: p["values"][0].update(history=[])),
            "profile normalization": lambda o: o["tables"]["decoder_cv_workspaces"][0].update(profile_text="Raw"),
            "invented hardware": lambda o: o["tables"]["decoder_cv_snapshots"][0].update(hardware_json='{"decoderId":"guess"}'),
            "complete measurement": lambda o: o["tables"]["decoder_cv_snapshots"][0].update(scope="complete_measurement"),
            "invented recording clock": lambda o: o["tables"]["decoder_cv_snapshots"][0].update(recorded_at="2026-10-20T08:30:00Z"),
            "source blob": lambda o: o["tables"]["decoder_cv_snapshot_sources"][0].update(blob_id="other"),
        }
        for name, change in mutations.items():
            with self.subTest(name=name):
                original, observed = fixture()
                change(observed)
                with self.assertRaises(AssertionError):
                    self.check(original, observed)

    def test_missing_sources_remain_unavailable_without_fabricated_original_metadata(self):
        original, observed = fixture()
        original["tables"]["vehicle_cv_files"] = []
        observed["tables"]["vehicle_cv_files"] = []
        source = observed["tables"]["decoder_cv_snapshot_sources"][0]
        source.update(content_state="unavailable", blob_id=None, metadata_json=json.dumps({
            "id": "file", "vehicleId": "", "fileName": "", "originalName": "", "sizeBytes": 0, "createdAt": "", "updatedAt": ""}))
        self.check(original, observed)


if __name__ == "__main__":
    unittest.main()
