# Änderungsprotokoll

[English](CHANGELOG.md)

Diese Datei dokumentiert alle wesentlichen Änderungen an RailKeeper.

## Unveröffentlicht

## [0.1.18] - 2026-08-16

### Hinzugefügt

- Die Fahrzeugübersicht bietet pro Benutzer gespeicherte Spaltenauswahl und -reihenfolge,
  zusätzliche Bestandsfilter, eine erweiterte Suche sowie kompakte, aufklappbare Mobilkarten.
- Fahrzeuge erfassen jetzt Erwerb, Kaufpreis und -datum, Lagerung, Zustand und Verpackung. Listen-
  und Kaufwerte werden für Fahrzeuge und Zubehör getrennt ausgewertet und in der Übersicht
  nachvollziehbar zusammengefasst.
- Stammdaten können sicher deaktiviert, reaktiviert und nur dann gelöscht werden, wenn keine
  fachlichen Referenzen bestehen. Herkunft und Status bleiben in Exporten und Sicherungen erhalten.
- Die Windows-Standalone-Ausgabe kann verfügbare Versionen anzeigen und das passende ZIP direkt
  herunterladen. Die Anwendung schützt bestehende Installationen durch dauerhafte Datenpfade,
  validierte SQLite-Sicherungen und automatische Sicherungskopien vor Migrationen.
- Das zweisprachige RailKeeper-Handbuch dokumentiert Einrichtung, Übersicht, Fahrzeuge, Wartung,
  Medien, Decoder/CV, Suche, Zubehör und Ausstellungsbetrieb.

### Geändert

- CSV-Import und -Export umfassen alle skalaren Fahrzeugfelder. Eindeutige deutsche und englische
  Überschriften, zusätzliche Aliase und die bestehende Änderungsvorschau ermöglichen einen
  verlustfreien RailKeeper-CSV-Rundlauf.
- Die Fahrzeugübersicht unterstützt `PluX12`; Aktionsmenüs bleiben in Tabellen- und Mobilansichten
  sichtbar und tastaturbedienbar.

### Behoben

- Kennzahlen der Übersicht bleiben auf breiten Ansichten in einer gemeinsamen Zeile.
- Fahrzeug-Kurzmenüs werden nicht mehr an Tabellen- oder Bildschirmrändern verdeckt.
- Windows-Aktualisierungen trennen Programmpaket und Benutzerdaten, statt Daten im
  Installationsordner durch entpackte Dateien zu gefährden.

### Sicherheit

- Vor Datenbankmigrationen wird eine konsistente SQLite-Sicherung erstellt; unsichere oder
  mehrdeutige Windows-Datenpfade blockieren den Start mit einer verständlichen Diagnose.
- Stammdatenimporte und -änderungen bewahren referenzierte historische Werte und erzwingen die
  serverseitigen Lebenszyklusregeln.

## [0.1.17.6] - 2026-08-15

### Geändert

- Die Konfiguration von Digitalzentralen führt jetzt in vier klaren Schritten von der Adapterwahl
  über Konfiguration und Verbindungstest bis zur bewussten Aktivierung.
- Verbindungstest, Diagnose, technische Details, ECoS-Meldungen und Sicherheitsumfang sind in
  einem gemeinsamen Arbeitsbereich zusammengeführt.
- Adapter können aktiviert, deaktiviert und vollständig entfernt werden. Der ECoS-Live-Monitor
  bleibt an einen erfolgreichen Test und eine aktive Verbindung gebunden.
- Die neue Oberfläche passt sich ohne horizontalen Überlauf an schmale Ansichten an.

## [0.1.17.5] - 2026-08-15

### Geändert

- Der Anlagen-Arbeitsbereich bleibt über den Direktzugriff verfügbar, wird aber nicht mehr in der
  Hauptnavigation angezeigt.

## [0.1.17.4] - 2026-08-14

### Behoben

- Die Spaltenüberschriften für Hersteller, Artikelnummer und Lagerung behalten ihre vorgesehenen
  Breiten und überlagern sich in der konfigurierbaren Zubehör-Tabelle nicht mehr.

## [0.1.17.3] - 2026-08-14

### Hinzugefügt

- Dateien für Repository-Zuständigkeit, Beiträge, Verhaltensregeln, Support, Projektkennzeichen und
  Drittanbieterhinweise wurden ergänzt.
- Die Zubehör-Tabelle bietet eine persistente Spaltenauswahl für Bild, Inventarnummer, Hersteller,
  Artikelnummer, Name, Art, Spur, Bestand und Lagerung. Inventarnummer und Name können nicht
  gleichzeitig ausgeblendet werden.

### Geändert

- Neue RailKeeper-Fassungen stehen unter `AGPL-3.0-only`, damit veränderte, über ein Netzwerk
  angebotene Fassungen offen bleiben und ihren Nutzern den korrespondierenden Quellcode anbieten.
  Bestehende Releases behalten ihre bisherigen Lizenzbedingungen; die AGPL erlaubt weiterhin
  kommerzielle Nutzung.
- Die Projektunterstützung verweist nur noch auf GitHub Sponsors und Ko-fi als freiwillige
  Zuwendungen ohne Vorteile, bezahlten Support, SLA oder besonderen Zugang.
- Die ECoS-Anbindung liest nur noch Lokstammdaten, CV-Werte und statische Funktionsdefinitionen.
  Der Schreibabgleich bleibt nach Prüfung und Bestätigung auf Name, Adresse und Protokoll begrenzt.
- Aktuelle Geschwindigkeit, Richtung, aktive Funktionszustände, Lokbilder, Schaltartikel,
  Fahrwege, S88, Booster und weitere ECoS-Objektmanager wurden aus Abfragen und UI-Verträgen
  entfernt.
- Archivieren, Wiederherstellen und endgültiges Löschen stehen in Tabellen-, Kachel- und
  Kompaktansicht als direkte, beschriftete Symbolaktionen zur Verfügung. Das bisherige
  Drei-Punkte-Menü entfällt.

## [0.1.17.2] - 2026-08-14

### Hinzugefügt

- Der Anlagen-Arbeitsbereich bietet jetzt einen interaktiven digitalen Zwilling, bearbeitbare
  technische Positionen, Modulports, Gleispläne, Revisionsvorschauen, Materialreservierungen und
  den Austausch von Gleisbibliotheken.
- Die Gleisplanung unterstützt geprüfte Tillig-Geometrien, Flexgleise, Übergangsbögen,
  Höhenprofile, freie Planobjekte, Abstandsprüfungen, Steigungsgrenzen und Verbindungsanalysen.
- Die Zubehörübersicht bietet eine persistente Tabellen- und Kachelansicht sowie eine kompakte
  Mobilansicht.
- Administratoren können vollständig unbenutzte Zubehörartikel nach ausdrücklicher Bestätigung
  endgültig löschen.

### Geändert

- Die Anlagen-Navigation ist wieder aktiv und öffnet den erweiterten Arbeitsbereich.
- Aktionen in Zubehörzeilen und -karten verwenden dasselbe tastaturbedienbare Menüverhalten.

### Behoben

- Hersteller importierter Gleise bleiben erhalten und Gleisbibliotheken werden stabil geladen.
- Zubehör-Aktionsmenüs bleiben an Tabellen- und Mobilkartenrändern sichtbar, lange Metadatenfelder
  in Karten bleiben klar getrennt.

### Sicherheit

- Die endgültige Zubehörlöschung wird serverseitig auf Administratoren begrenzt und abgewiesen,
  sobald Bestand, Käufe, Bewegungen, Reservierungen, Einbauten, Historie, Assets oder
  Anlagenreferenzen vorhanden sind.

## [0.1.17.1] - 2026-08-13

### Hinzugefügt

- Die Zubehör-Artikelsuche erfasst beschriftete Gleis-Fachangaben wie Gleissystem, Abmessungen,
  Richtung, Bettung, Anschlüsse und Digitaltauglichkeit als einzeln auswählbare, typisierte
  Vorschläge.
- Die Auswahl einer Spurweite trägt den konfigurierten Maßstab automatisch ein, bis das
  Maßstabsfeld manuell bearbeitet wird.
- Neue oder bisher unverschlagwortete Artikel erhalten synchronisierte Schlagwortvorschläge aus
  Bezeichnung, Hersteller, Artikelart und Unterart, bis das Schlagwortfeld manuell bearbeitet wird.

### Behoben

- Der Zubehör-Artikelsuchdialog wird jetzt über dem Editor angezeigt, statt hinter dessen
  Modalebene verborgen zu bleiben.
- Spurweiten-Mehrfachauswahlen reagieren auf eingegebene Optionsbezeichnungen und begrenzen Escape
  auf die geöffnete Auswahlliste, statt den gesamten Artikeleditor zu schließen.
- Fehlerhafte oder unpassende Gleis-Fachangaben externer Seiten lassen sich nicht auswählen oder
  importieren.

## [0.1.17] - 2026-08-13

### Hinzugefügt

- Artikeldaten- und Barcodesuche sind jetzt direkt beim Anlegen und Bearbeiten von Zubehör
  verfügbar, mit Quellenvorschau und ausdrücklicher felweiser Auswahl vor der Übernahme.
- Ausgewählte Produktbilder aus Suchtreffern lassen sich nach dem Speichern als private
  Zubehördokumente importieren.
- Die GitHub-Sponsoring-Links enthalten jetzt Buy Me a Coffee und PayPal.

### Geändert

- Der bisherige Bereich `Artikelübersicht` heißt jetzt `Zubehör`; die bestehende Route
  `/accessories` bleibt unverändert.
- Gemeinsame Artikelsuch- und Barcodedialoge verwenden in Fahrzeug- und Zubehörabläufen jetzt
  anwendungseigene Modal-, Tastatur-, Fokusfallen- und Fokuswiederherstellungslogik.
- `modernc.org/sqlite` wurde auf 1.56.0, `lucide-react` auf 1.29.0 und Vite auf 8.2.1 aktualisiert.

### Behoben

- Unbekannte oder inaktive Hersteller- und Spurweitenvorschläge lassen sich nicht mehr auswählen,
  wenn der entsprechende Stammdatenwert nicht verfügbar ist.
- Wiederholte Bildimporte sind idempotent, erhalten vorhandene Hauptbilder und liefern bereits
  gespeicherte Dokumente ohne erneuten Download zurück.
- Escape und Tab bleiben im aktiven untergeordneten Suchdialog, statt den dahinterliegenden
  Zubehördialog zu erreichen.

### Sicherheit

- Externe Zubehörbilder sind auf öffentliche HTTP(S)-Ziele beschränkt und werden vor der
  Speicherung anhand von URL, DNS, Weiterleitungen, MIME-Typ und Anhangsgröße geprüft.

## [0.1.16] - 2026-08-09

### Hinzugefügt

- Vollständiger Artikelkatalog und Artikelbestand mit generierten Inventarnummern, Mengen- oder
  Einzelverfolgung, kontrollierten Zusatzfeldern, Käufen und Bestandseinheiten.
- Lagerortbestände, transaktionale Bestandsbewegungen, Reservierungen, Einbauten, Zustandshistorie,
  Ausbauten, Dokumente und zusammengeführte Nutzungshistorie.
- Sortierbare Artikelübersicht mit Auswahlspalte an erster Stelle, Produktbildern und betrieblichen
  Bestandszusammenfassungen.
- Fundamente für Anlagen, Module, Aufbauten, Planvarianten und Planrevisionen mit Schutz vor
  Versionskonflikten.
- Backup-Format Version 3 für die erweiterten Artikeldaten bei erhaltener Importkompatibilität mit
  den Backup-Versionen 1 und 2.

### Geändert

- Artikelformulare verwenden konfigurierte Stammdaten-Dropdowns für Hersteller, Spurweiten,
  Artikelarten, Unterarten, Bestandseinheiten und kontrollierte Zusatzfelder.
- Kennzahlen der Artikelübersicht zeigen Summen deutlich und Einzelwerte zurückhaltender.
- Artikelbearbeitung, Bestandsbuchung, Reservierung, Einbau und gleisspezifische Formulare verwenden
  kompakte, ausgerichtete Mehrspaltenlayouts.
- Artikelstammdaten sind unter `Einstellungen > Daten` in getrennten Gruppen für allgemeine Daten
  und Artikeldaten zusammengeführt.
- Der Eintrag der Hauptnavigation heißt `Anlage` und bleibt sichtbar, aber vorübergehend deaktiviert,
  solange der Arbeitsbereich weiter verfeinert wird. Der Direktzugriff auf `/layouts` und die
  Anlagen-API bleiben verfügbar.

### Behoben

- Transaktionale Bestands- und Zuordnungsregeln bleiben bei Wiederholungen, Umbuchungen,
  Reservierungen, Einbauten, Ausbauten und Wiederherstellungen erhalten.
- Inaktive kontrollierte Zusatzfelder und historische Werte bleiben bei Bearbeitung und Backups
  erhalten.
- Artikeleingaben, Dokumentprüfung, Suchabfragen, ältere Aktualisierungen und die
  Fokuswiederherstellung nach Bestätigungsdialogen wurden gehärtet.
- Die Kompatibilität mit Backup-Version 1 und 2 bleibt erhalten, ohne die Vollständigkeitsprüfung von
  Version 3 abzuschwächen.

### Sicherheit

- API-Validierung und geschützte Stammdatenregeln beim Import wurden verschärft.
- Die Vorprüfung der Backup-Wiederherstellung bleibt konservativ, Authentifizierungsdaten bleiben
  aus Exporten ausgeschlossen.

[0.1.18]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.6...v0.1.18
[0.1.17.6]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.5...v0.1.17.6
[0.1.17.5]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.4...v0.1.17.5
[0.1.17.4]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.3...v0.1.17.4
[0.1.17.3]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.2...v0.1.17.3
[0.1.17.2]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.1...v0.1.17.2
[0.1.17.1]: https://github.com/ichwars/RailKeeper/compare/v0.1.17...v0.1.17.1
[0.1.17]: https://github.com/ichwars/RailKeeper/compare/v0.1.16...v0.1.17
[0.1.16]: https://github.com/ichwars/RailKeeper/compare/v0.1.15...v0.1.16
