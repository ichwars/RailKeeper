# Tillig-TT-Gleisplaner, Paket B

**Datum:** 2026-08-09

**Status:** Lokal umgesetzt und technisch abgenommen

**Abnahmestand 2026-08-09:** Paket B umfasst serverautoritatives Snapping, unmittelbar abgeleitete
Verbindungen und Prüfhinweise, symbolgestützte Anschlussmarker sowie eine Stückliste mit Soll,
physischem Bestand, Reserviert, Frei und Fehlend. Prüfhinweise wählen das betroffene Gleis direkt im
Plan aus. Der Materialabgleich verändert weder Katalog noch Bestand.

## Ziel

Paket B erweitert den maßhaltigen G1-Planer um magnetisches Anschluss-Snapping, eine jederzeit aus
dem aktuellen Plan abgeleitete Prüfung und eine Stückliste mit lokalem Bestandsstatus. Gespeichert
bleiben ausschließlich Gleisobjekte und ihre Pose. Verbindungen, Probleme und Materialbedarf werden
bewusst nicht dupliziert, damit sie nie vom Planstand abweichen können.

## Geometrische Regeln

- Alle Berechnungen arbeiten in Millimetern und Grad.
- Ein Snap-Kandidat liegt höchstens 8 mm vom bewegten Anschluss entfernt.
- Die Anschlussrichtungen müssen sich mit höchstens 5 Grad Abweichung gegenüberstehen.
- Beim Einrasten wird zuerst die Drehung korrigiert und anschließend der bewegte Anschluss exakt auf
  den Zielanschluss verschoben.
- Eine fertige Verbindung verwendet 0,25 mm Positions- und 0,5 Grad Richtungstoleranz.
- Ist ein fremder Anschluss höchstens 8 mm entfernt, zeigt aber in eine inkompatible Richtung, wird
  dies als Anschlusskonflikt statt als Verbindung bewertet.
- Paket B prüft Überlappungen gerader Fahrwege. Kurven-, Weichen- und Kreuzungssegmente folgen mit
  den entsprechenden verifizierten Geometriepaketen.

## Analysemodell

Die Analyse liefert Verbindungen, Probleme und Stücklistenzeilen. Probleme unterscheiden mindestens
`open_end`, `incompatible_connection`, `overlap` und `broken_geometry`. Offene Enden und fachliche
Warnungen blockieren einen Entwurf nicht. Beschädigte oder nicht auflösbare Geometrien gelten als
Fehler. Jede Meldung enthält betroffene Objekt- und Anschlusskennungen, damit die Oberfläche die
Geometrie direkt hervorheben kann.

Die Stückliste gruppiert nach konkreter Geometrieversion. Der Materialstatus verknüpft Tillig und
Artikelnummer deterministisch mit dem lokalen Zubehörkatalog. Er zeigt Bedarf, physischen Bestand,
aktive Reservierungen, frei verfügbare und fehlende Menge. Fehlende Katalogtreffer bleiben sichtbar
und werden nicht automatisch angelegt.

## Datenfluss und Zuständigkeiten

Reine Transformationen, Snapping und Planprüfung liegen in `backend/internal/domain`. Der
Anwendungsservice lädt den revisionsgebundenen Plan, wendet beim Verschieben den besten Snap an und
liefert die Analyse. Das Repository bleibt für SQLite-Abfragen, den Bestandsabgleich und versionierte
Schreibvorgänge zuständig. Ein neuer lesender API-Endpunkt liefert die Analyse für alle allgemeinen
Inventarrollen außer Messe.

Das Frontend spiegelt die Snap-Berechnung für die unmittelbare Ziehvorschau. Die Serverantwort bleibt
autoritativ. Anschlussmarker erhalten Symbolzustände für offen, verbunden und konfliktbehaftet.
Planprüfung und Stückliste erscheinen als ruhige, kompakte Seitenbereiche ohne native Browserdialoge.

## Abgrenzung

Paket B reserviert keinen Bestand und veröffentlicht keine Revision. Der bestätigte
Reservierungsworkflow und die Änderungsvorschau gegenüber der Basisrevision bilden Paket C. Flexgleis,
Höhenprofile und Modulports gehören weiterhin zu Stage 4.

## Prüfung

Golden-Tests decken Transformation, Toleranzgrenzen, beste Kandidaten, Rotationsnormalisierung,
Verbindungen, offene Enden und gerade Überlappungen ab. Repository- und API-Tests prüfen Rollen,
Bestandsaggregation und unveränderliche Revisionen. Frontendtests prüfen magnetische Vorschau,
Statusmarker, Warnungen, Stückliste, Leerzustände und schreibgeschützte Ansichten.
