---
title: Betriebseinträge prüfen und drucken
description: Betriebseinträge prüfen und drucken.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Betriebseinträge prüfen und drucken

**Ziel:** Eine Lok mit Besitzer, Steuerungsdaten und Fahrtagen erfassen und im Bericht wiederfinden.

<GuideFigure name="exhibition-entry" alt="Aktueller Eintragsdialog mit Allgemein, Fahrzeugbild und Funktionstasten." caption="Die tatsächliche Übung ergänzt die fehlende Zentrale/Schnittstelle. Der Betriebseintrag ändert keine Fahrzeug-Stammdaten." />

## Eintragsrechte und Speicherung

Admin, Messe und entsprechend kombinierte Konten dürfen Einträge in **Entwurf** oder **Offen**
ohne Sperre bearbeiten. Nur Admin löscht. Änderungen im Eintragsdialog bleiben bis **Speichern**
im Entwurf; **Abbrechen** verwirft sie. Server und Revision verhindern ungeprüftes Überschreiben.

## Eintrag anlegen oder bearbeiten

Wähle **Eintrag hinzufügen** bei **Teilnehmer und Fahrzeuge** oder das Bearbeiten-Symbol an der
Lokzeile. Der Dialog besitzt **Allgemein**, **Fahrzeugbild** und **Funktionstasten**.
Eine optionale Bestandsfahrzeugwahl steht Konten mit Bestandsleserecht zur Verfügung.
Die Auswahl ergänzt bekannte Angaben; prüfe besonders Adresse, Zentrale und Fahrtage.

## Allgemeine und Steuerungsdaten ausfüllen

### Grunddaten

Ergänze Besitzer, Lok-Bezeichnung, Hersteller, Baureihe, Gattung, Epoche, Bahnverwaltung und Verfügbarkeit.
Ein Gastfahrzeug kann ohne Verknüpfung zum Bestand eingetragen werden.
**Nicht verfügbar** ist ein Teilnahmehinweis und keine Löschung.

### Steuerung

Decoder-Typ und Adapter beschreiben die Lok. **Adresse DCC** und **Adresse SX** bleiben getrennt.
**Zentrale / Schnittstelle** benennt den vorgesehenen Steuerungskontext und ist nicht der Decoderstecker.
Bei digitalen Einträgen müssen eine geeignete Adresse und dieser Kontext vorhanden sein.
**Analog** kennzeichnet den entsprechenden Betrieb.

### Nr. / Beschriftung / Merkmale

Nutze das Notizfeld für betriebliche Hinweise. Es verändert keine individuellen Fahrzeug-Stammdaten.

## Messetage auswählen

Wähle **Alle Tage** oder die gewünschten Fahrtage der Veranstaltung. Die Tagesfilter im Arbeitsbereich
begrenzen die Ansicht; sie ändern die gespeicherte Teilnahme nicht. Prüfe Tagesauswahl nach einer Terminänderung.

## Betriebskonflikte lösen

**Konflikte prüfen** zeigt fehlende Angaben, mehrfach eingetragene Fahrzeuge und überschneidende
Adressen im gleichen Steuerungskontext an gemeinsamen Tagen. Korrigiere die betreffenden Einträge.
Eine dokumentierte Ausnahme bleibt sichtbar. Einträge ohne Bild oder Funktionstasten können auch ohne
Adresskonflikt weiter **Prüfen** anzeigen. **Bereit** ist eine Datenprüfung, keine Probefahrt.

## Bild hinzufügen oder entfernen

Öffne **Fahrzeugbild**, wähle eine geeignete Datei oder entferne die vorhandene Auswahl.
Erst **Speichern** sichert die Änderung des Betriebseintrags. Prüfe vor dem Drucken die Bilddarstellung.
Ein Bild am Bestandsfahrzeug wird durch diese Aktion nicht automatisch ausgetauscht.

## Funktionstasten F0 bis F31 konfigurieren

Öffne **Funktionstasten**. Ergänze eine Funktion, wähle Taste, Symbol und Bezeichnung und entferne
nicht benötigte Zeilen. Die Liste wird gemeinsam mit dem Betriebseintrag gespeichert.
Sie ist von den Decoder-CVs und der Steuerungsliste im normalen Fahrzeugeditor getrennt.

## Einträge lesen, sortieren und löschen

Nutze Tagesfilter, **Alle / Bereit / Prüfen** und **Einträge durchsuchen**. Die aktuelle Tabelle besitzt
keine frühere Sortier-Schaltfläche pro Spalte. Löschen verlangt Admin und eine bewusste Bestätigung;
bei geschützter Planung steht die Aktion nicht zur Verfügung.

## Ausstellungsliste drucken

Wähle **Drucken** im Veranstaltungskopf. Der aktuelle Arbeitsbereich erzeugt direkt den Bericht
und öffnet den Browser-Druckablauf; es gibt hier keinen vorgeschalteten Bildauswahl-Dialog.
Kontrolliere den Veranstaltungsnamen, alle vorgesehenen Einträge, Adressen und Funktionen im Bericht.
Die Ansicht enthält den gesamten Veranstaltungseintragsbestand, nicht nur einen aktiven Tabellenfilter.

<GuideFigure name="exhibition-print-report" alt="Tatsächlich erzeugter Bericht der Demo-Veranstaltung." caption="Das erzeugte Berichts-HTML wurde im Browser geprüft. Druckertreiber und native PDF-Ausgabe sind damit nicht abgenommen." />

Wähle Drucker oder PDF-Ziel und prüfe Format, Seitenumbrüche und Bilder. **Drucken** ändert keinen
Betriebseintrag. Die gespeicherte Standarddruckervorgabe wird nicht von jedem Druckablauf automatisch genutzt.

## Eintrags- und Druckfehler beheben

Bei veraltetem Stand neu laden und vergleichen. Bei fehlenden Aktionen Status und Rollen prüfen.
Bei einer leeren Druckansicht zunächst die gewählte Veranstaltung und deren Einträge kontrollieren.
Bei verdecktem Browserdialog diesen öffnen; physische Druckprobleme am Betriebssystem und Drucker prüfen.

## Verwandte Seiten

[Bebilderte Übung](practice) · [Veranstaltungsstatus](lists-and-locking).
