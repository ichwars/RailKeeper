# Fortschrittsmodell der Artikelsuche

- Datum: 2026-08-28
- Status: entschieden
- Bezug: Issue #150

## Entscheidung

RailKeeper behält für die Artikelsuche den synchronen Endpunkt `POST /api/v1/article-search`
und einen einzigen wahrheitsgetreuen Ladezustand bei. Das Frontend zeigt keine benannten
Zwischenphasen an, weil der Server während des Requests keine Zustandsereignisse veröffentlicht.

Die empfohlene deutsche Statusformulierung lautet:

> Externe Suchtreffer und verfügbare Detailseiten werden geprüft. Bis zum Ergebnis werden keine
> Daten übernommen.

Die englische Entsprechung lautet:

> External search results and available detail pages are being checked. No data is applied before
> the result is ready.

Die Formulierung beschreibt beobachtbares Verhalten, verspricht keine einzelne Quelle und nennt
keine Phase, die das Frontend nicht belegen kann. Sie wird in der Ergebnisfläche verwendet. Der
kürzere Buttontext lautet `Artikeldaten und Quellen werden gesucht...` beziehungsweise
`Searching article data and sources...`. Dieses Dokument und die Textänderung verändern weder API
noch Laufzeitverhalten.

Ein serverseitiger Suchauftrag mit Status-Polling wird derzeit nicht eingeführt. Die begrenzte
Requestdauer und der lokale Einprozessbetrieb rechtfertigen den zusätzlichen persistenten Zustand,
die neuen Endpunkte und die komplexere Fehlerbehandlung nicht.

## Aktueller Vertrag

Der Handler `searchArticleData` in `backend/internal/api/data_handlers.go` liest einen JSON-Request,
ruft `ArticleSearchService.Search` synchron mit dem Request-Kontext auf und liefert genau eine
abschließende Antwort:

- `200` mit `ArticleSearchResponse`, auch wenn einzelne Quellen oder Detailseiten ausgefallen sind,
  solange verwertbare Ergebnisse vorliegen,
- `400 invalid_json` bei ungültigem JSON,
- `400 article_search_validation` ohne verwendbares Suchfeld,
- `504 article_search_failed` bei einem nicht anderweitig klassifizierten Suchfehler.

Der Service begrenzt die Gesamtsuche mit einem 10-Sekunden-Kontext. Der Frontend-Adapter in
`frontend/src/shared/api.ts` bricht den Fetch nach 15 Sekunden ab. Die Detailseiten erhalten
zusätzlich jeweils höchstens 3 Sekunden, Dokumentabrufe jeweils höchstens 4 Sekunden. Alle diese
HTTP-Aufrufe bleiben zugleich an das übergeordnete Request-Kontextlimit gebunden.

Eine Ausnahme bildet der optionale PDF-OCR-Fallback in `article_search_pdf.go`: Seine externen
Prozesse verwenden eigene, von `context.Background()` abgeleitete Limits. Sie können daher nach
Abbruch des ursprünglichen Requests noch bis zu ihrem eigenen Limit weiterlaufen. Das ist eine
bestehende Timeout-Lücke, aber keine Grundlage für eine erfundene Fortschrittsphase. Eine spätere
Änderung sollte den Request-Kontext explizit bis in die OCR-Extraktion durchreichen.

Der OpenAPI-Vertrag beschreibt den synchronen Endpunkt und die Statuscodes. Seine aktuellen
Artikelsuch-Schemas bilden jedoch nicht alle bereits ausgelieferten Felder ab, unter anderem
Suchquellen, Hersteller-Domains, Query-Trace, Ergebnis-Trace, Dokumente und Ersatzteile. Diese
bestehende Vertragsabweichung sollte unabhängig von einer Fortschrittsanzeige korrigiert werden.

## Tatsächliche Backend-Stufen

Die Stufen sind eine Dokumentation des heutigen Codes. Sie sind keine vom Server publizierten
Statuswerte und dürfen deshalb nicht zeitgesteuert im Frontend angezeigt werden.

### 1. Validierung und Normalisierung

Code: `searchArticleData`, `cleanArticleSearchInput`, `cleanArticleSearchSources` und
`articleSearchQuery`.

- **Ablauf:** Der Handler dekodiert JSON. Der Service trimmt Kernfelder und Zusatzfelder,
  normalisiert die Quellenauswahl und bildet den Suchtext aus Bezeichnung, Artikelnummer, EAN,
  Hersteller und Spurweite. Ohne Suchtext endet die Anfrage mit einem Validierungsfehler.
- **Überspringen:** Diese Stufe wird nicht übersprungen. Die Metadatenfunktion wird im aktuellen
  Code vor der abschließenden Leerprüfung aufgerufen, ist bei komplett leerer Eingabe aber ein
  No-op.
- **Fehler:** Ungültiges JSON und fehlende Suchkriterien werden als getrennte `400`-Probleme
  geliefert. Unbekannte Suchquellen werden verworfen; eine leere Auswahl fällt auf die
  Standardquellen zurück.
- **Timeout:** Kein eigenes Stufenlimit. Der Request-Kontext gilt; die Arbeit ist überwiegend lokal.
- **Zwischendaten:** Bereinigte Eingabe, Quellenauswahl und Query existieren nur im Service. Das
  Frontend erhält sie erst in der finalen Antwort, soweit sie Teil des Response-Modells sind.

### 2. Hersteller-Metadaten und Domain-Anreicherung

Code: `withManufacturerMetadata`, `matchManufacturerEntry` und
`preferredManufacturerDomains`.

- **Ablauf:** Bei vorhandenem Hersteller liest RailKeeper aktive Hersteller-Stammdaten. Passende
  Aliase, `searchDomains` und die Domain der Hersteller-Website ergänzen die interne Suchanfrage.
- **Überspringen:** Ohne Stammdaten-Service, Hersteller, passenden Eintrag oder Metadaten bleibt die
  Eingabe unverändert.
- **Fehler:** Ein Fehler beim Lesen der Stammdaten wird bewusst toleriert. Die Suche läuft ohne
  zusätzliche Aliase und Domains weiter; der Fehler erscheint nicht in der Antwort.
- **Timeout:** Kein eigenes Limit. Der lokale Stammdatenaufruf verwendet den Request-Kontext.
- **Zwischendaten:** Hersteller-Aliase und bevorzugte Domains sind intern verfügbar. Die finalen
  Hersteller-Domains und der Query-Plan sind Teil des Go-Response-Modells, aber nicht als laufender
  Status abrufbar.

### 3. Quellenabfrage

Code: `searchPikoSpareParts`, `searchRocoSpareParts`,
`DuckDuckGoArticleSearchAdapter.Search`, `articleSearchQueries` und `searchDuckDuckGo`.

- **Ablauf:** Optionale direkte PIKO- oder ROCO-Ersatzteilsuchen laufen zuerst. Danach erzeugt der
  Adapter bis zu elf deduplizierte Suchanfragen für Hersteller-Domains, Modellbahn-Fokus,
  Händler, Wiki und allgemeines Web. Die HTTP-Aufrufe laufen derzeit sequenziell. Eine reine
  EAN-Suche verwendet zunächst die EAN und bei Bedarf eine Modelleisenbahn-Fallbackquery.
- **Überspringen:** Direkte Herstellerabfragen laufen nur mit passendem Hersteller und explizitem
  `sparePartLookup`-Feld. Quellengruppen ohne Auswahl werden nicht abgefragt. Herstellerqueries
  entfallen ohne bevorzugte Domain.
- **Fehler:** Direkte Herstellerfehler werden zu `kein Direktresultat` reduziert. Bei den
  Suchqueries wird der erste Fehler gemerkt, weitere Queries laufen weiter. Sobald irgendein
  Ergebnis vorhanden ist, kann die Suche trotz Teilfehlern erfolgreich enden. Nur wenn kein
  Ergebnis vorliegt, wird der erste Adapterfehler bis zum Handler durchgereicht und dort als `504`
  abgebildet.
- **Timeout:** Global 10 Sekunden, zusätzlich 10 Sekunden HTTP-Clientlimit. Ein abgebrochener
  Browser-Fetch beendet über `r.Context()` die angebundenen HTTP-Aufrufe.
- **Zwischendaten:** Trefferlisten pro Query, Quellenname, Query und erster Fehler existieren nur im
  Adapter. Teilfehler werden nicht als strukturierte Antwortdaten ausgegeben.

### 4. Erste Extraktion und Scoring

Code: `parseDuckDuckGoResults`, `buildArticleFields` und `scoreArticleResult`.

- **Ablauf:** Jeder Such-Response wird in Treffer zerlegt. Titel, URL und Snippet werden bereinigt,
  Felder extrahiert und ein erster Score aus Artikelnummer, EAN, Hersteller, Domainart,
  Suchrang und fachlichen Merkmalen berechnet.
- **Überspringen:** Leere, unvollständige, blockierte oder nicht parsebare Treffer erzeugen keinen
  Ergebnisdatensatz. Fachspezifische Feldparser laufen nur bei passendem Kontext.
- **Fehler:** Parser liefern keine eigenen Transportfehler. Nicht erkennbare Daten fehlen im
  Ergebnis; sie machen den gesamten Request nicht fehlerhaft.
- **Timeout:** Kein eigenes Limit; die Verarbeitung läuft im globalen Suchkontext, prüft dessen
  Ablauf aber nicht zwischen jedem Parser-Schritt.
- **Zwischendaten:** Vorläufige Felder, Confidence-Werte und Scores sind intern vorhanden. Sie
  werden erst nach Detailanreicherung und finaler Sortierung ausgegeben.

### 5. Detailabruf und Anreicherung

Code: `enrichResultsFromPages`, `articleResultEnrichmentIndices`, `fetchArticlePage`,
`articleSparePartsFromMatchingDocuments` und `fetchArticleDocument`.

- **Ablauf:** Bevorzugte Hersteller- und Katalogtreffer werden zuerst angereichert, danach weitere
  Treffer bis zu einem Limit von zehn. RailKeeper lädt öffentliche HTTP(S)-Detailseiten, folgt nur
  den Regeln des Safe-Fetch-Clients, extrahiert zusätzliche Felder, Bilder und Dokumentlinks und
  lädt bis zu vier priorisierte Ersatzteildokumente. PDF-Dokumente können optional OCR verwenden.
  Nach erfolgreicher Anreicherung wird der Treffer neu bewertet. Der aktuelle Adapter kann bereits
  priorisierte Teilergebnisse anreichern und die deduplizierte Gesamtliste anschließend erneut
  prüfen; ein Detailabruf ist deshalb nicht zwingend einmalig pro URL.
- **Überspringen:** Treffer außerhalb des Zehnerlimits werden nicht angereichert. Dokumentabrufe
  entfallen ohne Artikelnummer, ohne geeignete Dokumente oder nach Erreichen des Ersatzteillimits.
  Seiten-Ersatzteile werden nur für bevorzugte Hersteller- oder Katalogdomains extrahiert.
- **Fehler:** Nicht öffentliche URLs werden vor dem Request abgelehnt. Seitenfehler und leere
  Antworten werden pro Treffer in `trace.error` festgehalten; der Treffer bleibt als Suchtreffer
  erhalten. Fehler einzelner Dokumente werden übersprungen und derzeit nicht separat ausgewiesen.
- **Timeout:** 3 Sekunden pro Detailseite und 4 Sekunden pro Dokument, jeweils unter dem globalen
  10-Sekunden-Limit. Der OCR-Fallback besitzt die oben beschriebene Kontextlücke und eigene Limits
  bis 45 Sekunden beziehungsweise 20 Sekunden je Tesseract-Versuch.
- **Zwischendaten:** Detailfelder, Bilder, Dokumente, Ersatzteile, finale URL und Trace-Zähler sind
  nach jedem Treffer intern verfügbar. Sie werden nicht gestreamt und erscheinen erst in der
  Abschlussantwort.

### 6. Konflikte, Sortierung und Antwort

Code: `articleSearchConflicts`, `dedupeArticleResults` und
`ArticleSearchService.Search`.

- **Ablauf:** Der Service markiert Abweichungen zu bereits eingegebenen Werten, sortiert stabil
  nach Score, dedupliziert nach URL, begrenzt auf zehn Ergebnisse und erzeugt die Antwort mit
  Query, Quellen, Hersteller-Domains, Query-Plan und Treffern.
- **Überspringen:** Konflikte entfallen für leere Bestands- oder Ergebniswerte. Deduplizierung und
  Sortierung laufen auch bei einer leeren Ergebnisliste.
- **Fehler:** Diese Stufe erzeugt im Normalfall keinen eigenen Fehler. Ein vorheriger harter Fehler
  führt statt einer Teilantwort zum `504`-Problem.
- **Timeout:** Kein eigenes Limit. Der Code prüft nach der Adapterrückgabe nicht erneut explizit
  `searchCtx.Err()`; maßgeblich ist, ob die aufgerufenen Operationen den Kontext beachtet haben.
- **Zwischendaten:** Alle finalen Daten liegen jetzt vor. Erst `respondJSON` macht sie für das
  Frontend sichtbar.

## Bewertung der Varianten

### Variante A: synchroner Endpunkt und präziser Einphasenstatus

**Vorteile**

- passt zum vorhandenen lokalen Einprozessbetrieb und zum bereits implementierten Request-Kontext,
- besitzt keinen zusätzlichen Serverzustand, keine Aufräumfrist und keine neue Datenbanktabelle,
- ein Neustart beendet den Request eindeutig, ohne verwaiste oder scheinbar laufende Aufträge,
- parallele Suchanfragen bleiben unabhängige HTTP-Requests,
- der vorhandene Endpunkt und bestehende Clients bleiben rückwärtskompatibel,
- Service-, Adapter-, Handler- und Frontendtests bleiben direkt und deterministisch,
- der Ladehinweis kann exakt das versprechen, was der Client weiß: Die Gesamtsuche läuft und es
  wird noch nichts übernommen.

**Nachteile und Grenzen**

- Nutzer sehen nicht, welche interne Teiloperation gerade läuft.
- Das Schließen eines Dialogs bricht den Fetch derzeit nicht ausdrücklich ab. Der API-Adapter
  erzeugt zwar einen Timeout-`AbortController`, die Controller-Hooks besitzen aber kein eigenes
  Abbruchsignal.
- Mehrere programmgesteuert gestartete Suchen können parallel laufen. Die sichtbaren Buttons sind
  während `loading` deaktiviert, aber die Hooks besitzen keine Request-ID, die verspätete Antworten
  einer älteren Suche sicher verwirft. Batch-Ersatzteilsuchen führen bewusst mehrere Requests aus.
- Teilfehler werden nur eingeschränkt sichtbar: Detailseitenfehler stehen im Treffer-Trace,
  fehlgeschlagene Queries und Dokumente nicht.

Diese Grenzen rechtfertigen bei der heutigen Laufzeit noch keinen Jobserver. Expliziter Client-
Abbruch, `latest-request-wins` und die OCR-Kontextweitergabe sind kleinere, getrennt lösbare
Robustheitsverbesserungen.

### Variante B: serverseitiger Suchauftrag mit Status-Polling

Ein wahrheitsgetreues Pollingmodell müsste mindestens einen Suchauftrag anlegen, den aktuellen
serverseitigen Zustand speichern und ein finales Ergebnis getrennt bereitstellen. Reale Phasen
könnten nur an Stellen gesetzt werden, an denen der Backend-Code die Stufe tatsächlich betritt oder
verlässt.

**Vorteile**

- echte, testbare Phasen und strukturierte Teilfehler wären sichtbar,
- ein expliziter Abbruch könnte serverseitig modelliert werden,
- längere OCR- oder Mehrprovider-Suchen könnten den Lebenszyklus eines einzelnen HTTP-Requests
  verlassen,
- das Frontend könnte nach einem vorübergehenden Verbindungsabbruch denselben Auftrag erneut lesen.

**Erforderliche Komplexität**

- neue Status- und Ergebnisschemas sowie mindestens Erstellen-, Lesen- und Abbruch-Endpunkte,
- eindeutige Benutzerzuordnung und serverseitige Rollenprüfung für jeden Auftrag,
- Grenzen für parallele Aufträge pro Benutzer und Installation,
- TTL, Aufräumlogik und Schutz vor unbegrenzt wachsenden Ergebnisdaten,
- persistente SQLite-Speicherung, falls Aufträge einen Prozessneustart überleben sollen,
- definierte Neustartzustände wie `interrupted` statt dauerhaftem `running`,
- Synchronisation konkurrierender Updates sowie idempotente Abbruchsemantik,
- neue OpenAPI-, Handler-, Application-, Infrastructure-, Frontend- und Migrationstests,
- eine Kompatibilitätsentscheidung: bestehenden synchronen Endpunkt behalten oder Clients auf
  `202 Accepted` und Polling umstellen.

Ein rein speicherbasierter Jobmanager wäre einfacher, verlöre aber bei jedem Neustart Auftrag und
Ergebnis. Das widerspricht dem Nutzen des Pollings und erschwert die Fehlererklärung. Persistente
Jobs wären robust, erweiterten jedoch Backups, Restore-Kompatibilität und Aufräumregeln für einen
kurzlebigen Suchvorgang.

### Entscheidungsmatrix

| Kriterium | Synchron, eine Phase | Job und Polling |
|---|---|---|
| Wahrheitsgetreuer aktueller Status | Ja, als Gesamtsuche | Ja, nach Backend-Umbau |
| Expliziter Abbruch | Clientseitig ergänzbar | Eigener serverseitiger Zustand nötig |
| Parallele Suchen | Unabhängige Requests | Quoten, Besitz und Synchronisation nötig |
| Prozessneustart | Request endet eindeutig | Persistenz oder `interrupted`-Verlustmodell nötig |
| OpenAPI | Bestehender Endpunkt, Schema-Drift korrigieren | Mehrere neue Schemas und Endpunkte |
| Testbarkeit | Bestehende direkte Tests | Gute Phasentests, aber deutlich mehr Zustandsfälle |
| Rückwärtskompatibilität | Vollständig | Nur mit parallelem Alt-Endpunkt ohne Bruch |
| Aufwand im lokalen Einprozessbetrieb | Niedrig | Hoch |
| Nutzerwert bei 10-Sekunden-Servicegrenze | Ausreichend | Derzeit zu gering |

## Abbruch und parallele Suchen

Die Entscheidung für Variante A bedeutet nicht, Abbruch und Konkurrenz zu ignorieren:

- Der Browser-Timeout bricht den Fetch ab; dadurch wird normalerweise auch der Go-Request-Kontext
  abgebrochen.
- Ein späterer UI-Abbruch sollte dasselbe Signal beim Schließen oder erneuten Starten verwenden.
- Ein späterer `latest-request-wins`-Schutz sollte Antworten anhand einer lokalen Request-ID oder
  eines AbortControllers zuordnen. Er benötigt keinen serverseitigen Job.
- Bewusste Batch-Suchen müssen weiterhin mehrere unabhängige Requests ausführen dürfen.
- Der OCR-Pfad sollte den Request-Kontext erhalten, damit der serverseitige Abbruch lückenlos ist.

Diese Punkte sind Robustheitsarbeit am synchronen Vertrag. Sie dürfen nicht als Fortschrittsphasen
dargestellt werden.

## Teststrategie für die gewählte Variante

Für eine spätere Umsetzung der präziseren Texte genügen Frontendtests, die belegen:

- während `loading` erscheint genau der Einphasenstatus,
- der Text nennt weder simulierte Phasen noch Prozentwerte,
- Ergebnis, Leerzustand und Fehler ersetzen den Ladezustand eindeutig,
- deutsche und englische Texte sind sinngleich,
- keine Daten werden vor einer expliziten Auswahl übernommen.

Backendtests bleiben für die tatsächlichen Stufen zuständig: Validierung, Query-Plan, Teilfehler,
Timeout, Detail-Trace, Sortierung und Antwort. Eine Fortschritts-UI darf nicht aus Test-Timern oder
künstlichen Wartezeiten abgeleitet werden.

## Bedingungen für eine spätere Neubewertung

Polling wird neu bewertet, wenn mindestens einer der folgenden Fälle real eintritt:

- die Suche soll zuverlässig länger als das heutige Requestlimit laufen,
- OCR oder weitere Provider sollen im Hintergrund fortgesetzt werden,
- Nutzer müssen laufende Suchen nach Navigation oder Verbindungsabbruch wieder aufnehmen,
- ein expliziter serverseitiger Abbruch mit überprüfbarem Endzustand wird Produktanforderung,
- strukturierte Teilergebnisse besitzen bereits vor der Abschlussantwort eigenständigen Nutzen.

Dann ist vor der Implementierung ein eigener Protokollentwurf mit Statusschema, Besitzmodell,
Persistenzentscheidung, TTL, Endpunkten, OpenAPI-Migration und Rückwärtskompatibilität erforderlich.
Zeitgesteuerte Frontendphasen bleiben auch dann ausgeschlossen.

## Architektur- und Baseline-Abschluss

Diese Entscheidung bestätigt den bestehenden synchronen API- und Modulzuschnitt. Sie ändert weder
die öffentliche API noch die Abhängigkeitsrichtung zwischen `api`, `application`,
`infrastructure` und Frontend. `docs/architecture.md` bleibt deshalb als aktuelle Baseline gültig.
Das vorliegende Dokument hält die Begründung und die Grenzen fest, ohne ein paralleles ADR-System
oder nicht ausgeführte Jobarchitektur als aktuellen Zustand zu deklarieren.
