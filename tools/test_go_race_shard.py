import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

import go_race_shard as shards


class RaceShardTests(unittest.TestCase):
    def test_listing_retains_package_identity_and_rejects_duplicates(self):
        events = [
            {"Action": "start", "Package": "pkg/a"},
            {"Action": "output", "Package": "pkg/a", "Output": "TestBackupAlpha\nTestBackupBeta\n"},
            {"Action": "output", "Package": "pkg/b", "Output": "TestBackupAlpha\n"},
            {"Action": "output", "Package": "pkg/a", "Output": "ok pkg/a\n"},
        ]
        text = "\n".join(json.dumps(event) for event in events)
        self.assertEqual(shards.listing_inventory(text), [
            ("pkg/a", "TestBackupAlpha"), ("pkg/a", "TestBackupBeta"), ("pkg/b", "TestBackupAlpha")])
        with self.assertRaises(ValueError):
            shards.listing_inventory(text + "\n" + json.dumps(events[1]))
        for invalid in ["not JSON", "", json.dumps(events[0]), "[]"]:
            with self.assertRaises(ValueError):
                shards.listing_inventory(invalid)

    def test_partition_executes_every_original_case_once_in_each_group(self):
        inventory = [("pkg/a", f"TestBackup{index:03}") for index in range(63)]
        inventory += [("pkg/b", "TestBackup000")]
        parts = [shards.select_shard(inventory, index, 3) for index in range(3)]
        self.assertCountEqual(sum(parts, []), inventory)
        self.assertEqual(len(set(sum(parts, []))), len(inventory))
        self.assertLessEqual(max(map(len, parts)) - min(map(len, parts)), 1)
        for index in range(3):
            self.assertEqual(parts[index], shards.select_shard(list(reversed(inventory)), index, 3))
        for index, count in [(-1, 3), (3, 3), (0, 0)]:
            with self.assertRaises(ValueError):
                shards.select_shard(inventory, index, count)

    def manifests(self, directory):
        paths = []
        for group, (packages, pattern) in shards.GROUPS.items():
            prefix = "TestBackupFixture" if group == "backup" else "TestDecoderConfigurationFixture"
            inventory = [["railkeeper/backend/internal/application", f"{prefix}{n}"]
                         for n in range(10)]
            for index in range(3):
                selected = shards.select_shard(shards.case_list(inventory), index, 3)
                value = {"version": 1, "group": group, "count": 3, "index": index, "source": "same-source",
                         "packages": packages, "pattern": pattern, "inventory": inventory,
                         "selected": selected, "observed": selected, "passed": True}
                path = directory / f"{group}-{index}.manifest.json"
                path.write_text(json.dumps(value))
                paths.append(path)
        return paths

    def test_all_group_manifests_prove_complete_execution(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            self.manifests(directory)
            self.assertEqual(shards.verify_manifests(directory, list(shards.GROUPS), 3),
                             {"backup": 10, "decoder-cv": 10})
            with self.assertRaises(ValueError):
                shards.verify_manifests(directory, list(shards.GROUPS), 3, "another-required-source")

    def test_missing_duplicate_failed_or_inconsistent_manifests_fail(self):
        mutations = ["missing", "duplicate-index", "invalid-index", "failed", "wrong-source",
                     "changed-pattern", "changed-inventory", "missing-observation", "repeated-observation"]
        for mutation in mutations:
            with self.subTest(mutation=mutation), tempfile.TemporaryDirectory() as temporary:
                directory = Path(temporary)
                paths = self.manifests(directory)
                target = paths[1]
                value = json.loads(target.read_text())
                if mutation == "missing":
                    target.unlink()
                else:
                    if mutation == "duplicate-index": value["index"] = 0
                    if mutation == "invalid-index": value["index"] = True
                    if mutation == "failed": value["passed"] = False
                    if mutation == "wrong-source": value["source"] = "different-source"
                    if mutation == "changed-pattern": value["pattern"] = "^TestOneOnly$"
                    if mutation == "changed-inventory": value["inventory"].pop()
                    if mutation == "missing-observation": value["observed"].pop()
                    if mutation == "repeated-observation": value["observed"].append(value["observed"][0])
                    target.write_text(json.dumps(value))
                with self.assertRaises(ValueError):
                    shards.verify_manifests(directory, list(shards.GROUPS), 3)

    def test_go_listing_failure_is_not_reported_as_success(self):
        failed = subprocess.CompletedProcess(["go"], 17, stdout="", stderr="synthetic compiler failure")
        with tempfile.TemporaryDirectory() as temporary, patch.object(shards.subprocess, "run", return_value=failed) as run:
            self.assertEqual(shards.run_shard("backup", 0, 3, "source", Path(temporary) / "race.jsonl"), 17)
            self.assertEqual(run.call_count, 1)
            self.assertFalse(list(Path(temporary).glob("*.manifest.json")))


if __name__ == "__main__":
    unittest.main()
