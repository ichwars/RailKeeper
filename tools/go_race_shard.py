#!/usr/bin/env python3
"""Partition existing slow race groups and verify every listed root test ran once."""

from __future__ import annotations

import argparse
from collections import Counter
import json
import os
from pathlib import Path
import re
import subprocess
import sys


# These are the unchanged package/regex selections from the original CI groups.
GROUPS = {
    "backup": (("./internal/application",), r"^Test.*Backup"),
    "decoder-cv": (
        ("./internal/application",),
        r"^Test(DecoderConfiguration|CVDraft|CVFile|CVImport|CVProvenance|"
        r"CVDatabaseWorkspace|StoredCVValidation|VehiclePersistsCV)",
    ),
}
Case = tuple[str, str]
TEST_NAME = re.compile(r"Test\w*\Z")


def case_list(value: object) -> list[Case]:
    if not isinstance(value, list):
        raise ValueError("case inventory must be an array")
    result = []
    for item in value:
        if (not isinstance(item, (list, tuple)) or len(item) != 2
                or not all(isinstance(part, str) and part for part in item)
                or item[0].startswith("-") or not TEST_NAME.fullmatch(item[1])):
            raise ValueError("invalid package/root-test identity")
        result.append((item[0], item[1]))
    return result


def listing_inventory(text: str) -> list[Case]:
    cases = []
    for line in text.splitlines():
        if not line.strip():
            continue
        event = json.loads(line)
        if not isinstance(event, dict):
            raise ValueError("Go listing event must be an object")
        if event.get("Action") != "output":
            continue
        output = event.get("Output")
        if not isinstance(output, str):
            raise ValueError("Go listing output must be text")
        for name in output.splitlines():
            if TEST_NAME.fullmatch(name):
                cases.append((event.get("Package"), name))
    cases = case_list(cases)
    if not cases or len(set(cases)) != len(cases):
        raise ValueError("empty or duplicate Go test inventory")
    return sorted(cases)


def select_shard(inventory: list[Case], index: int, count: int) -> list[Case]:
    if count < 1 or not 0 <= index < count:
        raise ValueError("shard index must be within a positive shard count")
    if len(set(inventory)) != len(inventory):
        raise ValueError("duplicate root test in inventory")
    return [case for position, case in enumerate(sorted(inventory)) if position % count == index]


def observed_roots(text: str) -> list[Case]:
    result = []
    for line in text.splitlines():
        try:
            event = json.loads(line)
        except ValueError:
            continue  # Compiler diagnostics remain in the log; process failure still fails CI.
        name = event.get("Test", "")
        if event.get("Action") == "run" and TEST_NAME.fullmatch(name):
            result.append((event.get("Package"), name))
    return case_list(result)


def run_shard(group: str, index: int, count: int, source: str, log: Path) -> int:
    packages, pattern = GROUPS[group]
    if count < 1 or not 0 <= index < count:
        raise ValueError("invalid shard configuration")
    listing = subprocess.run(
        ["go", "test", "-json", f"-list={pattern}", *packages],
        capture_output=True, text=True, check=False,
    )
    if listing.returncode:
        print(listing.stderr or listing.stdout, file=sys.stderr)
        return listing.returncode
    inventory = listing_inventory(listing.stdout)
    selected = select_shard(inventory, index, count)
    if not selected:
        raise ValueError("empty shard; review its group/count configuration")
    log.parent.mkdir(parents=True, exist_ok=True)
    result = 0
    with log.open("w", encoding="utf-8") as output:
        for package in sorted({package for package, _ in selected}):
            names = [name for owner, name in selected if owner == package]
            exact = "^(" + "|".join(re.escape(name) for name in names) + ")$"
            process = subprocess.run(
                ["go", "test", "-race", "-timeout=20m", "-count=1", "-json",
                 package, "-run", exact], stdout=output, stderr=subprocess.STDOUT, check=False,
            )
            if process.returncode:
                result = process.returncode
                break
    text = log.read_text(encoding="utf-8", errors="replace")
    observed = observed_roots(text)
    if Counter(observed) != Counter(selected):
        print("Observed root tests do not match the selected shard.", file=sys.stderr)
        result = result or 1
    manifest = {
        "version": 1, "group": group, "index": index, "count": count, "source": source,
        "packages": packages, "pattern": pattern, "inventory": inventory,
        "selected": selected, "observed": observed, "passed": result == 0,
    }
    log.with_suffix(".manifest.json").write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
    print(f"{group} shard {index + 1}/{count}: {len(selected)}/{len(inventory)} group-scoped root tests")
    if result:
        print(text[-8000:], file=sys.stderr)
    else:
        timings = []
        for line in text.splitlines():
            try:
                event = json.loads(line)
            except ValueError:
                continue
            if event.get("Action") == "pass" and TEST_NAME.fullmatch(event.get("Test", "")):
                timings.append((event.get("Elapsed", 0), event["Test"]))
        for elapsed, name in sorted(timings, reverse=True)[:3]:
            print(f"  {elapsed:.3f}s {name}")
    return result


def verify_manifests(
    directory: Path, groups: list[str], count: int, expected_source: str | None = None,
) -> dict[str, int]:
    if count < 1 or not groups or len(set(groups)) != len(groups):
        raise ValueError("invalid expected group/count configuration")
    manifests = [json.loads(path.read_text(encoding="utf-8"))
                 for path in sorted(directory.rglob("*.manifest.json"))]
    if len(manifests) != len(groups) * count:
        raise ValueError("race shard manifests are missing or duplicated")
    if any(not isinstance(item, dict) or any(type(item.get(key)) is not int
           for key in ("version", "index", "count")) for item in manifests):
        raise ValueError("invalid race manifest structure")
    if len({item.get("source") for item in manifests}) != 1 or not manifests[0].get("source"):
        raise ValueError("race manifests do not share one source revision")
    if expected_source and manifests[0]["source"] != expected_source:
        raise ValueError("race manifests do not match the required source revision")
    if {item.get("group") for item in manifests} != set(groups):
        raise ValueError("unexpected or missing race group")
    totals = {}
    for group in groups:
        members = [item for item in manifests if item.get("group") == group]
        if len(members) != count or {item.get("index") for item in members} != set(range(count)):
            raise ValueError(f"{group}: missing or duplicate shard index")
        packages, pattern = GROUPS[group]
        inventory = sorted(case_list(members[0].get("inventory")))
        if not inventory or len(set(inventory)) != len(inventory):
            raise ValueError(f"{group}: empty or duplicate inventory")
        executed = []
        for item in members:
            if (item.get("version") != 1 or item.get("count") != count or item.get("passed") is not True
                    or tuple(item.get("packages", [])) != packages or item.get("pattern") != pattern
                    or sorted(case_list(item.get("inventory"))) != inventory):
                raise ValueError(f"{group}: failed, inconsistent or changed group selection")
            selected = case_list(item.get("selected"))
            observed = case_list(item.get("observed"))
            expected = select_shard(inventory, item["index"], count)
            if Counter(selected) != Counter(expected) or Counter(observed) != Counter(selected):
                raise ValueError(f"{group}: selected/observed root-test mismatch")
            executed.extend(observed)
        if Counter(executed) != Counter(inventory):
            raise ValueError(f"{group}: missing or repeated root-test execution")
        totals[group] = len(inventory)
    return totals


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)
    run = commands.add_parser("run")
    run.add_argument("--group", choices=GROUPS, required=True)
    run.add_argument("--index", type=int, required=True)
    run.add_argument("--count", type=int, default=3)
    run.add_argument("--source", default=os.environ.get("GITHUB_SHA", "local"))
    run.add_argument("--log", type=Path, required=True)
    verify = commands.add_parser("verify")
    verify.add_argument("--directory", type=Path, required=True)
    verify.add_argument("--groups", nargs="+", choices=GROUPS, default=list(GROUPS))
    verify.add_argument("--count", type=int, default=3)
    verify.add_argument("--source", default=os.environ.get("GITHUB_SHA"))
    args = parser.parse_args()
    try:
        if args.command == "run":
            return run_shard(args.group, args.index, args.count, args.source, args.log)
        totals = verify_manifests(args.directory, args.groups, args.count, args.source)
        print("Verified complete race selections: " + ", ".join(f"{group}={size}" for group, size in totals.items()))
        return 0
    except (OSError, ValueError, TypeError) as error:
        print(f"Race shard verification failed: {error}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
