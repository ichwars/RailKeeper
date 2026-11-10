## Deutsch

### Korrekturrelease

Version, behobene Fehler und zugehörige Issues nennen. In der v0.2.x-Testphase bleiben neue
Funktionen separat. Die Veröffentlichung erfolgt nach Prüfung und Merge dieses PRs über den
vorhandenen Release-Workflow, ohne direkten Release-Commit auf `main`.

### Nachweise

- [ ] Änderungen und Datenerhalt geprüft; genaue Tests und Ergebnisse im PR beschrieben.
- [ ] Laufzeitversion, DE/EN-Changelogs, DE/EN-Release-Notizen und stabile Doku-Version stimmen überein.
- [ ] Betroffene Handbuchtexte und Bilder sind aktuell, Bildnachweise vollständig.
- [ ] Alle erforderlichen CI-Prüfungen dieses PR-Stands sind erfolgreich.
- [ ] Nach dem Merge Main-CI und Paketprüfungen prüfen; erst danach das Release empfehlen.
- [ ] Go-Test-/Coverage-Nachweise des tatsächlichen Release-Commits dauerhaft anhängen und in beiden
      Sprachfassungen der Release-Notizen verlinken. Bestehende Tags und Pakete nicht überschreiben.
- [ ] Keine Zugangsdaten, Laufzeitdaten oder privaten Sicherungen enthalten.

## English

### Maintenance release

State the version, fixed defects and related issues. Keep new features separate during the
v0.2.x trial. Publish through the existing release workflow after reviewing and merging this PR,
without a direct release commit to `main`.

### Evidence

- [ ] Verify changes and data preservation; describe exact checks and results in this PR.
- [ ] Align runtime version, DE/EN changelogs, DE/EN release notes and stable documentation version.
- [ ] Update affected handbook text and pictures, including complete capture provenance.
- [ ] All required checks for this PR revision have passed.
- [ ] After merging, check Main CI and package qualification before recommending the release.
- [ ] Retain Go test/coverage evidence from the actual release commit as permanent release assets
      and link it from both release-note languages. Do not overwrite existing tags or packages.
- [ ] Include no credentials, runtime data or private backups.
