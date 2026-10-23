---
title: CV-Import mit Indexseiten ausprobieren
description: CV-Import mit Indexseiten ausprobieren.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# CV-Import mit Indexseiten ausprobieren

**Ziel:** Zwölf künstliche CV-Werte importieren und prüfen, dass gleiche CV-Nummern auf
verschiedenen Indexseiten getrennt bleiben. Du brauchst ein gespeichertes Testfahrzeug und
**Admin** oder **Editor**. Die Anleitung sendet keine Werte an einen Decoder.

<GuideDownload file="cv-index-demo.txt">CV-Beispiel als TXT herunterladen</GuideDownload>

Die Datei verwendet das LokProgrammer-TXT-Format, ist aber **künstlich erstellt**.
Ihre Werte sind keine Werkseinstellungen und keine Empfehlung zum Programmieren einer Lok.
Für ein übersichtliches Ergebnis beginne mit einem Fahrzeug ohne gespeicherte CVs.

## 1. Liste auswählen und Vorschau prüfen

1. Öffne das Testfahrzeug unter **Bearbeiten > CV**.
2. Wähle **CV-Liste importieren** und die unveränderte Datei `cv-index-demo.txt`.
3. Prüfe die Vorschau: **12 neue Werte**, keine ungültigen oder doppelten Einträge.
4. Als Decoderprofil wird **LokPilot 5 micro DCC** aus dem Dateikopf übernommen.
   Firmware und CV8 sind in dieser Übungsdatei ebenfalls vorgegeben.

<GuideFigure name="import-cv-preview" alt="CV-Importvorschau mit zwölf neuen Werten und zwei Indexseiten." caption="Die tatsächliche Importvorschau der künstlichen TXT-Datei. Gleiche CV-Nummern auf unterschiedlichen Seiten sind keine Dubletten." />

Du benötigst für dieses Beispiel keine ESUX-Datei. **Decoderdatei hochladen** ist eine getrennte
Aktion zum Hinterlegen eines Projekts; sie ersetzt die Werteliste nicht.

## 2. Übernehmen und die Indexseiten kontrollieren

Wähle **Ausgewählte Felder übernehmen** in der CV-Vorschau. Die Übernahme speichert die
markierten CVs am Fahrzeug. Öffne danach **Alle CV-Werte** und lass **Alle Indexseiten** ausgewählt.

<GuideFigure name="import-cv-values" alt="Gespeicherte CV-Werte mit getrennten Zeilen für CV257 auf zwei Indexseiten." caption="Zwölf Werte bleiben erhalten. CV257 und CV258 kommen auf beiden Seiten getrennt vor. Künstliche Demodaten." />

| Kontrolle | Erwartetes Ergebnis |
| --- | --- |
| Zahl der gespeicherten Werte | 12 |
| Basis-CVs | 8 Werte, ohne Indexseite |
| Indexseite `16/0` | CV257 = 12, CV258 = 0 |
| Indexseite `16/1` | CV257 = 44, CV258 = 1 |
| Exportdatum | 23.10.2026, 12:00:00, ohne erfundene Zeitzone |

`16/0` bedeutet **CV31=16, CV32=0**. Der TXT-Index 4096 ergibt sich daraus.
Für `16/1` ist der TXT-Index 4097. Die CV-Nummer allein genügt deshalb nicht als Identität.

## 3. Herkunft nachsehen

Öffne bei CV257 auf Seite `16/0` die Aktion **CV 257 erklären**.
Kontrolliere Wert, Seite und Herkunft. In der unveränderten Datei steht dieser Wert in
**Originalzeile 14**. CV257 auf Seite `16/1` stammt aus **Originalzeile 17**.

<GuideFigure name="import-cv-provenance" alt="Detailansicht von CV257 mit Indexseite, Exportdatum und Originalzeile." caption="Der Herkunftsnachweis bezieht sich auf die mitgelieferte künstliche Beispiel-TXT, nicht auf einen echten Decoderabruf." />

Eine fehlende Erklärung bedeutet nicht, dass der Rohwert fehlt. Eine Bibliotheksregel braucht
passende Modell-, Firmware- und Indexangaben. Der TXT-Export enthält keine ausgelesene
Decoder-Seriennummer, nur weil ein Profilname vorhanden ist.

## 4. Dieselbe Datei erneut prüfen

Importiere dieselbe Datei nochmals unverändert. Erwartet werden **12 bereits gleiche Werte**;
diese sind nicht automatisch zur Änderung ausgewählt. Wähle **Verwerfen**.
Der Bestand bleibt bei zwölf Werten. Du musst zum Ausprobieren keine vorhandenen CVs löschen.

Nach einem eigenen abweichenden Wert kann die Vorschau stattdessen **geändert** melden.
Prüfe dann den Vergleich und wähle bewusst aus, was übernommen werden soll.
Der Schutz manuell gepflegter Funktionstasten ist ein eigener Ablauf unter
[Steuerung](decoder-cv#digitalfunktionen-f0-f31-zuordnen).

[Dateien und Importregeln](decoder-cv#dateien-fur-einen-sauberen-import) · [Bestandsdateien ausprobieren](../import-export/examples)
