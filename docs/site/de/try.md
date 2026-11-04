---
title: RailKeeper ausprobieren
description: RailKeeper mit Windows oder Docker starten, Beispieldaten importieren und ein eigenes Fahrzeug mit Bild anlegen.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-11-04
---

# RailKeeper ausprobieren

**Ziel:** Zwei Beispielfahrzeuge importieren, ein drittes anlegen und ein Bild hinzufügen.
Du brauchst einen Windows-PC oder Docker, einen Browser und einen leeren Testbestand.
Eine Digitalzentrale und echte Sammlungsdaten sind für diese Übung nicht nötig.

## 1. Einen Startweg wählen

### Windows

1. Öffne das [aktuelle Release](https://github.com/ichwars/RailKeeper/releases/latest).
   Lade unter **Assets** das ZIP für Windows x64 herunter und entpacke es vollständig.
2. Öffne im entpackten Ordner ein PowerShell-Fenster. Starte mit einem eigenen Demo-Datenordner:

```powershell
$env:RAILKEEPER_DATA_DIR = Join-Path $env:LOCALAPPDATA 'RailKeeper-Demo\data'
.\start-railkeeper.bat
```

3. Lass das Programmfenster geöffnet. Der Browser öffnet sich automatisch; alternativ verwende
   die im Fenster angezeigte lokale Adresse.

RailKeeper benötigt unter Windows keine zusätzliche Laufzeitumgebung. Die Beispieldaten bleiben
im Ordner `%LOCALAPPDATA%\RailKeeper-Demo\data` erhalten. Ein neuer Programmordner allein
legt keinen neuen Bestand an. [Windows-Betrieb und Updates](administration/deployment-configuration).

### Docker

Mit laufendem Docker diesen Befehl im Terminal ausführen:

```bash
docker run -d --name railkeeper-demo --publish 127.0.0.1:18080:8080 --mount source=railkeeper_demo,target=/data ghcr.io/ichwars/railkeeper:latest
```

Öffne `http://localhost:18080` auf demselben Rechner.
Der Container und das Volume heißen bewusst anders als eine reguläre Installation.
Der Start ist nur über die lokale Adresse erreichbar. Auf einem entfernten Heimserver
verwende die [Installationsanleitung](administration/deployment-configuration).

Nach dem Versuch stoppt `docker stop railkeeper-demo` das Programm und erhält die Daten.
Mit `docker start railkeeper-demo` startest du denselben Bestand erneut.
Die Images unterstützen AMD64 und ARM64; ein Raspberry Pi braucht ein 64-Bit-Betriebssystem.

## 2. Dein lokales Konto anlegen

Beim ersten Aufruf erscheint die Ersteinrichtung. Lege dein eigenes Administratorkonto an
und melde dich an. Es gibt kein Standardpasswort. Für diesen lokalen Versuch brauchst du
keinen Mailserver. Prüfe unter **Einstellungen > Allgemein** den Datenordner, bevor du fortfährst.

## 3. Die Beispieldaten verwenden {#beispieldaten}

Lade diese beiden Dateien herunter. Alle Fahrzeuge und Bildinhalte sind künstliche Beispiele.

| Datei | Verwendung |
| --- | --- |
| <GuideDownload file="vehicle-import-demo.json">Zwei Beispielfahrzeuge als JSON</GuideDownload> | Wird als Fahrzeug-Import eingelesen, nicht als Anwendungssicherung. |
| <GuideDownload file="locomotive-demo.png">Bild einer Beispiel-Lok</GuideDownload> | Zum Hochladen am eigenen dritten Fahrzeug. |

1. Öffne **Import/Export > Profil anlegen**. Name: **Erster Versuch**, Richtung **Import**,
   Bereiche **Fahrzeuge**, Format **JSON**. Speichere mit **Profil anlegen**.
2. Wähle **Neuer Import**, dieses Profil und die heruntergeladene JSON-Datei.
3. Prüfe die Vorschau: `DEMO-JSON-001` und `DEMO-JSON-002`. Bestätige die Übernahme.
4. Öffne **Fahrzeugbestand** und suche nach `DEMO-JSON`. Beide Fahrzeuge sollten erscheinen.

Der Import legt Datensätze an. Verwende den leeren Demo-Bestand aus Schritt 1.
Bei wiederholtem Import melden dieselben Inventarnummern einen Konflikt; du kannst ihn abbrechen.
[Ausführliche Importübung mit CSV und Konflikten](guide/import-export/examples).

## 4. Ein eigenes Fahrzeug mit Bild anlegen

1. Wähle unter **Fahrzeugbestand** die Schaltfläche **Neues Fahrzeug**.
2. Wähle **Einzelfahrzeug** und trage **Meine erste Demo-Lok** als Bezeichnung ein.
   Wähle einen Hersteller, eine Spurweite und die Kategorie **Lokomotive**.
3. Wähle **Weiter zu Artikeldaten** und anschließend ohne Suche **Weiter zu Fahrzeugdetails**.
   Wähle dort eine Gattung, zum Beispiel **Diesellokomotive**, und bestätige mit **Anlegen**.
4. Der Fahrzeugeditor öffnet sich. Wähle **Uploads > Bild hochladen**, dann `locomotive-demo.png`
   und speichere die Änderungen.
5. Schließe den Editor über das **X**. Suche nach **Meine erste Demo-Lok** und öffne sie erneut.
   Das Bild sollte erhalten bleiben.

<GuideFigure name="vehicles" alt="Fahrzeugbestand mit Suchfeld und Schaltfläche Neues Fahrzeug." caption="Hier findest du deine importierten und selbst angelegten Fahrzeuge wieder. Die Aufnahme zeigt einen anderen künstlichen Demobestand." />

## 5. Ergebnis prüfen und Erfahrung teilen {#feedback}

- [ ] Die Ersteinrichtung und Anmeldung funktionieren.
- [ ] Beide importierten Fahrzeuge sind auffindbar.
- [ ] Die eigene Demo-Lok und ihr Bild sind nach erneutem Öffnen vorhanden.

Teile deine Erfahrung in den [Discussions](https://github.com/ichwars/RailKeeper/discussions/categories/general).
Diese Angaben helfen, auch wenn etwas nicht geklappt hat:

```text
System und RailKeeper-Version:
Bisherige Bestandsverwaltung:
Das hat funktioniert:
Hier bin ich hängen geblieben:
Würde ich RailKeeper weiterverwenden? Warum?
```

Die Version steht unter **Einstellungen > Allgemein > Updates**. Beschreibe bei einem Fehler
den letzten Schritt und die Fehlermeldung. Verwende im Bericht deine Demo-Daten und lasse
Passwörter, Zugangsdaten und persönliche Sammlungsdaten weg.

RailKeeper hilft dir? Mit einem [Stern auf GitHub](https://github.com/ichwars/RailKeeper)
unterstützt du das Projekt. Für die nächsten Aufgaben geht es zum [Handbuch](guide/)
oder zur Anleitung für [dein erstes Fahrzeug](guide/quick-start).
