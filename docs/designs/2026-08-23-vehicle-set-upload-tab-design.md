# Set-Hauptbild als Upload-Tab / Set main image as upload tab

## Ziel / Goal

Die Bildverwaltung eines Fahrzeugsets wird aus dem allgemeinen Bearbeitungsformular herausgelöst.
Der Dialog `Set bearbeiten` erhält eine kompakte Tab-Navigation mit `Allgemein` und `Upload`.
Dadurch bleibt das Formular ruhig und die visuell stärkere Hauptbild-Auswahl bekommt eine eigene,
ausreichend große Arbeitsfläche.

Move vehicle-set image management out of the general edit form. The `Edit set` dialog receives a
compact tab navigation with `General` and `Upload`. This keeps the form calm and gives the more
visual main-image workflow enough dedicated space.

## Dialogstruktur / Dialog structure

- `Allgemein / General` ist beim Öffnen aktiv.
- Der Tab enthält die bestehenden Bereiche `Grunddaten` und `Erwerb & Bestand` unverändert.
- `Upload` enthält ausschließlich die Verwaltung des Set-Hauptbilds.
- Der aktive Tab bleibt erhalten, solange der Dialog geöffnet ist.
- Schließen und erneutes Öffnen setzt den Dialog auf `Allgemein / General` zurück.
- Die Tab-Leiste steht direkt unter dem Dialogkopf und bleibt beim Scrollen sichtbar, sofern dies
  ohne zusätzliche verschachtelte Scrollfläche möglich ist.

- `General` is active when the dialog opens.
- It contains the existing `Basic data` and `Acquisition & stock` sections without behavioral
  changes.
- `Upload` contains only set-main-image management.
- The active tab is retained while the dialog remains open.
- Closing and reopening resets the dialog to `General`.
- The tab bar sits directly below the dialog header and remains visible while scrolling when this
  can be achieved without adding another nested scroll container.

## Upload-Tab / Upload tab

Der Upload-Tab verwendet die freigegebene Variante B, `Bildbühne`:

1. Kopfzeile mit Titel, kurzer Erklärung und sichtbarem Status des aktuell verwendeten Bilds.
2. Große, neutrale Bildfläche für die aktuelle Vorschau.
3. Das Bild wird proportional mit `object-fit: contain` dargestellt. Fahrzeugteile dürfen nicht
   abgeschnitten werden.
4. Unter der Vorschau stehen Bezeichnung, Quelle sowie die Aktionen `Eigenes Setbild` und
   `Automatische Auswahl`.
5. Darunter folgt ein kompakter horizontaler Streifen der verfügbaren Mitgliedsbilder. Die aktive
   Auswahl erhält Rahmen, Häkchen und Textstatus.
6. Auf schmalen Ansichten wird der Streifen zu einer einspaltigen Liste. Vorschau und Vorschaubilder
   bleiben vollständig sichtbar.

The upload tab uses the approved option B, `Image stage`:

1. A header shows the title, a short explanation, and the status of the current image.
2. A large neutral stage presents the current preview.
3. The image is scaled proportionally with `object-fit: contain`. No part of a vehicle may be
   cropped.
4. Name, source, `Dedicated set image`, and `Automatic selection` actions sit below the preview.
5. A compact horizontal strip lists available member images. The active choice uses a border,
   checkmark, and textual status.
6. On narrow screens, the strip becomes a single-column list. Preview and thumbnails remain fully
   visible.

## Verhalten und Rückmeldung / Behavior and feedback

- Bildaktionen behalten das bestehende unmittelbare Speichern über die Set-Bild-API bei.
- `Allgemein / General` zeigt weiterhin `Abbrechen` und `Speichern`; `Speichern` betrifft nur die
  Formulardaten.
- `Upload` zeigt im Dialogfuß nur `Schließen`, weil jede Bildaktion unmittelbar gespeichert wird.
- Während einer Bildaktion sind konkurrierende Bildaktionen deaktiviert.
- API-Fehler erscheinen im Dialog als verständliche Fehlermeldung und schließen den Dialog nicht.
- Nach erfolgreicher Auswahl aktualisieren sich Vorschau, Status, Galerieauswahl und Bestandsliste
  aus der zurückgegebenen Set-Antwort.
- Das Hochladen akzeptiert weiterhin JPEG, PNG und WebP.

- Image actions keep the existing immediate persistence through the set-image API.
- `General` continues to show `Cancel` and `Save`; `Save` applies to form data only.
- `Upload` shows only `Close` in the dialog footer because every image action persists immediately.
- Competing image actions are disabled while an image request is running.
- API failures appear as understandable dialog errors and do not close the dialog.
- After a successful selection, preview, status, gallery state, and inventory state update from the
  returned set response.
- Upload continues to accept JPEG, PNG, and WebP.

## Komponenten und Styling / Components and styling

- `VehicleSetEditorDialog` verwaltet den aktiven Tab und rendert nur den zugehörigen Inhalt.
- `VehicleSetMainImageEditor` bleibt die fachlich eigenständige Bildkomponente. Ihr Markup wird auf
  die Bildbühne und den Bildstreifen umgestellt.
- Bestehende RailKeeper-Schaltflächen, Abstände, Linien, Fokuszustände und Farbtokens werden
  wiederverwendet. Es werden keine neuen globalen Designsystem-Abstraktionen eingeführt.
- Tab-Schaltflächen sind per Tastatur erreichbar, besitzen einen sichtbaren Fokus und melden den
  aktiven Zustand über passende ARIA-Rollen und Attribute.

- `VehicleSetEditorDialog` owns the active tab and renders only its content.
- `VehicleSetMainImageEditor` remains the focused image-domain component. Its markup changes to the
  image stage and thumbnail strip.
- Existing RailKeeper buttons, spacing, borders, focus states, and color tokens are reused. No new
  global design-system abstraction is introduced.
- Tab controls are keyboard accessible, have visible focus, and expose active state through suitable
  ARIA roles and attributes.

## Tests und Abnahme / Tests and acceptance

- Komponententest für den Standard-Tab `Allgemein / General`.
- Komponententest für den Wechsel zu `Upload` und zurück.
- Bestehende Tests für Upload, automatische Auswahl, Mitgliedsbild-Auswahl und eigenes Bild bleiben
  erhalten und werden an die neue Tab-Navigation angepasst.
- Visuelle Prüfung in dunklem und hellem Theme sowie auf Desktop- und Mobilbreite.
- Abnahmebedingung: Das vollständige Setbild ist in Vorschau und Bildstreifen ohne Zuschnitt sichtbar.
- Frontend-Testlauf und Produktions-Build müssen erfolgreich sein.

- Component test for the default `General` tab.
- Component test for switching to `Upload` and back.
- Existing upload, automatic-selection, member-image, and dedicated-image tests remain and are
  adapted to the new tab navigation.
- Visual verification in dark and light themes at desktop and mobile widths.
- Acceptance condition: the complete set image is visible without cropping in both preview and strip.
- Frontend tests and the production build must pass.

## Nicht im Umfang / Out of scope

- Keine Änderung der Backend-API oder des Datenmodells.
- Kein Zuschneiden, Drehen oder Bearbeiten hochgeladener Bilder.
- Keine Änderung der Hauptbildlogik in Bestandszeilen oder Set-Zusammenfassungen außerhalb der
  erforderlichen Aktualisierung nach einer Auswahl.

- No backend API or data-model changes.
- No cropping, rotating, or editing uploaded images.
- No change to main-image behavior in inventory rows or set summaries beyond the existing refresh
  after a selection.
