# Erweiterte Geometrie, Paket A: Modulport-Fundament

**Datum:** 2026-08-10

**Status:** Paket A lokal umgesetzt und am 2026-08-10 abgenommen, keine Veröffentlichung

## Ziel

Paket A schafft das persistente Fundament für Modulübergänge. Planner und Admin können an einer
Anlageneinheit benannte, typisierte Ports mit millimetergenauer Position und Richtung dokumentieren.
Die Daten bleiben unabhängig von einer konkreten Aufbaukonfiguration und können dadurch von mehreren
Aufbauten wiederverwendet werden.

Das Paket berechnet noch keine Verbindungen und verschiebt keine Module automatisch. Es liefert das
stabile Modell, auf dem ein späteres Paket Kompatibilität, Snapping und Warnungen in einer
Aufbaukonfiguration ableitet.

## Fachliches Modell

Ein Modulport gehört genau zu einer Anlageneinheit und enthält:

- eine innerhalb der Einheit eindeutige Bezeichnung,
- die Art `track`, `power`, `digital`, `feedback`, `accessory` oder `other`,
- einen normalisierten Schnittstellenschlüssel für die spätere exakte Kompatibilitätsprüfung,
- Position `xMm` und `yMm` im lokalen Koordinatensystem der Einheit,
- eine normalisierte Richtung von 0 bis kleiner 360 Grad,
- optionale Hinweise,
- Version, Archivstatus und Zeitstempel.

Der Schnittstellenschlüssel ist absichtlich eine kurze dokumentierte Kennung statt einer neuen
globalen Stammdatenverwaltung. Beispiele sind `track:tillig-tt-modellgleis`, `power:16v-ac` oder
`feedback:s88-n`. Groß-/Kleinschreibung wird normalisiert. Eine spätere Kompatibilitätsprüfung darf
nur gleiche Art und gleichen Schlüssel automatisch als kompatibel behandeln.

Positionen müssen endlich und nicht negativ sein. Wenn eine Einheit eine Breite oder Höhe größer null
besitzt, darf der Port diese Grenze nicht überschreiten. Bezeichnung, Art und Schnittstellenschlüssel
sind Pflichtfelder. Die Richtung wird wie bei Planobjekten normalisiert.

## Speicherung und Konsistenz

Migration `0048_layout_unit_ports.sql` ergänzt `layout_unit_ports`. Ein Fremdschlüssel mit
`ON DELETE RESTRICT` schützt Ports vor dem Verlust ihrer Einheit. `(layout_unit_id, name)` ist ohne
Beachtung der Groß-/Kleinschreibung eindeutig. Änderungen verwenden `expectedVersion`; Konflikte
liefern den vorhandenen Layout-Versionskonflikt. Archivierung erfolgt über ein normales Update, es
gibt in diesem Paket keinen physischen Löschendpunkt.

Audit-Einträge verwenden `LayoutUnitPortCreated` und `LayoutUnitPortUpdated`. Backup-Format 7
enthält die neue Tabelle. Formate 1 bis 6 bleiben importierbar und erzeugen keine Ports.

## API und Berechtigungen

- `GET /api/v1/layout-units/{id}/ports` ist für Viewer, Editor, Planner und Admin lesbar.
- `POST /api/v1/layout-units/{id}/ports` ist nur für Planner und Admin schreibbar.
- `PUT /api/v1/layout-unit-ports/{id}` ist nur für Planner und Admin schreibbar.
- Die Messe-Rolle bleibt von allgemeinen Anlagenrouten ausgeschlossen.
- Schreibzugriffe bleiben CSRF-geschützt.

Die API verwendet `LayoutUnitPort`, `LayoutUnitPortInput` und `LayoutUnitPortUpdateInput`. Validierungs-
und Versionsfehler folgen den bestehenden Layout-Fehlerantworten.

## Oberfläche

Der vorhandene Bereich `Module` erhält unter der Einheitenverwaltung ein fokussiertes Port-Panel.
Nach Auswahl einer Einheit lädt es deren Ports. Eine kompakte Tabelle zeigt Bezeichnung, Art,
Schnittstelle, Position, Richtung und Status. Planner/Admin erhalten daneben ein Zweispaltenformular
mit app-eigenem `AppSelect` für die Art. Viewer sehen nur die Tabelle.

Das Panel unterscheidet Laden, keine Einheit ausgewählt, leer und Fehler. Archivierte Ports bleiben
sichtbar. Auf schmalen Ansichten fallen Tabelle und Formular untereinander. Deutsch und Englisch
werden gemeinsam ergänzt.

## Abgrenzung

Nicht Teil dieses Pakets sind:

- Verbindung oder Snapping von Ports in Aufbaukonfigurationen,
- automatische Modulreihenfolge,
- mehrere technische Schnittstellen in einem zusammengesetzten Port,
- Flexgleise, Höhenprofile, Steigungen oder Durchfahrtshöhen,
- digitale Steuerbefehle.

## Abnahme

- Planner/Admin können einen Port anlegen, auswählen, ändern und archivieren.
- Ungültige Koordinaten, Arten, Schlüssel, Duplikate und veraltete Versionen werden ohne Teilmutation
  abgelehnt.
- Viewer können Ports lesen, Editor darf sie nicht ändern, Messe erhält keinen Zugriff.
- Backup/Restore erhält Ports und Referenzen; ältere Backups bleiben importierbar.
- OpenAPI, Frontend-Client, deutsche und englische UI stimmen mit dem Backend überein.
- Go-Suite, Frontendtests, Produktionsbuild und lokale Browserabnahme sind grün.

### Abnahmeprotokoll vom 2026-08-10

- `go test ./...` im Backend: alle Pakete erfolgreich.
- `npm.cmd test -- --run` im Frontend: 65 Testdateien und 345 Tests erfolgreich.
- `npm.cmd run build`: 2.172 Module erfolgreich in den Produktionsbuild übernommen.
- Browserabnahme unter `http://127.0.0.1:18083/layouts`: Port anlegen, bearbeiten und archivieren
  erfolgreich; app-eigene Auswahlkomponente, lange Schnittstellenkennung und Dark Theme geprüft.
- Browserkonsole nach der Abnahme: keine Warnungen oder Fehler.
- Der QA-Port `QA West` bleibt ausschließlich lokal und archiviert in der Entwicklungsdatenbank.
