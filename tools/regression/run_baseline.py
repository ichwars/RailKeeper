"""Exercise a synthetic collection, app restore and a cold installation restore.

No existing data directory, external application URL, or production backup is accepted.
Only the Python standard library is required; --browser also uses the pinned Node tools.
"""

import argparse
import hashlib
import json
import os
import platform
import secrets
import shutil
import subprocess
from pathlib import Path

from client import Client, setup
from fixture import observations, seed
from matrix import inventory
from preservation import (
    assert_same, backup_manifest, digest, expected_backup_upgrade, expected_fresh_restore, installation_manifest, vehicle_edit_manifest,
)
from runtime import Runtime


def role_checks(client, fixture):
    for role in ("Editor", "Viewer", "Messe"):
        credentials = {"username": "reference-" + role.lower(), "password": secrets.token_urlsafe(32)}
        client.request("POST", "/api/v1/users", {**credentials, "roles": [role]}, 201)
        actor = Client(int(client.base.rsplit(":", 1)[1]))
        actor.login(credentials)
        actor.request("GET", "/api/v1/vehicles", expected=403 if role == "Messe" else 200)
        actor.request("GET", "/api/v1/users", expected=403)
        actor.request("POST", "/api/v1/vehicles", {}, expected=403 if role != "Editor" else 400)
        actor.request("GET", "/api/v1/exhibition-lists", expected=200 if role == "Messe" else 403)
    client.request("POST", "/api/v1/vehicles", {}, expected=403, csrf=False)
    Client(int(client.base.rsplit(":", 1)[1])).request("GET", "/api/v1/vehicles", expected=401)


def verify_media(client, fixture):
    prefix = "/api/v1/vehicles/" + fixture["vehicleId"]
    image = client.request("GET", prefix + "/images/" + fixture["imageId"] + "/file")
    if hashlib.sha256(image).hexdigest() != fixture["imageSHA256"]:
        raise AssertionError("Vehicle image content changed")
    text = client.request("GET", prefix + "/attachments/" + fixture["attachmentId"] + "/download")
    if text != b"Synthetic manual: no real manufacturer data.\n":
        raise AssertionError("Attachment content changed")


def restore(client, document, expected=200):
    return client.upload("/api/v1/backup/restore", "reference.json", json.dumps(document).encode(),
                         "application/json", expected=expected)


def verify_artifact(client, fixture):
    content = client.request("GET", "/api/v1/data-transfer/artifacts/" + fixture["artifactId"] + "/download", raw=True)
    if hashlib.sha256(content).hexdigest() != fixture["artifactSHA256"]:
        raise AssertionError("Stored transfer artifact content changed")


def run(args):
    root = args.output.resolve()
    root.mkdir(mode=0o700, parents=True, exist_ok=False)
    (root / ".railkeeper-regression").write_text("synthetic-only\n")
    matrix = inventory()
    report = {"synthetic": True, "platform": platform.platform(), "checks": [], "hardware": "not contacted",
              "routeCount": matrix["routeCount"], "routesSHA256": matrix["routesSHA256"]}
    previous = args.previous_binary or args.binary
    previous_backend = args.previous_backend or args.backend
    report["mode"] = "upgrade" if previous.read_bytes() != args.binary.read_bytes() else "restart"
    report["targetBinarySHA256"] = hashlib.sha256(args.binary.read_bytes()).hexdigest()
    report["sourceBinarySHA256"] = hashlib.sha256(previous.read_bytes()).hexdigest()
    if args.image:
        security = subprocess.check_output(["docker", "info", "--format", "{{json .SecurityOptions}}"], text=True)
        if "name=rootless" not in json.loads(security):
            raise AssertionError("Bind-mounted Docker regression requires a rootless test daemon")
        report["targetImage"] = subprocess.check_output(["docker", "image", "inspect", args.image,
                                                         "--format", "{{.Id}}"], text=True).strip()
        del report["targetBinarySHA256"]
        if not args.previous_binary:
            del report["sourceBinarySHA256"]

    def passed(name):
        report["checks"].append(name)
        print("PASS " + name, flush=True)

    source_data = root / "source-data"
    configuration = {"RAILKEEPER_MAX_IMAGE_MB": "2", "RAILKEEPER_MAX_ATTACHMENT_MB": "2"}
    config_path = root / "source-launcher-config.json"
    config_path.write_text(json.dumps(configuration, sort_keys=True) + "\n", encoding="utf-8")
    if os.name == "nt":
        default_data = root / "isolated-localappdata" / "RailKeeper" / "data"
        with Runtime(args.binary, args.backend, args.static_dir, default_data, root / "windows-default-runtime",
                     windows_default=True) as standalone:
            if not standalone.client.request("GET", "/api/v1/setup/status")["setupRequired"]:
                raise AssertionError("Windows default path did not create a fresh isolated installation")
            setup(standalone.client, "standalone-owner")
            if not (default_data / "railkeeper.db").is_file():
                raise AssertionError("Windows standalone did not use the isolated default data directory")
        passed("native-windows-standalone-default-data-path")
    with Runtime(previous, previous_backend, args.static_dir, source_data, root / "source-runtime",
                 image=args.image if not args.previous_binary else None, configuration=configuration) as source:
        client = source.client
        credentials = setup(client, "reference-admin")
        fixture = seed(client)
        role_checks(client, fixture)
        passed("reference-fixture-and-role-csrf-boundaries")
        settings = client.request("PUT", "/api/v1/profile/settings",
                                  {"settings": {"railkeeper.settings.defaultView": "vehicles"}})
        if settings["settings"].get("railkeeper.settings.defaultView") != "vehicles":
            raise AssertionError("Reference user preference was not persisted")
        backup = client.request("GET", "/api/v1/backup/export")
        baseline = backup_manifest(backup)
        observed = observations(client, fixture)
        verify_media(client, fixture)
        verify_artifact(client, fixture)
        validation = client.upload("/api/v1/backup/validate", "reference.json", json.dumps(backup).encode(),
                                   "application/json", expected=200)
        if not validation["compatible"]:
            raise AssertionError("Reference backup is incompatible")
        passed("reference-export-preflight-and-media")
        future = {**backup, "version": backup["version"] + 100}
        future_validation = client.upload("/api/v1/backup/validate", "future.json", json.dumps(future).encode(),
                                          "application/json", expected=200)
        if future_validation["compatible"]:
            raise AssertionError("Future backup was accepted")
        restore(client, future, expected=400)
        assert_same(baseline, backup_manifest(client.request("GET", "/api/v1/backup/export")), "Rejected restore")
        passed("incompatible-restore-does-not-change-application-data")
        marker = client.request("POST", "/api/v1/vehicles", {
            "manufacturer": "Referenz", "name": "NACH DER SICHERUNG", "gauge": "H0",
            "category": "Wagen", "gattung": "Reisezugwagen"}, 201)
        restore(client, backup)
        client.request("GET", "/api/v1/vehicles/" + marker["id"], expected=404)
        restored_backup = client.request("GET", "/api/v1/backup/export")
        if args.catalog_backfill and (previous_backend / "migrations/0074_vehicle_catalog_legacy.sql").is_file():
            assert_same(backup_manifest(expected_backup_upgrade(backup, restored_backup["version"], restored_backup)),
                        backup_manifest(restored_backup), "In-place restore with legacy catalog backfill")
            backup, baseline = restored_backup, backup_manifest(restored_backup)
        else:
            assert_same(baseline, backup_manifest(restored_backup), "In-place restore")
        assert_same(observed, observations(client, fixture), "In-place API observations")
        passed("app-restore-replaces-post-backup-marker-and-preserves-reference")

        with Runtime(args.binary, args.backend, args.static_dir, root / "app-restored-data",
                     root / "app-restored-runtime", image=args.image, configuration=configuration) as target:
            recovery = setup(target.client, "restore-owner")
            settings = target.client.request("PUT", "/api/v1/profile/settings",
                                             {"settings": {"railkeeper.settings.defaultView": "overview"}})
            if settings["settings"].get("railkeeper.settings.defaultView") != "overview":
                raise AssertionError("Target-local user preference was not persisted")
            before = installation_manifest(target.data_dir)["tables"]
            local_backup = target.client.request("GET", "/api/v1/backup/export")
            local_users = {user["id"] for user in target.client.request("GET", "/api/v1/users")}
            restore(target.client, backup)
            restored_backup = target.client.request("GET", "/api/v1/backup/export")
            expected = backup_manifest(expected_fresh_restore(backup, local_backup, local_users,
                restored_backup if args.catalog_backfill else None))
            after = installation_manifest(target.data_dir)["tables"]
            for table in ("users", "roles", "user_roles", "app_settings", "user_settings"):
                assert_same(before[table], after[table], "Local installation " + table)
            assert_same(expected, backup_manifest(restored_backup), "Fresh app restore")
            assert_same(observed, observations(target.client, fixture), "Fresh restore API observations")
            verify_media(target.client, fixture)
            target.client.login(recovery)
            users = target.client.request("GET", "/api/v1/users")
            if [user["username"] for user in users] != ["restore-owner"]:
                raise AssertionError("App restore imported installation users")
            passed("fresh-app-restore-preserves-target-login-and-all-reference-data")

            if args.browser:
                before_browser = target.client.request("GET", "/api/v1/backup/export")
                result = subprocess.run(["node", str(Path(__file__).with_name("browser.mjs"))],
                                        input=json.dumps({"base": target.client.base, "credentials": recovery,
                                                          "fixture": fixture, "output": str(root / "browser")}),
                                        text=True, capture_output=True, timeout=300)
                if result.returncode:
                    raise AssertionError("Target browser baseline failed: " + result.stderr[-2500:])
                passed("browser-login-navigation-edit-cancel-and-save")
                after_browser = target.client.request("GET", "/api/v1/backup/export")
                for name, document in (("before", before_browser), ("after", after_browser)):
                    (root / f"browser-{name}-export.json").write_text(json.dumps(document), encoding="utf-8")
                before_edit = vehicle_edit_manifest(before_browser, fixture["vehicleId"])
                after_edit = vehicle_edit_manifest(after_browser, fixture["vehicleId"])
                assert_same(before_edit["tables"], after_edit["tables"], "Unrelated table data after browser edit")
                assert_same({key: value for key, value in before_edit.items() if key != "tables"},
                            {key: value for key, value in after_edit.items() if key != "tables"},
                            "Unrelated backup data after browser edit")
                passed("browser-edit-preserves-all-other-reference-data")

            if args.browser:
                result = subprocess.run(["node", str(Path(__file__).with_name("editor-boundaries.mjs"))],
                                        input=json.dumps({"base": target.client.base, "credentials": recovery,
                                                          "fixture": fixture, "output": str(root / "editor-browser")}),
                                        text=True, capture_output=True, timeout=180)
                if result.returncode:
                    raise AssertionError("Editor save boundaries failed: " + result.stderr[-2500:])
                passed("target-browser-confirmed-action-preserves-drafts-and-cancel")

    # The source process has exited. Its main DB and WAL plus every stored file form a cold backup.
    original = installation_manifest(source_data)
    cold_backup = root / "installation-backup"
    shutil.copytree(source_data, cold_backup)
    backup_config = root / "installation-backup-launcher-config.json"
    shutil.copy2(config_path, backup_config)
    assert_same(original, installation_manifest(cold_backup), "Cold installation backup")
    restored_data = root / "installation-restored-data"
    shutil.copytree(cold_backup, restored_data)
    restored_config = root / "installation-restored-launcher-config.json"
    shutil.copy2(backup_config, restored_config)
    if config_path.read_bytes() != restored_config.read_bytes():
        raise AssertionError("External launcher configuration changed")
    assert_same(original, installation_manifest(restored_data), "Full installation restore")
    passed("cold-installation-backup-and-restore-all-tables-and-files")

    with Runtime(args.binary, args.backend, args.static_dir, restored_data, root / "installation-runtime",
                 image=args.image, configuration=json.loads(restored_config.read_text(encoding="utf-8"))) as target:
        startup_tables = installation_manifest(restored_data)["tables"]
        for name in ("users", "roles", "user_roles", "app_settings", "user_settings",
                     "data_transfer_jobs", "data_transfer_artifacts"):
            assert_same(original["tables"][name], startup_tables[name], "Installation startup " + name)
        target.client.login(credentials)
        upgraded = target.client.request("GET", "/api/v1/backup/export")
        assert_same(backup_manifest(expected_backup_upgrade(backup, upgraded["version"], upgraded if args.catalog_backfill else None)),
                    backup_manifest(upgraded), "Installation startup")
        assert_same(observed, observations(target.client, fixture), "Installation API observations")
        verify_media(target.client, fixture)
        verify_artifact(target.client, fixture)
        passed("restored-installation-startup-login-and-reference-workflows")
    report["backupVersion"] = backup["version"]
    report["freshRestorePolicies"] = ["keep target bundled workshop symbols", "clear missing local profile creator"]
    report["tables"] = {name: {"count": item["count"], "sha256": digest(item)}
                        for name, item in original["tables"].items()}
    report["externalFiles"] = original["files"]
    report["launcherConfigurationSHA256"] = hashlib.sha256(restored_config.read_bytes()).hexdigest()
    report["fixture"] = fixture
    report["passed"] = True
    (root / "report.json").write_text(json.dumps(report, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    # Synthetic only, ignored runtime artifact. Useful as input for the next preservation comparison.
    (root / "reference-backup.json").write_text(json.dumps(backup, ensure_ascii=False) + "\n", encoding="utf-8")
    print("Report: " + str(root / "report.json"))


if __name__ == "__main__":
    repo = Path(__file__).resolve().parents[2]
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, required=True)
    parser.add_argument("--backend", type=Path, default=repo / "backend")
    parser.add_argument("--static-dir", type=Path, default=repo / "frontend" / "dist")
    parser.add_argument("--previous-binary", type=Path)
    parser.add_argument("--previous-backend", type=Path)
    parser.add_argument("--output", type=Path, required=True, help="New directory; existing paths are refused")
    parser.add_argument("--browser", action="store_true")
    parser.add_argument("--catalog-backfill", action="store_true", help="Verify the separately checked legacy catalog backfill")
    parser.add_argument("--image", help="Optional existing image on a rootless local Docker test daemon")
    arguments = parser.parse_args()
    for field in ("binary", "backend", "static_dir", "previous_binary", "previous_backend"):
        value = getattr(arguments, field)
        if value is not None:
            setattr(arguments, field, value.resolve(strict=True))
    run(arguments)
