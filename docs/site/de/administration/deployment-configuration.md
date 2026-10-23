---
title: RailKeeper installieren
description: RailKeeper installieren.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
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
| `RAILKEEPER_COOKIE_SECURE` | Bei HTTPS `true`; muss zur tatsächlich benutzten Verbindung passen |
| `RAILKEEPER_TRUSTED_PROXY_CIDRS` | Nur die eigenen vertrauenswürdigen Proxy-Netze |

Die vollständige Konfiguration mit TLS, Uploadgrenzen, SMTP, OCR und Druckern steht im [Betriebshandbuch](https://github.com/ichwars/RailKeeper/blob/main/docs/production-runbook.md). Danach: [erstes Fahrzeug](../guide/quick-start) und [Sicherung](backup-restore).
