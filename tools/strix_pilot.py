"""Prepare disposable RailKeeper accounts and run a bounded Strix pilot (stdlib only)."""

import argparse
import http.cookiejar
import json
import math
import os
from pathlib import Path
import secrets
import signal
import subprocess
import sys
import urllib.parse
import urllib.request

ROOT = Path(__file__).resolve().parents[1]


def validate_url(value):
    url = urllib.parse.urlsplit(value)
    if (url.scheme not in {"http", "https"} or not url.hostname or url.username or url.password
            or url.path not in {"", "/"} or url.query or url.fragment):
        raise ValueError("Target must be an HTTP(S) origin without credentials, path, query or fragment")
    _ = url.port
    return value.rstrip("/")


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        raise ValueError("Redirect refused: test credentials must stay on the configured origin")


def seed(url, output):
    url = validate_url(url)
    output = Path(output)
    if output.exists():
        raise ValueError("Accounts file already exists; use a fresh disposable instance and file")
    client = urllib.request.build_opener(urllib.request.ProxyHandler({}), NoRedirect(),
                                        urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))

    def request(path, data=None, csrf=None):
        headers = {"Content-Type": "application/json"}
        if csrf:
            headers["X-CSRF-Token"] = csrf
        body = json.dumps(data).encode() if data is not None else None
        with client.open(urllib.request.Request(url + "/api/v1" + path, body, headers),
                         timeout=15) as response:
            return {} if response.status == 204 else json.load(response)

    if request("/setup/status").get("setupRequired") is not True:
        raise ValueError("Refusing to seed an existing installation: first-run setup is already complete")
    accounts = {role: {"username": "pilot-" + role, "email": role + "@pilot.invalid",
                       "password": secrets.token_urlsafe(24), "roles": [role.title()]}
                for role in ("admin", "editor", "viewer", "messe")}
    request("/setup/admin", accounts["admin"])
    session = request("/auth/login", accounts["admin"])
    for role in ("editor", "viewer", "messe"):
        request("/users", accounts[role], session["csrfToken"])
    request("/auth/logout", {}, session["csrfToken"])
    output.parent.mkdir(parents=True, exist_ok=True)
    with open(output, "x", encoding="utf-8", opener=lambda p, flags: os.open(p, flags, 0o600)) as file:
        json.dump(accounts, file)
    print("Four disposable role accounts created; credentials saved privately.")


def assess_run(directory, exit_code):
    records = list(Path(directory).glob("strix_runs/*/run.json"))
    if exit_code not in {0, 2} or len(records) != 1:
        return 3, "incomplete", 0
    try:
        record = json.loads(records[0].read_text())
        coverage = json.loads(records[0].with_name("coverage.json").read_text())
        if coverage.get("completeness", {}).get("complete") is not True:
            return 3, "incomplete", 0
        sarif_file = records[0].with_name("findings.sarif")
        runs = json.loads(sarif_file.read_text())["runs"] if sarif_file.exists() else None
        if runs is not None:
            if not isinstance(runs, list) or len(runs) != 1:
                return 3, "incomplete", 0
            if any(invocation.get("executionSuccessful") is False
                   for invocation in runs[0].get("invocations", [])):
                return 3, "incomplete", 0
        findings_file = records[0].with_name("vulnerabilities.json")
        if findings_file.exists():
            findings = json.loads(findings_file.read_text())
        else:
            # v1.7.0 omits the JSON index on clean scans; SARIF may still contain coverage entries.
            if not isinstance(runs, list) or len(runs) != 1:
                return 3, "incomplete", 0
            results = runs[0].get("results")
            if not isinstance(results, list) or any(
                    not isinstance(result, dict) or result.get("kind", "fail") not in {"pass", "notApplicable", "open"}
                    for result in results):
                return 3, "incomplete", 0
            findings = []
        if not isinstance(record, dict) or record.get("status") != "completed" or not isinstance(findings, list):
            return 3, "incomplete", 0
    except (OSError, ValueError, KeyError, TypeError, AttributeError):
        return 3, "incomplete", 0
    return (2 if findings or exit_code == 2 else 0), "completed", len(findings)


def scan(args):
    url = validate_url(args.url)
    if not math.isfinite(args.budget) or not 0 < args.budget <= 50:
        raise ValueError("Pilot budget must be greater than 0 and at most USD 50")
    if not 1 <= args.turns <= 300 or not 1 <= args.minutes <= 120:
        raise ValueError("Pilot requires 1-300 turns per agent and 1-120 minutes")
    if not os.environ.get("STRIX_LLM") or not os.environ.get("LLM_API_KEY"):
        raise ValueError("Set STRIX_LLM and LLM_API_KEY before starting the pilot")
    output = Path(args.output).resolve()
    if output.exists():
        raise ValueError("Scan output must be a new directory")
    source = Path(args.source).resolve()
    output.mkdir(parents=True, mode=0o700)
    snapshot = output / "source"
    snapshot.mkdir()
    # Only committed files enter the sandbox, never .git, .env, local data or work in progress.
    with subprocess.Popen(["git", "-C", str(source), "archive", "HEAD"], stdout=subprocess.PIPE) as archive:
        subprocess.run(["tar", "-x", "-C", str(snapshot)], stdin=archive.stdout, check=True)
        archive.stdout.close()
        if archive.wait() != 0:
            raise ValueError("Could not create the disposable source snapshot")
    accounts = json.loads(Path(args.accounts).read_text())
    instructions = (ROOT / "deploy/strix-pilot/instructions.md").read_text()
    instructions += "\n\nThe only authorized HTTP origin is " + url + ".\n"
    instructions += "Disposable test accounts (not production credentials):\n" + json.dumps(accounts)
    instructions_path = output / "instructions.private.md"
    instructions_path.write_text(instructions)
    instructions_path.chmod(0o600)
    mcp_config = output / "mcp-empty.json"
    mcp_config.write_text("[]\n")
    environment = dict(os.environ, STRIX_TELEMETRY="false")
    command = [args.strix, "-n", "-t", str(snapshot), "-t", url, "--scope-mode", "full",
               "--scan-mode", args.mode, "--max-budget", str(args.budget), "--max-turns", str(args.turns),
               "--instruction-file", str(instructions_path), "--mcp-config", str(mcp_config)]
    exit_code = 1
    try:
        with (output / "scan.private.log").open("w") as log:
            with subprocess.Popen(command, cwd=output, env=environment, stdout=log,
                                  stderr=subprocess.STDOUT, start_new_session=True) as process:
                try:
                    exit_code = process.wait(timeout=args.minutes * 60)
                except subprocess.TimeoutExpired:
                    os.killpg(process.pid, signal.SIGINT)
                    try:
                        process.wait(timeout=15)
                    except subprocess.TimeoutExpired:
                        os.killpg(process.pid, signal.SIGKILL)
                        process.wait()
                    print("Pilot time limit reached; assessment is incomplete.")
    finally:
        instructions_path.unlink(missing_ok=True)
    code, status, count = assess_run(output, exit_code)
    # Do not expose model-controlled finding titles, exploits or credentials in public Actions logs.
    count_label = str(count) if status == "completed" else "not evaluated"
    summary = f"Strix pilot: {status}; recorded findings: {count_label}; result code: {code}.\n"
    (output / "summary.txt").write_text(summary)
    print(summary, end="")
    if os.environ.get("GITHUB_STEP_SUMMARY"):
        with open(os.environ["GITHUB_STEP_SUMMARY"], "a") as file:
            file.write(summary)
    return code


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)
    prepare = commands.add_parser("seed", help="Create accounts on a fresh disposable instance only")
    prepare.add_argument("--url", required=True)
    prepare.add_argument("--output", required=True)
    run = commands.add_parser("scan", help="Scan a committed source snapshot and one authorized origin")
    run.add_argument("--url", required=True)
    run.add_argument("--source", default=str(ROOT))
    run.add_argument("--accounts", required=True)
    run.add_argument("--output", required=True)
    run.add_argument("--strix", default="strix")
    run.add_argument("--mode", choices=("quick", "standard", "deep"), default="quick")
    run.add_argument("--budget", type=float, default=10)
    run.add_argument("--turns", type=int, default=80)
    run.add_argument("--minutes", type=int, default=45)
    args = parser.parse_args()
    try:
        if args.command == "seed":
            seed(args.url, args.output)
            return 0
        return scan(args)
    except (OSError, ValueError, subprocess.SubprocessError) as error:
        # HTTP exceptions may contain target-controlled response text; keep the public error fixed.
        print(f"Pilot setup/execution failed ({type(error).__name__}); inspect private local output.", file=sys.stderr)
        return 3


if __name__ == "__main__":
    sys.exit(main())
