# ECoS-Anbindung mit begrenztem Funktionsumfang

## Ziel

RailKeeper behält eine bidirektionale ECoS-Anbindung für die Verwaltung von Lokomotivstammdaten.
Die Integration liest zusätzlich statische Decoder- und Funktionsinformationen. Laufende
Betriebszustände und alle nicht für Lokomotivstammdaten benötigten ECoS-Objektmanager werden aus
Abfragen, Datenverträgen und Oberfläche entfernt.

Die Änderung reduziert den technischen und rechtlichen Umfang. Sie behauptet keine Zustimmung der
ESU electronic solutions ulm GmbH & Co. KG.

## Erlaubter Umfang

### Verbindung und Objektliste

- Verbindungstest über die Basisinformationen der Zentrale
- Lokomotivliste über den Lokomotivmanager mit Objekt-ID, Name, Decoderadresse und Protokoll
- Zugriff auf einzelne Lokomotivobjekte nur für die nachfolgend erlaubten Felder
- Verbindungsmonitor nur für Basisinformationen der Zentrale, ohne Abonnement anderer
  Objektmanager oder einzelner Lokomotivzustände

### Lesen

RailKeeper darf folgende statische Informationen lesen und in einer Vorschau anzeigen:

- Lokomotivname
- Decoderadresse
- Digitalprotokoll und Decoderprofil
- statische Funktionstasten-Definitionen und Funktionssymbol-Codes, insbesondere `funcdesc` und
  dokumentierte Zuordnungsinformationen, jedoch ohne aktuellen Ein-/Aus-Zustand
- CV-Werte einschließlich statischer Konfigurationswerte wie CV 29
- ECoS-Objekt-ID als externe Zuordnung

CV-Werte, die Fahrverhalten oder Richtungsoptionen konfigurieren, bleiben erhalten. Sie sind
Decoderkonfiguration und keine aktuelle Geschwindigkeit oder Fahrtrichtung.

### Schreiben

Der bestehende Schreibumfang bleibt auf folgende Lokomotivstammdaten beschränkt:

- Name
- Decoderadresse
- Digitalprotokoll

Jeder Schreibvorgang benötigt weiterhin eine Vorschau, einen Dry Run und eine ausdrückliche
Bestätigung. Diese Arbeit erweitert den Schreibumfang nicht auf CV-Programmierung,
Funktionsbefehle, Symbole oder Bilder.

### Funktionssymbole

Die lokal mit RailKeeper ausgelieferten ESU/ECoS-Funktionssymbole und ihre Zuordnungscodes bleiben
vorerst erhalten. Sie werden als Drittanbieterbezug gekennzeichnet und im späteren Anschreiben an
ESU ausdrücklich zur Prüfung gestellt. RailKeeper lädt bei einer ECoS-Abfrage keine zusätzlichen
Symbol- oder Bilddateien automatisch von der Zentrale.

## Ausgeschlossener Umfang

RailKeeper fragt nicht ab, speichert nicht, zeigt nicht an und schreibt nicht:

- aktuelle Geschwindigkeit und Fahrstufe (`speed`, `speedstep`)
- aktuelle Fahrtrichtung (`dir`)
- aktive oder inaktive Funktionszustände (`funcset`, zustandsbezogene `func`-Werte)
- Fahrbefehle, Funktionsbefehle sowie STOP-/GO-Befehle
- Schaltartikel und Magnetartikel
- Fahrwege
- S88- und andere Rückmeldedaten
- Booster
- weitere ECoS-Objektmanager außerhalb der Basiszentrale und des Lokomotivmanagers
- Lokomotivbilder oder Bildreferenzen (`icon`, `image`, `picture`, `pic`, `userimage`)

Die generische Möglichkeit des Clients, einen Textbefehl zu senden, bleibt ein internes
Transportdetail. Produktionscode erzeugt ausschließlich Befehle aus einer überprüfbaren
Allowlist.

## Backend-Änderungen

`backend/internal/application/ecos.go` wird auf einen expliziten Feld- und Befehlsumfang reduziert:

- Detailabfragen enthalten keine ausgeschlossenen Laufzeitfelder.
- Rohprobes fragen keine Bildfelder ab.
- Live-Subscriptions enthalten weder Lokomotiv-, Schaltartikel-, S88-, Fahrweg-, Booster- noch
  sonstige Objektmanager.
- `ECoSLocomotive` und `ECoSRawLocomotive` enthalten keine Geschwindigkeit, Fahrstufe,
  Fahrtrichtung, Funktionszustandsmenge oder Bildkandidaten.
- `ECoSFunction` enthält Funktionsindex und statischen Beschreibungscode, aber keinen Aktivstatus.
- Parser ignorieren ausgeschlossene Zustandsfelder auch dann, wenn eine unerwartete Antwort sie
  enthält.

Die Schreiblogik bleibt bei Name, Adresse und Protokoll. API-Handler bleiben Admin-geschützt und
schreibende Requests bleiben CSRF-geschützt.

„Lesen und Schreiben bleibt erhalten“ bezeichnet damit den heute vorhandenen bidirektionalen
Stammdatenabgleich. Die Änderung führt weder CV-Programmierung noch das Auslösen von
Funktionstasten ein.

## API und Frontend

- `frontend/src/shared/api.ts` und `openapi/railkeeper.yaml` verlieren die ausgeschlossenen Felder.
- Funktionsvorschläge entstehen ausschließlich aus Funktionsdefinitionen, nicht aus aktiven
  Zuständen.
- Hinweise wie „aktuell aktiv“, Zustandssterne und Richtungsanzeigen entfallen.
- ECoS-Bildvorschau und Bildübernahme entfallen.
- CV-Vorschau, Funktionsvorschau, Zuordnung, Konfliktprüfung und bestätigter Stammdatenabgleich
  bleiben erhalten.
- Die Live-Oberfläche wird als Verbindungsmonitor beschrieben und suggeriert keine Überwachung von
  Anlagen- oder Fahrzeugzuständen.
- Deutsche und englische Texte benennen den erlaubten Umfang und die ausgeschlossenen Befehle.

Gespeicherte RailKeeper-Fahrzeugdaten, vorhandene Funktionsdefinitionen, CV-Werte und lokale Bilder
bleiben unverändert. Es ist keine Datenbankmigration erforderlich.

## Dokumentation und Drittanbieterhinweise

README, ECoS-Oberflächentexte und Drittanbieterhinweise erklären:

- RailKeeper ist ein unabhängiges Projekt ohne Verbindung zu ESU.
- ECoS ist eine Marke der ESU electronic solutions ulm GmbH & Co. KG.
- Die Integration dient dem vom Nutzer veranlassten Austausch ausgewählter Lokomotivstammdaten.
- RailKeeper ist keine Fahrsteuerung und verwaltet keine Betriebszustände oder Anlagenobjekte der
  ECoS.
- Fremde Marken, Grafiken, Dokumentationen und mögliche Protokollrechte werden nicht durch die
  RailKeeper-Lizenz freigegeben.

## Tests

Backend-Tests prüfen:

- die exakten erlaubten Abfragefelder
- das Fehlen von Geschwindigkeit, Richtung, Zustandsfunktionen und Bildfeldern
- das Fehlen aller nicht erlaubten Objektmanager in Live- und Importbefehlen
- Parserverhalten bei unerwarteten ausgeschlossenen Feldern
- weiterhin funktionierende Lokliste, CV-Auswertung und Funktionsbeschreibungen
- Dry Run, Bestätigung und Allowlist des Stammdaten-Schreibens

Frontend-Tests prüfen:

- Funktionsvorschläge nur aus Beschreibungscodes
- keine Darstellung aktiver Zustände, Geschwindigkeit, Richtung oder ECoS-Bilder
- fortbestehende CV-, Symbol-, Zuordnungs- und Sync-Vorschauen
- deutsche und englische Umfangshinweise

Abschließend laufen `go test ./...`, der Frontend-Testlauf und der Frontend-Produktionsbuild.

## Anschreiben an ESU

Nach Umsetzung und Verifikation wird ein englischer und ein deutscher Entwurf für ein sachliches
Anschreiben erstellt. Das Schreiben enthält:

- Projektbeschreibung und Link zum öffentlichen Repository
- exakte Liste der gelesenen und geschriebenen Felder
- ausdrückliche Liste der ausgeschlossenen Steuerungs- und Objektmanagerfunktionen
- Herkunft und Nutzung der lokal enthaltenen Funktionssymbole
- AGPL-3.0-Lizenz und Drittanbieterhinweise
- Bitte um Prüfung und Benennung konkreter Einwände oder gewünschter Anpassungen
- Bereitschaft, beanstandete Bestandteile anzupassen oder zu entfernen

Das Schreiben wird nur als Entwurf erstellt und nicht ohne eine weitere ausdrückliche Freigabe an
ESU versendet.
