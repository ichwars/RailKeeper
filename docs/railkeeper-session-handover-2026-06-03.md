# RailKeeper Übergabedokument

Stand: 03.06.2026  
Arbeitsverzeichnis: `C:\Users\droth\Documents\GitHub\RailKeeper`  
Lokale App zuletzt unter: `http://localhost:8081`

## Aktuelles Ziel und Teilziele

Ziel dieser Session war, RailKeeper nach der Umbenennung von RailKeeper2 weiter in Richtung produktionsfähiger Modellbahn-Inventarverwaltung zu bringen und die ECoS-Anbindung strukturiert vorzubereiten.

Teilziele:

- Projektumgebung auf `RailKeeper` umstellen und Altartefakte von `RailKeeper2` entfernen.
- Lokalen Server auf aktuellen Stand bringen.
- Ersteinrichtung und Login optisch angleichen, inklusive Passwortwiederholung.
- Import/Export, Backup/Restore und Update-Bereich prüfen und abrunden.
- Digitalzentralen unter Einstellungen vorbereiten, provider-offen denken und ECoS als ersten aktiven Adapter anbinden.
- ECoS-Lokdaten lesen, interpretieren und in Fahrzeugentwürfe bzw. Importprüfung übergeben.
- Offene/unfertige Bereiche sichtbar markieren.
- Übergabestand als PDF dokumentieren.

## Kontext, der mehrere Turns gebraucht hat

Die Arbeit begann mit dem Projektkontext aus `docs/` und der Umbenennung von RailKeeper2 zu RailKeeper. Danach wurden GitHub-Repo, Package-Namen, lokale Docker-/Volume-Fragen und Release-Ablauf geklärt.

Mehrere UI-Bereiche wurden iterativ angepasst:

- Ersteinrichtung bekam Passwortwiederholung und wurde optisch an das Login angeglichen.
- Updates wurden in Einstellungen verdrahtet, inklusive Versionserkennung, Update-Hinweis in der Sidebar und visuellem Effekt bei verfügbarer Version.
- Import/Export wurde geprüft. Backup/Restore und Stammdaten-Import/Export gelten als brauchbar verdrahtet.
- Benutzerverwaltung wurde funktional erweitert, aber die finale UX wurde nach Feedback bewusst offen gelassen.
- Digitalzentralen wurden als eigener Tab ergänzt. Die erste Version war zu versteckt, dann wurde ein Adapter-Konzept mit Konfigurationsdialog eingeführt.
- ECoS-Zugangsdaten wurden wieder entfernt, weil RailKeeper bereits eine Benutzerverwaltung besitzt und die ECoS-Verbindung nur Host/Port benötigt.
- Der Admin-Status wurde korrigiert, damit angemeldete Admins Digitalzentralen konfigurieren, speichern und testen können.
- `cbries/railessentials` wurde als Vorlage analysiert. Brauchbar waren vor allem das Interpretieren von `funcset`, `funcdesc`, Lok-Attributen und der View-/Query-Ablauf.

## Produzierte Artefakte

Wesentliche Code-Artefakte:

- `backend/internal/ecos/`  
  Neuer ECoS-Parser/Client-Unterbau für TCP-Blöcke und Protokollantworten.
- `backend/internal/application/ecos.go`  
  ECoS-Verbindungstest, Lokdaten-Probe, Live-Monitor-Status und Parsing von Lokdaten, Funktionen und CV-Hinweisen.
- `backend/internal/api/router.go` und `backend/internal/api/routes.go`  
  API-Endpunkte für ECoS-Test, Rohdaten-Probe und Live-Monitor.
- `openapi/railkeeper.yaml`  
  API-Dokumentation für die neuen Digitalzentralen-/ECoS-Endpunkte.
- `frontend/src/features/settings/SettingsDigitalTab.tsx`  
  Neuer Settings-Tab für Digitalzentralen mit Adapterliste, ECoS-Konfiguration, Verbindungstest, Lokdatenprobe und Status.
- `frontend/src/features/importExport/ImportExportView.tsx`  
  ECoS-Lesen im Import/Export-Bereich, Arbeitsliste, Weitergabe an Fahrzeugentwurf und Sammelimport.
- `frontend/src/features/importExport/importExportHelpers.tsx`  
  ECoS-Matching, externe Zuordnung, CV-Hinweise und Funktionsvorschläge aus `funcdesc`/`funcset`.
- `frontend/src/features/vehicles/VehiclesView.tsx` und `frontend/src/features/vehicles/vehicleViewModel.ts`  
  ECoS-Fahrzeugentwurf, Aktualisierung bestehender Fahrzeuge, Speicherung von Mapping, CVs und Funktionen.
- `frontend/src/features/settings/SettingsAuthTab.tsx`  
  Benutzerverwaltung mit Passwortwiederholung und sichtbarer Markierung, dass das UX-Konzept offen bleibt.
- `frontend/src/styles/settings.css`  
  Markierungs-Styles für offene, experimentelle und designoffene Bereiche.
- `frontend/src/shared/i18n/de.ts` und `frontend/src/shared/i18n/en.ts`  
  Sprachtexte für Digitalzentralen, ECoS-Import und Arbeitsstatus-Markierungen.

Dokumentations-Artefakte:

- `docs/ecos-locomotive-field-comparison-2026-05-13.md`  
  Feldvergleich und ECoS-Einschätzung.
- `docs/railkeeper-session-handover-2026-06-03.md`  
  Quelle dieses Übergabedokuments.
- `docs/railkeeper-session-handover-2026-06-03.pdf`  
  PDF-Ausgabe dieses Übergabedokuments.

## Getroffene Entscheidungen und warum

- Digitalzentralen wurden provider-offen modelliert.  
  Grund: ECoS ist nur der erste Adapter; Z21 und CS3 sollen später ohne Umbau der Hauptnavigation folgen können.

- ECoS speichert keine Zugangsdaten.  
  Grund: Die ECoS-Verbindung benötigt im aktuellen RailKeeper-Konzept nur Host/Port. Benutzer- und Rechteverwaltung bleibt Aufgabe von RailKeeper.

- Live-Monitor heißt bewusst nicht vollwertiger Live-Sync.  
  Grund: Aktuell werden Daten gelesen und Diagnoseinformationen gesammelt. Automatischer bidirektionaler Sync und schreibende Steuerung sind noch nicht sicher genug.

- ECoS-Import bevorzugt externe Zuordnung statt Änderung interner IDs.  
  Grund: RailKeeper-ID und Inventarnummer müssen stabil bleiben. ECoS-ID, Adresse und Protokoll werden als externe Mapping-Daten gespeichert.

- ECoS-Funktionen werden konservativ gemappt.  
  Grund: `funcdesc`-Codes lassen sich teilweise auf Symbole/Funktionstypen abbilden, echte Gerätewerte müssen aber weiter validiert werden.

- Benutzerverwaltung bleibt gestalterisch offen markiert.  
  Grund: Funktional ist Passwortwiederholung vorhanden, aber die gewünschte Zusammenführung von Benutzerverwaltung, aktuellem Benutzer und Passwortbereich ist noch nicht entschieden.

- Offene Bereiche sind jetzt sichtbar markiert.  
  Grund: Vor Releases sollen experimentelle oder unfertige Bereiche nicht versehentlich als fertig erscheinen.

## Was blockiert

- Der In-App-Browser-Connector scheitert lokal weiterhin an einem Windows-Sandbox-Fehler. Dadurch kann Codex UI-Klicktests und rote Browser-Kommentar-Markierungen nicht zuverlässig selbst entfernen.
- ECoS-Tests hängen von echter Hardware und echten Lokdaten ab. Ohne Gerät können Verbindung, Parsing und Mapping nur technisch, nicht fachlich vollständig validiert werden.
- Z21 und Märklin CS3 sind nur vorbereitet. Es fehlen Backend-Adapter, Protokollparser, Testdaten und UI-spezifische Detailflüsse.
- Schreibender ECoS-Sync ist bewusst blockiert, bis Lesen, Matching, Konfliktlogik und Sicherheitsmodell mit echter Zentrale stabil sind.
- Die finale UX der Benutzerverwaltung ist offen, weil die bisherige Lösung nicht den gewünschten Eindruck getroffen hat.

## Nächste Schritte für die Fortsetzung

1. Browser-Connector-Problem lösen oder alternativen UI-Testweg festlegen.
2. Markierte unfertige Bereiche visuell prüfen: Einstellungen > Digitalzentralen, Import/Export > ECoS Live-Sync, Einstellungen > Authentifizierung.
3. ECoS mit echter Zentrale weiter testen:
   - Verbindung testen
   - Lokdaten lesen
   - Funktionen prüfen
   - CV-Hinweise prüfen
   - bestehendes Fahrzeug aktualisieren
   - neues Fahrzeug anlegen
4. ECoS-Funktionsmapping anhand echter `funcdesc`-Werte erweitern.
5. Benutzerverwaltung neu denken:
   - Trennung oder Zusammenführung von aktuellem Benutzer und Benutzerverwaltung entscheiden
   - Passwortänderung logisch platzieren
   - Admin- und Nicht-Admin-Sicht prüfen
6. Danach erst schreibenden Sync oder Live-Aktionen planen.
7. Vor Commit/Release alle temporären Arbeitsmarkierungen prüfen und bewusst entscheiden, welche bleiben dürfen.

## Getroffene Design-Entscheidungen

- Settings bleiben als sachliche Arbeitsoberfläche gestaltet: kompakte Karten, klare Tabs, wenig dekorative Ablenkung.
- Digitalzentralen verwenden ein Adaptermodell: Hauptseite zeigt Status und Überblick; Details werden über Adapter-Klick im Dialog bearbeitet.
- Offene Bereiche nutzen Badges statt Fehlermeldungen:
  - `Experimentell` für ECoS-Lese-/Monitorpfade.
  - `Offen` für noch nicht implementierte Provider oder Sync-Lücken.
  - `Design offen` für die Benutzerverwaltungsoberfläche.
- ECoS-Lokdaten werden nicht automatisch blind übernommen. Nutzer prüfen Entwürfe oder Importprüfung bewusst.
- Pflichtfelder aus ECoS-Entwürfen bleiben sichtbar markiert, damit unvollständige ECoS-Daten nicht still zu schlechten Stammdaten werden.
- Die Sidebar-Update-Anzeige ist bewusst dezent, aber sichtbar, und hat keine äußere Umrandung mehr.

## Verifikation

- Frontend-Build: `npm.cmd run build` erfolgreich.
- Backend-Tests: `go test ./...` erfolgreich.
- Lokaler Server lieferte zuletzt ein aktuelles Vite-Bundle aus.

