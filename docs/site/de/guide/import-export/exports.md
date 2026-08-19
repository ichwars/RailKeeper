---
title: Bestand exportieren
description: CSV, RailKeeper-JSON und eine druckbare Fahrzeug-Bestandsausgabe erstellen.
audience: user
status: stable
reviewedVersion: 0.1.19.2
lastReviewed: 2026-08-16
---

# Bestand exportieren

Die Exportaktionen verwenden die vollständige Fahrzeugliste, die **Import/Export** lädt. Sie
übernehmen keine Suche oder Filter der Fahrzeugbestandsseite. Alle drei Bedienelemente bleiben
deaktiviert, während die Liste geladen wird oder wenn keine Fahrzeuge vorhanden sind.

Admin, Editor, Viewer und Planner dürfen exportieren. Der Export verändert keine RailKeeper-Daten.

## CSV-Export

**CSV exportieren** lädt `railkeeper-bestand.csv` mit folgenden Eigenschaften herunter:

- UTF-8-Text mit Byte-Reihenfolge-Markierung für Tabellenprogramm-Kompatibilität;
- durch Semikolon getrennte Spalten;
- die 62 skalaren Ziele aus [Fahrzeugdateien importieren](/de/guide/import-export/file-import);
- Spaltenbezeichnungen in der aktuell gewählten RailKeeper-Oberflächensprache;
- übersetzte Ja-/Nein-Texte für boolesche Felder;
- doppelte Anführungszeichen zur Maskierung von Semikolons, Anführungszeichen und Zeilenumbrüchen.

Dies ist der breiteste stabile Rundlauf für Fahrzeugfelder. Eine CSV aus einer Oberflächensprache
wird durch die eingebauten deutschen und englischen Spaltenaliasse erkannt. Trotzdem müssen
Zuordnung und Zeilen vor dem Schreiben in eine andere Installation geprüft werden.

Die CSV enthält keine Bilder, Beilagendateien, Wartungen, Ersatzteile, Funktionsdatensätze,
CV-Datensätze, Decoder-Dateien oder externe Zuordnungen.

## RailKeeper-JSON-Export

**JSON exportieren** lädt `railkeeper-bestand.json` herunter. Die oberste Struktur lautet:

```json
{
  "format": "railkeeper-vehicles",
  "version": 1,
  "vehicles": []
}
```

Das `vehicles`-Array enthält die Fahrzeugobjekte, welche an diese Seite geliefert wurden. Dadurch
eignet sich JSON zur Prüfung und für einen verlustärmeren Datenaustausch, ist aber keine
Anwendungssicherung und kein vollständiges Wiederherstellungsformat.

Der stabile JSON-Import von v0.1.19.2 liest pro Objekt nur 14 Felder. Er stellt nicht alle im
exportierten Objekt vorhandenen Eigenschaften, verschachtelten Datensätze oder gespeicherten
Dateibytes wieder her. Vor einem Rundlauf die genaue
[JSON-Teilmenge](/de/guide/import-export/file-import#teilmenge-beim-json-import) prüfen.

## PDF- und Druckansicht

**PDF/Druckansicht** öffnet ein neues Browserfenster und startet sofort den Browser-Druckdialog.
Das Dokument verwendet A4 im Querformat und enthält:

- Gesamtzahl sowie Anzahl digitaler und analoger Fahrzeuge;
- Erstellungsdatum und -uhrzeit;
- Inventarnummer, Hersteller, Artikel-Nr., Bezeichnung, Spurweite, Epoche, Kategorie,
  Digital/Analog und Listenpreis für jedes Fahrzeug.

RailKeeper erzeugt keine PDF-Datei auf dem Server. Im Browser-Druckdialog **Als PDF speichern**
wählen, wenn Browser oder Betriebssystem dieses Ziel anbieten.

Öffnet sich nichts, müssen Popups für die RailKeeper-Adresse erlaubt werden. Das Druckfenster
maskiert Fahrzeugtexte, bevor sie in den Bericht eingesetzt werden.

## Den Export wählen

| Bedarf | Empfohlene Ausgabe |
| --- | --- |
| Unterstützte skalare Felder in einem Tabellenprogramm prüfen oder bearbeiten | CSV |
| Die an die Seite gelieferten Fahrzeugobjekte im Rohformat prüfen | JSON |
| Eine kompakte, lesbare Bestandsliste ausgeben oder archivieren | PDF/Druckansicht |
| Anwendungsdaten, Uploads und zugehörige Datensätze wiederherstellen | Admin-Anwendungssicherung, nicht diese Exporte |

Exportdateien müssen entsprechend der Vertraulichkeit des Bestands behandelt werden. Artikelquellen,
Kaufdaten, Lagerorte, Werte, Notizen und andere private Sammlungsdetails können in CSV oder JSON
enthalten sein.

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert RailKeeper **v0.1.19.2** und wurde zuletzt am 16.08.2026 geprüft.
