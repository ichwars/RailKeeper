---
title: Einen Zentralenanschluss einrichten
description: Einen Zentralenanschluss einrichten.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Einen Zentralenanschluss einrichten

Eine Digitalzentrale ist optional. Du kannst Fahrzeuge, Bilder und CV-Dateien auch ohne angeschlossene Hardware verwalten. Für Einrichtung und Zentralenaktionen brauchst du **Admin**.


<GuideFigure name="digital-settings" alt="Konfiguriere die Zentrale, die dein RailKeeper-Server tatsächlich erreichen kann." caption="Konfiguriere die Zentrale, die dein RailKeeper-Server tatsächlich erreichen kann. Künstliche Demodaten, keine Werkseinstellungen. Anklicken für die volle Größe." />

## Verbindung einrichten

1. Öffne **Einstellungen > Digitalzentralen**.
2. Wähle den tatsächlichen Zentralentyp.
3. Trage die IP-Adresse der Zentrale und den vorgesehenen Port ein. Lies die Adresse an der Zentrale ab.
4. Speichere die Einstellungen und führe den Verbindungstest aus.
5. Öffne danach **Digitalzentralen** in der Seitenleiste für die angebotenen Arbeitsabläufe.

Ein erfolgreicher Ping bestätigt noch nicht, dass RailKeeper die benötigten Dienste erreichen kann. Entscheidend ist die Verbindung vom RailKeeper-Server zur Zentrale, nicht nur vom eigenen Browser.

## Was unterstützt welcher Anschluss?

| Zentrale | Aktueller Umfang |
| --- | --- |
| ESU ECoS | Lokdaten vergleichen, kontrolliert übernehmen, gespeicherte CVs lesen und bestätigte Schreibabläufe; gesonderter Bildtransfer |
| Märklin CS3 | Lokdaten lesen und vergleichen; kein allgemeiner Schreibabgleich |
| Z21 | Verbindungstest, Diagnose und begrenzter Abgleich bereits bekannter Adressen; keine vollständige Lokliste oder CVs |
| Intellibox 3 | Z21-kompatibler Verbindungstest und Diagnose; kein vollständiger Loklistenimport |

## ECoS: die zwei Anschlüsse nicht verwechseln

- **Port 15471:** TCP-Schnittstelle für RailKeeper. Dies ist keine Webseite für den Browser.
- **Port 80:** Weboberfläche und Bildtransfer. Im Browser normalerweise `http://IP-DEINER-ECOS/` öffnen, ohne `:15471`.

Bei getrennten Netzen müssen Route, Netzmaske, Gateway und gegebenenfalls Firewallfreigaben passen. Ändere nicht auf Verdacht die Zentraleinstellungen, wenn nur eine Anwendung die Verbindung nicht erreicht.

## Nächste Schritte

- [Lokdaten und gespeicherte CVs mit ECoS übernehmen](../guide/import-export/ecos-sync)
- [Lokbilder vorbereiten und bestätigt übertragen](../guide/vehicles/locomotive-images)
- [CS3-Lokdaten importieren](../guide/import-export/cs3-import)

[Kompatibilität und tatsächliche Gerätebelege](../reference/command-station-compatibility) · [Z21-Adressabgleich](../guide/import-export/z21-compare).
