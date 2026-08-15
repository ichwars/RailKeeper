# Zweisprachiges RailKeeper-Handbuch auf GitHub Pages

Datum: 15. August 2026

Status: fachlich freigegeben

Geltungsbereich: Dokumentationsplattform, Informationsarchitektur und Pflegeprozess

## Ziel

RailKeeper erhält ein öffentliches, zweisprachiges Handbuch auf GitHub Pages. Es soll den
vollständigen stabilen Funktionsumfang für Anwender und Administratoren sowie den aktuellen
Entwicklungsstand für Mitwirkende dokumentieren. Die Dokumentation muss auch bei wachsendem
Projektumfang auffindbar, überprüfbar und dauerhaft pflegbar bleiben.

Das Handbuch verwendet Englisch als öffentlichen Einstieg. Deutsch wird vollständig und fachlich
gleichwertig gepflegt. Eine Dokumentationsänderung gilt erst als abgeschlossen, wenn beide
Sprachfassungen vorliegen.

## Zielgruppen und Versionsbezug

Das Handbuch trennt drei Zielgruppen in der Hauptnavigation:

1. **Benutzerhandbuch:** Bedienung aller veröffentlichten Arbeitsbereiche und Funktionen.
2. **Installation und Administration:** Installation, Konfiguration, Betrieb, Sicherheit,
   Datensicherung und Fehlerbehebung.
3. **Entwicklung und Architektur:** Entwicklungsumgebung, Systemgrenzen, Verträge, Datenhaltung,
   Tests, Releases und Beiträge.

Benutzer- und Administrationsinhalte beschreiben den jeweils neuesten stabilen Release. Beim
ersten Handbuch-Release ist dies RailKeeper v0.1.17.5. Entwicklungs- und Architekturinhalt
beschreibt ausdrücklich den Stand von `main`.

Die öffentliche GitHub-Kennzeichnung „Latest Release“ verweist derzeit abweichend auf v0.1.14.
Vor Veröffentlichung des Handbuchs wird geprüft, ob Releases oder Release-Artefakte fehlen, und
v0.1.17.5 wird korrekt als neuester stabiler GitHub Release veröffentlicht beziehungsweise
gekennzeichnet. Diese Korrektur darf keine unvollständigen oder ungeprüften Release-Artefakte
veröffentlichen.

Historische Handbuchversionen sind in der ersten Ausbaustufe nicht vorgesehen. Release-Notizen
bleiben erhalten. Eine Versionsauswahl wird erst eingeführt, wenn mehrere RailKeeper-Versionen
parallel unterstützt werden oder relevante Unterschiede im Betrieb dies erforderlich machen.

## Technische Architektur

Das Handbuch wird mit VitePress als statische Website erzeugt und über GitHub Pages bereitgestellt.
Diese Wahl passt zum vorhandenen Node-, Vite- und TypeScript-Umfeld des Projekts und benötigt keinen
zusätzlichen externen Such- oder Hostingdienst.

VitePress stellt bereit:

- sprachabhängige Navigation und direkte Sprachumschaltung,
- lokale Volltextsuche im Browser,
- responsive Seiten mit heller und dunkler Darstellung,
- Markdown-basierte Inhalte mit Git-Historie und Pull-Request-Review,
- automatisierte statische Builds für GitHub Pages.

Die Dokumentationsquellen liegen unter `docs/` im Hauptrepository. Englisch verwendet den
Wurzelpfad der veröffentlichten Seite, Deutsch den Präfix `/de/`. Fachlich zusammengehörige Seiten
haben nach dem Sprachpräfix denselben Pfad, zum Beispiel:

```text
/guide/vehicles/maintenance
/de/guide/vehicles/maintenance
```

Die vorhandenen stabilen Dokumente in `docs/`, die README-Dateien, Release-Notizen und Screenshots
werden nicht blind dupliziert. Sie werden der neuen Informationsarchitektur zugeordnet, in beiden
Sprachen überarbeitet oder als weiterhin eigenständige Projektdateien verlinkt. Bestehende Links
sollen nach Möglichkeit erhalten oder gezielt weitergeleitet werden.

## Informationsarchitektur

### Benutzerhandbuch

Das Benutzerhandbuch umfasst mindestens:

- Schnellstart, Ersteinrichtung, Anmeldung und grundlegende Navigation,
- Übersicht, Kennzahlen, Suche, Filter und Datenqualitätsanzeigen,
- Fahrzeuge, Modelldaten, technische Daten, Eigentum und Leseansichten,
- Zubehör, Bestandsführung, Lagerorte, Reservierungen, Einbauten und Historie,
- Bilder, Anhänge, Webdokumente und Dokumentdownloads,
- Decoder, CV-Werte, Profile, Import, Export und Funktionstasten,
- Wartungen und Zustandshistorie,
- Artikel-, Produkt- und Ersatzteilsuche mit Quellenprüfung,
- PDF-Berichte, QR-Codes, Druck und Druckerauswahl,
- Ausstellung, Sperrstatus, Messe-Rolle und Listenansichten,
- CSV-, TSV-, XML- und JSON-Importe sowie kontrollierte Aktualisierungen,
- ECoS-Auslesung und die Grenzen der Z21- und CS3-Verbindungsadapter,
- Anlagenplanung, sobald der Arbeitsbereich als stabile Funktion veröffentlicht ist.

### Installation und Administration

Der Administrationsbereich umfasst mindestens:

- Windows-Portable- und Docker-Installation,
- Ersteinrichtung, Datenverzeichnis und Laufzeitmodell,
- Umgebungsvariablen und gespeicherte Systemeinstellungen,
- Benutzer, Rollen, Berechtigungsgrenzen und Sitzungen,
- SMTP und Passwort-Wiederherstellung,
- Backup, Validierung, Wiederherstellung und Kompatibilitätsprüfung,
- Updates, Release-Kanäle und Prüfungen nach Updates,
- TLS, Reverse Proxy, sichere Cookies und Schutz des Datenverzeichnisses,
- Upload-Grenzen, erlaubte Dateitypen, OCR und Druckerkonfiguration,
- Betrieb, Zustandsprüfung, Logs und Smoke-Tests,
- systematische Fehlerdiagnose und konservative Datenrettung,
- Sicherheits-, Datenschutz- und Local-first-Modell.

### Entwicklung und Architektur

Der Entwicklerbereich umfasst mindestens:

- lokale Entwicklungsumgebung und benötigte Werkzeuge,
- modularen Monolithen und Verantwortungsgrenzen,
- Backend-, Frontend- und Infrastrukturstruktur,
- OpenAPI-Vertrag, API-Konventionen und Fehlerformate,
- SQLite-Schema, Migrationen und Stammdaten-Seeds,
- Upload- und Dateispeicher sowie Pfadbegrenzung,
- Authentifizierung, Autorisierung, CSRF und Audit-Logging,
- Import-, Export- und Backupformate einschließlich Kompatibilität,
- Tests, Builds, statische Prüfungen und visuelle Qualitätssicherung,
- Release-Prozess, Versionsnummern, Artefakte und Release-Kanäle,
- Beitragsprozess und Pull-Request-Anforderungen,
- stabile Architekturentscheidungen und bewusst zurückgestellte Vorhaben.

### Bereichsübergreifende Inhalte

Volltextsuche, FAQ, Fehlerbehebung, Glossar und Release-Notizen sind aus allen drei Bereichen
erreichbar. Querverweise verbinden verwandte Abläufe. Fachliche Aussagen werden an einer
kanonischen Stelle gepflegt und von anderen Seiten verlinkt, um widersprüchliche Duplikate zu
vermeiden.

## Seitenstandard

Jede Funktionsseite folgt einem gemeinsamen Aufbau, soweit die einzelnen Abschnitte für das Thema
zutreffen:

1. Zweck und typische Einsatzfälle,
2. Voraussetzungen und benötigte Rolle,
3. Ablauf Schritt für Schritt,
4. sämtliche Felder, Aktionen, Statuswerte und Auswirkungen,
5. Validierungen, Grenzen und Sicherheitsabfragen,
6. häufige Fehler und deren Behebung,
7. verwandte Funktionen und weiterführende Seiten,
8. dokumentierter RailKeeper-Stand und letzte fachliche Prüfung.

Screenshots unterstützen Orientierung und komplexe Abläufe, ersetzen aber keine textliche
Erklärung. Wenn sichtbarer UI-Text für das Verständnis relevant ist, erhält jede Sprachfassung ein
Bild in der entsprechenden Sprache. Sprachneutrale Abbildungen dürfen gemeinsam verwendet werden.
Alle Bilder müssen im hellen oder dunklen Modus gut verständlich sein, relevante Zustände zeigen
und ohne personenbezogene oder produktive Daten auskommen.

## Vollständigkeitsnachweis

Eine Coverage-Matrix macht den Anspruch „alles“ überprüfbar. Sie wird aus folgenden Quellen
erstellt und bei funktionalen Änderungen fortgeführt:

- sichtbare Routen, Navigationseinträge, Ansichten, Tabs und Dialoge,
- deutsche und englische UI-Übersetzungsschlüssel,
- öffentliche API-Endpunkte und OpenAPI-Operationen,
- Rollen- und Berechtigungsprüfungen,
- Konfigurationswerte und Umgebungsvariablen,
- Migrationen, persistierte Hauptkonzepte und Dateispeicher,
- Import-, Export-, Backup- und Wiederherstellungsformate,
- bestehende README-, Betriebs-, Sicherheits-, Architektur- und Release-Dokumente.

Jeder Eintrag verweist auf seine englische und deutsche Zielseite oder trägt einen begründeten
Status, beispielsweise intern, noch nicht veröffentlicht oder bewusst ohne Benutzeroberfläche.
Ein Abschnitt gilt nicht allein deshalb als dokumentiert, weil sein Name in einer Übersicht
erscheint.

## Pflege- und Veröffentlichungsprozess

Funktionsänderungen aktualisieren die zugehörigen englischen und deutschen Seiten im selben Pull
Request. Die Pull-Request-Vorlage erhält dafür eine Dokumentationsprüfung. Automatische Prüfungen
ersetzen nicht die fachliche Übersetzungsprüfung.

Die Continuous-Integration-Prüfung umfasst mindestens:

- reproduzierbare Installation der festgelegten VitePress-Abhängigkeiten,
- erfolgreichen Produktions-Build,
- Prüfung interner Links und referenzierter Bilder,
- Prüfung, dass jede veröffentlichte Seite eine DE/EN-Gegenseite besitzt,
- Prüfung erforderlicher Seitenmetadaten,
- Erkennung versehentlich eingecheckter Build-Ausgaben.

Fehlende Sprachseiten, defekte Links, fehlende Medien oder ungültige Metadaten brechen den
Dokumentations-Build ab. Semantische Übersetzungsabweichungen werden im Review anhand einer kurzen
Checkliste geprüft. Kleine rein sprachliche Korrekturen dürfen nur eine Sprachfassung ändern, wenn
die fachliche Aussage nachweislich unverändert bleibt.

Nach erfolgreicher Prüfung veröffentlicht ein GitHub-Actions-Workflow die statische Seite aus
`main` auf GitHub Pages. Benutzer- und Administrationsseiten werden im Rahmen eines Releases auf
den neuen stabilen Stand gebracht. Entwicklerseiten können unabhängig davon mit `main`
fortgeschrieben werden. Die Startseite nennt klar den dokumentierten stabilen Release und verweist
separat auf die Entwicklerdokumentation.

## Umsetzungsetappen

### Etappe 1: Plattform und Bestandsaufnahme

VitePress, Sprachrouting, Suche, zurückhaltendes RailKeeper-Branding, GitHub-Pages-Workflow,
Qualitätsprüfungen und Coverage-Matrix werden eingerichtet. Die Navigation ist vollständig, auch
wenn einzelne Zielseiten zunächst eindeutig als noch nicht erstellt gekennzeichnet sind.

Abnahme: Beide Sprachen bauen lokal und in CI, Suche und Sprachwechsel funktionieren, die
Coverage-Matrix erfasst alle bekannten Quellen.

### Etappe 2: Benutzerhandbuch

Alle stabil veröffentlichten Bedienabläufe werden nach dem Seitenstandard dokumentiert. Komplexe
Abläufe erhalten geprüfte Screenshots und konkrete Beispiele.

Abnahme: Jeder stabile UI- und Benutzerablauf besitzt vollständige englische und deutsche Seiten;
offene Coverage-Einträge betreffen keine veröffentlichte Benutzerfunktion.

### Etappe 3: Installation und Administration

Installation, Konfiguration, Rollen, Backup und Restore, Updates, Sicherheit, Betrieb und
Fehlerbehebung werden zusammengeführt und vervollständigt.

Abnahme: Ein neuer Betreiber kann RailKeeper installieren, sicher konfigurieren, aktualisieren,
sichern, wiederherstellen und typische Fehler untersuchen, ohne auf interne Projektnotizen
angewiesen zu sein.

### Etappe 4: Entwicklung und Architektur

Die stabilen technischen Konzepte, Verträge und Entwicklungsabläufe werden auf Basis von `main`
dokumentiert. Flüchtige Implementierungsdetails werden nur aufgenommen, wenn sie für Verträge,
Sicherheit, Kompatibilität oder Beiträge relevant sind.

Abnahme: Ein neuer Mitwirkender kann das Projekt aufsetzen, die Modulgrenzen verstehen, Änderungen
testen und einen regelkonformen Pull Request vorbereiten.

### Etappe 5: Gesamtabnahme und Veröffentlichung

Coverage, Sprachgleichheit, Links, Medien, mobile und Desktop-Darstellung sowie helle und dunkle
Darstellung werden geprüft. Die GitHub-Release-Kennzeichnung wird korrigiert, bevor das Handbuch
öffentlich als Dokumentation für v0.1.17.5 bezeichnet wird.

Abnahme: Alle Coverage-Einträge sind dokumentiert oder begründet klassifiziert, beide Sprachen
sind vollständig, alle automatischen Prüfungen bestehen und GitHub Pages ist öffentlich
erreichbar.

## Qualitäts- und Sicherheitsgrenzen

- Das Handbuch darf keine Zugangsdaten, privaten Backups, produktiven Daten oder internen Pfade aus
  realen Installationen enthalten.
- Sicherheitskritische Anleitungen müssen mit den serverseitigen Rollen- und Schutzmechanismen
  übereinstimmen.
- Backup- und Restore-Anleitungen müssen Datenverlust-Risiken, Vorprüfung und Bestätigung klar
  benennen.
- Externe Artikel- und Suchdaten werden als Vorschläge dargestellt, nicht als vertrauenswürdige
  Stammdaten.
- Verborgene oder unfertige Funktionen werden nicht als stabiler Benutzerumfang dokumentiert.
- Die erzeugte Website und lokale Begleitartefakte werden nicht in Git eingecheckt.

## Nicht-Ziele der ersten Ausbaustufe

- keine separate native GitHub-Wiki-Instanz,
- kein externer Suchdienst und keine Dokumentations-Telemetrie,
- keine parallelen historischen Handbuchversionen,
- keine automatische maschinelle Übersetzung als veröffentlichte Endfassung,
- keine vollständige Referenz jedes internen Typs oder jeder privaten Funktion,
- keine Änderung des Local-first-, Sicherheits- oder Bereitstellungsmodells von RailKeeper.
