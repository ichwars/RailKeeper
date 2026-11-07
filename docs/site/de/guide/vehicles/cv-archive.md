---
title: Gespeicherte CV-Stände
description: Historische CV-Werte, ihre Herkunft und aufbewahrte Originalquellen lesen.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-11-07
---

# Gespeicherte CV-Stände

<GuideFigure name="cv-archive" alt="Gespeicherter CV-Teilstand mit damaligem Fahrzeug, aufbewahrter Originalquelle und Herkunft der Werte." caption="Der synthetische Referenzbestand zeigt einen bewusst gespeicherten Teilstand. Erfassungszeit und Exportzeit der Quelle bleiben getrennt." />

Öffne **CV-Archiv** in der Werkzeugleiste des Bestands oder im **CV**-Reiter eines gespeicherten
Fahrzeugs. Das Bestandsarchiv enthält auch historische Fahrzeugbezüge nach dem Entfernen eines
Fahrzeugs. Viewer, Editor und Admin dürfen lesen. Messe hat keinen Zugriff auf diese Bestandsdaten.

| Stand | Bedeutung |
| --- | --- |
| Gespeicherter Arbeitsbestand | Die aktuellen gespeicherten CVs, die durch bestätigte Änderungen und Importe weiterbearbeitet werden. |
| Offener CV-Entwurf | Noch nicht übernommene Änderungen im Fahrzeugdialog. |
| Gespeicherter Teilstand | Die bei der Erfassung vorhandenen Werte, tatsächlichen Historien und Quellenbezüge. |

Suche nach Fahrzeug, alter Inventarnummer, Arbeitsbezeichnung oder Notiz und wähle einen Teilstand.
Die Details zeigen Fahrzeug und gegebenenfalls Decoderbeschreibung zum damaligen Zeitpunkt.
Spätere Modellkorrekturen, Ausbau oder Fahrzeugänderungen schreiben diese Angaben nicht um.
Unbekannte Ein- und Ausbaudaten bleiben unbekannt.

Jeder Stand ist ein **Teilbestand**. Sein Erfassungszeitpunkt belegt weder eine vollständige Auslesung
noch einen gemeinsamen Quellenexport. Einzelne Quellenzeiten behalten ihre Bedeutung und erhalten
keine erfundene Zeitzone. Fehlende Indexangaben bleiben von Seite `0/0` getrennt. Historische
unbekannte Integerwerte bleiben sichtbar. Klappe die gespeicherten Änderungen auf, um die CV-Historie
zu lesen.

**Original aufbewahrt** bietet die ursprüngliche Datei als privaten Download an. **Original fehlt**
erhält die Quellenangaben und ersetzt die Datei nicht durch eine andere. Der JSON-Download enthält
Werte, Historien, Hardwarebeschreibungen und Quellenangaben. Lade benötigte Originaldateien gesondert
herunter. Eine vollständige Anwendungssicherung einschließlich aufbewahrter Originale erhältst du
über [Sicherung und Wiederherstellung](../../administration/backup-restore).

Lesen, Suchen und Herunterladen im Archiv erhalten deine offenen Fahrzeug- und CV-Entwürfe.
Ein leeres Archiv bedeutet, dass im aktuellen Bestand noch kein gespeicherter CV-Teilstand vorliegt.
