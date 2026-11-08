# Unabhängige Build- und Restore-Abnahme / Independent build and restore

Eine zweite fachkundige Person führt diese Schritte mit synthetischen Daten und einer eigenen
Testanmeldung aus. Quellprüfung, CI und eine weitere KI-Auswertung ersetzen diesen Nachweis nicht.
Das Bereitstellen dieser Anleitung ist noch keine bestandene menschliche Abnahme.

## Deutsch

### 1. Quelle und Build nachvollziehen

Auf einem eigenen Testrechner mit Git, Node.js 24 und dem Compiler aus `.go-version` beginnen.
Der separat unterstützte Mindestcompiler steht in `backend/go.mod`; beide Prüfwege sind in den
[Qualitätsprüfungen](quality-gates.md) beschrieben. Einen neuen Checkout verwenden:

```bash
git clone --branch main https://github.com/ichwars/RailKeeper.git RailKeeper-qualification
cd RailKeeper-qualification
git rev-parse HEAD
git status --short --branch
npm --prefix frontend ci
npm --prefix frontend run lint
npm --prefix frontend run build
go -C backend test ./...
mkdir -p output/independent-qualification
go -C backend build -o ../output/independent-qualification/railkeeper ./cmd/railkeeper
```

Den tatsächlichen Quellcommit, Betriebssystem, Compiler und Ergebnisse dokumentieren.
Für Windows nativ `tools/build_windows_standalone.ps1` verwenden; ein Cross-Build allein
belegt keine Windows-Laufzeit. Für Docker die [Betriebsanleitung](production-runbook.md) in einer
eigenen VM verwenden. Der Standard-Compose-Name und Port können mit einer laufenden Instanz
kollidieren; keine bestehende Compose-Installation für die Übung wiederverwenden.

### 2. Getrennte Linux-Testinstanz starten

Diese Befehle laufen aus dem eben angelegten Checkout und verwenden einen eigenen Datenpfad:

```bash
umask 077
RAILKEEPER_ADDR=127.0.0.1:18080 \
RAILKEEPER_DATA_DIR="$PWD/output/independent-qualification/source-data" \
RAILKEEPER_MIGRATIONS_DIR="$PWD/backend/migrations" \
RAILKEEPER_SEEDS_DIR="$PWD/backend/seeds" \
RAILKEEPER_STATIC_DIR="$PWD/frontend/dist" \
output/independent-qualification/railkeeper
```

Den Einmalcode ausschließlich lokal aus dem neuen Datenverzeichnis lesen und auf
`http://127.0.0.1:18080` ein eigenes Testkonto einrichten. Keine echte SMTP-Verbindung oder Zentrale
konfigurieren. Mit der [Restore-Übung](site/de/administration/restore-exercise.md) erkennbare
Fahrzeuge, Set, Funktionen, CV-Werte, Bild und Wartung anlegen. Zusätzlich zwei manuelle Bewertungen
mit getrennten Währungen und einen benannten CV-Teilstand mit Originalquelle erfassen.

### 3. Export und tatsächlichen Restore getrennt prüfen

App-Backup außerhalb des Datenverzeichnisses geschützt speichern. IDs, Beziehungen,
Erfassungszeiten, Profilnamen einschließlich Leerzeichen, fehlende gegenüber expliziten Indexseiten,
Bewertungen/Währungen und Dateiinhalte als Prüfpunkte festhalten.

Den gleichen Quellbuild in einem zweiten Terminal mit einem anderen Datenpfad starten:

```bash
umask 077
RAILKEEPER_ADDR=127.0.0.1:18081 \
RAILKEEPER_DATA_DIR="$PWD/output/independent-qualification/restore-data" \
RAILKEEPER_MIGRATIONS_DIR="$PWD/backend/migrations" \
RAILKEEPER_SEEDS_DIR="$PWD/backend/seeds" \
RAILKEEPER_STATIC_DIR="$PWD/frontend/dist" \
output/independent-qualification/railkeeper
```

Auf `http://127.0.0.1:18081` ein eigenes Konto erstellen. Backup vorprüfen, ausschließlich hier
mit `RESTORE` bestätigen und die notierten Werte sowie geöffneten Dateien vergleichen. Erneut
exportieren und auch diesen Export vorprüfen. Das Testkonto der Zielinstallation muss erhalten
bleiben; das Quellkonto wird nicht übernommen. Abgebrochene oder abgewiesene Vorprüfung darf
keine vorhandenen Zieldaten ändern. App-Backup und vollständige Installationssicherung sind
verschiedene Nachweise, siehe [Server-Restore](site/de/administration/server-restore.md).

### 4. Betrieb und Ergebnis ohne Session-Kontext nachvollziehen

Instanzen beenden und mit denselben getrennten Pfaden neu starten. Anmeldung, Bestand und Dateien
erneut prüfen. Für den Upgrade-Nachweis die [Upgrade-Prüfung](production-runbook.md) mit einer frischen
Sicherung und einem getrennten Kandidaten wiederholen. Keine ältere Sicherung über später bestätigte
Eingaben spielen. Diagnose anhand von Version, aktivem Datenpfad und redigiertem Fehlerkontext üben.

Prüfperson, Datum mit Zeitzone, Quellcommit, Plattform, Befehle, erwartete und tatsächliche Ergebnisse
sowie offene Fragen dokumentieren. Blockierende Punkte sind Datenverlust, stille Überschreibung,
falsche Decoderzuordnung, Rollenverletzung oder unbeabsichtigte Gerätebefehle. Zugangsdaten,
Datenbanken, private Backups und unredigierte Logs gehören nicht in Git oder öffentliche Issues.

## English

A second qualified person performs the procedure on a separate test machine using synthetic data
and their own account. Record the exact source commit, platform, compiler, commands, expected and
observed outcomes, and unresolved questions. Providing this document does not establish acceptance.

1. Clone a fresh `main` checkout, record its commit and run the build/test commands above. Use the
   compiler pinned in `.go-version`; the separately tested minimum is in `backend/go.mod`. Windows
   acceptance requires a native run, rather than a cross-build alone.
2. Start the source instance with the first explicit data directory and loopback port. Read its
   local setup code privately and create a test account. Follow the [restore exercise](site/administration/restore-exercise.md),
   adding distinct-currency valuations and a named partial CV snapshot with original provenance.
3. Export an app backup to protected storage outside both data directories. Record IDs, relationships,
   timestamps, exact profile labels, absent versus explicit index pages, currency/value entries and
   actual file contents. Start the second instance with the second directory and port, create its
   own administrator, validate the backup and confirm `RESTORE` only in that separate target.
4. Compare every checkpoint and opened file, then export and validate again. Local target accounts
   remain local. Aborted/rejected preflight must leave target data unchanged. Full installation
   recovery is separate from app restore; follow the [server-restore guide](site/administration/server-restore.md).
5. Restart both instances with the same separate paths. Check login, inventory and files again.
   Practice upgrade and diagnosis using the [runbook](production-runbook.md) and a fresh backup.
   Never replace later confirmed changes with an older backup.

Do not configure real SMTP or command-station writes for this exercise. Report data loss, silent
replacement, incorrect decoder ownership, role violations or unintended device writes as blockers.
Keep credentials, databases, private backups and unredacted logs outside Git and public issues.
