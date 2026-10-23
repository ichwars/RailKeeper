---
title: Digitalzentralen: Umfang und Prüfstand
description: Digitalzentralen: Umfang und Prüfstand.
audience: reference
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Digitalzentralen: Umfang und Prüfstand

Diese Übersicht trennt vorhandene Funktionen von tatsächlich dokumentierten Geräteprüfungen.
Sie beschreibt den geprüften Entwicklungsstand **main vom 23.10.2026**. Ältere Releases können abweichen.

## Funktionen je Anschluss

| Anschluss | Lesen und Vergleich | Zur Zentrale schreiben | Lokbilder |
| --- | --- | --- | --- |
| ECoS | Lokstammdaten, statische Funktionen, Vergleichsarbeitsliste; gespeicherter CV-Cache CV1–1024 | Bestätigter Lokdatenabgleich; keine CV-Programmierung | Erstellen, Download sowie bestätigter Upload und Zuordnung |
| CS3 / CS3 Plus | Lokliste mit UID, Name, Adresse und Protokoll; keine CVs | Nicht vorhanden | PNG erstellen und manuell in der CS3 importieren |
| Z21 / z21 | Test, Diagnose und begrenzter Adressabgleich für bekannte RailKeeper-Adressen | Nicht vorhanden | PNG für die Z21-App erstellen; kein Bildspeicher in der Zentrale |
| Intellibox 3 | Test und Diagnose im Z21-kompatiblen UDP-Modus | Nicht vorhanden | PNG für IB3 Configuration erstellen; kein direkter RailKeeper-Upload |

Eine lesbare Antwort oder eine passende Bildgröße belegt noch nicht die erfolgreiche Nutzung
jedes Gerätemodells. Ein Bildexport und ein direkter Upload sind getrennte Fähigkeiten.

## Tatsächlich belegter Gerätestand

| Gerät / Firmware | Beleg | Verbleibende Grenze |
| --- | --- | --- |
| ECoS, Anwendungsfirmware **4.3.3** | Dokumentierte Geräteabnahme des Lokdaten-Schreibablaufs nach PR #143; später vom Nutzer bestätigter Bildexport, Upload und automatische Zuordnung | Der damalige Nachweis benennt keine vollständige Modell-/Hardwarematrix und enthält kein anonymisiertes Rohantwort-Fixture |
| ECoS 4.3.3, belegte Loksteuerung | Einzelner freigegebener Protokollversuch bestätigte Fehler 25 bei belegter Steuerung | Die überarbeitete Reihenfolge mit Reservierung vor Upload, Dublettenerkennung und Pausieren/Fortsetzen besitzt automatisierte beziehungsweise simulierte Abdeckung; zusätzliche Geräteabnahme offen |
| Weitere ECoS-Firmwarestände | Vollständige passende Verbindungsantwort kann als **nicht geräteverifiziert** verwendet werden | Keine pauschale Freigabe aller Firmwarestände |
| CS3, API-Formen vor beziehungsweise ab 2.6 | Parser-/Regressionstests anhand nachgebildeter kompatibler Antwortformen | Keine direkte vollständige Hardwareabnahme durch RailKeeper belegt |
| Z21 / z21 | Protokollbasierte UDP-Fixtures und automatisierte Tests | Keine reale Modell-/Firmwareabnahme dieser Leseabläufe belegt |
| Intellibox 3 | Automatisierte Tests des Z21-kompatiblen Transports | Reale Intellibox-Abnahme offen; LocoNet-over-TCP ist geplant und nicht implementiert |

Die Modellbezeichnungen **ECoS 50200/50210/50220** im Bilddialog nennen das vorgesehene
Ausgabeformat. Sie stellen keine drei getrennten Hardwareabnahmen dar.
Die neuen Handbuchbilder zeigen eine isolierte Demo. Simulierte Zentralenantworten sind im
[Lokbildkapitel](../guide/vehicles/locomotive-images) ausdrücklich gekennzeichnet.

## Was bedeutet der Verbindungsstatus?

| ECoS-Ergebnis | Bedeutung und nächster Schritt |
| --- | --- |
| Nicht erreichbar oder abgewiesen | Adresse, Verbindung vom Server, Port und Meldung prüfen |
| Unvollständig | Benötigte Identifikationsfelder oder Abschluss fehlen; nicht aktivieren |
| Nicht geräteverifiziert | Vollständige Antwort, aber für diese Firmware fehlt die dokumentierte Abnahme |
| Geräteverifiziert | Die Firmware trifft den hinterlegten Abnahmestand; Sonderfälle behalten ihre eigenen Prüfgrenzen |

Ein erfolgreicher Ping oder eine erreichbare Webseite ersetzt den Protokolltest nicht.
ECoS benötigt TCP **15471**, für Bildtransfer zusätzlich HTTP **80**. CS3 nutzt ihren
konfigurierten HTTP-Port; Z21 und Intellibox 3 verwenden standardmäßig UDP **21105**.

## Einen weiteren Hardwarebeleg beitragen

Notiere Modell, Hardwarestand, Firmware, RailKeeper-Version, genutzten Anschluss und den konkreten
Ablauf. Halte erwartetes und tatsächliches Ergebnis sowie Abbruch-/Konfliktfälle fest.
Antwortmitschnitte müssen auf benötigte Felder begrenzt und von persönlichen Daten bereinigt werden.
Für Geräte-Schreibtests ist eine ausdrückliche Freigabe nötig; Verbindungstests beginnen lesend.

## Quellen und offene Arbeiten

- [ECoS-Kompatibilität](https://github.com/ichwars/RailKeeper/blob/main/docs/architecture/ecos-compatibility.md)
- [Bildtransfer und tatsächliche Protokollbelege](https://github.com/ichwars/RailKeeper/blob/main/docs/architecture/ecos-image-transfer.md)
- [CS3-Antwortformen und Grenzen](../guide/import-export/cs3-import)
- [Z21-Leseumfang](https://github.com/ichwars/RailKeeper/blob/main/docs/architecture/z21-read-scope.md)
- [Intellibox-Transportgrenzen](https://github.com/ichwars/RailKeeper/blob/main/docs/architecture/intellibox3-transports.md)
- Offene Integrationsthemen: [ECoS #132](https://github.com/ichwars/RailKeeper/issues/132),
  [Z21 #133](https://github.com/ichwars/RailKeeper/issues/133), [Intellibox 3 #134](https://github.com/ichwars/RailKeeper/issues/134).
