# Änderungsprotokoll

[English](CHANGELOG.md)

Diese Datei dokumentiert alle wesentlichen Änderungen an RailKeeper.

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

[0.1.16]: https://github.com/ichwars/RailKeeper/compare/v0.1.15...v0.1.16
