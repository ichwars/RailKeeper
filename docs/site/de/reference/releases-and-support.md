---
title: Versionen und Hilfe
description: Versionen und Hilfe.
audience: reference
status: stable
reviewedVersion: 0.2.0
lastReviewed: 2026-11-10
---

# Versionen und Hilfe

## Welchen Stand benutze ich?

RailKeeper zeigt seine Version in der Oberfläche und unter **Einstellungen > Allgemein > Updates**. Ein Release ist eine veröffentlichte Version. `main` ist der aktuelle Entwicklungsstand und kann bereits weitere Funktionen enthalten. Screenshots stammen aus isolierten Demos; jede Seite nennt ihren redaktionellen Prüfstand. Installierte ältere Releases können abweichen. Identifiziere einen Entwicklungsbuild zusätzlich anhand seines Commits oder Paketnachweises.

Die Zeile **Build** zeigt einen kurzen Commit sowie getrennt **Release**, **Entwicklungsstand**,
**Lokale Änderungen** oder **Build-Herkunft unbekannt**. Über dem Commit erscheint der vollständige
Wert als Hinweis. Gib bei einer Fehlermeldung Version und Build an. `railkeeper version`
(Windows: `RailKeeper.exe version`) liefert dieselbe Kennung samt Compiler und Plattform,
ohne Anwendungsdaten zu öffnen. Ein älterer Server ohne Buildangaben behält seine Versionsanzeige.

Die sichtbare **redaktionelle Prüfung** einer Seite ist das Datum der Inhaltsprüfung. **Letzte Änderung** am Seitenende ist das Git-Änderungsdatum. Beides sagt allein nichts über einen Test mit echter Hardware aus. Die [Abdeckungsübersicht](coverage) erklärt den jeweiligen Nachweis.

Die Handbuchbilder zeigen **v0.2.0**. Aufnahmeversion und redaktioneller Prüfstand werden getrennt
angegeben. Das vollständige [deutsche Änderungsprotokoll](https://github.com/ichwars/RailKeeper/blob/main/CHANGELOG.de.md#020---2026-11-09)
und die [englische Fassung](https://github.com/ichwars/RailKeeper/blob/main/CHANGELOG.md#020---2026-11-09)
beschreiben die Änderungen dieses Releases.

Die [Release-Nachweise zu v0.2.0](https://github.com/ichwars/RailKeeper/blob/main/docs/releases/v0.2.0.md#go-test--und-coverage-nachweis)
verlinken den erfolgreichen Go-Testlauf sowie dauerhaft gespeicherte Coverage-Berichte.

## Rückmeldung zur v0.2.x-Testphase

Nutze die kurze [Vorlage für Erfahrungsberichte](https://github.com/ichwars/RailKeeper/issues/new?template=experience_report.md).
Nenne Version/Build, Installationsweg, optionale Zentrale und Testzeitraum. Was hat funktioniert,
was war unklar und welche Funktion war schwer zu finden? Es genügt eine Sprachfassung der Vorlage.
Die [erste Auswertung der Vereinsrückmeldungen](https://github.com/ichwars/RailKeeper/issues/376)
ist für den 23. November bis 7. Dezember 2026 vorgesehen. Bis dahin bleiben neue Funktionen
zurückgestellt; geprüfte 0.2.x-Fehlerbehebungen und passende Dokumentation erfolgen über PRs.
Eine Auswertung ohne eingegangene Rückmeldungen gilt nicht als Praxisabnahme.

## Wo gehört meine Frage hin?

| Anliegen | Richtiger Ort |
| --- | --- |
| Wie verwende ich eine Funktion? | [Handbuch](../guide/) und [Fragen & Antworten](https://github.com/ichwars/RailKeeper/discussions/categories/q-a) |
| Eine Idee oder ein gewünschter Arbeitsablauf | [Ideen](https://github.com/ichwars/RailKeeper/discussions/categories/ideas) |
| Reproduzierbarer Fehler | [Neues Issue](https://github.com/ichwars/RailKeeper/issues/new/choose) |
| Erfahrung aus der Testphase | [Erfahrungsbericht](https://github.com/ichwars/RailKeeper/issues/new?template=experience_report.md) |
| Sicherheitslücke oder private Daten | [Private Sicherheitsmeldung](https://github.com/ichwars/RailKeeper/security/advisories/new) |
| Pakete und verbindliche Versionshinweise | [GitHub-Releases](https://github.com/ichwars/RailKeeper/releases) |
| Aktueller Projekteinstieg | [Willkommen in den Diskussionen](https://github.com/ichwars/RailKeeper/discussions/1) |

## Eine hilfreiche Frage stellen

Nenne Version, Windows/Docker, Sprache, Arbeitsschritte, erwartetes und tatsächliches Ergebnis. Bei einer Zentrale helfen Modell und Firmware. Ein zugeschnittener Screenshot kann den betroffenen Bereich zeigen. Entferne vorher private Fahrzeugdaten, Passwörter, Sitzungstoken und interne Adressen, die du nicht veröffentlichen möchtest. Eine vollständige Sicherung gehört nicht in ein öffentliches Issue.

Eine Antwort, die das Problem löst, kann im Q&A-Bereich als Lösung markiert werden. So finden spätere Leser sie wieder.
