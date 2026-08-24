---
title: Stammdatentransfer
description: RailKeeper-Stammdaten mit dem versionierten JSON-Dokument exportieren und abgleichen.
audience: admin
status: stable
reviewedVersion: 0.1.20.2
lastReviewed: 2026-08-16
---

# Stammdatentransfer

Der Bereich **Stammdatentransfer** unter **Einstellungen > Import/Export** überträgt kontrollierte
Einträge und ihre Beziehungen zwischen RailKeeper-Instanzen. Download und Upload sind ausschließlich
für Admins erlaubt. Dies ist ein spezieller Konfigurationstransfer, weder eine Anwendungssicherung
noch der Arbeitsbereich für den Fahrzeugdatei-Import und -Export.

## Inhalt des Dokuments

Die heruntergeladene Datei heißt `railkeeper-stammdaten-YYYYMMDD-HHMMSS.json`. RailKeeper v0.1.20.2
exportiert das Format `railkeeper-master-data` in Version 2 mit:

- Erstellungszeitpunkt;
- allen aktiven und inaktiven allgemeinen und Artikelstammdaten;
- Schlüsseln, Bezeichnungen, Quell-URLs, Metadaten, Sortierreihenfolge, Herkunft und Zeitstempeln;
- konfigurierten Stammdatenbeziehungen, beispielsweise Kategorie-Gattung-Zuordnungen;
- eingebetteten Bildern der Funktionssymbole, weil sie zu den Metadaten gehören.

Nicht enthalten sind Fahrzeuge, Zubehörartikel, Bestand, Lagerorte, Inventarnummernschemata,
Benutzer, Rollen, Sitzungen oder Anmeldedaten. Für eine wiederherstellbare Übertragung von Bestand
und Betriebsdaten ist die Anwendungssicherung und -wiederherstellung erforderlich. Lokale
Anmeldedaten bleiben auch dort getrennt.

## Stammdaten exportieren

1. **Einstellungen > Import/Export** als Admin öffnen.
2. Unter **Stammdatentransfer** die Downloadaktion wählen.
3. JSON zusammen mit der zugehörigen RailKeeper-Version und einer kurzen Änderungsnotiz ablegen.
4. Eine aktuelle Anwendungssicherung getrennt aufbewahren.

Der Export schreibt keine Daten. Er enthält eigene Änderungen und die zum Downloadzeitpunkt
vorhandenen inaktiven Zustände.

## Importabgleich verstehen

Der Import ist ein vollständiger Sollzustandsabgleich und keine ergänzende Zusammenführung.
RailKeeper akzeptiert die Dokumentversionen 1 und 2 und prüft das gesamte Dokument, bevor Einträge
und Beziehungen in einer Datenbanktransaktion ersetzt werden.

Dabei gelten diese Regeln:

- Eine vorhandene passende mitgelieferte Kennung bleibt unabhängig von der importierten Herkunft
  **Mitgeliefert**.
- Eine unbekannte Kennung wird **Eigen**, selbst wenn das Dokument sie als mitgeliefert bezeichnet.
- Im Dokument fehlende mitgelieferte Einträge bleiben mit ihrem aktuellen Zustand erhalten.
- Im Dokument fehlende ungenutzte eigene Einträge werden entfernt.
- Wird ein fehlender eigener Eintrag verwendet, scheitert der gesamte Import vor der Änderung.
- Die Standardschlüssel der Artikelarten bleiben geschützt und müssen eine gültige vollständige
  Konfiguration bilden.
- Fehlen Artikelarten oder Zubehörunterarten in älteren Dokumenten, bleiben die aktuellen
  geschützten Daten nach den Kompatibilitätsregeln erhalten.
- Beziehungen müssen eindeutig sein und auf Einträge verweisen, die nach dem Abgleich existieren.
- Eine vorhandene fehlende Beziehung bleibt nur erhalten, wenn beide Endpunkte mitgeliefert sind.
  Andere im Dokument fehlende Beziehungen werden entfernt.
- Definitionen eigener Felder und ihre gespeicherten Artikelverweise müssen gültig bleiben.

Erst nach allen erfolgreichen Prüfungen ersetzt RailKeeper die Stammdatentabellen und aktualisiert
den Laufzeit-Cache. Eine fehlgeschlagene Validierung oder Datenbankoperation lässt den zuvor
gespeicherten Zustand unverändert.

## Sicher importieren

1. Aktuelle Anwendungssicherung erstellen und prüfen.
2. Aktuelle Stammdaten zusätzlich als getrennte Rückfallreferenz exportieren.
3. Eingehende Datei mit dem aktuellen Export vergleichen, besonders fehlende eigene Einträge,
   Inaktivzustände, Artikelarten, Unterarten und Beziehungen.
4. JSON-Datei unter **Stammdatentransfer** auswählen.
5. Upload wählen und den destruktiven Abgleich bestätigen.
6. Auf die gemeldete Anzahl importierter Einträge und Beziehungen warten.
7. **Einstellungen > Daten** neu laden und repräsentative allgemeine und Artikelstammdaten prüfen.
8. Vor der Abnahme ein betroffenes Fahrzeug und einen Zubehörartikel anlegen oder bearbeiten.

Die stabile Oberfläche bietet keine Probelauf-Vorschau. Die Uploadgrenze beträgt 25 MiB. Eine
fehlerhafte Datei, nicht unterstützte Version, doppelte Kennung, defekte Beziehung, unzulässige
Änderung der Artikelarten oder ein fehlender verwendeter eigener Eintrag weist den Import ab.

## Abgewiesenen oder falschen Transfer beheben

Nach einer Abweisung die Instanz weiterlaufen lassen, Fehlermeldung lesen, Quelldokument oder
Quellinstanz korrigieren, erneut exportieren und wiederholen. Referenzierte Werte nicht nur deshalb
entfernen, damit ein Import erfolgreich ist.

Hat ein technisch gültiger Import den falschen Sollzustand erzeugt, das zuvor exportierte
Stammdatendokument importieren. Müssen auch Bestandszusammenhang, Nummernschemata, Lagerorte oder
andere Anwendungsdaten wiederhergestellt werden, stattdessen die geprüfte Anwendungssicherung
verwenden.

## Verwandte Seiten

- [Stammdaten-Administration](./master-data)
- [Allgemeine Stammdaten](./master-data-general)
- [Artikelstammdaten und Lagerorte](./master-data-articles)
- [Arbeitsbereich Import und Export](/de/guide/import-export/)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert RailKeeper **v0.1.20.2** und wurde zuletzt am 16.08.2026 geprüft.
