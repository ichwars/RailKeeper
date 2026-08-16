---
title: Übersicht, Kennzahlen und Datenqualität
description: RailKeeper-Dashboard auswerten, Datenlücken bearbeiten und Kacheln anordnen.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Übersicht, Kennzahlen und Datenqualität

Die **Übersicht** ist das Arbeits-Dashboard von RailKeeper für Bestandsgröße, erfassten Wert,
Digitalisierung, Wartung und Datenqualität. Dieses Kapitel erklärt die Bedeutung jeder Kennzahl und
den Weg von einem Hinweis zu den betroffenen Fahrzeugen. Es beschreibt den stabilen
RailKeeper-Stand v0.1.17.6.

Benutzer mit den Rollen Admin, Editor, Viewer oder Planner können die Übersicht öffnen. Ein Konto,
das ausschließlich die Rolle Messe besitzt, startet unter **Ausstellung** und kann dieses Dashboard
nicht aufrufen.

## Übersicht öffnen und aktualisieren

Wähle **Übersicht** in der Seitenleiste. RailKeeper lädt die für die aktuelle Sitzung verfügbaren
Fahrzeuge und berechnet die Dashboard-Werte im Browser. Die Übersicht ist damit eine aktuelle
Zusammenfassung der Fahrzeugdaten und kein separat gespeicherter Serverbericht.

Verwende das Aktualisieren-Symbol im Seitenkopf, nachdem Fahrzeug- oder Wartungsdaten in einem
anderen Bereich geändert wurden. Während des Ladens ist die Schaltfläche deaktiviert. Schlägt das
Laden fehl, zeigt RailKeeper den Fehler oberhalb des Dashboards. Zuvor geladene Werte bleiben, wenn
vorhanden, sichtbar. Prüfe deshalb den Fehler, bevor du dich auf diese Werte verlässt.

Das Import/Export-Symbol im selben Seitenkopf öffnet **Import/Export**.

## Die vier Hauptkennzahlen lesen

| Kennzahl | Bedeutung |
| --- | --- |
| **Gesamtbestand** | Anzahl der Fahrzeuge. Die Zeile darunter meldet, wie viele Kategorie- und Spurweitengruppen in den jeweils fünf häufigsten Werten vorkommen. Bei mehr als fünf Werten ist dies nicht die Anzahl aller unterschiedlichen Werte. |
| **Digitalisierung** | Digitale Fahrzeuge geteilt durch alle Fahrzeuge, auf ganze Prozent gerundet. Die Detailzeile zeigt die Anzahl digitaler und analoger Fahrzeuge. |
| **Erfasster Listenwert** | Summe der auswertbaren, gepflegten Fahrzeug-Listenpreise. RailKeeper verarbeitet übliche Zahlenformate mit Komma oder Punkt und zeigt die Euro-Summe auf ganze Euro gerundet. Fehlende oder nicht auswertbare Preise zählen als null. Die Kennzahl ist keine Marktwertschätzung. |
| **Wartung** | Anzahl nicht erledigter Wartungseinträge, deren Fälligkeitsdatum heute oder früher liegt. Die Detailzeile zeigt getrennt die in den nächsten 30 Tagen fälligen und alle offenen Einträge, einschließlich Einträgen ohne Fälligkeitsdatum. |

Ohne Fahrzeuge sind Bestand und Wert null, Prozentkennzahlen werden mit 0 % angezeigt.

## Dashboard-Kacheln verwenden

Die folgenden sieben Kacheln verbinden Bestandshinweise und Schnellzugriffe. Ihre Reihenfolge kann
geändert und jede Kachel kann ausgeblendet werden.

### Bestandsmix

**Bestandsmix** zeigt bis zu fünf Kategorien mit den meisten Fahrzeugen. Fehlende Kategorien werden
als **Ohne Kategorie** zusammengefasst. Der Balken zeigt den Anteil am Gesamtbestand, die Zahl die
exakte Fahrzeuganzahl.

### Datenqualität

**Datenqualität** zeigt den Anteil aller Fahrzeuge, die fünf einzelne Prüfungen erfüllen:

| Kennzahl | Gilt als erfüllt, wenn |
| --- | --- |
| **Bilder** | Mindestens ein Bild für das Fahrzeug gespeichert ist |
| **Decoder-Nummern** | Eines der beiden Decoder-Nummernfelder einen Wert enthält |
| **Artikelnummern** | Das Artikelnummernfeld einen Wert enthält |
| **EAN** | Das EAN-Feld einen Wert enthält |
| **Voll dokumentiert** | Artikelnummer, EAN und mindestens ein Bild gemeinsam vorhanden sind |

Jede Prozentangabe verwendet den Gesamtbestand als Bezugsgröße. **Decoder-Nummern** berücksichtigt
daher auch analoge Fahrzeuge, wenn eines der beiden Decoder-Nummernfelder gefüllt ist. Umgekehrt
erfordert **Voll dokumentiert** keine Decoder-Nummer und bestätigt nicht, dass jedes Fahrzeugfeld
vollständig ist. Ohne Fahrzeuge stehen alle fünf Werte bei 0 %.

### Handlungsbedarf

**Handlungsbedarf** zeigt Datenlücken mit einem Wert größer null. Wähle eine Zeile, um
**Fahrzeuge** mit dem passenden aktiven Filter zu öffnen.

| Datenlücke | Ausgewählte Fahrzeuge |
| --- | --- |
| **Ohne Hauptbild** | Fahrzeuge ohne ein gespeichertes Bild |
| **Ohne Artikel-Nr.** | Fahrzeuge ohne Artikelnummer |
| **Ohne EAN** | Fahrzeuge ohne EAN |
| **Digital ohne Decoder-Nr.** | Digitale Fahrzeuge ohne Wert in beiden Decoder-Nummernfeldern |

In v0.1.17.6 prüft **Ohne Hauptbild** technisch, ob irgendein Bild vorhanden ist. Die Prüfung
unterscheidet kein ausdrücklich ausgewähltes Hauptbild. Sind alle vier Werte null, meldet die Kachel,
dass keine größeren Datenlücken erkannt wurden.

Die Übersicht selbst besitzt keine allgemeine Suche und keine Filterleiste. Suche, kombinierte
Filter und Änderungen erfolgen unter **Fahrzeuge**, nachdem eine Datenlücke gewählt oder der Bestand
direkt geöffnet wurde.

### Hersteller

**Hersteller** ordnet bis zu fünf Hersteller nach Fahrzeuganzahl. Leere Herstellerangaben werden
unter **Ohne Hersteller** zusammengefasst.

### Schnellaktionen

**Schnellaktionen** öffnet den nächsten Arbeitsbereich, ohne Daten zu ändern:

- **Bestand pflegen** öffnet **Fahrzeuge**.
- **Import/Export** öffnet den Bereich zum Importieren, Exportieren und Drucken.
- **Stammdaten prüfen** öffnet **Einstellungen**. Dort können berechtigte Benutzer Auswahlwerte und
  Einstellungen für Bestandsnummern verwalten.

Am Ziel gelten weiterhin die Rechte des angemeldeten Benutzers. Eine Schnellaktion vergibt keine
zusätzliche Rolle.

### Wartungsradar

**Wartungsradar** zeigt bis zu vier nicht erledigte Wartungseinträge mit Fälligkeitsdatum, sortiert
nach dem nächsten Termin. Jede Zeile enthält Bestandsnummer, Fahrzeugname oder Wartungsart,
Wartungsart, Fälligkeitshinweis und Datum. Heute fällige und überfällige Einträge werden
hervorgehoben.

Die untere Zeile fasst zusammen:

- **Erledigt**: alle erledigten Wartungseinträge.
- **Kosten**: Summe der auswertbaren Kosten aller Wartungseinträge, einschließlich erledigter
  Einträge.
- **Zustände**: Anzahl unterschiedlicher Zustandsbewertungen innerhalb der fünf häufigsten Werte.
  Der angezeigte Wert ist deshalb auf fünf begrenzt.

Besitzt keine offene Wartung ein Fälligkeitsdatum, zeigt das Radar seinen Leerzustand. Offene
Wartungen ohne Datum zählen trotzdem zur Hauptkennzahl am oberen Seitenrand.

### Nächster Mehrwert

**Nächster Mehrwert** wählt nach einer festen Reihenfolge genau einen Hinweis aus den aktuellen
Daten:

1. Fahrzeuge anlegen oder importieren, wenn der Bestand leer ist.
2. Bilder ergänzen, solange die Bildabdeckung unter 70 % liegt.
3. Fällige Wartungen bearbeiten, sobald mindestens ein Eintrag heute fällig oder überfällig ist.
4. Ersatzteile und strukturierte Preis-/Wertpflege erwägen, wenn keine frühere Bedingung zutrifft.

Dies ist ein regelbasierter Hinweis und keine automatische Datenbewertung oder externe Empfehlung.

## Datenlücken bearbeiten

1. Öffne **Handlungsbedarf** in der Übersicht.
2. Wähle die passende Datenlücke.
3. RailKeeper öffnet **Fahrzeuge** und aktiviert den zugehörigen Bestands- oder Qualitätsfilter.
4. Öffne die betroffenen Datensätze und ergänze oder korrigiere die fehlenden Angaben.
5. Kehre zur **Übersicht** zurück und aktualisiere das Dashboard.

Das Zurücksetzen aller Filter unter **Fahrzeuge** entfernt den Datenlücken-Parameter aus der
Browseradresse. Auch die Wahl eines anderen Qualitätsfilters entfernt ihn. Andere manuelle
Filteränderungen können den ursprünglichen Parameter in der Adresse belassen, obwohl sich die
sichtbare Auswahl bereits geändert hat.

## Dashboard anordnen

Jeder Kachelkopf enthält drei Bedienelemente:

- Nach oben verschiebt die Kachel um eine Position nach vorn.
- Nach unten verschiebt sie um eine Position nach hinten.
- Ausblenden entfernt die Kachel aus dem Dashboard.

Sobald mindestens eine Kachel ausgeblendet wurde, erscheint das Zurücksetzen-Symbol im Seitenkopf.
**Layout zurücksetzen** zeigt wieder alle sieben Kacheln und stellt ihre Standardreihenfolge her.
Wurden alle Kacheln ausgeblendet, bietet auch der leere Dashboard-Bereich diese Schaltfläche an.

Reihenfolge und ausgeblendete Kacheln werden im lokalen Speicher des aktuellen Browsers abgelegt.
Sie werden nicht mit anderen Browsern geteilt und nicht als kontoweite Dashboard-Einstellung
gespeichert.

## Leere und besondere Zustände

| Situation | Anzeige in RailKeeper | Nächster Schritt |
| --- | --- | --- |
| Keine Fahrzeuge | Hauptkennzahlen mit null, leere Listen und Empfehlung zum Anlegen oder Importieren | Fahrzeuge anlegen oder eine Bestandsliste importieren |
| Keine offene Wartung mit Fälligkeitsdatum | Leerhinweis im Wartungsradar | Ein Fälligkeitsdatum ergänzen, wenn die Arbeit im Radar erscheinen soll |
| Keine größeren Datenlücken | Bestätigung unter **Handlungsbedarf** | Mit Wartung oder weiteren Bestandsangaben fortfahren |
| Alle Kacheln ausgeblendet | **Dashboard leer** mit **Layout zurücksetzen** | Layout zurücksetzen, um alle Kacheln wiederherzustellen |
| Laden der Fahrzeuge fehlgeschlagen | Fehlertext oberhalb des Dashboards | Verbindung und Sitzung prüfen, danach erneut aktualisieren |

## Verwandte Seiten

- [Überblick zum Benutzerhandbuch](/de/guide/)
- [Ersteinrichtung und Anmeldung](/de/guide/getting-started/)
- [Überblick zur Administration](/de/administration/)

## Dokumentierter RailKeeper-Stand

Diese Seite dokumentiert den stabilen RailKeeper-Stand **v0.1.17.6** und wurde zuletzt am
2026-08-16 geprüft.
