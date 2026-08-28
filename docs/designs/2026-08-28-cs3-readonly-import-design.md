# CS3: sicherer read-only Lokdatenimport

Stand: 28.08.2026

## Ziel

Der vorhandene Märklin-CS3-HTTP-Test wird zu einem sicheren, nachvollziehbaren read-only Adapter
ausgebaut. Eine konfigurierte und aktive CS3 kann ihre Lokliste in den bestehenden
Digitalzentralen-Vergleichsarbeitsbereich einlesen. RailKeeper schreibt keine Daten zur CS3 und
überwacht keine Fahr-, Richtungs-, Funktions- oder Anlagenzustände.

## Belastbare Schnittstellengrenze

Die CS3-Webanwendung stellt die Lokliste als JSON bereit. Die in öffentlich verfügbaren
Realgeräte-Fixtures beobachteten Endpunkte unterscheiden sich nach Firmwaregeneration:

| Firmwaregeneration | Endpunkt | RailKeeper-Behandlung |
| --- | --- | --- |
| vor 2.6 | `/app/api/loks` | unterstützter Legacy-Lesepfad |
| ab 2.6 | `/app/api/locos` | bevorzugter Lesepfad |

RailKeeper probiert zuerst den aktuellen Endpunkt und fällt nur bei einem klaren Not-Found-Ergebnis
auf den Legacy-Endpunkt zurück. Andere HTTP-Fehler werden nicht als Fallback kaschiert.

Referenzquellen:

- Märklin beschreibt CS3 und CS3 Plus als lokale Zentralen mit Lokdatenbank, veröffentlicht aber
  keinen stabilen Vertrag für die Web-App-Endpunkte.
- TrainControl unterstützt beide Endpunkte und dokumentiert den Wechsel für CS3-Firmware 2.6.0.
- Anonymisierte Regression-Fixtures in RailKeeper bilden nur die freigegebenen Felder realer
  Antwortformen nach. Sie enthalten keine privaten Anlagen- oder Lokdaten.

## Sicherheitsmodell

Externe CS3-Antworten gelten vollständig als nicht vertrauenswürdig.

- nur HTTP GET auf einen vom Administrator konfigurierten Host und Port
- keine Weiterleitungen
- kurze Verbindungs- und Lesetimeouts
- ausschließlich JSON-Content-Type für API-Antworten
- maximale Antwortgröße von 8 MiB
- maximal 5.000 Lokomotiven
- JSON muss ein Array aus Objekten sein
- je Lok nur `uid`, `name` beziehungsweise `internname`, `address` und `dectyp`
- keine Speicherung von Funktionen, Geschwindigkeit, Richtung, Icons, CVs oder Anlagenobjekten
- eindeutige, positive UID im 32-Bit-Bereich
- Decoderadresse innerhalb der bestehenden RailKeeper-Grenze
- begrenzte, bereinigte UTF-8-Namen und Protokollwerte

Ein HTTP-200-Ergebnis allein gilt nicht als Verbindungserfolg. Erfolg setzt einen passenden
JSON-Content-Type, ein valides Array und mindestens eine schema-konforme Antwortstruktur voraus.
Eine leere, aber valide Lokliste ist kompatibel und kann als leere Vorschau gelesen werden.

## Diagnose und Fehlerklassen

Ein eigener Admin-Endpunkt `/api/v1/digital-centers/cs3/probe` prüft beide bekannten API-Varianten
ohne Daten zu persistieren. Er liefert nur freigegebene Diagnosefelder:

- verwendeter API-Pfad
- erkannte API-Generation
- Content-Type
- HTTP-Status
- Anzahl gelesener Lokomotiven
- unterstützte read-only Fähigkeiten

Die Oberfläche unterscheidet mindestens:

- Netzwerk- oder Timeoutfehler
- Authentifizierungsfehler
- andere HTTP-Fehler
- Weiterleitung
- unerwartete Webseite beziehungsweise falscher Content-Type
- ungültiges oder zu großes JSON
- nicht unterstützte API-Version
- ungültige Lokdaten

## Importfluss

Der vorhandene generische Read-Session-Endpunkt bleibt unverändert. Die Workspace-Anwendung wählt
intern anhand des Providers den ECoS- oder CS3-Leser.

1. Serverkonfiguration lesen und aktive CS3 prüfen.
2. Read-Session mit den serverseitigen Fähigkeiten anlegen.
3. CS3-Lokliste einmal read-only abrufen und begrenzt dekodieren.
4. Felder in das bestehende providerneutrale Vergleichsformat normalisieren.
5. RailKeeper-Fahrzeuge read-only laden.
6. Bestehende Zuordnungen, eindeutige Adress-/Protokolltreffer, Abweichungen, neue Einträge,
   fehlende Einträge und Konflikte wie bei ECoS berechnen.
7. Nur die Vergleichsarbeitsliste persistieren.

Mehrdeutige Zuordnungen bleiben Konflikte. Namen allein erzeugen niemals einen automatischen
Treffer. Die externe CS3-UID wird als stabile Provider-ID verwendet, die Decoderadresse nur als
Vergleichsmerkmal.

## Fähigkeiten

Für `cs3` werden serverseitig genau diese Fähigkeiten aktiviert:

- `testConnection: true`
- `diagnose: true`
- `readLocomotives: true`

Diese Fähigkeiten bleiben deaktiviert:

- `liveMonitor`
- `writeLocomotives`
- `writeCVs`

Die vorhandene Digitalzentralen-Arbeitsfläche kann damit ohne zweiten Importdialog verwendet werden.
Die Einrichtungsseite zeigt Diagnose und read-only Import als verfügbar, Schreib- und
Monitoring-Funktionen weiterhin als gesperrt.

## API und Dokumentation

Backendroute, OpenAPI-Vertrag, strikter TypeScript-Client und deutsche sowie englische Texte werden
gemeinsam geändert. Die Betriebs- und Benutzerhilfe dokumentiert unterstützte Felder,
Firmwaregenerationen, bewusste Auslassungen und die fehlende Hardware-Verifikation je Fixture.

## Verifikation

- Parser- und HTTP-Tests mit anonymisierten Antwortformen für vor 2.6 und ab 2.6
- Grenztests für Content-Type, Redirect, HTTP-Status, Antwortgröße, JSON-Form, UID, Adresse,
  Duplikate und Unicode
- Workspace-Test für persistierte CS3-Vorschau und unveränderte Fahrzeugdaten
- API-, Rollen-, OpenAPI- und TypeScript-Clienttests
- Frontendtests für Diagnose- und Capability-Darstellung
- vollständige Backend-, Frontend-, Dokumentations- und Build-Prüfung
- manuelle Browser-QA bei 2580, 1440 und 820 px, jeweils hell und dunkel

