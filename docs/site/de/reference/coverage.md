---
title: Was ist dokumentiert und geprüft?
description: Was ist dokumentiert und geprüft?
audience: reference
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Was ist dokumentiert und geprüft?

Alle 20 Themen haben jetzt eine deutsche und eine englische Einstiegsseite. Das ist eine Inhaltsabdeckung, keine Behauptung, dass jedes Gerät, jede Firmware oder jeder Sonderfall vollständig getestet wurde.

## So liest du den Prüfstand

| Nachweis | Aussage | Grenze |
| --- | --- | --- |
| Quellabgleich | Anleitung mit aktuellem Code beziehungsweise Betriebsregeln abgeglichen | Kein eigener Lauf jeder Variante |
| Demo-Oberfläche | Zusätzlich echte Screenshots aus einer isolierten Testinstallation | Künstliche Daten; keine angeschlossene Zentrale |
| Hardwareprüfung | Nur dort, wo ein Kapitel einen konkreten Gerätetest ausdrücklich nennt | Gilt für dieses Gerät und diesen Firmwarestand |

Die Bilder wurden am **23.10.2026** mit künstlichen Fahrzeugen aufgenommen. CV-Werte darin sind keine Werkseinstellungen. Die Seiten nennen ihren Versionsbezug und die redaktionelle Prüfung. Am Seitenende steht zusätzlich das Git-Änderungsdatum.

## Themen und Nachweise

<CoverageTable />

## Was bleibt offen?

- CS3-Kompatibilitätsdaten sind über dokumentierte Antwortformen und Fixtures geprüft; eine eigene vollständige Hardwareabnahme ist nicht belegt.
- Z21 und Intellibox 3 bieten noch keinen gleichwertigen Lok-/CV-Import wie ECoS. Eine Anleitung erfindet diesen Funktionsumfang nicht.
- Die Decoderbibliothek enthält weiter gekennzeichnete Modell-, Firmware- und Quellenlücken. Ein Katalogeintrag bedeutet nicht, dass jede CV dieses Modells interpretiert werden kann.
- Die Demo-Aufnahmen zeigen die Einstiegspunkte. Nicht jeder Dialogzustand, Druckertreiber, Restore oder Netzwerkausfall ist durch ein Bild nachgewiesen.

## Wie halten wir das aktuell?

Die Themenliste und Prüfhinweise kommen aus [docs/coverage.json](https://github.com/ichwars/RailKeeper/blob/main/docs/coverage.json). Die automatische Prüfung erfasst Routen, API-Bereiche, Übersetzungsschlüssel und Konfigurationsvariablen. Sie prüft Sprachpaare, Metadaten, Bilddateien und die Zuordnung zu Handbuchthemen. Inhaltliche Richtigkeit und Hardwareverhalten bleiben eigene Prüfaufgaben.

Änderungen an einem Arbeitsablauf aktualisieren seine Anleitung und beide Sprachen im selben Pull Request. Unter `docs` prüft `npm run check` Tests, Zuordnungen und den Website-Build. [Eine Lücke melden](https://github.com/ichwars/RailKeeper/discussions/categories/q-a).

[Die Kompatibilitätstabelle](command-station-compatibility) nennt je Anschluss Funktionen, Belege und offene Geräteprüfungen. Z21 besitzt einen begrenzten Abgleich bekannter Adressen, Intellibox 3 weiterhin nur Test und Diagnose.
