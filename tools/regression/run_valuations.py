"""Qualify exact manual values only in fresh, synthetic loopback installations."""

import argparse
import copy
import hashlib
import json
import platform
import secrets
import subprocess
from pathlib import Path

from client import Client, setup
from preservation import assert_same, backup_manifest, expected_fresh_restore
from runtime import Runtime
from qualification import read_candidate


def run(args):
    root = args.output.resolve()
    root.mkdir(mode=0o700, parents=True, exist_ok=False)
    checks = []

    def passed(name):
        checks.append(name)
        print("PASS " + name, flush=True)

    def launch(data, name):
        return Runtime(args.binary, args.backend, args.static_dir, root / data, root / name)

    def backup(client):
        return client.request("GET", "/api/v1/backup/export")

    with launch("source-data", "source-runtime") as source:
        candidate = read_candidate(source.client, args.expected_commit)
        credentials = setup(source.client, "valuation-proof-owner")
        if args.browser:
            payload = {"base": source.client.base, "credentials": credentials, "output": str(root / "browser")}
            result = subprocess.run(["node", str(Path(__file__).with_name("valuations-browser.mjs"))],
                                    input=json.dumps(payload), text=True, capture_output=True, timeout=240)
            if result.returncode:
                message = result.stderr.replace(credentials["password"], "[redacted]")
                raise AssertionError("Valuation browser checks failed: " + message[-3500:])
            checks.extend(json.loads(result.stdout)["checks"])
            passed("valuation-browser-workflows")
        vehicles = []
        for name in ("Valued vehicle", "Unvalued vehicle"):
            vehicles.append(source.client.request("POST", "/api/v1/vehicles", {
                "name": name, "manufacturer": "Synthetic", "gauge": "H0", "category": "Lokomotive", "gattung": "Diesellok",
                "listPrice": "999.00", "purchasePrice": "500.00",
            }, 201))
        path = "/api/v1/vehicles/" + vehicles[0]["id"] + "/valuations"
        entry = {"id": secrets.token_hex(16), "amount": "0.123456", "currency": "EUR", "valuedOn": "2026-01-01",
                 "source": "Manual original", "notes": "Retain exact precision and original provenance"}
        first = source.client.request("POST", path, entry, 201)
        assert_same(first, source.client.request("POST", path, entry, 201), "Identical valuation replay")
        correction = {**entry, "id": secrets.token_hex(16), "replacesId": first["id"], "amount": "150.00", "source": "Reviewed correction"}
        source.client.request("POST", path, correction, 201)
        source.client.request("POST", path, {**correction, "id": secrets.token_hex(16)}, 409)
        summary = source.client.request("GET", "/api/v1/valuations/summary?asOf=2026-01-15")
        assert_same(summary, {"asOf": "2026-01-15", "vehicleCount": 2, "valuedVehicles": 1, "missingVehicles": 1,
                              "totals": [{"currency": "EUR", "amount": "150.00", "vehicles": 1}]}, "Manual value coverage")
        history = source.client.request("GET", path + "/export")
        if len(history["entries"]) != 2 or history["entries"][1]["recordedAt"] != first["recordedAt"]:
            raise AssertionError("Correction changed original recording time")
        passed("valuation-exact-amount-replay-correction-and-missing-coverage")
        for role in ("Editor", "Viewer", "Messe"):
            user = {"username": "valuation-proof-" + role.lower(), "password": secrets.token_urlsafe(32)}
            source.client.request("POST", "/api/v1/users", {**user, "roles": [role]}, 201)
            actor = Client(int(source.client.base.rsplit(":", 1)[1]))
            actor.login(user)
            expected = 403 if role == "Messe" else 200
            actor.request("GET", path, expected=expected)
            actor.request("GET", path + "/export", expected=expected)
            actor.request("GET", "/api/v1/valuations/summary", expected=expected)
            if role == "Editor":
                actor.request("POST", path, {**entry, "id": secrets.token_hex(16)}, expected=201)
            else:
                # Permission-only probes carry no unused body, matching the main reference runner.
                actor.request("POST", path, expected=403)
            actor.request("POST", path, expected=403, csrf=False)
        passed("valuation-real-login-role-and-csrf-boundaries")
        document = backup(source.client)
        if document["version"] != 28 or len(document["tables"]["vehicle_valuations"]) != 3:
            raise AssertionError("Application backup omitted valuation rows")
        preflight = source.client.upload("/api/v1/backup/validate", "valuations.json", json.dumps(document).encode(),
                                         "application/json", expected=200)
        if not preflight["compatible"]:
            raise AssertionError("Valid valuation backup failed preflight")
        invalid = copy.deepcopy(document)
        invalid["tables"]["vehicle_valuations"][0]["currency"] = "invalid"
        source.client.upload("/api/v1/backup/restore", "invalid.json", json.dumps(invalid).encode(), "application/json", expected=400)
        assert_same(backup_manifest(document), backup_manifest(backup(source.client)), "Rejected valuation restore")
        passed("valuation-backup-and-rejected-restore-preserve-every-application-row")
        observed = source.client.request("GET", path + "/export")
        source.crash()
    with launch("source-data", "source-restart") as restarted:
        read_candidate(restarted.client, args.expected_commit)
        for cookie in source.client.cookies:
            restarted.client.cookies.set_cookie(copy.copy(cookie))
        restarted.client.token = source.client.token
        session = restarted.client.request("GET", "/api/v1/auth/session")
        if session["username"] != credentials["username"]:
            raise AssertionError("Abrupt process stop replaced the existing local session")
        restarted.client.login(credentials)
        current = restarted.client.request("GET", path + "/export")
        assert_same(observed["entries"], current["entries"], "Valuation process restart")
        assert_same(backup_manifest(document), backup_manifest(backup(restarted.client)), "Restart application rows")
        passed("valuation-process-restart-preserves-entire-history")
        passed("valuation-abrupt-stop-preserves-existing-session-and-history")
    with launch("target-data", "target-runtime") as target:
        read_candidate(target.client, args.expected_commit)
        target_credentials = setup(target.client, "valuation-proof-target")
        local = backup(target.client)
        local_users = target.client.request("GET", "/api/v1/users")
        user_ids = [user["id"] for user in local_users]
        target.client.upload("/api/v1/backup/restore", "valuations.json", json.dumps(document).encode(), "application/json", expected=200)
        expected = expected_fresh_restore(document, local, user_ids)
        assert_same(backup_manifest(expected), backup_manifest(backup(target.client)), "Fresh valuation restore")
        assert_same(observed["entries"], target.client.request("GET", path + "/export")["entries"], "Restored valuation history")
        target.client.login(target_credentials)
        if target.client.request("GET", "/api/v1/auth/session")["username"] != target_credentials["username"]:
            raise AssertionError("Restore replaced target authentication")
        passed("valuation-fresh-restore-preserves-history-and-target-authentication")
    report = {"synthetic": True, "passed": True, "candidate": candidate, "platform": platform.platform(), "backupVersion": 28, "checks": checks,
              "binarySHA256": hashlib.sha256(args.binary.read_bytes()).hexdigest(), "hardware": "not contacted"}
    (root / "report.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    repo = Path(__file__).resolve().parents[2]
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, required=True)
    parser.add_argument("--backend", type=Path, default=repo / "backend")
    parser.add_argument("--static-dir", type=Path, default=repo / "frontend/dist")
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--browser", action="store_true")
    parser.add_argument("--expected-commit", help="Require the actual clean source commit in the running process")
    args = parser.parse_args()
    for field in ("binary", "backend", "static_dir"):
        setattr(args, field, getattr(args, field).resolve(strict=True))
    run(args)
