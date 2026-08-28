---
title: Import und Export
description: Fahrzeug-, Zubehör- und Messedaten über geprüfte Transferaufträge austauschen.
audience: user
status: stable
reviewedVersion: 0.1.20.3
lastReviewed: 2026-08-27
---

# Import und Export

Der Arbeitsbereich **Import/Export** tauscht Fahrzeug-, Zubehör- und Messedaten aus, ohne die
RailKeeper-Prüfungen und Berechtigungen zu umgehen. Dauerhafte Profile bestimmen Richtung,
Datenbereiche, Format und Optionen. Jede Ausführung erzeugt einen nachvollziehbaren Auftrag mit
Vorschau, Problemen, Entscheidungen, Ergebnis und bei Exporten einem herunterladbaren Artefakt.

Importe verwenden CSV für genau einen Datenbereich oder versioniertes RailKeeper-JSON. Sie bleiben
in Prüfung, bis alle blockierenden Probleme geklärt sind und ein berechtigter Benutzer die
transaktionale Übernahme ausdrücklich bestätigt. Digitalzentralen-Abgleich, Stammdatentransfer sowie
Anwendungssicherung und Wiederherstellung bleiben getrennte Abläufe.

## Zugriffsrechte

Admin, Editor, Viewer, Planner und Messe können **Import/Export** öffnen. Reine Messe-Konten sehen
und verwenden ausschließlich Profile und Aufträge für Messelisten. Der Server erzwingt diesen
Umfang unabhängig von der Oberfläche.

Admin und Editor können vollständige Transferprofile anlegen oder ändern und Importe übernehmen.
Viewer und Planner können Aufträge prüfen und Exporte erzeugen, aber keine importierten Daten
übernehmen. Messe darf Messelistentransfers im isolierten Umfang exportieren, prüfen und übernehmen.
Nur Admin darf Profile deaktivieren, Artefakte löschen oder deren Serverordner durch RailKeeper
öffnen lassen.

## Den passenden Ablauf wählen

| Ziel | Ablauf |
| --- | --- |
| Einen Datenbereich als CSV übertragen | Import- oder Exportprofil für Fahrzeuge oder Zubehör anlegen |
| Einen oder mehrere Bereiche mit Beziehungen übertragen | Versioniertes RailKeeper-JSON-Profil anlegen |
| Einen Import prüfen | Auftrag öffnen, Probleme und Lösungen prüfen, dann Übernahme bestätigen |
| Lokdaten lesen, vergleichen oder schreiben | Den getrennt geschützten Arbeitsbereich **Digitalzentralen** verwenden |
| CS3-Lokdaten ohne Schreibzugriff vergleichen | [CS3-Lokdaten read-only lesen](./cs3-import) |
| Hersteller, Spurweiten, Kategorien, Symbole oder andere Stammdaten übertragen | **Einstellungen > Stammdaten** verwenden, nicht diesen Arbeitsbereich |
| Anwendungsdaten und gespeicherte Uploads erhalten oder wiederherstellen | Admin-Anwendungssicherung verwenden, keinen Fahrzeugexport |

## Transferprofile verstehen

Ein Transferprofil speichert nur die wiederverwendbare Auswahl für einen Auftrag:

- **Richtung**: Import liest eine Datei ein, Export erzeugt eine Datei.
- **Bereiche**: Fahrzeuge, Zubehör oder Ausstellungslisten. CSV unterstützt genau einen Bereich,
  Ausstellungslisten benötigen RailKeeper-JSON.
- **Format**: CSV für tabellarische Einzelbereiche, RailKeeper-JSON für vollständige, auch
  bereichsübergreifende Pakete.
- **CSV-Zuordnung**: Bei Fahrzeugimporten kann die Zuordnung von Quellspalten zu 62 Fahrzeugfeldern
  und 26 Fahrzeugset-Feldern als Profilstandard gespeichert werden.

Die Tabelle **Transferprofile** zeigt Import-, Export- und deaktivierte Profile gemeinsam. Nur
aktive Profile besitzen eine Startaktion. Ein aktiver Profilname darf je Richtung nur einmal
vorkommen, derselbe Name darf aber einmal für Import und einmal für Export verwendet werden.

## Fahrzeugsets übertragen

Ein Fahrzeugprofil nimmt zugehörige Fahrzeugsets automatisch mit. Die Datensatzanzahl zählt dabei
weiterhin Fahrzeuge, nicht zusätzliche Set-Metadaten.

- RailKeeper-JSON verwendet ab Paketversion 3 die normalisierte Struktur `vehicleSets`. Dateien der
  Paketversionen 1 und 2 bleiben für Einzelfahrzeugimporte kompatibel.
- CSV wiederholt die `vehicleSet*`-Felder in jeder Mitgliedszeile. RailKeeper gruppiert die Zeilen
  über `vehicleSetInventoryNumber` und übernimmt Reihenfolge sowie individuelle Mitgliedslabels.
- Erkannte Sets erscheinen als eigene Gruppen in der Importprüfung. Bei einer vorhandenen
  Set-Inventarnummer kann das gesamte Set aktualisiert, als neues Set importiert oder übersprungen
  werden. Bei extern gebundenen Mitgliedern sind nur Set-Kopie oder Überspringen zulässig.
- Jede Set-Aktion gilt für die vollständige Gruppe. RailKeeper hängt kein bereits anderweitig
  gebundenes Fahrzeug stillschweigend um.

## Sicherer Arbeitsablauf

1. Vor einem großen oder unbekannten Import einen Admin um eine aktuelle Anwendungssicherung
   bitten.
2. Ein aktives Profil wählen, dessen Richtung, Bereiche und Format zum geplanten Transfer passen.
3. Beim Export den Auftrag anlegen, ausführen und das Artefakt nach Prüfung der Zusammenfassung
   herunterladen.
4. Beim Import den Auftrag anlegen und die passende CSV- oder RailKeeper-JSON-Datei hochladen.
5. Bei CSV jede erkannte Quellspalte prüfen. Offene Spalten zuordnen oder ausdrücklich ignorieren,
   bei wiederkehrenden Dateien die Zuordnung optional im Profil speichern.
6. Dauerhafte Vorschau sowie jede Warnung und jeden Fehler prüfen. Für alle entscheidungspflichtigen
   Probleme eine Lösung festhalten.
7. Nur den geprüften Stand bestätigen, dann abgeschlossenen Auftrag und repräsentative Datensätze
   kontrollieren.

## Wichtige Grenzen

- Transferdateien sind nicht vertrauenswürdige Eingaben. RailKeeper prüft Paketversion, Hashes,
  Pfade, Datensatzidentitäten, Verweise, kontrollierte Werte und Rollenumfang vor der Übernahme.
- CSV ist auf genau einen unterstützten Bereich begrenzt und kann keine Messelisten übertragen.
  Versioniertes JSON kann mehrere unterstützte Bereiche und ihre Beziehungen enthalten.
- Importvorschau und Problementscheidungen werden auf dem Server gespeichert. Die Bestätigung
  übernimmt den geprüften Stand transaktional; ein Fehler hinterlässt keinen Teilimport.
- Transferpakete enthalten keine Benutzer, Rollen, Sitzungen, Passwort-Hashes oder andere lokale
  Authentifizierungsdaten. Für vollständige lokale Bestands- und Upload-Wiederherstellung die
  Anwendungssicherung verwenden.
- Fahrzeugbilder und andere Unterdaten gehören weiterhin in die Anwendungssicherung und nicht in
  einen Fahrzeug-Transfer.
- Digitalzentralen-Lese- und Schreibvorgänge gelangen nie automatisch in einen Transferauftrag.

## Fehlerbehebung

| Symptom | Prüfen |
| --- | --- |
| Kein passendes Profil ist verfügbar | Editor oder Admin um ein aktives Profil mit benötigter Richtung, Bereichen und Format bitten. |
| CSV lässt sich nicht wählen | Genau einen Fahrzeug- oder Zubehörbereich verwenden oder auf RailKeeper-JSON wechseln. Messelisten benötigen JSON. |
| Eine CSV-Spalte bleibt offen | Ein RailKeeper-Feld auswählen oder die Spalte ausdrücklich ignorieren. Erst dann lässt sich die Zuordnung prüfen. |
| Ein Import bleibt in Prüfung | Auftragsdetails öffnen und jedes blockierende Problem vor der Bestätigung lösen. |
| Die Bestätigung meldet einen veralteten Stand | Auftrag neu laden. Ein anderer Bediener hat Revision, Zustand oder Problementscheidungen geändert. |
| Ein reines Messe-Konto sieht ein Profil nicht | Im isolierten Messeumfang sind nur Messelistenprofile sichtbar. |
| Ein Artefakt lässt sich nicht herunterladen | Auftrag neu laden und prüfen, ob ein Admin das Artefakt nach der Erzeugung gelöscht hat. |
| Ein fehlgeschlagener Auftrag bietet Wiederholen | Zuerst protokollierten Fehler und Quelle prüfen. Wiederholen erzeugt eine kontrollierte Fortsetzung, kein unbeobachtetes Duplikat. |

## Dokumentierte RailKeeper-Version

Dieses Kapitel dokumentiert RailKeeper **v0.1.20.3** und wurde zuletzt am 27.08.2026 geprüft.
