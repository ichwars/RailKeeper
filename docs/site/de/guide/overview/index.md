---
title: Übersicht, Kennzahlen und Datenqualität
description: RailKeeper-Dashboard auswerten, Datenlücken bearbeiten und Kacheln anordnen.
audience: user
status: stable
reviewedVersion: 0.1.19.2
lastReviewed: 2026-08-16
---

# Übersicht, Kennzahlen und Datenqualität

Die **Übersicht** ist das Arbeits-Dashboard von RailKeeper für Bestandsgröße, erfassten Wert,
Digitalisierung, Wartung und Datenqualität. Dieses Kapitel erklärt die Bedeutung jeder Kennzahl und
den Weg von einem Hinweis zu den betroffenen Fahrzeugen. Es beschreibt den stabilen
RailKeeper-Stand v0.1.19.2.

Benutzer mit den Rollen Admin, Editor, Viewer oder Planner können die Übersicht öffnen. Ein Konto,
das ausschließlich die Rolle Messe besitzt, startet unter **Ausstellung** und kann dieses Dashboard
nicht aufrufen.

Nur Benutzer mit der Rolle Admin oder Editor können Fahrzeugdaten ändern. Viewer und Planner
können das Dashboard und gefilterte Ergebnisse prüfen, müssen fehlende Daten aber durch einen Admin
oder Editor korrigieren lassen.

## Übersicht öffnen und aktualisieren

Wähle **Übersicht** in der Seitenleiste. RailKeeper lädt die für die aktuelle Sitzung verfügbaren
Fahrzeuge und ruft die aktuellen Wert-Summen für Fahrzeuge und Zubehör vom Server ab. Die Übersicht
ist eine aktuelle Zusammenfassung der gespeicherten Bestandsdaten und kein separat gespeicherter
Bericht.

Verwende das Aktualisieren-Symbol im Seitenkopf, nachdem Fahrzeug- oder Wartungsdaten in einem
anderen Bereich geändert wurden. Während des Ladens ist die Schaltfläche deaktiviert. Schlägt das
Laden der Fahrzeuge fehl, zeigt RailKeeper den Fehler oberhalb des Dashboards. Schlägt nur die
Wertabfrage fehl, bleiben die übrigen Kennzahlen verfügbar und der Wertebereich zeigt einen eigenen
Fehler. Zuvor geladene Fahrzeugwerte bleiben, wenn vorhanden, sichtbar. Prüfe deshalb den Fehler,
bevor du dich auf diese Werte verlässt.

Das Import/Export-Symbol im selben Seitenkopf öffnet **Import/Export**.

## Die vier Kennzahlenbereiche lesen

| Kennzahl | Bedeutung |
| --- | --- |
| **Gesamtbestand** | Anzahl der Fahrzeuge. Die Zeile darunter meldet, wie viele Kategorie- und Spurweitengruppen in den jeweils fünf häufigsten Werten vorkommen. Fehlende Angaben bilden die Gruppen **Ohne Kategorie** und **Ohne Spur**. Bei mehr als fünf Werten ist dies nicht die Anzahl aller unterschiedlichen Werte. |
| **Digitalisierung** | Digitale Fahrzeuge geteilt durch alle Fahrzeuge, auf ganze Prozent gerundet. Die Detailzeile zeigt die Anzahl digitaler und analoger Fahrzeuge. |
| **Erfasste Bestandswerte** | Vier centgenaue Euro-Summen: Listenwert der Fahrzeuge, Kaufpreis der Fahrzeuge, Listenwert des Zubehörs und Kaufkosten des Zubehörs. Die Werte bleiben getrennt und werden nicht zu einer Gesamtsumme vermischt. |
| **Wartung** | Anzahl nicht erledigter Wartungseinträge, deren Fälligkeitsdatum heute oder früher liegt. Die Detailzeile zeigt getrennt die in den nächsten 30 Tagen fälligen und alle offenen Einträge, einschließlich Einträgen ohne Fälligkeitsdatum. |

Die Fahrzeugwerte addieren den gepflegten Listen- und Kaufpreis einmal je Fahrzeug. Beim
Zubehör-Listenwert wird der gepflegte Einzel-Listenpreis mit der aktuell vorhandenen Menge
multipliziert. Die Zubehör-Kaufkosten verwenden erfasste Euro-Einkäufe mit Menge und Einzelpreis.
Zusätzlich zählen manuell gepflegte Kaufpreise einzelner Zubehör-Exemplare, die keinem Einkauf
zugeordnet sind. Ausdrücklich in Fremdwährung erfasste Einkäufe werden nicht eingerechnet und unter
den Werten als Hinweis ausgewiesen.

RailKeeper verarbeitet übliche Zahlenformate mit Komma oder Punkt. Fehlende oder nicht auswertbare
Preise zählen als null. Die Summen beschreiben erfasste Anschaffungsdaten und sind keine
Marktwertschätzung. Ohne Fahrzeuge sind Fahrzeugbestand, Fahrzeugwerte und Prozentkennzahlen null;
Zubehörwerte können weiterhin vorhanden sein.

## Dashboard-Kacheln verwenden

Die folgenden sieben Kacheln verbinden Bestandshinweise und Schnellzugriffe. Ihre Reihenfolge kann
geändert und jede Kachel kann ausgeblendet werden.

### Bestandsmix

**Bestandsmix** zeigt bis zu fünf Kategorien mit den meisten Fahrzeugen. Fehlende Kategorien werden
als **Ohne Kategorie** zusammengefasst. Der Balken zeigt den Anteil am Gesamtbestand, die Zahl die
exakte Fahrzeuganzahl. Ein Balken mit einem Wert größer null besitzt eine sichtbare Mindestbreite
von 8 %. Kleine Kategorien können deshalb breiter als ihr exakter Prozentanteil erscheinen.

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

**Handlungsbedarf** zeigt Datenlücken mit einem Wert größer null. Wähle eine Zeile, um den
**Fahrzeugbestand**, dessen Seitenüberschrift **Bestand** lautet, mit dem passenden aktiven Filter
zu öffnen.

| Datenlücke | Ausgewählte Fahrzeuge |
| --- | --- |
| **Ohne Hauptbild** | Fahrzeuge ohne ein gespeichertes Bild |
| **Ohne Artikel-Nr.** | Fahrzeuge ohne Artikelnummer |
| **Ohne EAN** | Fahrzeuge ohne EAN |
| **Digital ohne Decoder-Nr.** | Digitale Fahrzeuge ohne Wert in beiden Decoder-Nummernfeldern |

In v0.1.19.2 prüft **Ohne Hauptbild** technisch, ob irgendein Bild vorhanden ist. Die Prüfung
unterscheidet kein ausdrücklich ausgewähltes Hauptbild. Sind alle vier Werte null, meldet die Kachel,
dass keine größeren Datenlücken erkannt wurden.

Die Übersicht selbst besitzt keine allgemeine Suche und keine Filterleiste. Suche, kombinierte
Filter und Änderungen erfolgen im **Fahrzeugbestand**, nachdem eine Datenlücke gewählt oder der
Bestand direkt geöffnet wurde.

### Hersteller

**Hersteller** ordnet bis zu fünf Hersteller nach Fahrzeuganzahl. Leere Herstellerangaben werden
unter **Ohne Hersteller** zusammengefasst.

### Schnellaktionen

**Schnellaktionen** öffnet den nächsten Arbeitsbereich, ohne Daten zu ändern:

- **Bestand pflegen** öffnet den **Fahrzeugbestand**.
- **Import/Export** öffnet den Bereich zum Importieren, Exportieren und Drucken.
- **Stammdaten prüfen** öffnet **Einstellungen**. Dort können berechtigte Benutzer Auswahlwerte und
  Einstellungen für Bestandsnummern verwalten.

Am Ziel gelten weiterhin die Rechte des angemeldeten Benutzers. Eine Schnellaktion vergibt keine
zusätzliche Rolle.

### Wartungsradar

**Wartungsradar** zeigt bis zu vier nicht erledigte Wartungseinträge mit Fälligkeitsdatum, sortiert
nach dem frühesten Fälligkeitsdatum. Der am längsten überfällige Eintrag steht daher zuerst, danach
folgen später überfällige, heute fällige und zukünftige Einträge. Jede Zeile enthält Bestandsnummer,
Fahrzeugname oder Wartungsart, Wartungsart, Fälligkeitshinweis und Datum. Heute fällige und
überfällige Einträge werden hervorgehoben.

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
Das Anlegen, Importieren und Ändern von Fahrzeugen erfordert die Rolle Admin oder Editor.

## Datenlücken bearbeiten

1. Öffne **Handlungsbedarf** in der Übersicht.
2. Wähle die passende Datenlücke.
3. RailKeeper öffnet den **Fahrzeugbestand** und aktiviert den zugehörigen Bestands- oder
   Qualitätsfilter.
4. Öffne die betroffenen Datensätze. Ergänze oder korrigiere als Admin oder Editor die fehlenden
   Angaben. Verwende als Viewer oder Planner das Ergebnis, um die Datensätze zu identifizieren, und
   bitte einen Admin oder Editor um die Änderung.
5. Kehre zur **Übersicht** zurück und aktualisiere das Dashboard.

Bei einer Lücke für Artikelnummer, EAN oder Decoder entfernt das Schließen des aktiven
Qualitätsfilter-Eintrags den Datenlücken-Parameter aus der Browseradresse.
Andere Filtergruppen lassen diese Qualitätslücke aktiv und behalten den Parameter bei. Bei **Ohne
Hauptbild** ersetzt die Wahl eines anderen Bestandsfilters die aktive Bildlücke, lässt aber den nun
veralteten Parameter `gap=no-main-image` in der Adresse stehen. **Filter entfernen** entfernt den
Parameter bei jeder Datenlücke.

## Dashboard anordnen

Jeder Kachelkopf enthält drei Bedienelemente:

- **Nach vorn** verschiebt die Kachel um eine Position nach vorn.
- **Nach hinten** verschiebt sie um eine Position nach hinten.
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
| Keine Fahrzeuge | Hauptkennzahlen mit null, leere Listen und Empfehlung zum Anlegen oder Importieren | Als Admin oder Editor Fahrzeuge anlegen oder eine Liste importieren. Als Viewer oder Planner eine dieser Rollen um das Befüllen des Bestands bitten. |
| Keine offene Wartung mit Fälligkeitsdatum | Leerhinweis im Wartungsradar | Ein Fälligkeitsdatum ergänzen, wenn die Arbeit im Radar erscheinen soll |
| Keine größeren Datenlücken | Bestätigung unter **Handlungsbedarf** | Mit Wartung oder weiteren Bestandsangaben fortfahren |
| Alle Kacheln ausgeblendet | **Dashboard leer** mit **Layout zurücksetzen** | Layout zurücksetzen, um alle Kacheln wiederherzustellen |
| Laden der Fahrzeuge fehlgeschlagen | Fehlertext oberhalb des Dashboards | Verbindung und Sitzung prüfen, danach erneut aktualisieren |
| Laden der Bestandswerte fehlgeschlagen | Fehlertext unter **Erfasste Bestandswerte**, die übrigen Kennzahlen bleiben verfügbar | Verbindung und Sitzung prüfen, danach erneut aktualisieren |

## Verwandte Seiten

- [Überblick zum Benutzerhandbuch](/de/guide/)
- [Zubehörübersicht](/de/guide/accessories/)
- [Fahrzeugwartung und Zustand](/de/guide/vehicles/maintenance)
- [Ersteinrichtung und Anmeldung](/de/guide/getting-started/)
- [Überblick zur Administration](/de/administration/)

## Dokumentierter RailKeeper-Stand

Diese Seite dokumentiert den stabilen RailKeeper-Stand **v0.1.19.2** und wurde zuletzt am
2026-08-16 geprüft.
