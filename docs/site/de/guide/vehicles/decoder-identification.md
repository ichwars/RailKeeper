---
title: Den richtigen Decoder erkennen
description: Den richtigen Decoder erkennen.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Den richtigen Decoder erkennen

**Ziel:** Prüfen, welche Erklärungen zu deinen gespeicherten Werten passen. Du brauchst ein
bereits gespeichertes Testfahrzeug und Admin oder Editor für den Import.

## 1. Eine verlässliche Ausgangsdatei wählen

Für den ersten Versuch verwende [die CV-Übungsdatei](cv-import-example). Für die anschließende
Funktionsübung gibt es [eine zweite Datei mit F0, F1 und F3](functions-walkthrough).
Beide Dateien sind künstlich und keine Programmierempfehlung für einen Decoder.

Öffne **Fahrzeugbestand > Bearbeiten > CV > Dateien & Import**. Importiere die unveränderte TXT.
Prüfe Profil und Firmware in der Vorschau, bevor du die Werte übernimmst. Eine ESUX-Datei ist
für diesen Versuch nicht nötig; sie wird getrennt als Projektdatei aufbewahrt.

## 2. Erkennung und Quelle prüfen

<GuideFigure name="decoder-identification" alt="CV-Ansicht mit erkanntem ESU-Profil der künstlichen F0-/F1-/F3-Datei." caption="Die echte Demo zeigt das Profil zu den importierten Werten. Der Dateikopf benennt LokPilot 5 micro DCC und Firmware 5.14.183; dies sind Übungsangaben." />

Prüfe unter **CV** das gewählte Profil. Bei mehreren Profilen wähle den passenden Bestand.
Öffne die Herkunft von CV7 und CV8 über einen CV-Quellenlink oder **Alle CV-Werte**.

| Angabe | Was sie belegt |
| --- | --- |
| CV8 = 151 | ESU als Hersteller, noch kein bestimmtes ESU-Modell |
| `LokPilot 5 micro DCC` im Dateikopf | Profilname dieser Datei |
| `5.14.183` im Dateikopf | vollständige Versionsangabe der Datei |
| CV7 = 255 | Rohwert; daraus folgt allein keine Firmwareversion |
| CV31/CV32 | Kontext eines indizierten Wertes, keine Modellkennung |

Die Dateiangaben beschreiben den Export. Gleiche das genaue Modell bei Zweifeln mit Decoder,
LokProgrammer oder Originalunterlagen ab. Ein selbst umbenannter Profiltext ersetzt diesen Nachweis nicht.

## 3. Eine Erklärung nachschlagen

Öffne **Einstellungen > Daten > Allgemeine Stammdaten > Decoderbibliothek**. Suche nach dem Modell
oder einer CV. Prüfe **Prüfumfang**, **Erkennung**, **CV-Definitionen** und **Quellen**.
Modell, Firmware und Indexseite müssen zur jeweiligen Regel passen.

Eine Herstellerreferenz kann eine Tabellenzeile belegen, ohne dass RailKeeper daraus bereits
Adresse, Bremszustand oder Funktionstaste berechnet. Ein Anschluss allein bestimmt ebenfalls kein Modell.

## 4. Unbekannte und widersprüchliche Angaben behandeln

| Beobachtung | Vorgehen |
| --- | --- |
| Hersteller erkannt, Modell unbekannt | Genaues Modell am Fahrzeug prüfen; keine ähnliche Variante auf Verdacht wählen |
| Hersteller und Profil widersprechen sich | Importquelle und Fahrzeug prüfen; bis zur Klärung keine Modellregel erzwingen |
| Firmware fehlt | Originalexport prüfen; keine Version aus CV7 erfinden |
| CV vorhanden, Erklärung fehlt | Modell-, Firmware- und Indexgrenzen in der Quelle nachsehen |
| ECoS-Wert hat keine Indexseite | Als unindizierten Wert erhalten; die Bibliothek rekonstruiert keine fehlende Seite |

Bei neuen Bibliothekserklärungen werden gespeicherte Werte beim Anzeigen erneut ausgewertet.
Ein Neuimport ist dafür normalerweise nicht nötig. Er ist sinnvoll, wenn die ursprüngliche
Datei zusätzliche, bisher nicht gespeicherte Werte oder Herkunftsangaben enthält.

## 5. Fahrzeugangaben bewusst übernehmen

Unter **Modell > Digital** vergleiche die gespeicherten Fahrzeugangaben mit dem CV-Ergebnis.
**Aus CVs übernehmen** füllt die belegbaren Angaben in den Entwurf. Prüfe das Ergebnis und wähle
**Änderungen speichern**. DT für Doppeltraktion bleibt von der DCC-Verbundadresse getrennt.

Weiter: [Funktionstasten gemeinsam übernehmen und schützen](functions-walkthrough).
Die [technische Referenz](decoder-cv-reference) erklärt Formate und Sonderfälle.
