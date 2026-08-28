# Massen-Deaktivierung für Stammdaten

Datum: 28. August 2026  
Issue: #142

## Ziel

Editoren und Administratoren können viele aktive Stammdateneinträge eines Typs gezielt auswählen
und in einem einzigen, bestätigten Vorgang deaktivieren. Bestehende Datensätze behalten ihre
gespeicherten Werte. Lagerorte, dauerhafte Löschung und Massen-Reaktivierung gehören nicht zu
diesem Schnitt.

## Ausgangslage

Allgemeine und artikelbezogene Stammdaten bieten bislang nur Einzelaktionen. Für mehrere hundert
Einträge erzeugt das viele Klicks. Eine reine Frontend-Schleife über den bestehenden Einzelendpunkt
wäre langsam und könnte nach einem Teilfehler einen schwer nachvollziehbaren Zwischenstand
hinterlassen.

## Entscheidung

RailKeeper ergänzt eine explizite Auswahlspalte und einen atomaren Batch-Endpunkt.

- Jede betroffene Tabelle zeigt eine Checkbox für aktive Einträge, deren
  `capabilities.canDeactivate` gesetzt ist.
- Die Checkbox im Tabellenkopf markiert oder demarkiert alle sichtbaren, deaktivierbaren Einträge.
- Eine kompakte Aktionsleiste zeigt die Anzahl der ausgewählten Einträge und bietet
  **Ausgewählte deaktivieren** an.
- Vor dem Schreiben nennt ein Bestätigungsdialog Typ und Anzahl. Er weist darauf hin, dass
  bestehende Verwendungen unverändert bleiben.
- Nach Erfolg ersetzt die UI die betroffenen Zeilen durch die Serverantwort und leert die Auswahl.
- Ein Typwechsel, eine geänderte Suche oder ein explizites Neuladen leert die Auswahl ebenfalls.
- In der Artikelverwaltung besitzt jede Stammdatentabelle ihre eigene Auswahl. Der Abschnitt
  **Artikelarten und Unterarten** vermischt die beiden Typen daher nicht.

Reaktivierung und Löschung bleiben Einzelaktionen. Damit bleibt die Änderung auf das gemeldete
Problem begrenzt und eine versehentliche massenhafte Reaktivierung oder Löschung wird nicht
eingeführt.

## API und Datenfluss

Der neue Endpunkt lautet:

```http
PATCH /api/v1/master-data/{type}/active
Content-Type: application/json

{
  "keys": ["märklin", "roco"],
  "active": false
}
```

Die Antwort ist eine Liste der aktualisierten `MasterDataEntry`-Objekte einschließlich ihrer
Management-Fähigkeiten. Der Endpunkt bleibt Editor/Admin vorbehalten und CSRF-geschützt.

Der Anwendungsdienst:

1. normalisiert Typ und Schlüssel,
2. fordert zwischen 1 und 5.000 eindeutige Schlüssel,
3. öffnet eine SQLite-Transaktion,
4. aktualisiert jeden Eintrag innerhalb derselben Transaktion,
5. bricht bei einem unbekannten Schlüssel vollständig ab,
6. bestätigt die Transaktion, invalidiert den Cache und liest die aktualisierten
   Management-Einträge zurück.

Das Limit deckt große lokale Stammdatenbestände ab und begrenzt zugleich Requestgröße und
Transaktionsdauer. Doppelte Schlüssel werden vor dem Limit und vor der Transaktion entfernt.

## Fehlerbehandlung

- Fehlender Zustand, leerer Typ, leere Schlüssel oder mehr als 5.000 eindeutige Schlüssel ergeben
  `400 master_data_validation`.
- Sobald ein Schlüssel nicht existiert, ergibt der gesamte Vorgang
  `404 master_data_not_found`; kein Eintrag wird geändert.
- Datenbankfehler ergeben `500 master_data_update_failed`; die Transaktion wird zurückgerollt.
- Die UI behält die Auswahl nach einem Fehler bei, zeigt die vorhandene Fehlermeldungsfläche und
  ermöglicht einen erneuten Versuch.
- Während des Requests sind Auswahl und Batch-Aktion deaktiviert. Einzelaktionen derselben Tabelle
  sind ebenfalls gesperrt.

## Bedienung und Barrierefreiheit

- Zeilen- und Kopfcheckboxen besitzen vollständige zugängliche Namen.
- Der Tabellenkopf zeigt bei einer Teilmenge den nativen `indeterminate`-Zustand.
- Nicht deaktivierbare oder bereits inaktive Einträge erhalten keine auswählbare Checkbox.
- Der Auswahlstatus wird nicht allein über Farbe vermittelt.
- Die Aktionsleiste bleibt kompakt und verwendet vorhandene Buttons, Abstände und Farbtokens.
- Mobile Tabellen bleiben in ihrem vorhandenen begrenzten horizontalen Scrollbereich.

## Abgelehnte Varianten

### Viele Einzelrequests aus dem Frontend

Diese Variante benötigt keinen neuen Vertrag, verursacht aber bis zu hunderte Requests und kann
bei einem Fehler nach einzelnen erfolgreichen Updates stoppen. Sie erfüllt die geforderte
verlässliche Massenfunktion nicht.

### Alle gefilterten Einträge ohne Auswahl deaktivieren

Diese Variante spart Checkboxen, koppelt eine weitreichende Änderung aber an den aktuellen
Suchfilter. Ein übersehener oder leerer Filter könnte wesentlich mehr Einträge als beabsichtigt
deaktivieren.

## Verifikation

- Anwendungstests prüfen Validierung, Deduplizierung, vollständigen Rollback bei unbekanntem
  Schlüssel, Cache-Invalidierung und die zurückgegebenen Fähigkeiten.
- API-Tests prüfen Editorzugriff, Requestfehler, atomaren Erfolg und den OpenAPI-Vertrag.
- Frontendtests prüfen Kopf- und Zeilenauswahl, Teilzustand, Bestätigungstext, genau einen
  Batch-Request, erfolgreiche Zeilenaktualisierung, Auswahlrücksetzung und Fehlerversuch.
- Der vollständige Go-Testlauf, Frontendtests, Frontend-Build und Dokumentationsprüfung laufen vor
  der PR.

## Nicht-Ziele

- keine Massenlöschung,
- keine Massen-Reaktivierung,
- keine typübergreifende Auswahl,
- keine Änderungen an Lagerorten,
- kein Hintergrundjob oder Fortschrittsprotokoll für den lokalen, begrenzten Batch.
