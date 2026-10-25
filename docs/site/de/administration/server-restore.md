---
title: Eine vollständige Serversicherung wiederherstellen
description: Eine private Serversicherung getrennt wiederherstellen, prüfen und die Testkopie sicher entfernen.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-25
---

# Eine vollständige Serversicherung wiederherstellen

**Ziel:** Nachweisen, dass du RailKeeper mit Bestand, gespeicherten Dateien und vorhandener Anmeldung
wieder starten kannst. Die laufende Installation bleibt dabei unverändert.

Diese Anleitung gehört zur **privaten Datei-/Volumensicherung des Servers**. Die JSON-Datei aus
**Einstellungen > Backup** wird in der [bebilderten Anwendungssicherungs-Übung](restore-exercise)
behandelt. Ein Serverarchiv lässt sich dort nicht hochladen.

## 1. Sicherung und passenden Softwarestand bereitlegen

Du benötigst das vollständige Archiv, seine getrennt aufbewahrte SHA-256-Prüfsumme und ein vorhandenes
Admin-Konto aus dem gesicherten Stand. Passwort und gegebenenfalls den zweiten Faktor privat bereithalten.

| Bestandteil | Was damit zurückkommt |
| --- | --- |
| Gesamter Datenordner `data/` | Datenbank, gespeicherte Dateien und lokale Konten/Rollen |
| `.env` und Compose-Datei | Start- und Sicherheitskonfiguration der gesicherten Installation |
| Passendes Dockerimage, separat verfügbar | Das Programm, mit dem die Sicherung wieder gestartet wird |

Ein Archiv dieses Ablaufs enthält `.env`, `compose.yaml` und `data/railkeeper.db`.
Dateinamen und Speicherpfade können bei deiner Installation abweichen. Passe dann die Befehle an.
Ein Dockerimage gehört nicht automatisch zum Archiv: Halte Image-ID oder Digest fest und stelle sicher,
dass dieser Stand noch verfügbar ist. Verwende für den ersten Lauf möglichst den ursprünglichen Stand.

## 2. Eine private Testkopie prüfen

Arbeite als Betreiber auf einem Linux-Rechner mit Docker Compose, Python **3.12 oder neuer** und
`nsenter`. Für Dateirechte und den Netzwerk-Test sind Administratorrechte nötig.
Die folgenden Beispiele werden in derselben Shell ausgeführt. Ersetze den Produktions-Containernamen
`railkeeper` in den Kontrollbefehlen, falls deine Installation einen anderen Namen verwendet.

Setze `RK_BACKUP` auf die tatsächliche Archivkopie. In diesem Beispiel liegt die zugehörige
Prüfsumme unter demselben Namen mit angehängtem `.sha256`:

```sh
RK_BACKUP="/pfad/zur/privaten/serversicherung.tar.gz"
(cd "$(dirname "$RK_BACKUP")" && sha256sum -c "$(basename "$RK_BACKUP").sha256")
gzip -t "$RK_BACKUP"
```

**Erwartung:** Prüfsumme stimmt und `gzip -t` meldet keinen Fehler. Bei Abweichungen hier stoppen und
eine unveränderte Kopie beschaffen. Die Prüfsumme muss aus deinem vertrauenswürdigen Sicherungsnachweis stammen.

## 3. In einen neuen Ordner wiederherstellen

Entpacke ausschließlich in einen neu angelegten privaten Testordner. Die ursprüngliche Installation,
Originalsicherung und bisherigen Datenordner bleiben erhalten. Der folgende Block prüft außerdem
Archivpfade, alle entpackten Dateien und die SQLite-Datenbank:

<details>
<summary>Technische Schritte: Archiv prüfen und entpacken</summary>

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

**Erwartung:** Keine Fehlermeldung; die ausgegebenen Zahlen passen zum Sicherungsstand.
`indexPages` zählt CV-Indexseiten, `provenance` Werte mit Originalzeile und Exportdatum.
Auch in der Datenbank gespeicherte Medien gehören zur Sicherung, selbst wenn eine Dateianzeige null meldet.

## 4. Nur die Testkopie starten

Die gespeicherte Compose-Datei dient hier zum **Lesen der Konfiguration**. Starte nicht einfach
`docker compose up` aus der Kopie: Containername, Ports oder absolute Volumes könnten zur laufenden
Installation gehören. Das folgende Beispiel startet stattdessen einen eigenen Container mit genau
einem Datenpfad, ohne Netzwerk und ohne freigegebene Ports.

<details>
<summary>Technische Schritte für den Betreiber: Konfiguration übernehmen und isoliert starten</summary>

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

**Erwartung:** Der Healthcheck ist erfolgreich. Fehlt das ursprüngliche Image, beschaffe genau diesen
Stand oder wähle einen ausdrücklich zur Sicherung passenden Digest. Wechsle nicht ungeprüft auf `latest`.
Eine weitere Version wird mit einer **frisch entpackten Kopie** geprüft, damit deren Änderungen den
ursprünglichen Vergleich nicht verfälschen.

## 5. Bestand und Anmeldung wirklich kontrollieren

Der Testcontainer hat keine Browseradresse im Heimnetz. Greife nicht versehentlich auf deine normale
RailKeeper-Adresse zu. Im geprüften Ablauf liefen die HTTP-Prüfungen innerhalb seines eigenen
Netzwerkbereichs; dort ist nur `127.0.0.1` erreichbar.

| Kontrolle | Erwartung |
| --- | --- |
| Vorhandenes Konto | Anmeldung funktioniert mit dem gesicherten Konto; keine neue Ersteinrichtung |
| Benutzer und Rollen | Bekannte Konten und Rechte sind vorhanden |
| Fahrzeuge und Funktionen | Anzahlen und bekannte Angaben stimmen mit der Sicherung überein |
| CVs | Werte, Indexseiten, Originalzeilen, Exportdatum und manuelle Angaben sind erhalten |
| Bilder und Dokumente | Gespeicherte Dateien tatsächlich öffnen oder herunterladen |
| Decoderbibliothek | Unterstütztes Profil und seine Erklärungen sind verfügbar |
| Neue Anwendungssicherung | Exportieren und ohne Einspielen erfolgreich validieren |

<details>
<summary>Technisches Beispiel: Anmeldung und Daten im isolierten HTTP-Test prüfen</summary>

Dieser Konsolentest fragt ein vorhandenes Konto verdeckt am Terminal ab. Er vergleicht Bestand,
Funktionen, CVs und Herkunft mit den zuvor ermittelten Zahlen, lädt gespeicherte Fahrzeugdateien
und prüft einen neuen App-Backup-Export.

Er bildet die geprüfte lokale HTTP-Konfiguration ab. Bei sicheren HTTPS-Cookies oder Zwei-Faktor-Anmeldung
benötigt der Betreiber einen passenden getrennten TLS-/Anmelde-Test. Behalte die Sicherheitskonfiguration
bei und schalte weder Cookieschutz noch zweiten Faktor für die Übung ab.

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

**Erwartung:** Die Anzahlen stimmen; Dateien lassen sich öffnen und der neue Export ist kompatibel.
Wenn ein Konto fehlt, du eine leere Datenbank siehst oder eine Prüfung scheitert, stoppe die Probe.
Prüfe Archiv, eingebundenen Datenpfad und Imageversion. Setze nicht die laufende Installation zurück.

## 6. Die laufende Installation unverändert bestätigen

Vergleiche auf dem Betreiberrechner die vorher festgehaltene Container-ID, Startzeit und Image-ID mit
dem laufenden RailKeeper. Konfiguration und Bestandsdaten dürfen durch die Probe nicht verändert worden sein.
Es werden keine Verbindungs-, Fahr-, Programmier- oder Uploadaktionen an einer Digitalzentrale ausgelöst.
Der Testcontainer hat keinen Zugang zu diesen Geräten.

```sh
docker inspect --format '{{.Id}} {{.State.StartedAt}} {{.Image}}' railkeeper > "$RK_TEST_ROOT/live-after.txt"
diff -u "$RK_TEST_ROOT/live-before.txt" "$RK_TEST_ROOT/live-after.txt"
```

Keine Differenz bedeutet: derselbe Container, dieselbe Startzeit und dasselbe Image.

## 7. Ergebnis festhalten und Testdaten entfernen

Notiere Sicherungsdatum und Prüfsumme, getestete Imageversionen, kontrollierte Anzahlen und das Ergebnis.
Bewahre den Prüfbericht neben der privaten Sicherung auf. Veröffentliche keine Sicherungen,
Anmeldedaten oder unbereinigten Konfigurations-/Protokollausgaben in GitHub-Issues.

Entferne anschließend ausschließlich den eigenen Testcontainer und seinen temporären Ordner:

<details>
<summary>Technische Schritte: Testkopie entfernen</summary>

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

Das ursprüngliche Archiv bleibt erhalten. Eine echte Wiederinbetriebnahme nach einem Serverausfall
ist ein eigener Schritt: Stelle erst nach erfolgreicher Probe Programm, Daten und Konfiguration passend
zusammen und richte den vorgesehenen Zugriff wieder ein. Ein älteres Programm darf nicht ungeprüft eine
bereits migrierte Datenbank öffnen.

## Was wurde nachgewiesen?

Am **24.10.2026** wurde eine vollständige private Serversicherung aus der externen Kopie wiederhergestellt
und mit dem ursprünglichen Softwarestand sowie dem veröffentlichten `main`-Image zu `b02f060d` geprüft.
Anmeldung, Rollen, Bestand, CVs samt Herkunft, Dateiabrufe und neuer App-Backup-Export mit Validierung
funktionierten. Die laufende Installation blieb unverändert; Testcontainer und entpackte Daten wurden entfernt.

Der Nachweis gilt für diese Linux-/Docker-Wiederherstellung mit bereits verfügbaren Images.
Eine Neuinstallation des Hostbetriebssystems, ein Windows-Umzug oder das erneute Beschaffen eines
Images waren nicht Teil der Probe. Deine Erfolgszahlen richten sich nach deiner eigenen Sicherung.

[Welche Sicherung brauche ich?](backup-restore) · [Anwendungssicherung üben](restore-exercise) · [Updates und Betrieb](system-operations)
