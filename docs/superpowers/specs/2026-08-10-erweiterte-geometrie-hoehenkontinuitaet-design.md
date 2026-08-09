# Erweiterte Geometrie, Paket D: Höhenkontinuität

**Datum:** 2026-08-10

**Status:** zur lokalen Umsetzung freigegeben, keine Veröffentlichung

## Ziel

Paket D prüft, ob geometrisch verbundene Gleise an ihrem gemeinsamen Anschluss dieselbe Höhe
besitzen. Ein Höhenversatz bleibt als bearbeitbarer Warnhinweis im Entwurf sichtbar und blockiert die
Veröffentlichung nicht. Damit wird aus einzelnen Höhenprofilen erstmals ein prüfbarer durchgehender
Gleisverlauf.

## Gewählter Ansatz

Die vorhandene geometrische Verbindungserkennung bleibt unverändert. Nachdem zwei Anschlüsse als
verbunden erkannt wurden, ermittelt die Domänenanalyse für beide Anschluss-IDs die jeweilige Höhe.
Bei Geometrien mit genau zwei geordneten Anschlüssen gilt:

- erster Anschluss: `elevationStartMm`,
- zweiter Anschluss: `elevationEndMm`.

Geometrien mit einer anderen Anschlussanzahl erhalten in diesem Paket keine Höhenkontinuitätsprüfung.
Eine pauschale Zuordnung des Endwerts zu mehreren Weichenanschlüssen wäre fachlich mehrdeutig und wird
nicht geraten.

## Analyse und Toleranz

Ab einer absoluten Differenz von mehr als 0,01 mm erzeugt die Analyse den Warnhinweis
`elevation_mismatch`. Die Toleranz dient ausschließlich der Stabilität von Fließkommaberechnungen und
ist kein fachlicher Anlagen-Grenzwert. Der Hinweis enthält beide Objekt- und Anschluss-IDs sowie die
absolute Differenz in Millimetern.

Die Verbindung selbst bleibt bestehen. Offene Enden, Stückliste, Materialverfügbarkeit und
Snapping-Verhalten ändern sich nicht. Die vorhandene Änderungsvorschau kann neue und behobene
Höhenversätze wie andere Prüfhinweise vergleichen.

## API und UI

`TrackPlanIssue` erhält das optionale Feld `elevationDifferenceMm`. OpenAPI und Frontend-Typen führen
den neuen Code und das Detailfeld. Die Planprüfung zeigt die Anzahl der Höhenversätze in der
Zusammenfassung. Der Warnhinweis nennt die lokalisierte Differenz und fokussiert beim Anklicken das
erste betroffene Gleis, dessen Höhenprofil anschließend direkt korrigiert werden kann.

## Fehler- und Grenzfälle

- Gleiche Höhen und Differenzen bis einschließlich 0,01 mm erzeugen keinen Hinweis.
- Negative Höhen werden wie positive Höhen verglichen.
- Ungültige Geometrien bleiben beim vorhandenen Fehler `broken_geometry` und werden nicht zusätzlich
  als Höhenversatz gemeldet.
- Mehrport-Geometrien werden nicht spekulativ bewertet.
- Die Sortierung der Hinweise bleibt deterministisch.

## Abgrenzung

Nicht Teil dieses Pakets sind konfigurierbare Steigungsgrenzen, automatische Höhenangleichung,
Höheninterpolation über mehrere Objekte, Weichenprofile mit Höhen je Anschluss, Durchfahrtshöhen,
Ebenenkollisionen und Flexgleise.

## Abnahme

- Eine verbundene G1-Gleiskante mit mehr als 0,01 mm Höhenversatz erzeugt genau einen Warnhinweis.
- Gleiche Höhen und die Toleranzgrenze erzeugen keinen Warnhinweis.
- Der Hinweis enthält Differenz, Objekte und Anschlüsse in stabiler Reihenfolge.
- Die Planer-Zusammenfassung zählt Höhenversätze und zeigt die Differenz lokalisiert an.
- Anklicken des Hinweises fokussiert das betroffene Gleis.
- Backendtests, Frontendtests und Produktionsbuild laufen erfolgreich.
- Die lokale Browserprüfung bestätigt Erzeugen und Beheben eines Höhenversatzes.

