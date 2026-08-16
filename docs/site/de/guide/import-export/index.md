---
title: Import und Export
description: Fahrzeugbestand sicher austauschen und den getrennten ECoS-Ablauf verstehen.
audience: user
status: stable
reviewedVersion: 0.1.18
lastReviewed: 2026-08-16
---

# Import und Export

Der Arbeitsbereich **Import/Export** tauscht Fahrzeugbestand aus, ohne die normalen RailKeeper-
Prüfungen und Berechtigungen zu umgehen. Dateiimporte werden zuerst als Prüftabelle aufbereitet.
Exporte erstellen eine CSV-Datei, eine RailKeeper-JSON-Datei oder eine Druckansicht im Browser. Ein
getrennt kontrollierter ECoS-Ablauf kann Lokdaten lesen und jede Lok einzeln an den Fahrzeugeditor
übergeben.

Dieses Kapitel dokumentiert RailKeeper v0.1.18. Es behandelt ausschließlich den
Fahrzeugdatenaustausch. Stammdaten-Transfer, Digitalzentralen-Konfiguration sowie
Anwendungssicherung und Wiederherstellung bleiben getrennte Administrationsaufgaben.

## Zugriffsrechte

Admin, Editor, Viewer und Planner können **Import/Export** öffnen. Ein reines Messe-Konto bleibt im
isolierten Messearbeitsbereich und kann diese Seite nicht öffnen. Der Server prüft weiterhin jeden
Lese- und Schreibvorgang einzeln.

| Aktion | Admin | Editor | Viewer | Planner |
| --- | --- | --- | --- | --- |
| Aktuelle Fahrzeugliste laden | Ja | Ja | Ja | Ja |
| CSV, JSON oder Druckansicht exportieren | Ja | Ja | Ja | Ja |
| Datei einlesen und lokale Vorschau prüfen | Ja | Ja | Ja | Ja |
| Fahrzeuge aus der Vorschau anlegen oder aktualisieren | Ja | Ja | Nein | Nein |
| ECoS konfigurieren, lesen oder beschreiben | Ja | Nein | Nein | Nein |

Die Seite blendet **Auswahl speichern** für Viewer und Planner nicht aus. Die geschützte Fahrzeug-
API lehnt deren Schreibvorgang dennoch ab. Verwende für einen tatsächlichen Dateiimport ein
Editor- oder Admin-Konto.

## Den passenden Ablauf wählen

| Ziel | Ablauf |
| --- | --- |
| CSV, TSV, XML oder RailKeeper-JSON prüfen und Fahrzeuge anlegen oder aktualisieren | [Fahrzeugdateien importieren](/de/guide/import-export/file-import) |
| Aktuellen Fahrzeugbestand herunterladen oder kompakten Bestandsbericht drucken | [Bestand exportieren](/de/guide/import-export/exports) |
| Loks, statische Funktionen und CV-Werte aus einer ESU ECoS lesen | [ECoS-Lokabgleich](/de/guide/import-export/ecos-sync) |
| Hersteller, Spurweiten, Kategorien, Symbole oder andere Stammdaten übertragen | **Einstellungen > Stammdaten** verwenden, nicht diesen Arbeitsbereich |
| Anwendungsdaten und gespeicherte Uploads erhalten oder wiederherstellen | Admin-Anwendungssicherung verwenden, keinen Fahrzeugexport |

## Sicherer Arbeitsablauf

1. Vor einem großen oder unbekannten Import einen Admin um eine aktuelle Anwendungssicherung
   bitten.
2. CSV als lesbare Vergleichsdatei exportieren, wenn die aktuellen skalaren Fahrzeugfelder in
   einem Tabellenprogramm geprüft werden sollen.
3. Quelldatei laden und jedes Pflichtfeld, jede Validierungsmeldung und jedes Duplikat klären.
4. Nur geprüfte Zeilen auswählen. Besonders auf als Überschreibung gekennzeichnete Felder achten.
5. Auswahl speichern und anschließend sowohl gespeicherte als auch fehlgeschlagene Zeilen prüfen.
6. Nach einem großen Import einige repräsentative Fahrzeuge öffnen und deren Grunddaten prüfen.

## Wichtige Grenzen

- Der Dateiaustausch behandelt Fahrzeugfelder. Er stellt keine Bilder, Beilagendateien, Wartung,
  Ersatzteile, Decoder-Dateien, Funktionszuordnungen, CV-Datensätze oder Benutzerkonten wieder her.
- Die Dateivorschau wird im Browser berechnet. Eine Zeile wird erst durch **Auswahl speichern** mit
  einem berechtigten Konto geschrieben.
- Ausgewählte Zeilen werden nacheinander gespeichert, nicht als Alles-oder-nichts-Transaktion. Ein
  späterer Fehler macht bereits als **gespeichert** markierte Fahrzeuge nicht rückgängig.
- CSV ist das breiteste stabile Rundlaufformat für die 62 unterstützten skalaren Felder. Der
  JSON-Import liest absichtlich weniger Felder, auch wenn der JSON-Export weitere Eigenschaften
  enthält.
- ECoS-Daten verwenden eine getrennte Arbeitsliste und Fahrzeugeditor-Übergabe. Sie gelangen nie
  automatisch in die allgemeine Datei-Importprüfung.

## Fehlerbehebung

| Symptom | Prüfen |
| --- | --- |
| Export-Schaltflächen sind deaktiviert | Warten, bis die Fahrzeugliste geladen wurde. Bei leerem Bestand bleiben alle Export-Schaltflächen deaktiviert. |
| Eine Spalte wird als offen angezeigt | Unter **Spaltenzuordnung** ein Zielfeld wählen oder bei einer nicht benötigten Spalte **Ignorieren** belassen. |
| Eine Duplikatzeile ist nicht ausgewählt | Update-Vorschau prüfen und nur dann ausdrücklich auswählen, wenn der erkannte Inventarnummer-Treffer stimmt. |
| Speichern scheitert trotz gültiger Vorschau | Admin- oder Editor-Rolle prüfen und die zeilenbezogene Servermeldung lesen. Die Servervalidierung bleibt maßgeblich. |
| Einige Zeilen wurden vor einem Fehler gespeichert | Das ist beim sequenziellen Import zu erwarten. Gespeicherte Zeilen prüfen und nur fehlgeschlagene oder noch offene Zeilen wiederholen. |
| Druckansicht öffnet sich nicht | Popups für die RailKeeper-Adresse erlauben und erneut versuchen. |

## Dokumentierte RailKeeper-Version

Dieses Kapitel dokumentiert RailKeeper **v0.1.18** und wurde zuletzt am 16.08.2026 geprüft.
