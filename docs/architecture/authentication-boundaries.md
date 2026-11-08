# Anmelderegeln und Persistenz

Die Authentifizierung besitzt abgegrenzte Anwendungsfälle und benannte Speicherports.
`internal/authentication` besitzt die Entscheidung über Kontobudget, Passwortprüfung,
zweiten Faktor, Hash-Upgrade und ein erfolgreiches Ergebnis. Das Modul importiert weder
SQL noch HTTP oder Dateizugriffe. Sein `Store`-Vertrag beschreibt die benötigten Abfragen
und den atomaren Abschluss; eine Passwortpolitik und eine Uhr sind gezielt austauschbar.
Tests prüfen Arbeit und Parallelität über kontrollierte Prüffunktionen, ohne SQLite und
ohne instabile Laufzeitvergleiche.

`internal/credentials` enthält die bestehenden Argon2id-, TOTP-, Token- und ID-Primitiven.
Hashformate, Tokenkodierung, Zufallsquellen und TOTP-Toleranz bleiben erhalten.
`internal/authentication/management` besitzt Kontoanlage/-änderung, Rollenregeln,
Sitzungen/CSRF, Passwortwechsel/-Reset und Zwei-Faktor-Verwaltung. Seine typisierten
Speicherports enthalten vorbereitete Änderungen; Passwortberechnung und reine
Entscheidungen haben keine SQL-Abhängigkeit.

`internal/infrastructure/authdata` implementiert Anmelde- und Verwaltungsports mit SQLite.
Der Programmstart stellt diesen Adapter ausdrücklich bereit. `application.AuthService`
erhält als Fassade seine vorhandenen Methoden und Transporttypen. Für ältere interne
Aufrufer ohne expliziten Port bleibt die bestehende Datenbank-Konstruktion kompatibel.
Diese Übergangskonstruktion ist auf die Fassade begrenzt.

## Transaktionsbesitzer

| Änderung | Besitzer und Grenze |
| --- | --- |
| Anmeldebudget reservieren | `infrastructure/ratelimit.Store.Allow`: abgelaufene Einträge, Grenze und neuer Versuch gemeinsam. |
| Erfolgreiche Anmeldung | `infrastructure/authdata.Store.CommitLogin`: aktuellen Hash/2FA-Zustand prüfen, TOTP verbrauchen, gegebenenfalls Hash aktualisieren, Sitzung und Audit schreiben, eigenes Budget leeren, gemeinsam committen. |
| Konto-/Rollenänderung | `authdata.Store`: Rollenprüfung, Schutz des letzten Administrators, Konto/Rollen, nötiger Sitzungswiderruf und Audit gemeinsam. |
| Passwortwechsel | `authdata.Store.ChangePassword`: geprüften bisherigen Hash erneut vergleichen, Passwort ändern, andere Sitzungen widerrufen, Browserbelege widerrufen und Audit gemeinsam. |
| Passwort-Reset | `authdata.Store.CreateResetRequest`/`CommitPasswordReset`: Anforderung einschließlich vorheriger Links und Audit atomar; Abschluss prüft Token, Gültigkeit und einmaligen Verbrauch erneut im selben Commit wie Passwort und Sitzungswiderruf. |
| Sitzung widerrufen/Logout | `authdata.Store`: Widerruf und Audit atomar; ungültige Sitzungen bleiben beim Logout idempotent, technische Fehler werden gemeldet. |
| Zwei-Faktor-Verwaltung | `authdata.Store`: Vorbereitung oder atomarer Aktivierungs-/Deaktivierungsnachweis mit TOTP-Verbrauch, Audit und gegebenenfalls Sitzungswiderruf. |
| Anmeldebudget freigeben | `authdata.Store.ResetLoginBudget`: Administratorrolle im Commit prüfen, Budgets/Browserbelege ändern und Audit gemeinsam. |
| Ersteinrichtung | `authentication/bootstrap` besitzt Einmalcode-/Administratorregeln und die Transaktionsfolge; `authdata` führt dieselben atomaren SQL-Änderungen aus. |

Eine zwischen Prüfung und Abschluss geänderte Anmeldung wird zurückgewiesen. Das gilt für
Passwort, 2FA-Secret und Aktivierung. Der Adapter überschreibt eine parallele Änderung
nicht. Fehler und abgebrochene Transaktionen geben keine Sitzung zurück. Es gibt weder
einen zweiten Anmeldeschreibpfad noch eine automatische Wiederholung einer Speicherung.

Die bisherige `RateLimitService`-Fassade delegiert ihre SQLite-Operationen an denselben
fokussierten Adapter wie das Kontobudget. Die IP-Begrenzung bleibt bestehen.

## Ersteinrichtung

`authentication/bootstrap` prüft den lokalen Einmalcode vor der Passwortableitung.
Innerhalb einer einzigen Transaktion werden bestehende Konten erneut geprüft, der Code
atomar verbraucht und genau das erste Konto samt Admin-/Editor-/Viewer-Rollen und Audit
angelegt. Fehlgeschlagene Rollen-, Audit- oder Commit-Schritte erhalten den ursprünglichen
Code für einen erneuten Versuch. Der vertrauenswürdige lokale Bootstrap bleibt ein eigener
Aufruf; die öffentliche HTTP-Einrichtung verlangt weiterhin den lokalen Code und ihr
bestehendes IP-Limit. Bestehende Konten lassen ihre Einrichtungserlaubnis unverändert.
Anwendungsbackups enthalten weiterhin keine lokalen Authentifizierungsdaten.

## Verbleibender Bestand und Nachweise

Benutzerverwaltung, Rollen, Passwort-Reset und Zwei-Faktor-Verwaltung führen in
`application` keine SQL-Befehle mehr aus. Auch die übrigen bisherigen SQL-Anwendungsfälle sind in fokussierten Fachmodulen und
Speicheradaptern getrennt. Die SQL-Restliste der Anwendung ist leer und bleibt CI-geprüft.

`go test ./internal/architecture` prüft die direkten Produktionsimporte der neuen Module.
Reine Regeln dürfen keine Persistenz- oder Transportpakete importieren; die Adapter dürfen
nicht auf die alte Anwendung oder das gesamte Infrastrukturpaket zurückgreifen. Go selbst
prüft zusätzlich Importzyklen. Diese Prüfung gehört über `go test ./...` zur regulären CI.

Die Sicherheitsfälle aus der [Qualitätsplanung](../quality-improvements.md) bleiben
verbindlich. SQLite-Integrationstests prüfen unveränderte Kontozustände, konkurrierende
Logins, verbrauchte TOTP-Schritte und vollständigen Rollback. Die synthetischen
[Upgrade-/Restorevergleiche](../testing-baseline.md) ergänzen diese gezielten Prüfungen.

## English summary

The independent authentication modules own login, account, session, password/reset and
second-factor and one-use first-run setup decisions behind consumer-defined ports.
SQLite adapters own the explicitly named atomic commits. The existing application facade
preserves callers and transport contracts; startup wires the production store. Credential
formats, existing accounts/sessions and public response fields remain compatible. Migration 81
adds private browser proofs without changing application backup format 28. AST import checks,
pure rule tests and real SQLite concurrency/rollback tests enforce these boundaries;
the application SQL ratchet is empty and rejects new concrete SQL calls.
