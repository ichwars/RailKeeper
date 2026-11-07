# Modernisierung erproben und ausliefern

Dieser Ablauf gilt für den Entwicklungsstand mit Schema 80 und App-Backupformat 28.
Er definiert die technische Qualifizierung und die verbleibende Anwendererprobung.
Ein zusammengeführter PR ist noch kein veröffentlichter oder produktiv installierter Stand.
Eine Veröffentlichung oder Produktivumstellung erfolgt mit einem eigenen Auftrag.

## Einen eindeutigen Kandidaten qualifizieren

Halte Commit-SHA, Quell- und Zielversion, Binary-Prüfsummen beziehungsweise Image-Digest,
Plattform und Prüfbericht gemeinsam fest. Eine gleiche sichtbare Versionsnummer beweist
keinen identischen Entwicklungsstand. Windows-Nachweise müssen den exakten Kandidaten
betreffen. Native Dockerprüfungen für AMD64 und ARM64 bleiben erforderlich; Emulation
ersetzt keinen nativen Plattformnachweis.

Die [Referenzprüfung](testing-baseline.md) verwendet ausschließlich neue synthetische
Installationen. Ihr Upgrade startet mit der tatsächlichen Vorgängerversion und vergleicht
Dateninhalte, IDs, Beziehungen, CV-Indexseiten, Herkunft, Historien und Datei-Prüfsummen.
SQLite-Integrität und Fremdschlüssel werden zusätzlich geprüft. Der aktuelle Referenzlauf
kann die getrennten Katalog- und Decoderübernahmen einschließen:

Auf `codex-dev` aus dem Repository mit installiertem Go, Node und den
[Browser-Voraussetzungen](testing-baseline.md#synthetic-reference-collection-and-process-checks).
Der Starter legt temporäre Quellen und Browserprofile über `TMPDIR` auf der Datenplatte
ab und räumt sie nach dem Lauf auf. Berichte und Programme bleiben unter `output/` erhalten:

```bash
codex-test-tmp run -- bash -euo pipefail <<'RK_TEST'
rk_previous_source="$(mktemp -d "$TMPDIR/railkeeper-modernization-source.XXXXXX")"
git archive debfad6198299052f07220883c1476581be691cc backend | tar -x -C "$rk_previous_source"
mkdir -p output/regression
go -C "$rk_previous_source/backend" build -o "$PWD/output/regression/railkeeper-v0.1.22" ./cmd/railkeeper
go -C backend build -o "$PWD/output/regression/railkeeper" ./cmd/railkeeper
npm --prefix frontend run build
python3 tools/regression/run_baseline.py \
  --binary output/regression/railkeeper \
  --previous-binary output/regression/railkeeper-v0.1.22 \
  --previous-backend "$rk_previous_source/backend" \
  --catalog-backfill --catalog-authority --catalog-writes \
  --decoder-model --decoder-backfill \
  --browser --output output/regression-modernization
RK_TEST
```

Auf anderen Hosts gilt deren eigene Temp-Konfiguration. Der innere Bash-Block benötigt
ein gesetztes `TMPDIR` auf einem geeigneten Datenträger; der lokale Starter wird nicht
als Bestandteil von RailKeeper oder automatisch auf Windows installiert.

Die Pfade müssen auf die separat gebauten, geprüften Programme und deren zugehörige
Migrationen zeigen. Der Ausgabeordner muss neu sein. Der bestehende Runner nimmt keine
Produktionsadresse, private Sicherung oder vorhandenen Datenordner als Testziel an.
Unter Windows verwendet der Paket-Workflow die echte EXE und einen getrennten
`LOCALAPPDATA`-Pfad. Der Rootless-Docker-Lauf verwendet ausschließlich eigene kurzlebige
Testcontainer. Programmdateien im Image sind unabhängig von der Build-Host-Umask lesbar;
private Laufzeitdaten erhalten dadurch keine zusätzlichen Leserechte.

Prüfberichte, Datenbanken, Screenshots und private Zugänge bleiben außerhalb von Git.
Der vollständige Frontend-Testlauf, Build, API-Vertrag, Backendtests, Sicherheitsprüfungen
und Dokumentationsprüfung ergänzen diese Laufzeitnachweise. Ein Fehler bei Bestand,
Herkunft, Rollen, Dateiinhalt oder bisheriger Bedienung verhindert die Freigabe des Kandidaten.

## Anwendererprobung mit denselben Aufgaben

Ein kleiner Anwenderkreis prüft den qualifizierten Kandidaten auf einer getrennten
Pilotinstallation. Verwende einen synthetischen Bestand oder eine private, ausdrücklich
dafür bereitgestellte Kopie. Kopien erhalten keine Verbindung zu realen Zentralen,
Maildiensten oder Meldediensten. Eine neue Installation erhält ihre eigene Anmeldung;
ein App-Import überträgt keine lokalen Konten. Während der Erprobung bleiben produktive
Installationen unverändert.

| Aufgabe | Erwartetes Ergebnis |
| --- | --- |
| Zwei Fahrzeuge desselben Artikels und ein Set mit individuellen Angaben bearbeiten | Gemeinsamer Artikel und individuelle Abweichungen bleiben unterscheidbar; nicht ausgewählte Exemplare bleiben unverändert. |
| Modell-, Funktions-, CV- und Datei-Eingaben offen lassen, Reiter wechseln und eine unabhängige Aktion bestätigen | Offene Eingaben bleiben erhalten; die Bestätigung nennt genau die gespeicherte Wirkung. Abbrechen verwirft unbestätigte Eingaben. |
| Eine Änderung in einer zweiten Sitzung bestätigen und den alten Entwurf speichern | Keine stille Überschreibung. Aktueller und eigener Stand lassen sich prüfen; unberührte Angaben bleiben erhalten. |
| Einen Teilimport vorbereiten, Zeilen abwählen, abbrechen und erneut übernehmen | Auswahl, bestehende Werte, Quellen und Fehler sind nachvollziehbar; es gibt keine Teilübernahme eines gescheiterten Stapels. |
| Zwei gleich beschriebene Decoder erfassen, Einbau ohne bekanntes Datum und späteren Ausbau dokumentieren | Getrennte Exemplare, unbekannte tatsächliche Daten und Erfassungszeiten bleiben unterscheidbar. |
| Eine Arbeitsgruppe ausdrücklich wählen, CVs vormerken und danach eine andere Voreinstellung bestätigen | Vorgemerkte Änderungen, geöffnete CV-Zeilen und vorbereitete Dateien behalten ihr Ziel. Quellenprofil und Ziel bleiben sichtbar. |
| Während eines vorbereiteten CV-/Dateivorgangs die Decoderzuordnung in einer zweiten Sitzung ändern | Bestätigung wird blockiert. Erneuter Vergleich erhält Eingaben und lädt die aktuelle Zuordnung; er startet keine Speicherung. |
| Einen Teilstand erfassen, Hardwareangaben später ändern und die damalige Quelle herunterladen | Der gespeicherte Zustand und die Originalbytes bleiben erhalten. Ein Teilbestand erscheint nicht als vollständige Auslesung. |
| Export in einer frischen Installation vorprüfen und wiederherstellen | CVs einschließlich unbekannter Werte, Indexseiten, Quellen und Archiv bleiben erhalten; die Anmeldung der Zielinstallation bleibt eigenständig. |
| Dieselben Aufgaben als Viewer und mit der Rolle Messe versuchen | Viewer liest erlaubte Bestandsdaten; Messe erhält keinen allgemeinen Bestands- oder Archivzugriff. Schreibaktionen bleiben geschützt. |

Erfasse je Aufgabe Kandidat, Plattform, Sprache, Rolle, Schritte, erwartetes und tatsächliches
Ergebnis. Ein Bedienabbruch oder Missverständnis ist ein Befund mit Reproduktionsschritten.
Behebe betroffene Alltagsabläufe und wiederhole deren Aufgaben am neuen Kandidaten.
Automatisierte Browserprüfungen können die Aufgaben technisch absichern, ersetzen aber
keine beobachtete Anwendererprobung. Physische Zentralen-, Drucker-, Kamera- und
Scannerabnahmen benötigen die tatsächlich verwendeten Geräte und deren Firmware.
Ohne diesen Nachweis bleiben sie ausdrücklich offen. Reale Geräteschreibzugriffe
benötigen weiterhin eine eigene Bestätigung.

## Produktivumstellung und Rückfallgrenze

Vor einer beauftragten Umstellung werden Zielhost, Startart, aktiver Datenpfad,
Softwarestand und private Konfiguration aktuell geprüft. Sichere zuerst den App-Export
und eine konsistente vollständige Installation. Die vollständige Sicherung umfasst
Datenbank, gespeicherte Dateien, lokale Anmeldung und Konfiguration; bewahre die passende
alte Software getrennt auf. Eine laufende SQLite-Datenbank wird nicht durch Kopieren nur
der DB-Datei gesichert. Erprobe Sicherung und Migration zuerst isoliert.

Für die Umstellung werden Schreibzugriffe pausiert. Nach Sicherung und Migration werden
Erhaltungsbericht, Gesundheit, Anmeldung, Berechtigungen, Bestandsansichten, Dateiabruf
und aktueller Backup-Export geprüft. Erst danach beginnt die normale Bearbeitung.
Der [Betriebsablauf](production-runbook.md) und die
[vollständige Wiederherstellung](site/de/administration/server-restore.md) beschreiben
die Betriebsgrenzen.

Solange keine neue Nutzereingabe bestätigt wurde, kann die geprüfte vollständige alte
Installation gemeinsam mit ihrer alten Software wiederhergestellt werden. Seit der
Sicherung bestätigte Eingaben müssen hingegen erhalten und vollständig übernommen
werden; eine Rücksicherung allein wäre Datenverlust. In diesem Fall aktuellen Zustand
zusätzlich sichern und eine kompatible Korrektur vorwärts einsetzen oder alle Änderungen
nachweislich übertragen. Neue Decoder- und Archivdaten lassen sich nicht verlustfrei in
ein altes Modell oder Backupformat zurückschreiben. Keine Formatnummer manuell ändern
und kein altes Programm ungeprüft auf einer migrierten Datenbank starten.

Release Notes nennen tatsächliche Migrationen, Format 28, erhaltene ältere Importformate,
unabhängige lokale Anmeldung, geprüfte Plattformen und diese Rückfallgrenze. Sie unterscheiden
technische Nachweise, Anwendererprobung und physische Geräteabnahme. Eine neue Version
oder ein Veröffentlichungstermin wird erst im gesonderten Releaseauftrag festgelegt.
