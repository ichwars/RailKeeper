# Etappe 5: Gleisbibliotheken und Austauschformate

**Datum:** 2026-08-10

**Status:** lokal freigegeben für Issue #33, keine Veröffentlichung

## Ziel

RailKeeper kann zusätzliche, versionierte Gleisbibliotheken als dokumentiertes RailKeeper-JSON
prüfen, importieren, freigeben, stilllegen und wieder exportieren. Bestehende Planobjekte behalten
den beim Platzieren gültigen Geometriesnapshot, auch wenn eine Bibliothek später stillgelegt oder
eine neue Version importiert wird.

## Produktgrenze

Etappe 5 liefert das verlässliche Austauschfundament. Proprietäre AnyRail-, SCARM- und
WinTrack-Konverter werden erst ergänzt, wenn je Format eine belastbare, testbare Spezifikation oder
ein rechtlich nutzbarer Referenzkorpus vorliegt. Bis dahin ist
`railkeeper.track-library` Schema 1 das maßgebliche portable Format.

## Bibliotheksdokument

Das Dokument enthält:

- `format: "railkeeper.track-library"`
- `schemaVersion: 1`
- Hersteller, Gleissystem, Spurweite, Maßstab, Bibliotheksversion und primäre Quellen-URL
- Bibliotheks-Prüfstatus und Exportzeitpunkt
- 1 bis 500 Geometriedefinitionen mit Artikelnummer, Name, Art, Länge, optionalem Mindestradius,
  eigener Quellen-URL, Prüfstatus und Geometrie Schema 1

Unbekannte Felder sind für Schema 1 unzulässig. Namen und URLs besitzen feste Längenlimits.
Zahlen müssen endlich und Maße positiv sein. Ports, Routen und Punkte sind begrenzt und ihre IDs
müssen innerhalb einer Geometrie eindeutig sein. Es werden ausschließlich HTTP- und HTTPS-URLs
gespeichert, aber beim Import keine externen Inhalte abgerufen.

## Vertrauens- und Freigabemodell

Ein Import ist zweistufig:

1. Die Vorschau parst und validiert das gesamte Dokument ohne Schreibzugriff. Sie zeigt Metadaten,
   Definitionszahl, Warnungen und vorhandene Versionskonflikte.
2. Ein explizit bestätigter Import legt Bibliothek und Definitionen immer als `draft` an. Ein Admin
   muss die Bibliothek anschließend mit einer Prüfnotiz freigeben. Erst dann setzt RailKeeper
   Bibliothek und Definitionen auf `verified` und bietet ihre Geometrien im Planer an.

Eine Freigabe oder Stilllegung erzeugt Audit-Ereignisse. Versionen werden nicht überschrieben. Die
Kombination Hersteller, Gleissystem, Spurweite und Version bleibt eindeutig. Stillgelegte
Bibliotheken verschwinden aus der Palette, bereits platzierte Objekte bleiben nutzbar.

## Unveränderliche Plangeometrie

`plan_track_objects.geometry_snapshot_json` speichert beim Platzieren die vollständige
`TrackGeometryDefinition`. Revisionsklone kopieren diesen Snapshot. Lesen und Analyse verwenden
den Snapshot; für alte Datensätze bleibt ein kontrollierter Fallback auf die Definition bestehen.
Migration und Restore ergänzen fehlende Snapshots aus der zusammengehörigen Bibliotheksdefinition.

Damit verändern neue Bibliotheksversionen, Statuswechsel oder administrative Datenkorrekturen
weder Zeichnung, Prüfung, Diff noch Stückliste bestehender Pläne.

## API

- `GET /api/v1/track-libraries`
- `GET /api/v1/track-libraries/{id}/export`
- `POST /api/v1/track-libraries/import/preview`
- `POST /api/v1/track-libraries/import`
- `PUT /api/v1/track-libraries/{id}/status`

Lesen und Exportieren folgen dem vorhandenen Anlagen-Zugriff. Vorschau, Import und Statuswechsel
sind Admin-Aktionen und CSRF-geschützt. Fehler unterscheiden ungültiges Dokument, Konflikt und
nicht gefundene Bibliothek.

## Oberfläche

Der Planer erhält ein kompaktes Gleisbibliotheks-Panel. Es zeigt Hersteller, System, Spurweite,
Maßstab, Version, Quelle, Definitionszahl und Prüfstatus. Admins können eine JSON-Datei über die
app-eigene Dateiauswahl prüfen, die Vorschau bestätigen, Entwürfe mit Prüfnotiz freigeben und
Bibliotheken nach Bestätigung stilllegen. Export bleibt als RailKeeper-JSON verfügbar.

Alle Dialoge und Bestätigungen verwenden RailKeeper-Komponenten. Status, Warnung und Fehler werden
nicht nur über Farbe vermittelt. Deutsch und Englisch werden gemeinsam gepflegt.

## Kompatibilität

Die App-Backupversion steigt auf 14. Backups bis Version 13 werden akzeptiert und erhalten beim
Restore fehlende Snapshots. Bibliotheksimporte ändern keine Authentifizierungsdaten und greifen
nicht auf externe URLs zu.

## Abnahme

- Ungültige, übergroße, doppelte oder nicht endliche Dokumente werden vollständig abgewiesen.
- Vorschau schreibt keine Daten; Import erzeugt ausschließlich einen nicht platzierbaren Entwurf.
- Erst eine bestätigte Admin-Freigabe macht Geometrien sichtbar und platzierbar.
- Eine gleiche Bibliotheksversion erzeugt einen Konflikt, eine neue Version bleibt parallel.
- Export und erneute Vorschau erhalten alle Fachwerte.
- Eine nach Platzierung geänderte oder stillgelegte Definition verändert das Planobjekt nicht.
- Revisionsklon und Backup/Restore erhalten den Snapshot.
- OpenAPI, strikter TypeScript-Client, deutsche und englische UI, Tests und Build sind synchron.

Alles bleibt lokal auf `dev/issue-36-advanced-geometry`. Es erfolgt kein Push, Pull Request, Merge
oder Release.
