# Bei RailKeeper mitpflegen / Help maintain RailKeeper

## Deutsch

Der in [CODEOWNERS](.github/CODEOWNERS) genannte Maintainer koordiniert Reviews und
Abnahmen. Du kannst mit einem Quellenfund oder einem Testbericht einsteigen.

1. **Ersten Beitrag liefern.** Wähle eine Aufgabe in der [GitHub-Discussion zum Einstieg](https://github.com/ichwars/RailKeeper/discussions/427).
   Halte dein Ergebnis im jeweiligen Issue fest. Für einen Codebeitrag öffne einen kleinen
   PR und beschreibe die Änderung und ihre Prüfung.
   [Beitragsregeln](CONTRIBUTING.md)

2. **Ein Review übernehmen.** Stimme den Umfang in einem konkreten PR ab. Bei Änderungen an
   Authentifizierung, Backup und Restore, `safefetch` oder Release-Berechtigungen prüft eine
   zweite Person nach Möglichkeit den Diff und die relevanten Auswirkungen. Halte Commit,
   geprüfte Fälle, Befunde und offene Punkte im Review fest. Melde Sicherheitslücken privat.
   [Prüfschritte](docs/quality-gates.md) · [Sicherheitsmeldungen](SECURITY.md)

3. **Gemeinsam abnehmen.** Vereinbare mit dem Maintainer einen konkreten Build oder Tag.
   Wiederhole den vereinbarten Windows-, Upgrade- oder Restore-Test in einer getrennten
   Umgebung mit Beispieldaten. Dokumentiere Version und Commit, Plattform, Prüfschritte,
   erwartete und tatsächliche Ergebnisse sowie offene Punkte.
   [Unabhängige Build- und Restore-Abnahme](docs/independent-qualification.md)

Wenn du regelmäßig Reviews oder Abnahmen übernehmen möchtest, schreib in der Discussion,
welche Bereiche und Testumgebungen du abdecken kannst. Der Maintainer stimmt Zuständigkeit
und benötigte Repository-Berechtigungen mit dir ab.

## English

The maintainer named in [CODEOWNERS](.github/CODEOWNERS) coordinates reviews and qualification.
You can start with a source reference or a test report.

1. **Make a first contribution.** Pick a task from the [starter discussion](https://github.com/ichwars/RailKeeper/discussions/427).
   Record the result in its issue. For code contributions, open a focused PR and describe
   the change and how you verified it.
   [Contribution guidelines](CONTRIBUTING.md)

2. **Take on a review.** Agree on the scope in a specific PR. For changes involving
   authentication, backup and restore, `safefetch`, or release permissions, a second
   person reviews the diff and its relevant effects where possible. Record the commit,
   cases checked, findings and unresolved points in the review. Report vulnerabilities privately.
   [Quality checks](docs/quality-gates.md) · [Security reporting](SECURITY.md)

3. **Qualify a build together.** Agree on a specific build or tag with the maintainer.
   Repeat the agreed Windows, upgrade, or restore test in an isolated environment with
   sample data. Record the version and commit, platform, checks, expected and observed
   results, and unresolved points.
   [Independent build and restore qualification](docs/independent-qualification.md)

If you would like to review changes or qualify builds regularly, use the discussion to share
which areas and test environments you can cover. The maintainer will coordinate responsibilities
and any repository permissions you need.
