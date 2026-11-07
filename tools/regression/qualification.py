"""Candidate identity from the local process, independent of update-feed status."""

from pathlib import Path


def read_candidate(client, expected_commit=None, expected_go_version=None):
    value = client.request("GET", "/api/v1/version")
    supplied = value.get("build")
    build = None
    if isinstance(supplied, dict):
        keys = ("commit", "shortCommit", "status", "goVersion", "platform")
        build = {key: supplied[key] for key in keys if key in supplied}
    if expected_commit:
        if not build or build.get("commit") != expected_commit or build.get("status") not in ("release", "development"):
            raise AssertionError("Running process does not match the clean qualified source commit")
        if expected_go_version is None:
            expected_go_version = "go" + (Path(__file__).resolve().parents[2] / ".go-version").read_text().strip()
        if build.get("goVersion") != expected_go_version or not build.get("platform"):
            raise AssertionError("Running process does not match the qualified compiler/platform identity")
    return {"version": value["version"], "build": build}
