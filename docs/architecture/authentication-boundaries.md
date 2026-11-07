# Anmelderegeln und Persistenz

Die Anmeldung ist der erste abgegrenzte Anwendungsfall der Qualitätsverbesserungen.
`internal/authentication` besitzt die Entscheidung über Kontobudget, Passwortprüfung,
zweiten Faktor, Hash-Upgrade und ein erfolgreiches Ergebnis. Das Modul importiert weder
SQL noch HTTP oder Dateizugriffe. Sein `Store`-Vertrag beschreibt die benötigten Abfragen
und den atomaren Abschluss; eine Passwortpolitik und eine Uhr sind gezielt austauschbar.
Tests prüfen Arbeit und Parallelität über kontrollierte Prüffunktionen, ohne SQLite und
ohne instabile Laufzeitvergleiche.

`internal/credentials` enthält die bestehenden Argon2id-, TOTP-, Token- und ID-Primitiven.
Hashformate, Tokenkodierung, Zufallsquellen und TOTP-Toleranz bleiben erhalten.
`internal/infrastructure/authdata` implementiert den schmalen Anmeldevertrag mit SQLite.
Der Programmstart stellt diesen Adapter ausdrücklich bereit. `application.AuthService`
erhält als Fassade seine vorhandenen Methoden und Transporttypen. Für ältere interne
Aufrufer ohne expliziten Port bleibt die bestehende Datenbank-Konstruktion kompatibel.
Diese Übergangskonstruktion ist auf die Fassade begrenzt.

## Transaktionsbesitzer

| Änderung | Besitzer und Grenze |
| --- | --- |
| Anmeldebudget reservieren | `infrastructure/ratelimit.Store.Allow`: abgelaufene Einträge, Grenze und neuer Versuch gemeinsam. |
| Erfolgreiche Anmeldung | `infrastructure/authdata.Store.CommitLogin`: aktuellen Hash/2FA-Zustand prüfen, TOTP verbrauchen, gegebenenfalls Hash aktualisieren, Sitzung und Audit schreiben, eigenes Budget leeren, gemeinsam committen. |
| Andere Konten-/2FA-/Setup-Aktionen | Die bisherigen Anwendungsdienste behalten ihre vorhandenen Transaktionen. Sie sind noch nicht auf den Anmeldeport umgestellt. |

Eine zwischen Prüfung und Abschluss geänderte Anmeldung wird zurückgewiesen. Das gilt für
Passwort, 2FA-Secret und Aktivierung. Der Adapter überschreibt eine parallele Änderung
nicht. Fehler und abgebrochene Transaktionen geben keine Sitzung zurück. Es gibt weder
einen zweiten Anmeldeschreibpfad noch eine automatische Wiederholung einer Speicherung.

Die bisherige `RateLimitService`-Fassade delegiert ihre SQLite-Operationen an denselben
fokussierten Adapter wie das Kontobudget. Die IP-Begrenzung bleibt bestehen.

## Verbleibender Bestand und Nachweise

Benutzerverwaltung, Rollenänderungen, Passwort-Reset und andere Fachbereiche verwenden
weiterhin die dokumentierten älteren Anwendungs-/SQL-Grenzen. Eine erfolgreiche Umstellung
der Anmeldung ist kein Nachweis, dass der gesamte Anwendungsbereich bereits SQL-frei ist.
Weitere Bereiche werden einzeln mit ihren tatsächlichen Transaktionen übernommen.

`go test ./internal/architecture` prüft die direkten Produktionsimporte der neuen Module.
Reine Regeln dürfen keine Persistenz- oder Transportpakete importieren; die Adapter dürfen
nicht auf die alte Anwendung oder das gesamte Infrastrukturpaket zurückgreifen. Go selbst
prüft zusätzlich Importzyklen. Diese Prüfung gehört über `go test ./...` zur regulären CI.

Die Sicherheitsfälle aus der [Qualitätsplanung](../quality-improvements.md) bleiben
verbindlich. SQLite-Integrationstests prüfen unveränderte Kontozustände, konkurrierende
Logins, verbrauchte TOTP-Schritte und vollständigen Rollback. Die synthetischen
[Upgrade-/Restorevergleiche](../testing-baseline.md) ergänzen diese gezielten Prüfungen.

## English summary

The independent authentication module owns login decisions behind consumer-defined ports.
SQLite adapters own the explicitly named atomic commits. The existing application facade
preserves callers and transport contracts; startup wires the production store. Credential
formats, schema, backup format and public behavior remain unchanged. AST import checks,
pure rule tests and real SQLite concurrency/rollback tests enforce this first boundary;
other legacy application SQL remains explicitly outside this completed slice.
