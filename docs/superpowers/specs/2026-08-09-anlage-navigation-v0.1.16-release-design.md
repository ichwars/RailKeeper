# Anlage-Navigation und Release v0.1.16

**Datum:** 2026-08-09
**Status:** Zur Dokumentfreigabe
**Branch:** `dev/stage1-acceptance`

## Ziel

RailKeeper bereitet die vorhandenen Stage-1-Änderungen als stabile Version `0.1.16` vor. Der
Anlagenbereich bleibt technisch erhalten, wird in der Hauptnavigation vorübergehend deaktiviert und
erhält eine einheitliche Bezeichnung im Singular. Release-Dokumentation und öffentliche Hinweise
werden vollständig auf Deutsch und Englisch bereitgestellt.

## Navigationsverhalten

- Der deutsche Navigationstext lautet `Anlage` statt `Anlagen`.
- Der englische Navigationstext lautet `Layout` statt `Layouts`.
- Der Eintrag bleibt in der Sidebar an seiner benutzerdefinierten Position sichtbar.
- Der Eintrag ist nicht fokussierbar und löst keine Navigation aus.
- Der deaktivierte Zustand ist durch reduzierte Text- und Symbolintensität sowie einen neutralen
  Mauszeiger erkennbar. Hover- und Aktivmarkierungen werden nicht angezeigt.
- Ein lokalisierter Hinweis erklärt, dass der Bereich vorübergehend nicht über das Menü verfügbar
  ist. Für Hilfstechnologien wird der Zustand mit `aria-disabled` ausgezeichnet.
- In der eingeklappten Sidebar bleibt das Symbol sichtbar und erhält denselben lokalisierten Hinweis.
- Die Route `/layouts`, die React-Ansicht, Backend-Routen und vorhandene Daten bleiben unverändert.
  Direkte Aufrufe sind damit für Entwicklung und spätere Reaktivierung weiterhin möglich.

## Startseiten- und Einstellungsverhalten

- `Anlage` bleibt in der Sidebar-Konfiguration sichtbar, damit Reihenfolge und spätere Reaktivierung
  keine Benutzereinstellungen verlieren.
- `Anlage` kann vorübergehend nicht als Standardansicht ausgewählt werden.
- Ist `layouts` bereits als Standardansicht gespeichert, verwendet die Anwendung beim Start
  `Übersicht`. Die gespeicherte Einstellung darf dabei ohne ausdrückliche Benutzeraktion nicht
  destruktiv überschrieben werden.

## Version und Dokumentation

- Die Anwendungsversion in der Go-Anwendung wird von `0.1.15` auf `0.1.16` erhöht.
- `README.md` bleibt die englische Hauptdokumentation und verweist sichtbar auf die deutsche Fassung.
- `README.de.md` wird als vollständige deutsche Fassung ergänzt und verweist zurück auf Englisch.
- Beide README-Dateien beschreiben den Artikelbestand, die Inventarnummern, Lager- und
  Zuordnungsabläufe sowie den vorübergehend deaktivierten Anlagen-Menüpunkt korrekt.
- Veraltete fest angegebene Beispielversionen werden auf `0.1.16` aktualisiert, sofern sie die
  aktuelle stabile Installation beschreiben. Historische oder ausdrücklich als Beispiel markierte
  Betaversionen bleiben erhalten.
- `CHANGELOG.md` enthält das englische Changelog, `CHANGELOG.de.md` die inhaltlich äquivalente
  deutsche Fassung. Beide beginnen mit `0.1.16` und verlinken gegenseitig.
- Die Release-Notizen zu `v0.1.16` fassen die nutzerrelevanten Änderungen zweisprachig zusammen.

## Git- und Release-Ablauf

1. Den aktuellen Stand von `origin/main` abrufen und in `dev/stage1-acceptance` integrieren. Wegen der
   langen veröffentlichten Branch-Historie wird ein Merge bevorzugt, damit keine bereits
   veröffentlichten Commits umgeschrieben werden.
2. Konflikte eng am bestehenden Verhalten auflösen und fremde Änderungen vollständig erhalten.
3. Navigation, Version und Dokumentation implementieren und lokal vollständig prüfen.
4. Den Branch pushen und einen PR gegen `main` erstellen. Der PR beschreibt Stage 1, die
   Navigationsdeaktivierung, Migrationen, Backup-Version 3, die Importkompatibilität mit Version 1
   und 2 sowie den geplanten Release.
5. Alle verpflichtenden GitHub-Prüfungen abwarten. Fehler werden auf dem Branch behoben und erneut
   geprüft.
6. Den geprüften PR nach `main` mergen.
7. Den lokalen `main`-Branch per Fast-Forward auf den gemergten Remote-Stand synchronisieren. Der
   saubere Haupt-Worktree darf erst dann auf `main` wechseln, wenn sein vorhandener Bootstrap-Commit
   im Zielstand enthalten oder nachweislich inhaltsgleich ist.
8. Den signierten oder annotierten Tag `v0.1.16` exakt auf dem geprüften Merge-Commit erstellen und
   pushen.
9. GitHub Release, Windows-Portable-Paket und Docker-Image bis zum Abschluss überwachen.
10. Prüfen, dass das Release-Artefakt vorhanden ist und GHCR die Tags `latest`, `0.1.16` und
    `v0.1.16` ausliefert. Die vorhandene Trivy-Prüfung muss erfolgreich sein.

## Prüfung

Vor dem PR:

- gezielte Tests für Singularbezeichnung, deaktivierten Link, Tastaturverhalten, Startseiten-Fallback
  und englische Übersetzung
- vollständiger Frontend-Testlauf
- Frontend-Produktionsbuild
- vollständiger Go-Testlauf
- `git diff --check`
- visueller Test der Sidebar in Hell und Dunkel, ausgeklappt und eingeklappt sowie auf Desktop und
  Mobilgerät
- Smoke-Test von Anmeldung, Artikelübersicht, Artikelbearbeitung, Bestandsbuchung und `/health`
- Backup-Export und Backup-Validierung für das durch Stage 1 erweiterte Backup-Format

Nach dem PR-Merge:

- verpflichtende GitHub-Prüfungen auf dem Merge-Commit kontrollieren
- Tag-Workflows vollständig abwarten
- GitHub-Release, Windows-ZIP und Container-Tags kontrollieren
- nach Möglichkeit einen isolierten Start des veröffentlichten Containers mit temporärem Datenpfad
  durchführen und `/health` prüfen

## Abnahmekriterien

- Die Sidebar zeigt `Anlage` beziehungsweise `Layout` sichtbar, aber nicht bedienbar.
- Maus, Tastatur und Touch können die deaktivierte Navigation nicht auslösen.
- Ein gespeichertes `layouts` als Startansicht öffnet stattdessen die Übersicht.
- Direkte Anlagenroute und Backend bleiben unverändert funktionsfähig.
- Anwendung, README-Dateien, Changelogs und Release-Notizen weisen konsistent `0.1.16` aus.
- Alle lokalen und verpflichtenden Remote-Prüfungen sind erfolgreich.
- Der PR ist in `main` gemergt und der lokale `main` entspricht dem Remote-Stand.
- `v0.1.16` ist als stabiles GitHub-Release mit Windows-Artefakt und veröffentlichtem Docker-Image
  verfügbar.

## Nicht-Ziele

- Keine Entfernung oder Änderung der Anlagen-Datenmodelle, API oder Datenbankmigrationen.
- Keine allgemeine Überarbeitung der Sidebar oder ihrer Personalisierung.
- Keine neuen Cloud-, Mehrmandanten- oder Synchronisationsfunktionen.
- Kein Umschreiben der bestehenden Stage-1-Commit-Historie.
