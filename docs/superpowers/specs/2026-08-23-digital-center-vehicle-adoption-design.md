# Digital Center Vehicle Adoption / Fahrzeugübernahme aus Digitalzentralen

Date: 2026-08-23
Status: Approved design / Freigegebenes Design

## Deutsch

### Ziel

Eine aus der ECoS gelesene, noch nicht zugeordnete Lok kann aus dem Arbeitsbereich
„Digitalzentralen“ entweder als neues RailKeeper-Fahrzeug vorbereitet oder einem bestehenden
Fahrzeug zugeordnet werden. Der Ablauf bleibt ein sicherer Lese- und Zuordnungsprozess. Er schreibt
keine Daten zur Digitalzentrale.

### Bedienablauf

Bei einem Arbeitslisteneintrag ohne `vehicleId` zeigt der Vergleichsdialog zwei primäre Aktionen:

1. **Neues Fahrzeug anlegen** öffnet den vorhandenen vollständigen Fahrzeugeditor. Name,
   Decoderadresse, Protokoll und ECoS-Objekt-ID werden als überprüfbarer ECoS-Entwurf übernommen.
   RailKeeper-Pflichtfelder, die ECoS nicht verlässlich liefert, bleiben sichtbar zu ergänzen.
2. **Bestehendem Fahrzeug zuordnen** öffnet eine durchsuchbare Fahrzeugauswahl. Plausible Treffer
   anhand Decoderadresse und Name werden zuerst angezeigt, aber nie automatisch gespeichert. Der
   Benutzer bestätigt das ausgewählte Fahrzeug ausdrücklich.

Nach erfolgreichem Speichern oder Zuordnen kehrt die Anwendung zur Digitalzentralen-Seite zurück,
liest die aktuelle Arbeitsliste erneut und zeigt die neue RailKeeper-Zuordnung. Abbruch oder
Zurücknavigation verändert keine Daten.

### Technische Gestaltung

- Der vorhandene `ECoSVehicleDraftPayload` und `useVehicleECoSDraftController` bleiben die einzige
  Logik für einen vorbefüllten Fahrzeugeditor.
- Der Digitalzentralen-Arbeitsbereich erzeugt aus dem `DigitalCenterWorkItem` einen begrenzten
  ECoS-Entwurf. Nicht vorhandene CV- oder Funktionsdaten werden nicht erfunden.
- Die Rückkehrinformation verweist auf `/digital-centers` und die betroffene Sitzung bzw.
  Objekt-ID. Die bisherige Rückkehr zum alten Import/Export-Ablauf bleibt kompatibel.
- Beim Speichern des Fahrzeugeditors wird zuerst das Fahrzeug gespeichert und danach über den
  vorhandenen Endpunkt `/vehicles/{id}/external-mappings` die Zuordnung aus Provider und externer
  Objekt-ID angelegt. Danach wird das Fahrzeug erneut geladen.
- Für die Zuordnung zu einem bestehenden Fahrzeug wird derselbe Mapping-Endpunkt verwendet. Ein
  Konflikt wird sichtbar gemeldet und die Arbeitsliste nicht optimistisch verfälscht.
- Die bestehende Eindeutigkeit `(provider, external_id)` bleibt maßgeblich. Eine vorhandene
  Zuordnung darf nicht still auf ein anderes Fahrzeug verschoben werden. Der Service erhält dafür
  eine explizite Konfliktprüfung, falls die aktuelle Upsert-Semantik dies noch nicht gewährleistet.
- Fahrzeuganlage und Zuordnung stehen nur Rollen mit Bearbeitungsrecht zur Verfügung. Viewer sehen
  die Zuordnung und einen verständlichen Hinweis, aber keine schreibenden Aktionen.

### Fehler- und Randfälle

- Fehlende Pflichtfelder blockieren nur das Speichern im Fahrzeugeditor, nicht das Öffnen des
  Entwurfs.
- Ist die ECoS-Objekt-ID ungültig oder fehlt sie, wird keine Aktion angeboten.
- Wurde das Objekt zwischenzeitlich bereits zugeordnet, fordert RailKeeper ein erneutes Lesen an.
- Netzwerk- und API-Fehler bleiben im jeweiligen Dialog sichtbar. Die Auswahl und der Entwurf
  bleiben erhalten, damit der Benutzer erneut versuchen oder abbrechen kann.
- Die Zuordnung schreibt keine Stammdaten zur ECoS und löst keinen Schreib-Grant aus.

### Oberfläche und Barrierefreiheit

- Alle neuen Texte werden auf Deutsch und Englisch gepflegt.
- Dialoge verwenden die vorhandene Fokusführung, Escape-Behandlung und semantische
  Dialogauszeichnung.
- Treffer werden nicht nur farblich gekennzeichnet. Decoderadresse, Name und Inventarnummer machen
  die Auswahl nachvollziehbar.
- Lange Namen sowie Desktop- und schmale Ansichten werden berücksichtigt.

### Verifikation

- Komponenten- und Hook-Tests für sichtbare Aktionen, Rollen, Vorbefüllung, Abbruch, Rückkehr und
  Refresh.
- Tests für Fahrzeugauswahl, Suche, vorgeschlagene Treffer, Bestätigung und API-Fehler.
- Backend-Tests für eine neue Zuordnung, idempotentes erneutes Speichern und Konflikte mit einer
  bereits einem anderen Fahrzeug zugeordneten ECoS-ID.
- Vollständige Go-Tests, vollständige Frontend-Tests und Produktionsbuild.
- Visueller Test im lokalen Browser mit einer real gelesenen ECoS-Arbeitsliste.

## English

### Goal

An ECoS locomotive read into the Digital Centers workspace without a RailKeeper assignment can be
prepared as a new RailKeeper vehicle or assigned to an existing vehicle. This remains a safe read
and assignment workflow and never writes data to the command station.

### User flow

For a work-list item without a `vehicleId`, the comparison dialog exposes two primary actions:

1. **Create new vehicle** opens the existing full vehicle editor. Name, decoder address, protocol,
   and ECoS object ID are transferred as a reviewable ECoS draft. Required RailKeeper fields that
   cannot be obtained reliably from ECoS remain visible and must be completed.
2. **Assign existing vehicle** opens a searchable vehicle picker. Plausible matches based on decoder
   address and name appear first, but are never saved automatically. The user explicitly confirms
   the selected vehicle.

After a successful save or assignment, the app returns to Digital Centers, reloads the current work
list, and shows the new RailKeeper assignment. Cancelling or navigating back does not change data.

### Technical design

- The existing `ECoSVehicleDraftPayload` and `useVehicleECoSDraftController` remain the single path
  for opening a prefilled vehicle editor.
- Digital Centers creates a bounded ECoS draft from `DigitalCenterWorkItem`. Missing CV or function
  data is never fabricated.
- Return metadata points to `/digital-centers` and the affected session or object ID. The existing
  return path to the legacy import/export flow remains compatible.
- Saving the vehicle first persists the vehicle, then creates the provider/object mapping through
  `/vehicles/{id}/external-mappings`, and finally reloads the vehicle.
- Assigning an existing vehicle uses the same mapping endpoint. Conflicts are shown explicitly and
  the work list is not updated optimistically.
- The existing `(provider, external_id)` uniqueness remains authoritative. An existing mapping must
  not silently move to another vehicle. The service gains an explicit conflict check if the current
  upsert semantics do not yet enforce this product rule.
- Create and assignment actions require edit permission. Viewers can inspect the state and see a
  clear explanation, but do not receive write actions.

### Errors and edge cases

- Missing required fields block saving in the vehicle editor, not opening the draft.
- No action is offered when the ECoS object ID is missing or invalid.
- If another action assigned the object in the meantime, RailKeeper requests a fresh read.
- Network and API errors stay visible in the active dialog. The draft or selection remains intact
  for retry or cancellation.
- Assignment never writes master data to ECoS and does not request a device write grant.

### UI and accessibility

- All new copy is maintained in German and English.
- Dialogs use the existing focus management, Escape handling, and semantic dialog markup.
- Suggested matches are not identified by color alone. Decoder address, name, and inventory number
  make the choice auditable.
- Long labels, desktop layouts, and narrow layouts are covered.

### Verification

- Component and hook tests cover action visibility, roles, prefill, cancellation, return, and
  refresh.
- Tests cover vehicle search, suggested matches, confirmation, and API errors.
- Backend tests cover a new assignment, idempotent reassignment to the same vehicle, and conflicts
  when an ECoS ID already belongs to another vehicle.
- Run all Go tests, all frontend tests, and the production build.
- Visually verify the workflow in the local browser using a real ECoS work list.
