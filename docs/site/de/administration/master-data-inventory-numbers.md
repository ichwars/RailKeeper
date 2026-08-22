---
title: Inventarnummernschemata
description: Automatische Inventarnummern für Fahrzeuge und Zubehörartikel kollisionsfrei konfigurieren.
audience: admin
status: stable
reviewedVersion: 0.1.20
lastReviewed: 2026-08-16
---

# Inventarnummernschemata

Unter **Einstellungen > Allgemein > Inventarnummern** werden die automatischen Kennungen für neue
Fahrzeuge und Zubehörartikel gesteuert. Ein Schema enthält eindeutige Kategorie, Präfix, nächste
Nummer, Stellenzahl, Aktivstatus und Vorschau.

RailKeeper v0.1.20 stellt normalerweise diese aktiven Schemata bereit:

| Kategorie | Standardpräfix | Standardbeispiel | Verwendung |
| --- | --- | --- | --- |
| Fahrzeug | `RK-FAH` | `RK-FAH-000001` | Fahrzeug-Rückfall |
| Lokomotive | `RK-LOK` | `RK-LOK-000001` | Kategorien mit `lok` |
| Wagen | `RK-WAG` | `RK-WAG-000001` | Kategorien mit `wagen` oder `waggon` |
| Artikel | `RK-ART` | `RK-ART-000001` | Zubehörartikel |

## Vergaberegeln

RailKeeper reserviert die Nummer innerhalb derselben Datenbanktransaktion, in der der Datensatz
angelegt wird. Eine erfolgreiche Reservierung erhöht **Nächste Nummer**. Scheitert das Anlegen,
wird die Nummer nicht verbraucht.

Bei einem Fahrzeug sucht RailKeeper zuerst ein aktives Schema, dessen Kategorie exakt der gewählten
Fahrzeugkategorie entspricht. Danach folgen der erkannte Rückfall Lokomotive oder Wagen und zuletzt
das Fahrzeugschema. Ein Zubehörartikel benötigt das aktive Artikelschema und besitzt keinen
Rückfall.

Das erzeugte Format lautet immer `PRÄFIX-NUMMER`. RailKeeper wandelt das gespeicherte Präfix in
Großbuchstaben um, entfernt äußere Leerzeichen und ersetzt Leerzeichen durch Bindestriche. Die
Stellenzahl muss zwischen 1 und 12 liegen, die nächste Nummer mindestens 1 betragen. Die
Stellenzahl ist eine Mindestbreite, größere Zahlen werden niemals abgeschnitten.

Vor der Vergabe prüft RailKeeper die passende Datensatztabelle. Ist ein Kandidat bereits belegt,
wird der nächste Wert versucht und der Zähler über die Kollision hinaus weitergestellt. Nach 500
Versuchen endet die Suche. Fahrzeug- und Artikelnummern werden in ihren jeweiligen Tabellen und
nicht in einem gemeinsamen globalen Namensraum geprüft.

## Schema anlegen oder bearbeiten

1. Vor dem Anlegen die vorhandenen Kategorien prüfen. Pro Kategorie ist nur ein Schema erlaubt.
2. Kategorie, Präfix, nächste Nummer, Stellenzahl und Aktivstatus in der neuen Zeile eingeben.
3. Vorschau prüfen und **Anlegen** wählen.
4. Bei einem vorhandenen Schema die Werte direkt ändern und **Speichern** wählen.

In v0.1.20 gibt es keinen Löschendpunkt. Ein ungenutztes Schema stattdessen deaktivieren. Das
Bearbeiten oder Deaktivieren nummeriert vorhandene Datensätze niemals um.

Nur Admin und Editor dürfen Schemata anlegen oder speichern. Viewer und Planner können die Tabelle
lesen. Die stabile Oberfläche kann ihnen trotzdem bearbeitbar wirkende Felder zeigen, der Server
weist den Schreibvorgang jedoch ab.

## Nächste Nummer sicher ändern

Wird **Nächste Nummer** nach vorne gesetzt, entsteht bewusst eine Lücke. Beim Zurücksetzen wird eine
vorhandene Nummer wegen der Kollisionsprüfung nicht stillschweigend wiederverwendet. Der Vorgang
kann aber viele Versuche verursachen und Prüfungen erschweren. Eine nur vorwärts laufende Folge ist
vorzuziehen.

Vor einer Änderung an Präfix oder Kategorie:

1. Anwendungssicherung exportieren.
2. Aktuelle Vorschau und nächste Nummer notieren.
3. Prüfen, dass sich das Präfix im selben Datensatzbereich nicht mit einem anderen Schema
   überschneidet.
4. Schema speichern und einen Testdatensatz anlegen.
5. Vergebene Nummer und aktualisierten nächsten Wert kontrollieren.

Kann beim Anlegen eines Artikels keine Nummer vergeben werden, muss das Artikelschema reaktiviert
oder korrigiert werden. Bei Fahrzeugen das exakte Kategorieschema sowie die Rückfälle Lokomotive,
Wagen und Fahrzeug prüfen.

## Sicherung und Wiederherstellung

Inventarnummernschemata gehören zur Anwendungsdatenbank und sind in Anwendungssicherung und
Wiederherstellung enthalten. Der Stammdaten-JSON-Transfer enthält sie nicht. Für ältere
Artikelzeilen ohne Inventarnummer stellt die Wiederherstellung ein Artikelschema sicher. Fehlende
Nummern werden nur bei aktivem Schema vergeben, andernfalls wird die Wiederherstellung abgewiesen.
Dabei entstehen keine Dubletten zu anderen wiederhergestellten Artikelnummern. Das Schema kann
numerisch hinter vorhandenen Artikelnummern liegen, spätere Artikelanlagen überspringen diese
Kollisionen sicher.

## Verwandte Seiten

- [Stammdaten-Administration](./master-data)
- [Stammdatentransfer](./master-data-transfer)
- [Fahrzeugbestand](/de/guide/vehicles/)
- [Artikelstammdaten und Fachangaben](/de/guide/accessories/article-records)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert RailKeeper **v0.1.20** und wurde zuletzt am 16.08.2026 geprüft.
