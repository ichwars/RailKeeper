import json
import os
from pathlib import Path
import subprocess
import tempfile
import textwrap
import unittest


ROOT = Path(__file__).resolve().parents[2]


def audit_script():
    workflow = (ROOT / '.github/workflows/security-audit.yml').read_text()
    step = workflow.split('      - name: Scan runtime and development dependencies\n', 1)[1]
    return textwrap.dedent(step.split('        run: |\n', 1)[1])


class SecurityAuditWorkflowTests(unittest.TestCase):
    def test_audits_both_projects_and_propagates_failures(self):
        for failures in [[], ['frontend'], ['docs'], ['frontend', 'docs']]:
            with self.subTest(failures=failures), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                for name in ['frontend', 'docs', 'bin', 'reports']:
                    (root / name).mkdir()
                npm = root / 'bin/npm'
                npm.write_text('''#!/usr/bin/env python3
import json
import os
from pathlib import Path
import sys

assert sys.argv[1:] == ['audit', '--include=dev', '--audit-level=high']
project = Path.cwd().name
with Path(os.environ['AUDIT_CALLS']).open('a') as calls:
    calls.write(project + '\\n')
print('private audit details')
sys.exit(1 if project in json.loads(os.environ['AUDIT_FAILURES']) else 0)
''')
                npm.chmod(0o755)
                calls = root / 'calls.txt'
                environment = dict(os.environ, PATH=f'{root / "bin"}:{os.environ["PATH"]}',
                                   RUNNER_TEMP=str(root / 'reports'), AUDIT_CALLS=str(calls),
                                   AUDIT_FAILURES=json.dumps(failures))
                result = subprocess.run(['bash', '-e', '-o', 'pipefail', '-c', audit_script()],
                                        cwd=root, env=environment, capture_output=True,
                                        text=True, check=False)
                self.assertEqual(result.returncode, 1 if failures else 0, result.stderr)
                self.assertEqual(calls.read_text().splitlines(), ['frontend', 'docs'])
                for name in ['frontend', 'docs']:
                    self.assertEqual((root / 'reports' / f'npm-audit-{name}.txt').read_text(),
                                     'private audit details\n')
                    self.assertEqual(f'audit failed in {name}' in result.stdout, name in failures)
                self.assertNotIn('private audit details', result.stdout + result.stderr)


if __name__ == '__main__':
    unittest.main()
