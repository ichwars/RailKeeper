# RailKeeper AGPL- und Community-Baseline

## Ziel

RailKeeper wechselt für alle ab dem Lizenzwechsel veröffentlichten Fassungen von der bisherigen
MIT Self-Hosting License zur GNU Affero General Public License Version 3. Die Umstellung soll
geschlossene Weiterentwicklungen und geschlossene gehostete Varianten verhindern, ohne ein
Bezahl-, SLA-, Mitgliedschafts- oder Supportmodell einzuführen.

## Lizenzentscheidung

- Der Repository-Inhalt wird unter `AGPL-3.0-only` veröffentlicht.
- `LICENSE.md` enthält den unveränderten offiziellen Text der GNU AGPL Version 3.
- Der Projektname, das Logo, fremde Marken, fremde Grafiken, fremde Dokumentationen und mögliche
  Protokollrechte werden nicht durch die Softwarelizenz freigegeben.
- Bereits veröffentlichte Fassungen bleiben unter den Lizenzbedingungen verfügbar, unter denen sie
  veröffentlicht wurden. Der Wechsel wirkt nicht rückwirkend.
- Die AGPL erlaubt kommerzielle Nutzung. RailKeeper selbst führt mit dieser Änderung jedoch keine
  kostenpflichtigen Funktionen, Leistungen oder bevorzugten Zugänge ein.

## Begründung des Wechsels

README, deutsche README und Changelog erklären den Wechsel inhaltlich übereinstimmend:

> RailKeeper ist eine lokale, selbst gehostete Anwendung. Die AGPL-3.0 stellt sicher, dass
> Weiterentwicklungen offen bleiben und dass Nutzer einer über ein Netzwerk bereitgestellten,
> veränderten Fassung Zugang zum korrespondierenden Quellcode erhalten. Die Umstellung schützt
> damit die dauerhafte Offenheit des Projekts besser als die bisherige freizügige Lizenz.

Die Erklärung behauptet weder ein Verbot kommerzieller Nutzung noch eine rückwirkende Änderung
bereits veröffentlichter Fassungen.

## Funding

`.github/FUNDING.yml` enthält ausschließlich:

```yaml
github: ichwars
ko_fi: ichwars
```

PayPal und Buy Me a Coffee werden aus Funding-Datei und READMEs entfernt. Die Supportabschnitte
verweisen auf freiwillige Tips ohne Gegenleistung. Tips begründen keinen Anspruch auf Software,
Support, Reaktionszeiten, Funktionen oder besonderen Zugang.

## Rechtliche Hinweise

`THIRD_PARTY_NOTICES.md` grenzt Drittanbieterrechte ab und nennt insbesondere:

- ECoS ist eine Marke der ESU electronic solutions ulm GmbH & Co. KG.
- RailKeeper ist ein unabhängiges Projekt und steht in keiner Verbindung zu ESU.
- Die RailKeeper-Lizenz überträgt keine Rechte an Marken, Grafiken, Dokumentationen oder
  Protokollrechten Dritter.
- Drittanbieterkomponenten und eingebundene Inhalte behalten ihre jeweiligen Rechte und Lizenzen.

`TRADEMARKS.md` erklärt, dass die AGPL keine Rechte am Namen oder Logo von RailKeeper einräumt.
Veränderte Fassungen dürfen ihre Herkunft nicht falsch darstellen oder eine offizielle Verbindung
zum RailKeeper-Projekt suggerieren. Eine Registrierung als Marke wird nicht behauptet.

## Community-Dateien

### CODEOWNERS

`.github/CODEOWNERS` weist `@ichwars` als Standardinhaber aus und enthält zusätzliche Regeln für:

- `backend/`, insbesondere Authentifizierung, Rollen, Sitzungen, Audit und Migrationen
- `frontend/` und die Abhängigkeits-Lockdatei
- `openapi/`
- `Dockerfile`, `docker-compose.yml`, `deploy/` und `.github/`
- Lizenz-, Marken-, Security- und Community-Dateien

Die CODEOWNERS-Datei besitzt sich über `/.github/ @ichwars` selbst. Eine verpflichtende
Codeowner-Freigabe wird für den alleinigen Maintainer nicht als Teil dieser Änderung aktiviert.

### Beiträge und Support

`CONTRIBUTING.md` beschreibt den vorhandenen Entwicklungs- und Prüfablauf. Es hält ausdrücklich
fest, dass eingereichte Beiträge unter `AGPL-3.0-only` lizenziert werden und dass Beitragende die
erforderlichen Rechte an ihren Änderungen besitzen. Ein CLA oder eine Rechteübertragung wird nicht
eingeführt.

`SUPPORT.md` verweist für Fehler und Funktionswünsche auf GitHub Issues beziehungsweise
Discussions, trennt Sicherheitsmeldungen auf `SECURITY.md` ab und stellt klar, dass kein SLA oder
Anspruch auf individuellen Support besteht.

`.github/PULL_REQUEST_TEMPLATE.md` verlangt eine kurze Änderungsbeschreibung, Testnachweise,
Dokumentationsauswirkungen, Sicherheits-/Datenhinweise und die Bestätigung der Beitragslizenz.

Ein kurzer `CODE_OF_CONDUCT.md` definiert respektvolle Zusammenarbeit und einen privaten
Meldeweg. Es werden nur Regeln aufgenommen, die der Maintainer praktisch durchsetzen kann.

## Metadaten und Dokumentation

- README-Badges und Lizenzabschnitte in Deutsch und Englisch nennen `AGPL-3.0-only`.
- `frontend/package.json` und der Root-Eintrag in `frontend/package-lock.json` erhalten die
  SPDX-Lizenzkennung.
- Das Runtime-Image erhält OCI-Labels für Projektquelle und Lizenz.
- Der Changelog dokumentiert Lizenzwechsel, Funding-Umstellung und Community-Dateien.
- `SECURITY.md` bleibt die verbindliche Meldestelle für Schwachstellen.

## Nicht Bestandteil

- keine kostenpflichtigen Funktionen, Sponsorstufen, Mitgliedschaften oder Supportpakete
- keine Änderung der Anwendungslogik oder Datenbank
- kein CLA, keine Übertragung von Contributor-Urheberrechten
- keine automatische Aktivierung eines GitHub-Rulesets
- keine Behauptung einer anwaltlichen oder behördlichen Freigabe

## Verifikation

- Suche nach verbliebenen Angaben zu MIT Self-Hosting, PayPal und Buy Me a Coffee
- Prüfung der YAML-, CODEOWNERS- und Package-Metadaten
- Abgleich der deutschen und englischen Lizenz- und Funding-Texte
- Kontrolle, dass der AGPL-Lizenztext unverändert ist
- Frontend-Build zur Prüfung der geänderten Package-Metadaten
