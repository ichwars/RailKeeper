---
title: Dateiimport mit Beispieldaten ausprobieren
description: Dateiimport mit Beispieldaten ausprobieren.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Dateiimport mit Beispieldaten ausprobieren

**Ziel:** Zwei Fahrzeuge importieren, eine Spalte selbst zuordnen und anschließend einen
Inventarnummernkonflikt lösen. Du brauchst **Admin** oder **Editor**. Verwende einen Testbestand;
die Bestätigung legt echte Datensätze in deiner jeweiligen Installation an.

Alle Angaben sind künstlich. Die Dateien enthalten keine Bilder, CVs oder Decoderprojekte.
Für Decoderwerte gibt es ein [eigenes CV-Beispiel](../vehicles/cv-import-example).

## Dateien herunterladen

| Datei | Verwendung |
| --- | --- |
| <GuideDownload file="vehicle-import-demo.csv">Fahrzeuge als CSV</GuideDownload> | Zwei neue Fahrzeuge; die zusätzliche Spalte `myNote` wird manuell zugeordnet. |
| <GuideDownload file="vehicle-import-conflict.csv">CSV mit einem Konflikt</GuideDownload> | Nach dem ersten Import: eine vorhandene Inventarnummer und ein neuer Wagen. |
| <GuideDownload file="vehicle-import-demo.json">Fahrzeuge als RailKeeper-JSON</GuideDownload> | Zwei weitere Fahrzeuge im versionierten Austauschformat. |

Speichere die Dateien mit ihren ursprünglichen Endungen. Die CSV ist UTF-8 und verwendet Semikolons.
Die JSON-Datei ist ein RailKeeper-Transferpaket der Version 3, keine Anwendungssicherung.

## 1. CSV-Profil und erste Vorschau

1. Öffne **Import/Export > Profil anlegen**.
2. Vergib den Namen **Demo CSV**, wähle **Import**, **Fahrzeuge** und **CSV** und lege das Profil an.
3. Wähle **Neuer Import**, das Profil **Demo CSV** und `vehicle-import-demo.csv`.
4. Warte auf die **Erkannte CSV-Zuordnung**. Bisher wurde kein Fahrzeug importiert.

<GuideFigure name="import-csv-mapping" alt="CSV-Zuordnung mit der noch offenen Quellspalte myNote." caption="Die bekannten Spalten werden erkannt. myNote ist absichtlich noch offen. Künstliche Demodaten; anklicken für die volle Größe." />

## 2. Eine Spalte selbst zuordnen

Wähle bei `myNote` das Zielfeld **Beschreibung**. Aktiviere **Zuordnung im Profil speichern**,
wenn du die Zuordnung für die zweite Datei wiederverwenden möchtest. Wähle **Zuordnung prüfen**.
Jede Spalte muss zugeordnet oder ausdrücklich ignoriert sein.

Prüfe in der Vorschau diese zwei neuen Inventarnummern:

| Inventarnummer | Bezeichnung |
| --- | --- |
| `DEMO-LOK-001` | Demo-Lok 01 |
| `DEMO-WAG-001` | Demo-Wagen 01 |

<GuideFigure name="import-csv-review" alt="Geprüfte CSV mit zwei neuen Fahrzeugen vor der Übernahme." caption="Zwei neue Datensätze sind bereit. Erst die Import-Schaltfläche übernimmt sie. Künstliche Demodaten." />

## 3. Import bestätigen und Ergebnis kontrollieren

Wähle **2 Datensätze importieren**. Der Dialog schließt nach erfolgreicher Übernahme.
Öffne den abgeschlossenen Auftrag und kontrolliere danach die beiden Fahrzeuge im Bestand.
Ihre Beschreibung lautet durch die Zuordnung **Nur Übungsdaten / Synthetic example**.

<GuideFigure name="import-csv-result" alt="Auftragsdetails nach dem erfolgreichen CSV-Import." caption="Der abgeschlossene Auftrag zeigt das Ergebnis; kontrolliere zusätzlich die Fahrzeuge im Bestand. Künstliche Demodaten." />

## 4. Einen Konflikt bewusst überspringen

1. Starte mit **Demo CSV** einen weiteren Import und wähle `vehicle-import-conflict.csv`.
2. Prüfe die Spaltenzuordnung erneut. Falls sie nicht im Profil gespeichert wurde, ordne `myNote`
   wieder **Beschreibung** zu. Gehe zur Prüfung weiter.
3. Bei `DEMO-LOK-001` erscheint ein Konflikt, weil die Inventarnummer bereits vorhanden ist.
   Die Quelldatei nennt diese Lok jetzt **Demo-Lok 01 Update**.
4. Wähle für diese Zeile **Diesen Datensatz nicht importieren**.
5. Prüfe den neuen Wagen `DEMO-WAG-002` und wähle **1 Datensatz importieren**.

<GuideFigure name="import-csv-conflict" alt="Inventarnummernkonflikt vor der Auswahl einer Lösung." caption="Eine vorhandene Inventarnummer braucht eine Entscheidung. Eine zweite Zeile kann neu angelegt werden. Künstliche Demodaten." />

**Erwartetes Ergebnis:** Drei CSV-Demofahrzeuge stehen im Bestand. Die ursprüngliche Lok heißt
weiterhin **Demo-Lok 01**. Es gibt keine zusätzliche Lok mit derselben Inventarnummer.
Ersetzen oder als Kopie importieren sind andere Entscheidungen mit anderen Ergebnissen.

## 5. Das JSON-Paket ausprobieren

Lege ein weiteres Importprofil **Demo JSON** mit Bereich **Fahrzeuge** und Format **JSON** an.
Wähle bei **Neuer Import** dieses Profil und `vehicle-import-demo.json`. Eine CSV-Spaltenzuordnung
ist hier nicht erforderlich. Prüfe die zwei Datensätze und bestätige ihre Übernahme.

<GuideFigure name="import-json-review" alt="Vorschau des versionierten JSON-Beispielpakets mit zwei Fahrzeugen." caption="RailKeeper-JSON wird als Paket geprüft. Die Inventarnummern unterscheiden sich vom CSV-Beispiel. Künstliche Demodaten." />

Danach sind zusätzlich `DEMO-JSON-001` und `DEMO-JSON-002` vorhanden. Wenn du alle Schritte einmal
in einem zuvor leeren Demobestand ausführst, erhältst du insgesamt **fünf Demofahrzeuge**.
Bereits vorhandene Demo-Inventarnummern führen bei Wiederholung erneut zu Konflikten.

[Importregeln und Fehlerbehebung](file-import) · [Unterschied zwischen Export und Sicherung](exports)
