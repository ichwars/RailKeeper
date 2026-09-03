---
title: Installation und Administration
description: RailKeeper installieren, konfigurieren, absichern und betreiben.
audience: admin
status: stable
reviewedVersion: 0.1.20.4
lastReviewed: 2026-08-16
---

# Installation und Administration

Dieser Bereich behandelt Windows Standalone, Docker, Laufzeitkonfiguration, Benutzer und Rollen,
SMTP, Sicherung und Wiederherstellung, Updates, TLS, Uploads, OCR, Drucker,
Betriebsprüfungen und konservative Fehlerbehebung.

Die Administrationsanleitungen beschreiben die stabile Laufzeit v0.1.20.3 und erhalten das lokale,
selbst gehostete Sicherheitsmodell von RailKeeper.

## Administrative Abläufe

- [Stammdaten-Administration](./master-data) behandelt kontrollierte Werte, Lebenszyklus,
  Lagerorte, Inventarnummernschemata und den JSON-Transfer.

## Sichere Updates unter Windows

Das ZIP-Paket von Windows Standalone enthält ausschließlich die Anwendung. Datenbank, Uploads,
Anhänge, Vorschaubilder und Sicherungen sind niemals enthalten. Dauerhafte Daten liegen
standardmäßig unabhängig vom austauschbaren Programmordner unter
`%LOCALAPPDATA%\RailKeeper\data`.

Nach erfolgreicher Prüfung unter **Einstellungen > Allgemein > Updates** zeigt Windows Standalone
**Version X herunterladen** nur an, wenn das passende ZIP zum erkannten GitHub-Release gehört. Der
Browser lädt dieses ZIP direkt herunter. RailKeeper entpackt, installiert, ersetzt, migriert oder
startet dabei nichts neu. Fehlt das vertrauenswürdige Paket, bleibt die GitHub-Release-Seite
verlinkt.

So wird RailKeeper aktualisiert:

1. Das passende ZIP über die Schaltfläche mit Versionsangabe oder die Release-Seite herunterladen.
2. Eine aktuelle Anwendungssicherung erstellen und RailKeeper beenden.
3. Das ZIP in einen neuen Programmordner entpacken. Keine Datenbank in diesen Ordner kopieren.
4. Die neue `RailKeeper.exe` starten.
5. Anmelden, unter **Einstellungen > Datenspeicher** den aktiven Speicherort prüfen und den Bestand
   kontrollieren, bevor der bisherige Programmordner gelöscht wird.

Ältere Standalone-Versionen speicherten Daten in einem Ordner `data` neben `RailKeeper.exe`. Beim
ersten Start kopiert RailKeeper diese Altdaten an den sicheren Speicherort und lässt die Quelle
unverändert. Enthalten beide Orte bereits eine Datenbank, beendet RailKeeper den Start, bevor eine
der Datenbanken geöffnet oder migriert wird, und zeigt beide Pfade an. RailKeeper dann beenden,
getrennte Kopien beider Ordner erstellen und die aktuelle Datenbank bestimmen. Den vorhandenen
sicheren Ordner umbenennen, statt ihn zu überschreiben oder Ordner zusammenzuführen. Anschließend
den gewählten vollständigen Datenordner nach `%LOCALAPPDATA%\RailKeeper\data` kopieren. Beide
Quellkopien aufbewahren, bis Bestand und Anhänge geprüft wurden.

Ein ausdrücklich konfiguriertes `RAILKEEPER_DATA_DIR` hat immer Vorrang und deaktiviert die
automatische Altdatenübernahme. Der Pfad darf nicht auf einen austauschbaren Programmordner oder
einen unzuverlässigen Wechseldatenträger zeigen. Administratoren sehen den exakten aktiven Pfad
unter **Einstellungen > Datenspeicher**. Das Öffnen im Explorer steht nur bei einer lokalen Windows
Standalone-Instanz zur Verfügung.

Vor ausstehenden Datenbankmigrationen erstellt RailKeeper unter `safety-backups` eine geprüfte,
private Kopie. Sie enthält die vollständige Datenbank einschließlich lokaler Anmeldedaten. Diese
Startabsicherung ersetzt weder eine Anwendungssicherung noch eine Datei- oder Volumensicherung.

## Docker-Updates

Docker bewahrt dauerhafte Daten im eingebundenen Volume `/data` auf. Die Anwendung wird so
aktualisiert:

```powershell
docker compose pull
docker compose up -d
```

Das Volume vor dem Update sichern. Anschließend `/health`, Anmeldung, Bestand und die
Sicherungsprüfung kontrollieren. Die automatische Datenbankkopie vor einer Migration liegt
ebenfalls innerhalb von `/data` und schützt daher nicht vor dem Verlust des Volumes selbst.

## Lebenszyklus der Stammdaten

RailKeeper unterscheidet zwischen mitgelieferten Einträgen und eigenen Einträgen, die ein
Administrator oder Editor angelegt hat. Mitgelieferte Einträge lassen sich bearbeiten und
deaktivieren, aber nicht endgültig löschen. Ein ungenutzter eigener Eintrag kann endgültig gelöscht
werden. Sobald ein eigener Eintrag verwendet wird, lässt er sich nur noch deaktivieren.

Deaktivierte Einträge werden für neue Fahrzeuge, Zubehörartikel oder andere Datensätze nicht mehr
angeboten. Bestehende Datensätze behalten ihren gespeicherten Wert und zeigen ihn weiterhin als
inaktiv an. Eine Deaktivierung macht historische Bestandsdaten daher nicht ungültig.

Bearbeitungen und Deaktivierungen bleiben nach einem Neustart, einem Abgleich der mitgelieferten
Stammdaten, einem Update, dem Export und Import von Stammdaten sowie einer Sicherung und
Wiederherstellung der Anwendung erhalten. Vor umfangreichen Stammdatenänderungen wird dennoch eine
aktuelle RailKeeper-Sicherung empfohlen.

Die vollständigen Feld-, Berechtigungs-, Nummerierungs-, Lagerort- und Importregeln beschreibt die
[Stammdaten-Administration](./master-data).
