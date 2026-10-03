import json
import os
from pathlib import Path
import subprocess
import tempfile
import textwrap
import unittest

ROOT = Path(__file__).resolve().parents[2]
CHECKED_SHA = '1111111111111111111111111111111111111111'
OTHER_SHA = '2222222222222222222222222222222222222222'

GH_FIXTURE = r'''#!/usr/bin/env python3
import json
import os
from pathlib import Path
import sys

args = sys.argv[1:]
fixture = json.loads(Path(os.environ['RELEASE_FIXTURE']).read_text())
state_path = Path(os.environ['RELEASE_STATE'])
state = json.loads(state_path.read_text())
if args[0] == 'api':
    route = next(arg for arg in args if arg.startswith('repos/'))
    if route.endswith('/git/ref/heads/main'):
        print(fixture['main_sha'])
    elif route.endswith('/actions/workflows/ci.yml/runs'):
        assert 'branch=main' in args and 'event=push' in args
        assert 'head_sha=1111111111111111111111111111111111111111' in args
        print(fixture['ci'])
    elif '/releases?' in route:
        if fixture['lookup_error']:
            sys.exit(1)
        if fixture['existing_release']:
            print('v' + fixture['version'])
    elif '/git/matching-refs/tags/' in route:
        if fixture['existing_tag']:
            print('refs/tags/v' + fixture['version'])
    elif '/commits/' in route:
        print(fixture['tag_sha'])
    else:
        raise AssertionError('Unexpected GitHub API request: ' + route)
elif args[:2] == ['release', 'create']:
    notes_path = args[args.index('--notes-file') + 1]
    state['release'] = {'args': args, 'notes': Path(notes_path).read_text()}
    state_path.write_text(json.dumps(state))
elif args[:2] == ['workflow', 'run']:
    state['dispatches'].append(args)
    state_path.write_text(json.dumps(state))
else:
    raise AssertionError('Unexpected GitHub CLI call: ' + repr(args))
'''


def release_script():
    workflow = ROOT / '.github/workflows/release.yml'
    if not workflow.exists():
        return ''
    lines = workflow.read_text().splitlines()
    index = next(i for i, line in enumerate(lines) if line.strip() == 'run: |')
    indentation = len(lines[index]) - len(lines[index].lstrip())
    body = []
    for line in lines[index + 1:]:
        if line.strip() and len(line) - len(line.lstrip()) <= indentation:
            break
        body.append(line)
    return textwrap.dedent('\n'.join(body))


class ReleaseWorkflowTests(unittest.TestCase):
    def run_release(self, *, version='0.1.21', main_sha=CHECKED_SHA, ci='success',
                    notes=True, languages=('Deutsch', 'English'), existing_release=False, existing_tag=False,
                    tag_sha=CHECKED_SHA, attempt='1', event='workflow_run', lookup_error=False):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'backend/cmd/railkeeper').mkdir(parents=True)
            (root / 'backend/cmd/railkeeper/main.go').write_text(
                f'const (\n\tversion = "{version}"\n)\n')
            (root / 'docs/releases').mkdir(parents=True)
            if notes:
                (root / f'docs/releases/v{version}.md').write_text(
                    f'# RailKeeper v{version}\n\n' + '\n\n'.join(
                        f'## {language}\n\nReviewed release notes.' for language in languages) + '\n')
            (root / 'bin').mkdir()
            gh = root / 'bin/gh'
            gh.write_text(GH_FIXTURE)
            gh.chmod(0o755)
            fixture = root / 'fixture.json'
            fixture.write_text(json.dumps({
                'main_sha': main_sha, 'ci': ci, 'existing_release': existing_release,
                'existing_tag': existing_tag, 'tag_sha': tag_sha, 'version': version,
                'lookup_error': lookup_error,
            }))
            state = root / 'state.json'
            state.write_text(json.dumps({'release': None, 'dispatches': []}))
            environment = dict(os.environ, PATH=f'{root / "bin"}:{os.environ["PATH"]}',
                               RELEASE_FIXTURE=str(fixture), RELEASE_STATE=str(state),
                               GITHUB_REPOSITORY='ichwars/RailKeeper', RELEASE_SHA=CHECKED_SHA,
                               GITHUB_RUN_ATTEMPT=attempt, GITHUB_EVENT_NAME=event)
            result = subprocess.run(['bash', '-c', release_script()], cwd=root,
                                    env=environment, capture_output=True, text=True, check=False)
            return result, json.loads(state.read_text())

    def test_new_stable_release_uses_checked_commit_and_starts_both_builds(self):
        result, state = self.run_release()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIsNotNone(state['release'])
        args = state['release']['args']
        self.assertEqual(args[2], 'v0.1.21')
        self.assertEqual(args[args.index('--target') + 1], CHECKED_SHA)
        self.assertIn('--latest', args)
        self.assertNotIn('--prerelease', args)
        self.assertIn('Reviewed release notes.', state['release']['notes'])
        self.assertEqual(state['dispatches'], [
            ['workflow', 'run', 'docker-image.yml', '--repo', 'ichwars/RailKeeper',
             '--ref', 'v0.1.21'],
            ['workflow', 'run', 'windows-standalone.yml', '--repo', 'ichwars/RailKeeper',
             '--ref', 'v0.1.21'],
        ])

    def test_failed_ci_never_publishes(self):
        result, state = self.run_release(ci='failure')
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(state, {'release': None, 'dispatches': []})

    def test_superseded_commit_never_publishes(self):
        result, state = self.run_release(main_sha=OTHER_SHA)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(state, {'release': None, 'dispatches': []})

    def test_missing_release_notes_never_publishes(self):
        result, state = self.run_release(notes=False)
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(state, {'release': None, 'dispatches': []})

    def test_missing_language_section_never_publishes(self):
        for language in ('Deutsch', 'English'):
            with self.subTest(language=language):
                result, state = self.run_release(languages=(language,))
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(state, {'release': None, 'dispatches': []})

    def test_existing_release_remains_unchanged(self):
        result, state = self.run_release(existing_release=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(state, {'release': None, 'dispatches': []})

    def test_existing_tag_at_another_commit_is_not_reused(self):
        result, state = self.run_release(existing_tag=True, tag_sha=OTHER_SHA)
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(state, {'release': None, 'dispatches': []})

    def test_rerun_resumes_builds_without_recreating_release(self):
        result, state = self.run_release(existing_release=True, existing_tag=True, attempt='2')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIsNone(state['release'])
        self.assertEqual(len(state['dispatches']), 2)

    def test_beta_release_does_not_become_latest(self):
        result, state = self.run_release(version='0.1.22-beta.1')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIsNotNone(state['release'])
        self.assertIn('--prerelease', state['release']['args'])
        self.assertIn('--latest=false', state['release']['args'])

    def test_malformed_version_never_publishes(self):
        result, state = self.run_release(version='0.1.21;echo injected')
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(state, {'release': None, 'dispatches': []})

    def test_empty_prerelease_identifier_never_publishes(self):
        for version in ('0.1.22-beta..1', '0.1.22-beta.', '0.1.22-.'):
            with self.subTest(version=version):
                result, state = self.run_release(version=version)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(state, {'release': None, 'dispatches': []})

    def test_release_lookup_failure_never_publishes(self):
        result, state = self.run_release(lookup_error=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(state, {'release': None, 'dispatches': []})


if __name__ == '__main__':
    unittest.main(verbosity=2)
