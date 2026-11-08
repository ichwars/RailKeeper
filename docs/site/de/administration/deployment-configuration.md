---
title: RailKeeper installieren
description: RailKeeper installieren.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-11-01
---

# RailKeeper installieren

Wähle zuerst, wo dein Bestand liegen soll: auf einem Windows-PC oder auf einem eigenen Server. RailKeeper benötigt keinen Cloud-Dienst. Ein Browser greift auf die laufende Anwendung zu.

## Windows: ohne zusätzliche Installation

1. Öffne die [GitHub-Releases](https://github.com/ichwars/RailKeeper/releases) und lade das Windows-x64-ZIP der gewünschten Version herunter.
2. Entpacke das gesamte ZIP in einen eigenen Programmordner.
3. Starte `start-railkeeper.bat`.
4. Lege beim ersten Aufruf deinen Admin an. Es gibt kein Standardpasswort.
5. Prüfe unter **Einstellungen > Allgemein > Speichernutzung** den aktiven Datenordner.

Der Standard ist `%LOCALAPPDATA%\RailKeeper\data`. Er liegt außerhalb des austauschbaren Programmordners. Bei einem alten `data`-Ordner neben der EXE übernimmt RailKeeper die Daten nur nach den [Sicherheitsregeln für Updates](system-operations). Vermische keine Datenbanken.

## Docker: auf deinem eigenen Rechner oder Server

Voraussetzung sind Git und Docker Compose. Die [Compose-Datei](https://github.com/ichwars/RailKeeper/blob/main/docker-compose.yml) bindet den Dienst zunächst nur an den eigenen Rechner.

Docker-Builds unterstützen AMD64 und ARM64, auch Raspberry Pi mit einem 64-Bit-ARM-Betriebssystem.
32-Bit-ARM wird nicht unterstützt. Docker wählt die Architektur automatisch; entferne auf ARM64
ein gesetztes `platform: linux/amd64`. Ab v0.1.22 enthalten die stabilen Images sowohl unter
`latest` als auch unter versionierten Release-Tags ARM64. Ältere Release-Tags bleiben unverändert.

```sh
git clone https://github.com/ichwars/RailKeeper.git
cd RailKeeper
docker compose pull
docker compose up -d
```

Öffne auf diesem Rechner `http://localhost:8080` und richte den Admin ein. Dauerhafte Daten liegen im Volume `railkeeper_data`, im Container unter `/data`. Lösche dieses Volume nicht beim Austausch der Anwendung.

## Zugriff von anderen Geräten

Für das Heimnetz oder einen Reverse Proxy muss der Betreiber die Erreichbarkeit ausdrücklich konfigurieren. Verwende für Netzwerkzugriff HTTPS und die passende öffentliche URL. Sichere Cookies benötigen HTTPS; bei einem bewusst nur lokalen HTTP-Test muss die Cookie-Einstellung dazu passen. Deaktiviere keine Anmeldung oder Rollenprüfung, um ein Verbindungsproblem zu umgehen.

| Einstellung | Bedeutung |
| --- | --- |
| `RAILKEEPER_DATA_DIR` | Dauerhafter Datenordner; kein austauschbarer Programmordner |
| `RAILKEEPER_PUBLIC_URL` | Adresse, unter der Benutzer RailKeeper erreichen, auch für Rücksetzlinks |
| `RAILKEEPER_COOKIE_SECURE` | Standard `auto`: sichere Cookies bei konfigurierter öffentlicher HTTPS-URL; `true`/`false` bleiben explizite Vorgaben |
| `RAILKEEPER_LOGIN_PARALLEL` | Aktive Passwortprüfungen, Standard 2, Bereich 1–4; je Vorgang bis 256 MiB einplanen |
| `RAILKEEPER_LOGIN_QUEUE` | Begrenzte Warteplätze, Standard 8, Bereich 0–64 |
| `RAILKEEPER_LOGIN_WAIT_MS` | Maximale Wartezeit, Standard 1000 ms, Bereich 0–2000; 0 weist sofort ab |
| `RAILKEEPER_LOGIN_MEMORY_MB` | Budget aktiver Passwortarbeit, Standard 512 MiB, Bereich 256–1024; mindestens Parallelität × 256, zuzüglich Prozess-/GC-Reserve |
| `RAILKEEPER_TRUSTED_PROXY_CIDRS` | Nur die eigenen vertrauenswürdigen Proxy-Netze |

Die vollständige Konfiguration mit TLS, Uploadgrenzen, SMTP, OCR und Druckern steht im [Betriebshandbuch](https://github.com/ichwars/RailKeeper/blob/main/docs/production-runbook.md). Danach: [erstes Fahrzeug](../guide/quick-start) und [Sicherung](backup-restore).

## Erststart im Browser prüfen

<GuideFigure name="installation-first-admin" alt="Ersteinrichtung einer tatsächlich leeren lokalen Installation." caption="Die echte leere Demo verlangt ein eigenes erstes Administratorkonto. Sie zeigt den Browserablauf, keine Windows-Installation oder Docker-Inbetriebnahme." />

Lies zuerst den einmaligen Code aus `setup-code.txt` im lokalen Datenordner des Servers.
Bei Docker: `docker compose exec railkeeper cat /data/setup-code.txt`. Unter Windows
Standalone liegt die Datei standardmäßig unter `%LOCALAPPDATA%\RailKeeper\data`.
Ein eigener konfigurierter Datenpfad hat Vorrang. Der Code erscheint nicht in der
öffentlichen Status-API, Backups oder Diagnoseexporten.

Trage den Code, einen eigenen Benutzernamen, eine gültige E-Mail-Adresse und ein Passwort
mit mindestens zwölf Zeichen samt Wiederholung ein. Der Code wird nur bei erfolgreicher
Einrichtung verbraucht. Bereits eingerichtete Installationen behalten ihre Anmeldung.
Prüfe danach die Anmeldung und unter **Einstellungen > Allgemein > Speichernutzung** den Datenpfad.

[Konten und Zwei-Faktor-Anmeldung einrichten](account-security-example) · [SMTP und Rücksetzlinks](smtp-password-reset) · [Update-Schrittfolge](system-operations).

## Anmeldeverfügbarkeit und Browserbelege

Nach vollständiger Anmeldung erhält der Browser einen zufälligen, HttpOnly-geschützten
Beleg für ein eigenes Anmeldebudget (höchstens 30 Tage, zehn Browser pro Konto).
Passwort und gegebenenfalls Zwei-Faktor-Code bleiben bei jeder Anmeldung erforderlich.
Unbekannte Browser teilen das bisherige Kontobudget von zehn Versuchen in fünf Minuten.
Ein unbekannter Angreifer kann das eigene Budget eines bekannten Browsers nicht verbrauchen.
IP-Limit und globale Kapazitätsgrenzen gelten weiterhin, auch bei gemeinsamer NAT-Adresse.

Unter **Einstellungen → Authentifizierung → Benutzerverwaltung → Anmeldung freigeben** kann
ein angemeldeter Administrator die Kontobudgets zurücksetzen und Browserbelege widerrufen.
Die Aktion wird protokolliert; Passwort, Zwei-Faktor-Schutz und bestehende Sitzungen bleiben
erhalten. Konto-, Passwort-, Rollen- und Zwei-Faktor-Änderungen widerrufen die Belege ebenfalls.
App-Backups enthalten diese lokalen Authentifizierungsdaten nicht. Neue Browser und
andauernde Überlast bleiben durch die vorhandenen Limits begrenzt.
