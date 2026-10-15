import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

import strix_pilot


class ScanResultTests(unittest.TestCase):
    def assess(self, status="completed", exit_code=0, findings=None):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            run = root / "strix_runs" / "pilot"
            run.mkdir(parents=True)
            (run / "run.json").write_text(json.dumps({"status": status}))
            (run / "coverage.json").write_text('{"completeness": {"complete": true}}')
            (run / "vulnerabilities.json").write_text(json.dumps(findings or []))
            return strix_pilot.assess_run(root, exit_code)

    def test_completed_without_findings_is_success(self):
        self.assertEqual(self.assess(), (0, "completed", 0))

    def test_findings_are_reported_without_their_contents(self):
        result = self.assess(exit_code=2, findings=[{"title": "sensitive", "severity": "high"}])
        self.assertEqual(result, (2, "completed", 1))

    def test_budget_stop_is_incomplete_even_with_success_exit(self):
        self.assertEqual(self.assess(status="stopped"), (3, "incomplete", 0))

    def test_running_interrupted_failed_and_unknown_are_incomplete(self):
        for status in ("running", "interrupted", "failed", "unknown"):
            with self.subTest(status=status):
                self.assertEqual(self.assess(status=status)[0], 3)

    def test_process_error_cannot_be_overridden_by_completed_record(self):
        self.assertEqual(self.assess(exit_code=1), (3, "incomplete", 0))

    def test_findings_do_not_depend_on_process_exit_code(self):
        self.assertEqual(self.assess(findings=[{"severity": "low"}]), (2, "completed", 1))

    def test_missing_multiple_or_malformed_records_are_incomplete(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.assertEqual(strix_pilot.assess_run(root, 0), (3, "incomplete", 0))
            for name in ("a", "b"):
                run = root / "strix_runs" / name
                run.mkdir(parents=True)
                (run / "run.json").write_text("{}")
            self.assertEqual(strix_pilot.assess_run(root, 0), (3, "incomplete", 0))
            (root / "strix_runs/b/run.json").unlink()
            (root / "strix_runs/a/run.json").write_text("not json")
            self.assertEqual(strix_pilot.assess_run(root, 0), (3, "incomplete", 0))

    def test_pinned_strix_clean_scan_without_json_index_accepts_sarif_coverage(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            run = root / "strix_runs/pilot"
            run.mkdir(parents=True)
            (run / "run.json").write_text('{"status": "completed"}')
            (run / "coverage.json").write_text('{"completeness": {"complete": true}}')
            (run / "findings.sarif").write_text('{"version": "2.1.0", "runs": [{"results": []}]}')
            self.assertEqual(strix_pilot.assess_run(root, 0), (0, "completed", 0))
            coverage = {"runs": [{"results": [{"ruleId": "coverage", "kind": "pass"},
                                               {"ruleId": "coverage", "kind": "open"},
                                               {"ruleId": "coverage", "kind": "notApplicable"}]}]}
            (run / "findings.sarif").write_text(json.dumps(coverage))
            self.assertEqual(strix_pilot.assess_run(root, 0), (0, "completed", 0))
            (run / "findings.sarif").write_text('{"runs": [{"results": [{"ruleId": "vuln-1"}]}]}')
            self.assertEqual(strix_pilot.assess_run(root, 0), (3, "incomplete", 0))

    def test_completed_root_cannot_hide_failed_child_or_missing_completeness(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            run = root / "strix_runs/pilot"
            run.mkdir(parents=True)
            (run / "run.json").write_text('{"status": "completed"}')
            (run / "vulnerabilities.json").write_text('[]')
            self.assertEqual(strix_pilot.assess_run(root, 0), (3, "incomplete", 0))
            (run / "coverage.json").write_text('{"completeness": {"complete": false}}')
            self.assertEqual(strix_pilot.assess_run(root, 0), (3, "incomplete", 0))
            (run / "coverage.json").write_text('{"completeness": {"complete": true}}')
            (run / "findings.sarif").write_text('{"runs": [{"invocations": [{"executionSuccessful": false}]}]}')
            self.assertEqual(strix_pilot.assess_run(root, 0), (3, "incomplete", 0))


class URLTests(unittest.TestCase):
    def test_credentials_paths_and_non_http_targets_are_rejected(self):
        for url in ("file:///tmp/app", "http://user:pass@localhost", "http://localhost/api",
                    "http://localhost/?query=secret", "http://localhost/#fragment"):
            with self.subTest(url=url), self.assertRaises(ValueError):
                strix_pilot.validate_url(url)

    def test_lan_origin_is_preserved(self):
        self.assertEqual(strix_pilot.validate_url("http://192.168.4.143:8080/"),
                         "http://192.168.4.143:8080")


class ScanCommandTests(unittest.TestCase):
    def invoke(self, *, status="completed", budget="10"):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / "repository"
            source.mkdir()
            subprocess.run(["git", "init", "-q", str(source)], check=True)
            (source / "tracked.txt").write_text("committed fixture")
            subprocess.run(["git", "-C", str(source), "add", "tracked.txt"], check=True)
            subprocess.run(["git", "-C", str(source), "-c", "user.name=Pilot test",
                            "-c", "user.email=pilot@test.invalid", "commit", "-qm", "fixture"], check=True)
            (source / ".env").write_text("SECRET=must-not-be-mounted")
            accounts = root / "accounts.json"
            accounts.write_text('{"viewer": {"username": "fixture", "password": "private-fixture"}}')
            scanner = root / "scanner"
            scanner.write_text(
                "#!/usr/bin/env python3\n"
                "import json, os, pathlib, sys\n"
                "args = sys.argv[1:]\n"
                "source = pathlib.Path(args[args.index('-t') + 1])\n"
                "assert (source / 'tracked.txt').read_text() == 'committed fixture'\n"
                "assert not (source / '.env').exists() and not (source / '.git').exists()\n"
                "assert '--max-budget' in args and '--max-turns' in args\n"
                "run = pathlib.Path('strix_runs/pilot'); run.mkdir(parents=True)\n"
                "(run / 'run.json').write_text(json.dumps({'status': os.environ['FIXTURE_STATUS']}))\n"
                "(run / 'coverage.json').write_text('{\"completeness\": {\"complete\": true}}')\n"
                "(run / 'findings.sarif').write_text('{\"runs\": [{\"results\": []}]}')\n"
                "print('private-fixture: simulated private scanner output')\n"
            )
            scanner.chmod(0o700)
            output = root / "result"
            result = subprocess.run(
                [sys.executable, str(Path(strix_pilot.__file__)), "scan", "--source", str(source),
                 "--url", "http://localhost:18080", "--accounts", str(accounts),
                 "--output", str(output), "--strix", str(scanner), "--budget", budget],
                env=dict(os.environ, STRIX_LLM="fixture/model", LLM_API_KEY="fixture-key",
                         FIXTURE_STATUS=status), capture_output=True, text=True,
            )
            self.assertNotIn("private-fixture", result.stdout + result.stderr)
            self.assertFalse((output / "instructions.private.md").exists())
            return result, output.exists()

    def test_committed_snapshot_and_private_output_with_clean_release_artifacts(self):
        result, _ = self.invoke()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("completed; recorded findings: 0", result.stdout)

    def test_zero_exit_from_budget_stopped_scanner_does_not_pass(self):
        result, _ = self.invoke(status="stopped")
        self.assertEqual(result.returncode, 3)
        self.assertIn("incomplete", result.stdout)

    def test_nonfinite_and_excessive_budgets_do_not_launch_scanner(self):
        for budget in ("nan", "inf", "0", "51"):
            with self.subTest(budget=budget):
                result, created = self.invoke(budget=budget)
                self.assertEqual(result.returncode, 3)
                self.assertFalse(created)


if __name__ == "__main__":
    unittest.main()
