# Strix-Pilot / Strix pilot

## Deutsch

Der Workflow **Strix-Pilot / Strix pilot** startet ausschließlich manuell über
**Actions → Strix-Pilot / Strix pilot → Run workflow**. Nach dem Merge muss die
Workflow-Datei auf dem Default-Branch liegen, damit GitHub den Startknopf anbietet.
Der ausgewählte Branch bestimmt den gebauten und geprüften Commit.

### Einrichtung und erster Lauf

Unter **Settings → Secrets and variables → Actions** konfigurieren:

- Variable `STRIX_LLM`: Modellkennung des gewünschten Strix-Providers.
- Secret `LLM_API_KEY`: zugehöriger API-Schlüssel, möglichst mit eigenem Provider-Budget.

Modellkennungen stehen in der [Strix-Provider-Dokumentation](https://docs.strix.ai/llm-providers/overview).
Der Pilot verwendet die Open-Source-CLI, keinen Strix-Cloud-Account. Bei einem externen
Modell werden Quelltextauszüge, Testzugänge und HTTP-Inhalte an den Provider übertragen.
Strix-Telemetrie ist deaktiviert; das ändert den Datenfluss zum Modell nicht.

Für den ersten Lauf `quick`, USD `10` und `publish_reports: false` verwenden.
Das Budget ist eine nach Modellantworten geprüfte Kostenschätzung mit möglicher
Überschreitung. Zusätzlich gelten 80 Turns **pro Agent**, 45 Minuten Scanzeit und
65 Minuten für den gesamten Job. Der Workflow erlaubt maximal USD 50.

### Isolation und Ergebnisse

- Ein frischer GitHub-Runner baut RailKeeper aus dem ausgewählten Commit.
- Eine eigenständige Compose-Datei verwendet ein internes Docker-Netz, ohne Standardroute
  ins LAN oder Internet. Ein separater Nginx-Zugang leitet ausschließlich an RailKeeper
  weiter. Sein veröffentlichter Port ist vom Runner und vom Strix-Sandbox-Container
  über das Docker-Bridge-Gateway erreichbar. Der Proxy kann Uploadgrenzen und andere
  HTTP-Eigenschaften beeinflussen; der Test prüft diese konkrete Deployment-Variante.
- `/data` und `/tmp` liegen im RAM. Es gibt keine vorhandenen Datenvolumes oder
  Host-Verzeichnisse in der Anwendung. Neustart oder Entfernen verwirft die Daten.
- Der normale Setup-/Login-/CSRF-Ablauf erstellt vier zufällige Testzugänge für
  `admin`, `editor`, `viewer`, `messe`. Bestehende Installationen werden beim Seeding abgelehnt.
- Strix erhält einen separaten Snapshot ausschließlich eingecheckter Dateien,
  keinen Zugriff auf `.git`, lokale `.env`-Dateien oder laufende Arbeitsänderungen.
- Der Scanner benötigt Internetzugriff für sein Modell und seine Werkzeuge. Seine Docker-Sandbox
  ist keine vollständige Netzsperre. Deshalb ist der GitHub-Pilot an einen kurzlebigen
  GitHub-Runner gebunden; ein Runner im privaten LAN ist nicht vorgesehen.
- Es gibt keine automatischen Pushes, Issues, Fix-PRs, Merges oder Gerätezugriffe.

Die Zusammenfassung zeigt nur Abschlussstatus und Befundanzahl. Exitcode `0` bedeutet
abgeschlossen ohne gemeldete Befunde, `2` bedeutet Befunde und `3` bedeutet unvollständig
oder technisch fehlgeschlagen. Budgetabbruch, Timeout, fehlende oder unlesbare Ergebnisse
zählen niemals als bestandene Prüfung. Ein abgeschlossener Lauf beweist keine vollständige
Prüfabdeckung und keine Schwachstellenfreiheit.

**RailKeeper ist öffentlich.** Mit `publish_reports: true` werden Markdown-Berichte als
Artefakte für drei Tage hochgeladen und können für Leser des Repositorys zugänglich sein.
Nur bewusst nach Abwägung aktivieren. Rohlogs, Zugangsdaten-Datei, Agententranskripte und
Quellcode-Snapshot werden nicht hochgeladen. Ohne Upload werden detaillierte Befunde am
Jobende verworfen; vertrauliche Detailanalysen daher auf einer eigenen kurzlebigen VM
ausführen und Ergebnisse privat aufbewahren. Befunde vor einer Veröffentlichung validieren.

### Lokal starten, einschließlich 192.168.4.143

Die vorhandene Installation unter `http://192.168.4.143/` bleibt ein
möglicher separater Testgegenstand.
Ein GitHub-gehosteter Runner kann private LAN-Adressen nicht direkt erreichen. Der Pilot
öffnet dafür keinen Tunnel und stellt keinen öffentlichen Zugang bereit.

Die isolierte Testinstanz kann auch auf diesem Host neben der vorhandenen Installation laufen.
Für einen Scanner auf einer **eigenen, kurzlebigen Linux-VM mit rootful Docker** bereits
vor dem ersten Start die Bridge-Bindung setzen. Für reine Browserprüfungen diese Zeile
weglassen; dann gilt Loopback:

```bash
export STRIX_BIND_IP="$(docker network inspect bridge --format '{{(index .IPAM.Config 0).Gateway}}')"
docker compose -p railkeeper-strix-pilot -f deploy/strix-pilot/compose.yml up --build -d --wait
mkdir -p .strix-pilot
python3 tools/strix_pilot.py seed --url "http://${STRIX_BIND_IP:-127.0.0.1}:18080" --output .strix-pilot/accounts.json
```

Für den Scan auf dieser VM anschließend:

```bash
bash .github/scripts/install_strix.sh "$PWD/.strix-pilot/bin"
# STRIX_LLM und LLM_API_KEY vorher privat in der Shell setzen.
python3 tools/strix_pilot.py scan --strix "$PWD/.strix-pilot/bin/strix" \
  --url http://host.docker.internal:18080 --accounts .strix-pilot/accounts.json \
  --output .strix-pilot/scan --mode quick --budget 10
docker compose -p railkeeper-strix-pilot -f deploy/strix-pilot/compose.yml down --volumes
```

Die lokale Docker-/Firewall-Konfiguration muss Zugriffe auf diese Bridge-Bindung aus dem
LAN blockieren. Auf Rootless-Docker-Hosts können Bridge-Gateway und Portweiterleitung
anders funktionieren; der automatisierte Workflow verwendet rootful Docker auf Ubuntu.

Für einen späteren Test **der vorhandenen Installation** benötigt der Scanner deren genaue
URL und eigene Testkonten. Die Accounts-Datei ist ein privates JSON-Objekt mit Rollen als
Schlüsseln und je `username`/`password` als Werten. `scan --url` akzeptiert dann die reale
HTTP(S)-Origin statt `host.docker.internal`. `seed` ist dafür ausdrücklich ungeeignet.
Vorher echte Zentralen und andere Geräte auf Netzwerkebene vom Testzugriff trennen.
Die Scope-Anweisung allein erzwingt keine Netzisolation. Der hier vorbereitete GitHub-Workflow
testet immer seine eigene Instanz und verwendet `192.168.4.143` nicht automatisch.

### Verifikation

```bash
python3 -m unittest discover -s tools -p test_strix_pilot.py
docker compose -f deploy/strix-pilot/compose.yml config --quiet
bash -n .github/scripts/install_strix.sh
```

## English

The **Strix-Pilot / Strix pilot** workflow is manual only (`workflow_dispatch`). Merge it
to the default branch to expose GitHub's **Run workflow** button. Configure repository
variable `STRIX_LLM` and secret `LLM_API_KEY`, then start with `quick`, USD `10` and
report publication disabled. External model providers receive source excerpts and test
traffic; disabled Strix telemetry does not prevent this model-provider data transfer.

Each run builds the selected commit on a fresh GitHub-hosted Ubuntu runner. RailKeeper
uses an internal Docker network, RAM-backed disposable storage, resource limits and four
random accounts covering admin/editor/viewer/messe. No existing installation or device is
connected. Strix sees a disposable snapshot of committed files and has Internet access;
its sandbox is not a network allowlist. The supplied instructions restrict testing scope.

The pilot caps estimated model spend, per-agent turns and wall time. Model budget estimates
can overshoot; use a provider-side budget too. Exit `0` means completed with no reported
findings, `2` means findings, and `3` means incomplete/error. Budget stops and missing results
never pass. A completed scan is not evidence of full coverage or security.

This repository is public. Enabling `publish_reports` uploads potentially sensitive
Markdown findings for three days. Raw logs, account files, transcripts and source snapshots
are excluded. With uploads disabled, details are discarded when the runner is destroyed.
Use a disposable private VM for confidential detailed assessments and keep results private.

The commands above also start a separate instance on a local host. The default port binds
to loopback; a scanner container needs the explicitly configured Docker bridge binding on
rootful Linux. Validate local firewall isolation before scanning. Private address
`http://192.168.4.143` requires LAN connectivity and dedicated test
credentials; the GitHub workflow always targets its own disposable instance. Never seed an
existing installation or connect real railway controllers to the testing network.
