---
title: Lokbilder für die Digitalzentrale
description: Lokbilder für die Digitalzentrale.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Lokbilder für die Digitalzentrale

Ein Foto im Fahrzeugbestand ist oft zu groß oder hat breite Ränder. RailKeeper erstellt daraus eine eigene Datei für die gewählte Zentrale. Dein Original bleibt erhalten.


<GuideFigure name="media" alt="Der Einstieg liegt bei den Bildern unter Uploads; zuerst ein geeignetes Foto hinterlegen." caption="Der Einstieg liegt bei den Bildern unter Uploads; zuerst ein geeignetes Foto hinterlegen. Künstliche Demodaten, keine Werkseinstellungen. Anklicken für die volle Größe." />

## Ein Bild vorbereiten

1. Öffne das gespeicherte Fahrzeug unter **Bearbeiten > Uploads**.
2. Wähle bei einem vorhandenen Fahrzeugbild **Zentralenbild erstellen**.
3. Wähle das Ziel, zum Beispiel **ECoS**.
4. Prüfe den vorgeschlagenen Ausschnitt. Die Lok sollte möglichst die verfügbare Breite nutzen, ohne abgeschnitten oder verzerrt zu werden.
5. Prüfe die Vorschau in tatsächlicher Größe und lade die Datei herunter.

RailKeeper prüft Dateiformat, Auflösung, Ausschnitt und Flächennutzung. Es erkennt nicht zuverlässig, ob das Foto die richtige Lok zeigt, scharf ist oder aus einer geeigneten Perspektive aufgenommen wurde. Diese Prüfung bleibt bei dir.

## Wohin geht die Datei?

| Ziel | Verwendung |
| --- | --- |
| ECoS 50200 / 50210 / 50220 | BMP-Datei für den manuellen Import oder den getrennt bestätigten Direkttransfer |
| CS3 / CS3 plus | PNG für den Bildimport in der Weboberfläche |
| Intellibox 3 | PNG zur weiteren Verarbeitung mit IB3 Configuration |
| Z21-App | PNG für die Bildauswahl der App, kein Protokoll-Upload an die Z21 |

## Direkt an die ECoS übertragen

Nur Admins können **Direkt an ECoS übertragen** verwenden. Wähle die Ziel-Lok und prüfe Name, Decoderadresse, Objekt-ID und Bildvorschau. Bestätige erst danach die Übertragung. RailKeeper kann ein identisches vorhandenes Bild wiederverwenden.

Ist die Loksteuerung belegt, pausiert der Vorgang. Wähle die Lok an der ECoS ab und verwende bewusst **Fortsetzen**. Geänderte Vorschauen müssen neu bestätigt werden. Wiederhole bei unklaren Antworten nicht einfach den Upload, sondern prüfe Bildplatz und Lokzuordnung.

Die ECoS-Verbindung verwendet Port 15471; der Bildtransfer benötigt zusätzlich die Weboberfläche auf Port 80. [Zentrale einrichten](../../administration/digital-centers).
