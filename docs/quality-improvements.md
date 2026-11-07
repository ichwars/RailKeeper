# Qualität, Datenerhalt und schrittweise Verbesserung

RailKeeper verbessert bestehende Abläufe in kleinen, getrennt geprüften Schritten.
Fahrzeuge, Zubehör, Sets, Decoder/CVs, Quellen, Historien, Dateien und lokale Konten
bleiben erhalten. SQLite und der lokale Betrieb bleiben die Grundlage.

## Umfang und Abnahme

| Paket | Ergebnis | Erforderlicher Nachweis |
| --- | --- | --- |
| Q0 | Eindeutiger Kandidat und Zuordnung aller Befunde | Commit, Plattform, Schema, Backupformat und offene Prüfungen benennen. |
| Q1 | Hook-Linting, belastbare Coverage und abgestimmte Werkzeuge | Verpflichtende CI-Prüfungen; keine still wachsenden Ausnahmen. |
| Q2 | Kontobezogener Loginschutz, Dummy-Hash, einmaliges TOTP, kompatible Passwortprüfung, sichere Einrichtung/Cookies | Fehlversuche, Parallelität, Neustart und bestehende Konten prüfen. |
| Q3 | Verständliche Bestandsfilter und persönliche Ansichten | DE/EN, hell/dunkel, Tastatur, 2580/1440/1280/820/390 px und Vergrößerung prüfen. |
| Q3c | Manuell erfasste Wertentwicklung und ergänzende Sammlungsauswertung | Datierte Werte mit Herkunft; unbekannte Werte und unvollständige Abdeckung sichtbar halten. |
| Q4 | Klare fachliche Module und Persistenzgrenzen | Authentifizierung, Transfer, Zentralen, Backup und ausdrücklich Decoder/CV einzeln umstellen. |
| Q5 | Alltag, Wiederherstellung und nachvollziehbare Auslieferung | Plattformnachweise, echte Anwendererprobung und verbleibende Hardwaregrenzen dokumentieren. |

Wertentwicklung verwendet ausdrücklich eingegebene Bewertungen mit Bewertungsdatum,
Währung, Herkunft und Erfassungszeit. Listen- und Kaufpreise werden nicht überschrieben
oder als historische Marktwerte ausgegeben. Frühere Zeitpunkte und unbekannte Bewertungen
werden nicht erfunden. Auswertungen nennen Bewertungsumfang und fehlende Angaben.
Es gibt keine automatische Verbindung zu Preisportalen oder Cloud-Diensten.

Bewertungen werden als unveränderliche Einträge ergänzt. Eine Berichtigung oder Rücknahme
verweist auf ihren Vorgänger und bewahrt dessen Betrag, Herkunft und Erfassungszeit.
Konkurrierende Berichtigungen erfordern einen erneuten Vergleich. Bei einer verlorenen
Speicherantwort prüft die Oberfläche die bereits vergebene Eintragskennung, bevor weitere
Änderungen möglich sind. Die Stichtagsauswertung verwendet den heute bekannten Verlauf,
zählt Fahrzeugexemplare und summiert unterschiedliche Währungen getrennt und ohne Rundung.
Ein vollständiger JSON-Verlaufsexport und App-Backupformat 28 erhalten auch ersetzte Werte.
Schema 80 ergänzt ausschließlich eine zunächst leere Bewertungstabelle.

Im Fahrzeugbestand bleiben häufige Filter direkt sichtbar. „Weitere Filter“ bündelt
Gattung, Bahngesellschaft, Epoche, Adapter, Datenprüfung und Ausstellungstauglichkeit.
Aktive Filter erscheinen auch bei geschlossenem Bereich mit einzeln entfernbaren
Beschriftungen. „Spalten“ bietet eine ausdrücklich auswählbare kompakte Ansicht;
vorhandene Spalten und Breiten bleiben bis zur bewussten Änderung erhalten. Breite
Tabellen scrollen innerhalb ihres gekennzeichneten, per Tastatur erreichbaren Bereichs.

Unter Einstellungen bietet „Persönliche Ansicht“ die ausdrücklich auswählbaren Profile
„Sammlung“ und „Alle Funktionen“. Sammlung bündelt Ausstellung, Import/Export und
Digitalzentralen unter „Weitere Funktionen“; diese Werkzeuge bleiben innerhalb der
jeweiligen Berechtigungen erreichbar. Die Auswahl verändert weder die Rolle noch die
persönliche Reihenfolge oder offene Eingaben. Bestehende Ansichten werden nicht automatisch
umgestellt. CV-Arbeitsgruppen enthalten aktuelle bearbeitbare Werte, gespeicherte Teilstände
bewahren frühere Werte und Quellen. Beides ersetzt keinen Beleg für ein Decoderexemplar.

## Codequalität als eigener Bereich

- Fachliche Module haben schmale Verträge und überprüfbare Abhängigkeiten.
- Reine Regeln benötigen weder HTTP noch SQLite; Persistenzadapter besitzen Integrationstests.
- Fachlich atomare Änderungen behalten einen ausdrücklich benannten Transaktionsbesitzer.
- Neue oder wesentlich überarbeitete Dateien bleiben möglichst unter 500 Zeilen.
  Funktionen ab etwa 80–120 Zeilen werden auf eine fachlich sinnvolle Zerlegung geprüft.
- `VehiclesView` zeigt Bestand und Hilfsdialoge, `VehicleEditorHost` verbindet die Reiter
  mit den vorhandenen Controllern und `useVehicleWorkspace` koordiniert deren Lebensdauer.
  Die frühere, ausschließlich von ihrem eigenen Test geladene `importExportHelpers.tsx`
  entfällt. Aktive Transfer-Vorschauen und ECoS-Übernahmen behalten ihre vorhandenen Module.
  `ecos.go` und `backup.go` werden getrennt nach Verantwortungen entkoppelt.
  Dateilänge allein ist kein Qualitätsnachweis.
- TypeScript bleibt strikt; OpenAPI, Backend und generierte Frontendverträge bleiben synchron.
- Validierung, Parallelkonflikt, Abbruch und technische Fehler bleiben unterscheidbar.
  Eine verlorene Antwort darf keine blinde Wiederholung einer Speicherung auslösen.
- Go-Linting und React-Hook-Regeln sind verpflichtend. Coverage wird auf einem festen
  Messumfang abgesichert und gezielt auf 80 % Statements/Zeilen sowie 75 %
  Verzweigungen/Funktionen erhöht. Kritische Fallprüfungen gelten unabhängig davon.

## Vor jedem nächsten Schritt

Anmeldungen behalten die bestehende IP-Begrenzung. Zusätzlich gilt pro exakt getrimmter
Anmeldekennung ein Datenbankbudget von zehn Versuchen in fünf Minuten. Eine vollständig
erfolgreiche Anmeldung leert nur das eigene Budget. Abgewiesene weitere Versuche verlängern
das Fenster nicht. Gespeichert wird der SHA-256-Schlüssel der Kennung. Eine vorbereitete
Dummy-Passwortprüfung behandelt unbekannte Konten ebenfalls mit Argon2id. Höchstens zwei
Passwortprüfungen pro Authentifizierungsdienst laufen gleichzeitig; Überlastung liefert
denselben vorübergehenden HTTP-429-Status. Kontoerstellung oder Passwortänderung folgen
daraus nicht.

Passwort-Hashes enthalten ihre Argon2id-Parameter und werden mit diesen Werten geprüft.
Ungültige Formate und unbeschränkte Ressourcenanforderungen werden vor der Ableitung
abgelehnt. Bisherige reguläre Hashes bleiben unverändert gültig. Ein notwendiges Upgrade
erfolgt erst nach vollständiger Anmeldung samt zweitem Faktor; stärkere vorhandene
Parameter werden dabei nicht abgesenkt. Die Standardparameter bleiben bei 64 MiB,
einer Iteration, Parallelität vier und 32 Byte Schlüssel.

Erfolgreich verwendete TOTP-Zeitschritte werden je Konto dauerhaft gespeichert.
Aktivierung, Anmeldung und Deaktivierung verlangen jeweils einen noch nicht verbrauchten
Code. Die bestehende Toleranz von einem Zeitschritt bleibt erhalten. Bei der Anmeldung
werden Codeverbrauch, gegebenenfalls Hash-Upgrade, Sitzung, Audit und Budgetrücksetzung
gemeinsam bestätigt oder vollständig zurückgerollt. Schema 78 ergänzt ausschließlich
einen zunächst leeren Verbrauchszähler; bestehende Konten, Secrets und Sitzungen bleiben
erhalten. Lokale Authentifizierung bleibt aus App-Backups ausgeschlossen.

Eine leere Installation verlangt für die öffentliche Ersteinrichtung einen lokal erzeugten
Einmalcode aus `setup-code.txt`. Schema 79 speichert ausschließlich Hash und Verbrauch.
Codeverbrauch, erstes Konto, Rollen und Audit gehören zu einer Transaktion. Bestehende
Konten überspringen diese Vorbereitung. Interne Bootstrap-Aufrufe werden nicht als
öffentlicher Setup-Endpunkt angeboten. Session-Cookies verwenden standardmäßig `auto`:
Eine konfigurierte öffentliche HTTPS-URL aktiviert `Secure`, lokale HTTP-Nutzung bleibt
möglich. Explizite Vorgaben bleiben wirksam, Forwarding-Header verändern die Wahl nicht.

`cd frontend && npm run lint` prüft Produktionscode einschließlich React-Hook-Regeln.
Die Syntaxwerkzeuge liegen mit Babels TypeScript-/JSX-Parser in `tools/frontend-lint`;
die eigentliche Frontend-Typprüfung und der Build verwenden den Compiler aus
`frontend/package.json`. Es gibt keine Peer-Overrides und keine pauschal abgeschaltete
Hook-Regel. Typgenerierung bleibt separat unter `tools/api-types`.

Die isolierte Lintwerkzeugkette verwendet vorübergehend `minimatch` 11.0.5 als gezieltes
transitives Sicherheitsupdate für [GHSA-g3hv-75wg-m249](https://github.com/advisories/GHSA-g3hv-75wg-m249).
ESLint und seine Konfigurationsbibliothek fordern noch Major 10 an. Die verwendeten
Glob-APIs und Modul-Exporte bleiben passend; die Werkzeugkette wird unter Node 24 geprüft.
Diese Überbrückung betrifft keine Peer-Abhängigkeiten oder Laufzeitpakete und entfällt,
sobald ESLint selbst eine gepatchte Version anfordert.

Strukturänderungen und neue Verhaltensregeln werden getrennt geprüft. Vor Datenänderungen
werden Backup-Kompatibilität und vollständige Wiederherstellung berücksichtigt.
App-Backups enthalten weiterhin keine lokale Anmeldung. Migrationen sind additiv und
vorwärtsgerichtet; historische Migrationen werden nicht nachträglich geändert.

Die [Referenzprüfung](testing-baseline.md) vergleicht Inhalte, IDs, Beziehungen,
CV-Indexseiten, Quellen, Historien und Dateibytes. Sie ergänzt Rollen-/CSRF-, Fehler-,
Konflikt- und Wiederherstellungsprüfungen. Den exakten Auslieferungskandidaten und die
offenen Nachweise hält der [Pilot- und Rückfallablauf](modernization-rollout.md) auseinander.

Eine technische Prüfung ersetzt keine Anwendererprobung. Als Ziel dienen fünf Anwender
mit unterschiedlicher Erfahrung und ein 14-tägiger Pilot. Mindestens 90 % der zugewiesenen
Kernaufgaben sollen ohne Entwicklereingriff gelingen. Tatsächliche Teilnahme und
Geräteabnahmen werden erst nach ihrer Durchführung als bestanden bezeichnet.
Schreibtests an realen Zentralen brauchen weiterhin eine ausdrückliche Gerätefreigabe.

## English summary

The incremental quality work covers code boundaries, frontend lint/coverage, authentication,
clearer inventory workflows, dated manual valuations and decoder/CV domain logic.
Existing data, public contracts and capabilities must survive each step. Synthetic upgrade,
restore and browser checks complement platform qualification; actual user pilots and
hardware acceptance remain separate evidence. Application backups continue to exclude
local authentication. Valuations never fabricate past market values or contact external services.
Inventory filters have visible labels and individually removable active chips. Additional
filters can be collapsed without losing their effect. The column picker offers an explicit
compact preset while retaining existing preferences until the user changes them. Wide tables
remain in a labelled, keyboard-accessible scrolling region.
Personal view presets are opt-in. Collection groups specialist tools under More functions;
All functions shows every permitted workspace. Existing ordering, roles and open inputs
remain unchanged. Editable CV working values, recorded partial states and physical decoder
identity are described separately; stored profile labels are never renamed.
