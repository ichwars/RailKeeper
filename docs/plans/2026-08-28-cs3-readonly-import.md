# Umsetzungsplan: sicherer CS3-Lokdatenimport

Issue: #135

## 1. CS3-HTTP-Grenze testgetrieben absichern

- Anonymisierte Fixtures für die API-Formen vor 2.6 und ab 2.6 ergänzen.
- Fehlschlagende Tests für aktuellen Endpunkt, Legacy-Fallback und Feldnormalisierung schreiben.
- Fehlschlagende Tests für Redirect, Authentifizierung, HTTP-Fehler, Content-Type, Größenlimit,
  ungültiges JSON, doppelte UID, ungültige Adresse und Unicode schreiben.
- Einen fokussierten CS3-HTTP-Reader mit begrenzter Dekodierung implementieren.
- Verbindungstest so umstellen, dass nur eine kompatible API-Antwort als Erfolg gilt.
- Eigenen Diagnoseaufruf mit freigegebenen Feldern implementieren.
- Go-Formatierung und fokussierte Anwendungstests ausführen.
- Commit: sichere CS3-HTTP-Grenze und Diagnose.

## 2. Read-only Import an den Vergleichsarbeitsbereich anbinden

- Fehlschlagende Capability- und Workspace-Tests für CS3 ergänzen.
- Providerabhängige Leserwahl in der bestehenden Read-Session implementieren.
- CS3-Lokomotiven in das vorhandene Vergleichsformat normalisieren.
- Sicherstellen, dass weder Fahrzeugdaten noch CS3-Daten geschrieben werden.
- Tests für bestehende Zuordnung, Vorschlag, Konflikt, neuen und fehlenden Eintrag ausführen.
- Commit: CS3-Lokliste als persistierte Vergleichsvorschau.

## 3. Backendroute und API-Vertrag synchronisieren

- Fehlschlagende Handler-, Rollen- und OpenAPI-Tests für `/digital-centers/cs3/probe` ergänzen.
- Adminroute und dünnen Handler implementieren.
- OpenAPI-Pfad und Antwortschemas präzisieren.
- Routenanzahl und CSRF-/Rollenprüfungen aktualisieren.
- Backendtests vollständig ausführen.
- Commit: CS3-Diagnose- und Importvertrag.

## 4. Strikten Client und zweisprachige Oberfläche aktualisieren

- Fehlschlagende Clienttests für den CS3-Diagnoseaufruf ergänzen.
- TypeScript-Client um `probeCS3Connection` erweitern.
- Einrichtungsdiagnose für CS3 aktivieren und HTTP-Anfragen verständlich darstellen.
- Capability-Darstellung auf CS3 read-only anpassen, Monitoring und Schreiben gesperrt lassen.
- Deutsche und englische Texte synchron ergänzen.
- Betroffene Frontendtests und Build ausführen.
- Commit: CS3-Diagnose und read-only Status in der Oberfläche.

## 5. Dokumentation und Kompatibilitätsmatrix abschließen

- README-Abgrenzung von Verbindungstest auf read-only Vorschau aktualisieren.
- Digitalzentralen-Hilfe um unterstützte Firmwaregenerationen, Felder und Auslassungen ergänzen.
- Herkunft und Grenzen der anonymisierten Fixtures dokumentieren.
- Dokumentationstests und VitePress-Build ausführen.
- Commit: CS3-Kompatibilität und sichere Nutzung dokumentieren.

## 6. Gesamtprüfung, PR und Merge

- `go test ./...` im Backend ausführen.
- `npm.cmd run test:coverage` und `npm.cmd run build` im Frontend ausführen.
- `npm.cmd run check` in `docs` ausführen.
- Browser-QA bei 2580, 1440 und 820 px in Hell und Dunkel durchführen.
- Read-only Vorschau, leere Liste, Fehlerzustand, lange deutsche Texte, Fokus und Überlauf prüfen.
- Arbeitsbaum und Diff prüfen, Branch pushen und zweisprachigen PR erstellen.
- GitHub-Checks, Reviews sowie allgemeine und Inline-Kommentare prüfen.
- Nur bei vollständig grünem Status regulär mergen, danach `main` synchronisieren und Branches
  bereinigen.

