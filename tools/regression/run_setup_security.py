"""Qualify setup and login only in a new synthetic installation, including process restarts."""

import argparse
import json
import secrets
import subprocess
from pathlib import Path

from preservation import assert_same, database_manifest
from runtime import Runtime


def run(args):
    root = args.output.resolve()
    root.mkdir(mode=0o700, parents=True, exist_ok=False)
    data = root / "data"
    checks = []
    credentials = {"username": "setup-owner", "password": secrets.token_urlsafe(32)}

    def runtime(name):
        return Runtime(args.binary, args.backend, args.static_dir, data, root / name)

    def browser(server, **extra):
        payload = {"base": server.client.base, "credentials": credentials, "output": str(root / "browser"), **extra}
        completed = subprocess.run(["node", str(Path(__file__).with_name("setup-security.mjs"))],
                                   input=json.dumps(payload), text=True, capture_output=True, timeout=180)
        if completed.returncode:
            raise AssertionError("Setup browser checks failed: " + completed.stderr[-2000:])
        result = json.loads(completed.stdout)
        checks.extend(result["checks"])
        return result

    with runtime("first-start") as server:
        code = server.client.setup_code
        if len(code) != 43:
            raise AssertionError("Fresh installation did not provision its local setup code")
        status = server.client.request("GET", "/api/v1/setup/status")
        if status != {"setupRequired": True, "cookieSecure": False}:
            raise AssertionError("Public status is unexpected or leaks setup information")
        server.client.request("POST", "/api/v1/setup/admin", {**credentials, "email": "setup-owner@example.test"}, 403)
    with runtime("pending-restart") as server:
        if server.client.setup_code != code:
            raise AssertionError("Restart changed the pending setup code")
        checks.append("pending-code-survives-restart-and-missing-code-is-rejected")
        session = browser(server, mode="setup", setupCode=code)["sessionState"]
    auth_tables = ("users", "roles", "user_roles", "sessions", "setup_authorizations")
    manifest = database_manifest(data / "railkeeper.db")
    before = {name: manifest[name] for name in auth_tables}
    sentinel = "Preserve this invalid file: existing users must skip code provisioning.\n"
    (data / "setup-code.txt").write_text(sentinel, encoding="utf-8")
    with runtime("existing-restart") as server:
        if server.client.request("GET", "/api/v1/setup/status")["setupRequired"]:
            raise AssertionError("Existing administrator was lost")
        if (data / "setup-code.txt").read_text(encoding="utf-8") != sentinel:
            raise AssertionError("Existing account startup changed the local file")
        manifest = database_manifest(data / "railkeeper.db")
        after = {name: manifest[name] for name in auth_tables}
        assert_same(before, after, "Existing authentication after process restart")
        browser(server, mode="existing", sessionState=session)
    if any(code in log.read_text(encoding="utf-8") for log in root.glob("*/server.log")):
        raise AssertionError("Startup logs leaked the local setup code")
    checks.append("code-excluded-from-status-and-logs-existing-authentication-unchanged")
    (root / "report.json").write_text(json.dumps({"checks": checks}, indent=2) + "\n", encoding="utf-8")
    for check in checks:
        print("PASS " + check)


if __name__ == "__main__":
    repo = Path(__file__).resolve().parents[2]
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, required=True)
    parser.add_argument("--backend", type=Path, default=repo / "backend")
    parser.add_argument("--static-dir", type=Path, default=repo / "frontend" / "dist")
    parser.add_argument("--output", type=Path, required=True)
    run(parser.parse_args())
