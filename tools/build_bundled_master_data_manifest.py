#!/usr/bin/env python3
import argparse
import json
import sqlite3
from pathlib import Path


def build_manifest(migrations_dir: Path, seed_path: Path) -> dict[str, object]:
    connection = sqlite3.connect(":memory:")
    try:
        for migration in sorted(migrations_dir.glob("*.sql")):
            connection.executescript(migration.read_text(encoding="utf-8"))
        seed = json.loads(seed_path.read_text(encoding="utf-8"))
        for entry in seed["entries"]:
            connection.execute(
                """
                INSERT OR IGNORE INTO master_data_entries(
                  id, type, key, label, active, sort_order, source_url,
                  metadata_json, created_at, updated_at, origin
                ) VALUES(?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'), 'bundled')
                """,
                (
                    entry["id"],
                    entry["type"],
                    entry["key"],
                    entry["label"],
                    int(entry.get("active", True)),
                    entry.get("sortOrder", 0),
                    entry.get("sourceUrl", ""),
                    json.dumps(
                        entry.get("metadata", {}),
                        ensure_ascii=False,
                        separators=(",", ":"),
                    ),
                ),
            )
        rows = connection.execute(
            "SELECT DISTINCT type, key FROM master_data_entries ORDER BY type, key"
        ).fetchall()
        return {
            "version": 1,
            "entries": [
                {"type": type_name, "key": key} for type_name, key in rows
            ],
        }
    finally:
        connection.close()


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--migrations", type=Path, required=True)
    parser.add_argument("--seed", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    document = build_manifest(args.migrations, args.seed)
    args.output.write_text(
        json.dumps(document, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
        newline="\n",
    )


if __name__ == "__main__":
    main()
