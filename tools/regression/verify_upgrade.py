"""Rehearse an additive migration on a marked synthetic cold-backup copy only."""

import argparse
import hashlib
import json
import shutil
import sqlite3
from contextlib import closing
from pathlib import Path

from preservation import assert_same, database_manifest, digest, installation_manifest
from runtime import Runtime
from catalog import TABLES, expected_catalog_backfill


CATALOG_TABLES = {"catalog_articles", "catalog_article_revisions", "vehicle_article_links", "vehicle_set_article_links"}


def run(args):
    reference = args.reference.resolve(strict=True)
    if (reference / ".railkeeper-regression").read_text() != "synthetic-only\n":
        raise ValueError("Only a marked synthetic reference run is accepted")
    source = reference / "installation-backup"
    output = args.output.resolve()
    if source.is_symlink() or output.is_relative_to(reference):
        raise ValueError("The backup must be a directory and the new probe must be outside its reference run")
    before = installation_manifest(source)
    output.mkdir(mode=0o700, parents=True, exist_ok=False)
    (output / ".railkeeper-regression").write_text("synthetic-only\n")
    target = output / "data"
    shutil.copytree(source, target)
    assert_same(before, installation_manifest(target), "Cold probe copy")
    configuration = json.loads((reference / "installation-backup-launcher-config.json").read_text())
    with Runtime(args.binary, args.backend, args.static_dir, target, output / "runtime", configuration=configuration):
        after = installation_manifest(target)
        for name, original in before["tables"].items():
            if name == "schema_migrations":
                current = after["tables"][name]
                assert_same({"columns": original["columns"]}, {"columns": current["columns"]}, "Migration ledger columns")
                if not set(original["rows"]).issubset(current["rows"]) or current["count"] <= original["count"]:
                    raise AssertionError("Migration ledger did not retain all existing entries")
            elif not args.catalog_backfill or name not in CATALOG_TABLES:
                assert_same(original, after["tables"].get(name, {}), "Migration preserved table " + name)
            else:
                assert_same({"columns": original["columns"]}, {"columns": after["tables"][name]["columns"]}, "Catalog schema columns")
        added = set(after["tables"]) - set(before["tables"])
        if args.catalog_backfill:
            if added != CATALOG_TABLES - set(before["tables"]):
                raise AssertionError("Unexpected table added by catalog backfill")
            document = json.loads((reference / "reference-backup.json").read_text())
            with closing(sqlite3.connect((target / "railkeeper.db").as_uri() + "?mode=ro", uri=True)) as db:
                db.row_factory = sqlite3.Row
                actual = {"tables": {table: [dict(row) for row in db.execute('SELECT * FROM "' + table + '"')] for table in TABLES}}
            expected = expected_catalog_backfill(document, actual)
            for table in TABLES:
                assert_same({"rows": sorted(digest(row) for row in expected["tables"][table])},
                            {"rows": sorted(digest(row) for row in actual["tables"][table])}, "Complete expected catalog backfill " + table)
        elif added != CATALOG_TABLES or any(after["tables"][name]["count"] for name in added):
            raise AssertionError("Catalog schema must add exactly four empty tables")
        for name, value in before["files"].items():
            assert_same(value, after["files"].get(name, {}), "Migration preserved file " + name)
        new_files = set(after["files"]) - set(before["files"])
        if len(new_files) != 1:
            raise AssertionError("Migration must create exactly one safety snapshot")
        for name in new_files:
            if not name.startswith("safety-backups/railkeeper-pre-migration-") or not name.endswith(".db"):
                raise AssertionError("Unexpected file created during migration")
            assert_same(before["tables"], database_manifest(target / name), "Automatic pre-migration snapshot")
    report = {"passed": True, "synthetic": True, "hardware": "not contacted",
              "targetBinarySHA256": hashlib.sha256(args.binary.read_bytes()).hexdigest(),
              "preservedTables": {name: {"count": item["count"], "sha256": digest(item)}
                                  for name, item in before["tables"].items() if name != "schema_migrations"
                                  and (not args.catalog_backfill or name not in CATALOG_TABLES)},
              "preservedFiles": before["files"], "addedTables": sorted(added),
              "catalogTables": {name: {"count": after["tables"][name]["count"], "sha256": digest(after["tables"][name])}
                                for name in sorted(CATALOG_TABLES)},
              "validatedSafetySnapshots": sorted(new_files), "catalogBackfillVerified": args.catalog_backfill}
    (output / "report.json").write_text(json.dumps(report, indent=2) + "\n")
    print(f"PASS migration preserved all {len(report['preservedTables'])} legacy tables, files and safety snapshot")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--reference", type=Path, required=True)
    parser.add_argument("--binary", type=Path, required=True)
    parser.add_argument("--backend", type=Path, required=True)
    parser.add_argument("--static-dir", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--catalog-backfill", action="store_true")
    arguments = parser.parse_args()
    for field in ("binary", "backend", "static_dir"):
        setattr(arguments, field, getattr(arguments, field).resolve(strict=True))
    run(arguments)
