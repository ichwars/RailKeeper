---
title: Veranstaltungen, Status und Sperren
description: Veranstaltungen, Status und Sperren.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Veranstaltungen, Status und Sperren

**Ziel:** Eine Veranstaltung vorbereiten und ihren Bearbeitungsstand bewusst festlegen.

<GuideFigure name="exhibition-locked" alt="Veranstaltungsarbeitsbereich mit gesperrter Demo-Liste." caption="Die tatsächliche Statusänderung ist gespeichert. Drucken und die zum Status passenden weiteren Aktionen bleiben sichtbar." />

## Rollen für Betrieb und Verwaltung

Admin und Konten mit **Editor plus Messe** verwalten Planung und Status. Messe pflegt bestehende
Einträge bei offener Planung. Das Löschen einer Veranstaltung oder eines Eintrags bleibt Admin vorbehalten.
[Zugriff im Überblick](./#zugriffsrechte-und-trennung).

## Liste anlegen

Wähle **Neue Veranstaltung**. Fülle Bezeichnung, Beginn und Ende aus; ergänze Ort, Beschreibung
und interne Organisation. **Speichern** legt die Veranstaltung an und wählt sie aus.
Prüfe den Zeitraum, bevor du Fahrtage bei Einträgen zuweist. **Abbrechen** verwirft den Entwurf.

## Listen auswählen, sortieren und prüfen

Wähle die Veranstaltung in der linken Liste. **Alle Veranstaltungen anzeigen** blendet archivierte
Veranstaltungen ein. Die aktuelle Oberfläche verwendet Veranstaltungsschaltflächen statt der früheren
sortierbaren Listentabelle. Datum und Status stehen im Veranstaltungskopf.

## Bezeichnung oder Datum bearbeiten

Wähle **Weitere Aktionen > Veranstaltung bearbeiten**, solange die Planung bearbeitbar ist.
Änderungen werden beim Speichern zusammen mit der erwarteten Revision geprüft. Ein fremder
zwischenzeitlicher Stand kann die Änderung abweisen; lade dann neu und vergleiche.
Ein geänderter Zeitraum kann die Bedeutung vorhandener Fahrtage verändern, daher danach alle Einträge prüfen.

## Liste sperren oder entsperren

Wähle **Liste sperren** im Veranstaltungskopf. Ohne Konflikte wird die Statusänderung direkt gespeichert.
Sind Konflikte vorhanden, öffnet sich eine Rückfrage.

<GuideFigure name="exhibition-lock-conflicts" alt="Sperrdialog bei noch vorhandenen Konflikten mit verpflichtender Begründung." caption="Der Negativtest zeigt, dass die Liste bei Konflikten nicht still gesperrt wird. In der Übung wird zunächst abgebrochen und die fehlende Schnittstelle ergänzt." />

Wähle vorzugsweise **Abbrechen**, korrigiere den Eintrag und prüfe erneut.
Eine bewusste Ausnahme verlangt eine Begründung und **Trotzdem sperren**. Die Probleme bleiben sichtbar;
die Ausnahme macht den Eintrag nicht vollständig bereit.
**Liste entsperren** öffnet die Planung wieder. **Weitere Aktionen** bietet unter anderem Start,
Abschluss, Wiederöffnung und Archivieren, soweit der jeweilige Ausgangsstatus dies erlaubt.

## Liste löschen

Wähle als Admin **Weitere Aktionen > Löschen** und lies die Bestätigung für die konkrete Veranstaltung.
Ihre Betriebseinträge gehören zu ihr. Sichere benötigte Daten vorher. Archivieren ist für eine spätere
Nachschau geeignet und wird von Löschen getrennt angeboten.

## Listenfehler beheben

| Meldung oder Zustand | Nächster Schritt |
| --- | --- |
| Fehlende Pflichtangaben | Bezeichnung und gültigen Zeitraum prüfen |
| Veralteter Stand | Neu laden und eigene Änderungen mit der aktuellen Revision vergleichen |
| Konflikte beim Sperren | Konflikte prüfen; korrigieren oder ausdrücklich begründete Ausnahme dokumentieren |
| Aktion fehlt | Rollen und den aktuellen Status prüfen |
| Veranstaltung scheinbar verschwunden | Archiv einblenden und Auswahl prüfen |

## Verwandte Seiten

[Einträge und Drucken](entries-and-printing) · [Bebilderte Übung](practice).
