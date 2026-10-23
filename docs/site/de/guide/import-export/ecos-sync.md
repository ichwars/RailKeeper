---
title: ECoS-Loks und gespeicherte CVs übernehmen
description: ECoS-Loks und gespeicherte CVs übernehmen.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# ECoS-Loks und gespeicherte CVs übernehmen

**Ziel:** Lokdaten aus der ECoS mit deinem Bestand vergleichen und gespeicherte CV-Werte übernehmen. Der Ablauf benötigt Admin-Rechte und eine [eingerichtete ECoS-Verbindung](../../administration/digital-centers).


<GuideFigure name="central-workspace" alt="Digitalzentralen ist ein eigener Arbeitsbereich; die Demo hat keine Zentrale verbunden." caption="Digitalzentralen ist ein eigener Arbeitsbereich; die Demo hat keine Zentrale verbunden. Künstliche Demodaten, keine Werkseinstellungen. Anklicken für die volle Größe." />

## 1. Lokliste lesen und die richtige Lok wählen

1. Öffne **Digitalzentralen** in der Seitenleiste.
2. Prüfe die aktive ECoS und wähle **Daten lesen**.
3. Wähle die gewünschte Lok in der Arbeitsliste.
4. Vergleiche Name, Adresse, Protokoll und ECoS-Objekt-ID mit dem tatsächlichen Fahrzeug.

Eine gespeicherte externe Zuordnung ist ein wichtiger Anhaltspunkt. Adresse und Name allein beweisen keine Identität. Prüfe vorgeschlagene RailKeeper-Treffer vor jeder Übernahme. Die Vergleichsarbeitsliste liegt auf dem Server und kann später erneut geöffnet werden.

## 2. Alle gespeicherten CVs lesen

Öffne den Bereich **Gespeicherte CV-Werte der ECoS** und starte **Alle gespeicherten CVs lesen**. RailKeeper fragt CV1 bis CV1024 in kleinen Gruppen ab und zeigt den Fortschritt. Du kannst stoppen und fortsetzen. Das ist kein erneutes Auslesen des Decoders auf dem Programmiergleis: Übernommen wird ausschließlich der gespeicherte CV-Cache der ECoS.

Nicht gelieferte Werte bleiben unbekannt; sie werden nicht zu null. Bei Anlage oder Zuordnung eines Fahrzeugs wird der vollständige Abruf berücksichtigt. Für ein bereits zugeordnetes Fahrzeug steht nach vollständigem Abruf **CV-Werte ergänzen** bereit.

## 3. Übernahme prüfen

Kontrolliere die Anzahl, die Werte und den Zielbestand. Vorhandene Werte werden nicht ungefragt ersetzt; abweichende Werte bleiben nachvollziehbar. Indizierte CVs und CVs ohne Indexangabe sind getrennte Datensätze. ECoS ist als Herkunft erkennbar.

CV8 kann den Hersteller belegen. Modellregeln benötigen zusätzlich eine passende Modellkennung, etwa den gespeicherten Decoder-Typ am Fahrzeug. Die Bibliothek kann fehlende Indexseiten nicht rekonstruieren: Mehrdeutige Werte bleiben ohne erfundene Zuordnung. [CVs verstehen](../vehicles/decoder-cv) und [Decoderbibliothek](../vehicles/decoder-library).

## 4. Ein Fahrzeug anlegen oder zuordnen

Wähle die passende Übernahmeaktion der Arbeitsliste. Bei einem neuen Fahrzeug ergänzt du die Pflichtfelder, die die Zentrale nicht kennt, etwa Hersteller, Spurweite und Gattung. Prüfe die Zuordnung zu einem bestehenden Fahrzeug besonders sorgfältig. Speichere erst danach.

Die Fahrzeuganlage mit externer Zuordnung ist ein mehrstufiger Ablauf. Falls eine spätere Aktion scheitert, prüfe den bereits gespeicherten Fahrzeugstand und den Auftrag, bevor du neu anlegst. So vermeidest du Duplikate.

## Was kann zur ECoS zurückgeschrieben werden?

Lesen und Übernehmen in RailKeeper verändert keine Decoder-CVs. Der separate Schreibablauf zeigt Änderungen an Lokdaten vorab und verlangt eine Bestätigung. Bildübertragung und Bildzuordnung besitzen eine [eigene Vorschau](../vehicles/locomotive-images). Die CV-Übernahme programmiert keinen Decoder und ersetzt kein LokProgrammer-Projekt.

## Wenn etwas fehlt

| Beobachtung | Nächster Schritt |
| --- | --- |
| Weboberfläche erreichbar, Protokoll nicht | TCP-Port 15471 vom RailKeeper-Server prüfen; die Weboberfläche nutzt einen anderen Port |
| Weniger CVs als im LokProgrammer-Export | Die ECoS liefert nur ihren gespeicherten Cache; bei Bedarf den vollständigen TXT-Export im Fahrzeug importieren |
| Ein Wert hat keine Erklärung | Modell, Firmware und Indexkontext prüfen; fehlende Daten bleiben offen |
| Falscher Fahrzeugtreffer | Übernahme unterbrechen und die Zuordnung korrigieren |
| Abruf unvollständig | Fortschritt und Fehler prüfen, dann kontrolliert fortsetzen |

Der [CS3-Leseablauf](cs3-import) übernimmt andere Daten. Z21 und Intellibox 3 bieten noch keinen gleichwertigen Lok-/CV-Import.
