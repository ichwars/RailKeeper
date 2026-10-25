---
title: Restore a full server backup
description: Restore and verify a private server backup separately, then remove the test copy safely.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-25
---

# Restore a full server backup

**Goal:** Prove that RailKeeper can start again with its inventory, saved files and existing sign-in.
The running installation stays unchanged throughout the exercise.

This guide covers a **private server filesystem/volume backup**. The JSON file from
**Settings > Backup** is covered by the [illustrated application-backup exercise](restore-exercise).
A server archive cannot be uploaded there.

## 1. Prepare the backup and matching software

You need the complete archive, its separately retained SHA-256 checksum and an existing Admin account
from the backed-up state. Keep the password and any second factor private and available.

| Component | What it restores |
| --- | --- |
| Entire `data/` directory | Database, saved files and local accounts/roles |
| `.env` and Compose file | The installation's startup and security configuration |
| Matching Docker image, available separately | The program used to start the restored backup |

The archive used in this procedure contains `.env`, `compose.yaml` and `data/railkeeper.db`.
Names and paths may differ in your installation; adapt the commands accordingly.
Docker images are not automatically included: record the image ID or digest and ensure that version
remains available. Prefer the original software version for the first run.

## 2. Verify a private test copy

Work as the operator on a Linux machine with Docker Compose, Python **3.12 or newer** and `nsenter`.
Administrator privileges are needed for file ownership and the network test.
Run the following examples in the same shell. Replace the production container name `railkeeper`
in the comparison commands if your installation uses another name.

Set `RK_BACKUP` to the actual archive copy. This example expects its checksum in a file with the same
name followed by `.sha256`:

```sh
RK_BACKUP="/path/to/private/server-backup.tar.gz"
(cd "$(dirname "$RK_BACKUP")" && sha256sum -c "$(basename "$RK_BACKUP").sha256")
gzip -t "$RK_BACKUP"
```

**Expected:** The checksum matches and `gzip -t` reports no error. Stop on a mismatch and obtain an
unchanged copy. The checksum must come from your trusted backup record.

## 3. Restore into a new directory

Extract only into a newly created private test directory. Keep the original installation, backup and
existing data directories. This block also checks archive paths, all extracted files and SQLite integrity:

<details>
<summary>Operator steps: verify and extract the archive</summary>

```sh
umask 077
RK_TEST_ROOT="$(mktemp -d /tmp/railkeeper-restore.XXXXXX)"
export RK_TEST_ROOT RK_BACKUP
docker inspect --format '{{.Id}} {{.State.StartedAt}} {{.Image}}' railkeeper > "$RK_TEST_ROOT/live-before.txt"
python3 - <<'PY'
import hashlib, json, os, sqlite3, tarfile
from pathlib import Path, PurePosixPath
root = Path(os.environ['RK_TEST_ROOT'])
restored = root / 'restored'
with tarfile.open(os.environ['RK_BACKUP'], 'r:gz') as archive:
    members = archive.getmembers()
    for member in members:
        name = PurePosixPath(member.name)
        assert not name.is_absolute() and '..' not in name.parts
        assert not member.issym() and not member.islnk()
    archive.extractall(restored, filter='data')
    for member in members:
        if member.isfile():
            with archive.extractfile(member) as source:
                digest = hashlib.sha256(source.read()).hexdigest()
            assert hashlib.sha256((restored / member.name).read_bytes()).hexdigest() == digest
assert all((restored / name).is_file() for name in ('.env', 'compose.yaml', 'data/railkeeper.db'))
with sqlite3.connect(f'file:{restored}/data/railkeeper.db?mode=ro', uri=True) as db:
    assert db.execute('PRAGMA integrity_check').fetchall() == [('ok',)]
    assert not db.execute('PRAGMA foreign_key_check').fetchall()
    counts = {name: db.execute('SELECT count(*) FROM ' + name).fetchone()[0]
              for name in ('vehicles', 'vehicle_cv_values', 'vehicle_functions', 'users', 'roles')}
    counts['indexPages'] = db.execute('SELECT count(*) FROM (SELECT DISTINCT index_cv31, '
                                    'index_cv32 FROM vehicle_cv_values WHERE index_cv31 IS NOT NULL)').fetchone()[0]
    counts['provenance'] = db.execute("SELECT count(*) FROM vehicle_cv_values "
                                    "WHERE source_line IS NOT NULL AND source_exported_at<>''").fetchone()[0]
(root / 'expected.json').write_text(json.dumps(counts))
print(counts)
PY
```

</details>

**Expected:** No error; the printed counts match the backup.
`indexPages` counts CV index pages; `provenance` counts values retaining their original line and export date.
Media stored inside the database are included even when a file counter displays zero.

## 4. Start only the test copy

The saved Compose file is used here to **read the configuration**. Do not simply run `docker compose up`
from the copy: its container name, ports or absolute volumes may refer to the running installation.
The example starts a separate container with exactly one data path, no network and no published ports.

<details>
<summary>Operator steps: load the configuration and start an isolated container</summary>

```sh
docker compose --project-directory "$RK_TEST_ROOT/restored" \
  --env-file "$RK_TEST_ROOT/restored/.env" -f "$RK_TEST_ROOT/restored/compose.yaml" \
  config --format json > "$RK_TEST_ROOT/compose-resolved.json"
python3 - <<'PY'
import json, os
from pathlib import Path
root = Path(os.environ['RK_TEST_ROOT'])
service = json.loads((root / 'compose-resolved.json').read_text())['services']['railkeeper']
env = service.get('environment', {})
lines = []
for key, value in env.items():
    if value is not None:
        assert '\n' not in str(value) and '\r' not in str(value)
        lines.append(key + '=' + str(value))
(root / 'runtime.env').write_text('\n'.join(lines) + '\n')
(root / 'image.txt').write_text(service['image'])
PY
RK_TEST_IMAGE="$(cat "$RK_TEST_ROOT/image.txt")"
docker image inspect "$RK_TEST_IMAGE" > /dev/null
RK_TEST_NAME="$(basename "$RK_TEST_ROOT")"
sudo chown -R 10001:10001 "$RK_TEST_ROOT/restored/data"
docker run -d --name "$RK_TEST_NAME" --pull never --network none --read-only \
  --cap-drop ALL --security-opt no-new-privileges --user 10001:10001 \
  --env-file "$RK_TEST_ROOT/runtime.env" \
  --env RAILKEEPER_ADDR=127.0.0.1:8080 --env RAILKEEPER_DATA_DIR=/data \
  --mount "type=bind,src=$RK_TEST_ROOT/restored/data,dst=/data" "$RK_TEST_IMAGE"
for RK_TRY in $(seq 1 30); do
  if docker exec "$RK_TEST_NAME" railkeeper healthcheck; then break; fi
  sleep 1
done
docker exec "$RK_TEST_NAME" railkeeper healthcheck
```

</details>

**Expected:** The healthcheck succeeds. If the original image is missing, obtain that version or select
an explicitly compatible digest. Do not switch to `latest` without checking compatibility.
Test another version against a **freshly extracted copy**, so its changes cannot alter the original
comparison.

## 5. Check the actual records and sign-in

The test container has no browser address on your home network. Avoid accidentally checking your normal
RailKeeper URL. In the verified procedure, HTTP checks ran inside the test container's own network
namespace, where only `127.0.0.1` is reachable.

| Check | Expected result |
| --- | --- |
| Existing account | Sign-in works with the saved account; no new initial setup |
| Users and roles | Known accounts and permissions are present |
| Vehicles and functions | Counts and known details match the backup |
| CVs | Values, index pages, original lines, export dates and manual entries are retained |
| Images and documents | Actually open or download the saved files |
| Decoder library | Supported profile and explanations are available |
| New application backup | Export and successfully validate it without restoring it |

<details>
<summary>Technical example: check sign-in and records in the isolated HTTP test</summary>

This console test prompts privately for an existing account. It compares inventory, functions, CVs
and provenance with the recorded counts, downloads saved vehicle files and validates a new application
backup export.

It follows the verified local HTTP configuration. Secure HTTPS cookies or two-factor sign-in require
an appropriate separate TLS/sign-in test. Keep the security configuration; do not disable cookie
protection or the second factor for this exercise.

```sh
RK_TEST_PID="$(docker inspect -f '{{.State.Pid}}' "$RK_TEST_NAME")"
sudo nsenter --target "$RK_TEST_PID" --net -- python3 - "$RK_TEST_ROOT" <<'PY'
import getpass, http.cookiejar, json, secrets, socket, sys, urllib.request
from pathlib import Path
root = Path(sys.argv[1])
assert {name for _, name in socket.if_nameindex()} == {'lo'}
expected = json.loads((root / 'expected.json').read_text())
env = json.loads((root / 'compose-resolved.json').read_text())['services']['railkeeper']['environment']
assert str(env.get('RAILKEEPER_COOKIE_SECURE', 'false')).lower() == 'false', 'Use isolated TLS for secure cookies'
client = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()),
                                   urllib.request.ProxyHandler({}))
csrf = ''
def request(path, body=None, content_type='application/json'):
    headers = {'Origin': env.get('RAILKEEPER_PUBLIC_URL') or 'http://127.0.0.1:8080',
               'Content-Type': content_type}
    if csrf:
        headers['X-CSRF-Token'] = csrf
    with client.open(urllib.request.Request('http://127.0.0.1:8080' + path, data=body,
                                           headers=headers), timeout=30) as response:
        return response.read()
try:
    auth = json.loads(request('/api/v1/auth/login', json.dumps({
        'username': getpass.getpass('Account: '), 'password': getpass.getpass('Password: ')
    }).encode()))
    csrf = auth['csrfToken']
    vehicles = json.loads(request('/api/v1/vehicles'))
    details = [json.loads(request('/api/v1/vehicles/' + v['id'])) for v in vehicles]
    cvs = [cv for v in details for cv in v.get('cvValues', [])]
    actual = {'vehicles': len(details), 'vehicle_cv_values': len(cvs),
              'vehicle_functions': sum(len(v.get('functions', [])) for v in details),
              'users': len(json.loads(request('/api/v1/users'))),
              'roles': len(json.loads(request('/api/v1/roles'))),
              'indexPages': len({(cv['indexCV31'], cv['indexCV32']) for cv in cvs if cv.get('indexCV31') is not None}),
              'provenance': sum(cv.get('sourceLine') is not None and bool(cv.get('sourceExportedAt')) for cv in cvs)}
    assert actual == expected, 'Restored records differ from the archive'
    downloads = 0
    for v in details:
        prefix = '/api/v1/vehicles/' + v['id']
        for kind, route in [('images', 'file'), ('attachments', 'download'), ('cvFiles', 'download')]:
            for item in v.get(kind, []):
                if kind == 'images' and not item.get('fileName'):
                    continue
                segment = 'cv-files' if kind == 'cvFiles' else kind
                assert request(prefix + '/' + segment + '/' + item['id'] + '/' + route)
                downloads += 1
    assert json.loads(request('/api/v1/decoder-library'))['profiles']
    backup = request('/api/v1/backup/export')
    boundary = 'rkrestore' + secrets.token_hex(12)
    body = (f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="test.json"\r\n'
            'Content-Type: application/json\r\n\r\n').encode() + backup + f'\r\n--{boundary}--\r\n'.encode()
    result = json.loads(request('/api/v1/backup/validate', body, 'multipart/form-data; boundary=' + boundary))
    assert result['compatible'] and not result['errors']
    print({'records': actual, 'storedFileDownloads': downloads, 'backupWarnings': result['warnings']})
finally:
    if csrf:
        request('/api/v1/auth/logout', b'{}')
PY
```

</details>

**Expected:** Counts match, files open and the new export is compatible.
If an account is missing, the database is empty or a check fails, stop the exercise. Check the archive,
mounted data path and image version. Do not reset the running installation.

## 6. Confirm the running installation is unchanged

On the operator machine, compare the recorded container ID, start time and image ID with the running
RailKeeper instance. The exercise must not change its configuration or business data.
Do not trigger connection, driving, programming or upload actions on a command station.
The test container cannot access these devices.

```sh
docker inspect --format '{{.Id}} {{.State.StartedAt}} {{.Image}}' railkeeper > "$RK_TEST_ROOT/live-after.txt"
diff -u "$RK_TEST_ROOT/live-before.txt" "$RK_TEST_ROOT/live-after.txt"
```

No difference means the same container, start time and image.

## 7. Record the result and remove test data

Record the backup date and checksum, tested image versions, checked counts and result.
Keep the report beside the private backup. Do not publish backups, sign-in data or unredacted
configuration/log output in GitHub issues.

Remove only your own test container and temporary directory:

<details>
<summary>Operator steps: remove the test copy</summary>

```sh
test "$RK_TEST_NAME" = "$(basename "$RK_TEST_ROOT")" || exit 1
docker rm -f "$RK_TEST_NAME"
sudo python3 - "$RK_TEST_ROOT" <<'PY'
import shutil, sys
from pathlib import Path
root = Path(sys.argv[1])
assert root.parent == Path('/tmp') and root.name.startswith('railkeeper-restore.')
assert not root.is_symlink()
shutil.rmtree(root)
PY
```

</details>

Keep the original archive. An actual restart after server failure is a separate step: after a successful
exercise, bring together matching program, data and configuration and restore the intended access.
An older application must not open an already migrated database without checking compatibility.

## What was verified?

On **24 October 2026**, a full private server backup was restored from its off-system copy and tested
against its original software version and the published `main` image at `b02f060d`.
Sign-in, roles, inventory, CV provenance, file downloads and a new application-backup export with
validation succeeded. The running installation remained unchanged; test containers and extracted data
were removed.

This evidence covers that Linux/Docker recovery using already available images. Reinstalling the host
operating system, moving a Windows installation and obtaining an image again were outside the exercise.
Use your own backup's counts as the expected result.

[Which backup do I need?](backup-restore) · [Practice application backup](restore-exercise) · [Updates and operation](system-operations)
