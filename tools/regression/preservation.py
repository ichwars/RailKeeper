"""Read-only, content-based comparisons for isolated RailKeeper regression runs."""

import base64
import copy
import hashlib
import json
import sqlite3
from contextlib import closing
from pathlib import Path


def encoded(value):
    if isinstance(value, bytes):
        return {"sqlite_blob": base64.b64encode(value).decode("ascii")}
    return value


def digest(value):
    content = json.dumps(value, sort_keys=True, ensure_ascii=True, separators=(",", ":"))
    return hashlib.sha256(content.encode()).hexdigest()


def database_manifest(path):
    """Hash column metadata and sorted full rows, retaining NULLs and SQLite types."""
    uri = Path(path).resolve().as_uri() + "?mode=ro"
    with closing(sqlite3.connect(uri, uri=True)) as db:
        db.execute("PRAGMA query_only=ON")
        db.execute("BEGIN")
        if db.execute("PRAGMA integrity_check").fetchall() != [("ok",)]:
            raise AssertionError("SQLite integrity check failed")
        if db.execute("PRAGMA foreign_key_check").fetchall():
            raise AssertionError("SQLite foreign-key check failed")
        tables = [r[0] for r in db.execute(
            "SELECT name FROM sqlite_schema WHERE type='table' AND name NOT LIKE 'sqlite_%'"
        )]
        result = {}
        for table in sorted(tables):
            identifier = '"' + table.replace('"', '""') + '"'
            columns = db.execute(f"PRAGMA table_info({identifier})").fetchall()
            rows = [digest([encoded(v) for v in row]) for row in db.execute(f"SELECT * FROM {identifier}")]
            result[table] = {"columns": columns, "rows": sorted(rows), "count": len(rows)}
        return result


def file_manifest(root):
    root = Path(root).resolve()
    files = {}
    for path in sorted(root.rglob("*")):
        if path.is_symlink():
            raise AssertionError("Symlinks are not supported in a regression installation")
        if not path.is_file():
            continue
        relative = path.relative_to(root).as_posix()
        # SQLite contents are compared logically, including committed WAL state.
        if relative in {"railkeeper.db", "railkeeper.db-wal", "railkeeper.db-shm", "railkeeper.db-journal"}:
            continue
        files[relative] = {"size": path.stat().st_size, "sha256": hashlib.sha256(path.read_bytes()).hexdigest()}
    return files


def installation_manifest(data_dir):
    return {"tables": database_manifest(Path(data_dir) / "railkeeper.db"), "files": file_manifest(data_dir)}


def assert_same(before, after, label):
    """Fail on any unexplained difference; print names, never row contents/secrets."""
    if before == after:
        return
    keys = sorted(set(before) | set(after))
    changed = [key for key in keys if before.get(key) != after.get(key)]
    raise AssertionError(f"{label}: content differs in {', '.join(changed)}")


def backup_manifest(document):
    """Ignore only the export wall clock and ordering, not entity IDs or timestamps."""
    return {
        "format": document["format"],
        "version": document["version"],
        "tables": {name: sorted(digest(row) for row in rows) for name, rows in document["tables"].items()},
        "profiles": sorted(digest(row) for row in document.get("profiles", [])),
        "files": sorted(digest(row) for row in document.get("files", [])),
    }


def expected_backup_upgrade(document, target_version):
    """Describe the checked additive 24 -> 25 schema step; never ignore actual rows."""
    expected = copy.deepcopy(document)
    if document["version"] == target_version:
        return expected
    if (document["version"], target_version) != (24, 25):
        raise AssertionError("No checked reference rule for this backup-format upgrade")
    expected["version"] = 25
    for table in ("catalog_articles", "catalog_article_revisions", "vehicle_article_links", "vehicle_set_article_links"):
        if expected["tables"].get(table):
            raise AssertionError("Legacy backup unexpectedly contains catalog data")
        expected["tables"][table] = []
    return expected


def expected_fresh_restore(document, local_backup, local_user_ids):
    """Apply the checked schema upgrade and two existing cross-installation policies.

    BackupService retains the target's bundled workshop symbols and clears a
    transfer-profile creator if that installation does not contain that user.
    All remaining fields, including timestamps and IDs, must match exactly.
    """
    expected = expected_backup_upgrade(document, local_backup["version"])
    symbols = {}
    for row in local_backup["tables"]["master_data_entries"]:
        if (row["type"] == "symbols" and row["origin"] == "bundled"
                and json.loads(row["metadata_json"] or "{}").get("library") == "railkeeper-workshop-line"):
            symbols[(row["type"], row["key"])] = row
    expected["tables"]["master_data_entries"] = [
        row for row in expected["tables"]["master_data_entries"] if (row["type"], row["key"]) not in symbols
    ] + list(symbols.values())
    for row in expected.get("profiles", []):
        if row.get("created_by_user_id") not in local_user_ids:
            row["created_by_user_id"] = None
    return expected


def vehicle_edit_manifest(document, vehicle_id):
    """A name edit may refresh only this vehicle's and submitted images' edit clocks."""
    expected = copy.deepcopy(document)
    for table, identity in (("vehicles", "id"), ("vehicle_images", "vehicle_id")):
        for row in expected["tables"].get(table, []):
            if row[identity] == vehicle_id:
                row.pop("updated_at", None)
    return backup_manifest(expected)
