import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
CHECKED_SHA = '1111111111111111111111111111111111111111'
OTHER_SHA = '2222222222222222222222222222222222222222'

GH_FIXTURE = r'''#!/usr/bin/env python3
import hashlib
import json
import os
from pathlib import Path
import sys

args = sys.argv[1:]
fixture = json.loads(Path(os.environ['RELEASE_FIXTURE']).read_text())
state_path = Path(os.environ['RELEASE_STATE'])
state = json.loads(state_path.read_text())
repo = 'ichwars/RailKeeper'
tag = 'v' + fixture['version']
sha = os.environ['RELEASE_SHA']
zip_name = f'RailKeeper-windows-x64-{tag}.zip'

if args[0] == 'api':
    route = next(arg for arg in args if arg.startswith('repos/'))
    if route.endswith('/git/ref/heads/main'):
        print(fixture['later_main'] if 'attest' in state['events'] else fixture['main_sha'])
    elif route.endswith('/actions/workflows/ci.yml/runs'):
        assert 'branch=main' in args and 'event=push' in args and f'head_sha={sha}' in args
        print(fixture['ci'])
    elif '/releases?' in route:
        if fixture['lookup_error']:
            sys.exit(1)
        if fixture['existing_release']:
            print(tag + '\t' + fixture['existing_release'])
    elif '/git/matching-refs/tags/' in route:
        if fixture['existing_tag']:
            print(f'refs/tags/{tag}')
    elif route.endswith('/git/refs'):
        assert 'POST' in args and f'ref=refs/tags/{tag}' in args and f'sha={sha}' in args
        state['events'].append('stage-tag')
    elif '/commits/' in route:
        print(fixture['tag_sha'])
    elif route.endswith('/actions/workflows/windows-standalone.yml/dispatches'):
        assert 'POST' in args and f'ref={tag}' in args
        state['events'].append('windows')
        print(421)
    elif route.endswith('/actions/runs/421'):
        result = fixture['windows_result']
        print(f'workflow_dispatch\t{tag}\t{fixture["windows_sha"]}\tcompleted\t{result}')
    else:
        raise AssertionError('Unexpected GitHub API request: ' + route)
elif args[:2] == ['run', 'download']:
    assert args[2] == '421' and args[args.index('--name') + 1] == 'railkeeper-windows-verified'
    state['events'].append('download')
    directory = Path(args[args.index('--dir') + 1])
    payload = b'Test Windows package'
    if fixture['package'] != 'missing':
        (directory / zip_name).write_bytes(payload)
    digest = hashlib.sha256(payload).hexdigest()
    if fixture['package'] == 'bad-hash':
        digest = '0' * 64
    (directory / 'checksums.txt').write_text(f'{digest}  {zip_name}\n')
    if fixture['package'] != 'missing-bundle':
        (directory / 'windows-provenance.jsonl').write_text('{}\n')
elif args[:2] == ['attestation', 'verify']:
    assert Path(args[2]).name == zip_name and Path(args[2]).is_file()
    assert Path(args[args.index('--bundle') + 1]).is_file()
    assert args[args.index('--repo') + 1] == repo
    assert args[args.index('--signer-workflow') + 1] == repo + '/.github/workflows/windows-standalone.yml'
    assert args[args.index('--source-ref') + 1] == f'refs/tags/{tag}'
    assert args[args.index('--source-digest') + 1] == sha
    state['events'].append('attest')
    if not fixture['valid_attestation']:
        state_path.write_text(json.dumps(state))
        sys.exit(1)
elif args[:2] == ['release', 'create']:
    assert args[2] == tag and '--verify-tag' in args
    assert Path(args[3]).name == zip_name and Path(args[3]).is_file()
    assert Path(args[4]).name == 'checksums.txt' and Path(args[4]).is_file()
    assert 'attest' in state['events']
    notes_path = args[args.index('--notes-file') + 1]
    state['release'] = {'args': args, 'notes': Path(notes_path).read_text()}
    state['events'].append('release')
elif args[:2] == ['workflow', 'run']:
    assert args == ['workflow', 'run', 'docker-image.yml', '--repo', repo, '--ref', tag]
    state['events'].append('docker')
else:
    raise AssertionError('Unexpected GitHub CLI call: ' + repr(args))
state_path.write_text(json.dumps(state))
'''


class ReleaseWorkflowTests(unittest.TestCase):
    def run_release(self, *, version='0.1.21', main_sha=CHECKED_SHA, later_main=CHECKED_SHA,
                    ci='success', notes=True, languages=('Deutsch', 'English'),
                    existing_release='', existing_tag=False, tag_sha=CHECKED_SHA,
                    windows_sha=CHECKED_SHA, windows_result='success', package='ok',
                    valid_attestation=True, lookup_error=False, main_source=None):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'backend/cmd/railkeeper').mkdir(parents=True)
            (root / 'backend/cmd/railkeeper/main.go').write_text(
                main_source if main_source is not None else f'const (\n\tversion = "{version}"\n)\n')
            (root / 'docs/releases').mkdir(parents=True)
            if notes:
                (root / f'docs/releases/v{version}.md').write_text(
                    f'# RailKeeper v{version}\n\n' + '\n\n'.join(
                        f'## {language}\n\nReviewed release notes.' for language in languages) + '\n')
            (root / '.github/scripts').mkdir(parents=True)
            for script in ('publish_release.sh', 'verify_windows_release.sh'):
                shutil.copyfile(ROOT / '.github/scripts' / script, root / '.github/scripts' / script)
            (root / 'bin').mkdir()
            gh = root / 'bin/gh'
            gh.write_text(GH_FIXTURE)
            gh.chmod(0o755)
            settings = {
                'main_sha': main_sha, 'later_main': later_main, 'ci': ci,
                'existing_release': existing_release, 'existing_tag': existing_tag,
                'tag_sha': tag_sha, 'version': version, 'lookup_error': lookup_error,
                'windows_sha': windows_sha, 'windows_result': windows_result,
                'package': package, 'valid_attestation': valid_attestation,
            }
            fixture = root / 'fixture.json'
            fixture.write_text(json.dumps(settings))
            state = root / 'state.json'
            state.write_text(json.dumps({'release': None, 'events': []}))
            environment = dict(os.environ, PATH=f'{root / "bin"}:{os.environ["PATH"]}',
                               RELEASE_FIXTURE=str(fixture), RELEASE_STATE=str(state),
                               GITHUB_REPOSITORY='ichwars/RailKeeper', RELEASE_SHA=CHECKED_SHA)
            result = subprocess.run(['bash', '.github/scripts/publish_release.sh'], cwd=root,
                                    env=environment, capture_output=True, text=True, check=False)
            return result, json.loads(state.read_text())

    def test_runtime_source_version_can_publish(self):
        source = (ROOT / 'backend/cmd/railkeeper/main.go').read_text()
        version = re.search(r'\bversion\s*=\s*"([^"]+)"', source).group(1)
        result, state = self.run_release(version=version, main_source=source)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(state['release']['args'][2], f'v{version}')

    def test_verified_windows_assets_precede_publication(self):
        result, state = self.run_release()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(state['events'], ['stage-tag', 'windows', 'download', 'attest', 'release', 'docker'])
        args = state['release']['args']
        self.assertIn('--latest', args)
        self.assertNotIn('--prerelease', args)
        self.assertIn('Reviewed release notes.', state['release']['notes'])

    def test_invalid_release_inputs_never_prepare_a_tag(self):
        for settings in ({'ci': 'failure'}, {'notes': False}, {'languages': ('Deutsch',)},
                         {'languages': ('English',)}, {'lookup_error': True}):
            with self.subTest(settings=settings):
                result, state = self.run_release(**settings)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(state, {'release': None, 'events': []})

    def test_superseded_commit_never_publishes(self):
        for settings in ({'main_sha': OTHER_SHA}, {'later_main': OTHER_SHA}):
            with self.subTest(settings=settings):
                result, state = self.run_release(**settings)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertIsNone(state['release'])
                self.assertNotIn('docker', state['events'])

    def test_existing_release_and_draft_remain_untouched(self):
        for draft in ('true', 'false'):
            with self.subTest(draft=draft):
                result, state = self.run_release(existing_release=draft)
                self.assertEqual(result.returncode == 0, draft == 'false')
                self.assertEqual(state, {'release': None, 'events': []})

    def test_existing_tag_at_another_commit_is_not_reused(self):
        result, state = self.run_release(existing_tag=True, tag_sha=OTHER_SHA)
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(state, {'release': None, 'events': []})

    def test_retry_can_reuse_unpublished_tag_at_same_commit(self):
        result, state = self.run_release(existing_tag=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(state['events'], ['windows', 'download', 'attest', 'release', 'docker'])

    def test_missing_or_unverified_windows_evidence_blocks_release(self):
        for settings in ({'package': 'missing'}, {'package': 'bad-hash'}, {'package': 'missing-bundle'},
                         {'valid_attestation': False}, {'windows_sha': OTHER_SHA}, {'windows_result': 'failure'}):
            with self.subTest(settings=settings):
                result, state = self.run_release(**settings)
                self.assertNotEqual(result.returncode, 0)
                self.assertIsNone(state['release'])
                self.assertNotIn('docker', state['events'])

    def test_beta_release_does_not_become_latest(self):
        result, state = self.run_release(version='0.1.22-beta.1')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn('--prerelease', state['release']['args'])
        self.assertIn('--latest=false', state['release']['args'])

    def test_malformed_version_never_publishes(self):
        for version in ('0.1.21;echo injected', '0.1.22-beta..1', '0.1.22-beta.', '0.1.22-.'):
            with self.subTest(version=version):
                result, state = self.run_release(version=version)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(state, {'release': None, 'events': []})


if __name__ == '__main__':
    unittest.main(verbosity=2)
