# Implementierungsplan: Anpassbare Spaltenbreiten

Issue: #141

Design: `docs/designs/2026-08-28-table-column-widths-design.md`

## Aufgabe 1: Gemeinsames Layoutmodell und Profilpersistenz

Dateien:

- Neu: `frontend/src/shared/tableColumnLayout.ts`
- Neu: `frontend/src/shared/tableColumnLayout.test.ts`
- Neu: `frontend/src/shared/useProfileTableLayout.ts`
- Neu: `frontend/src/shared/useProfileTableLayout.test.tsx`

Schritte:

1. Zuerst fehlschlagende Modelltests für versionierte Werte, Legacy-Arrays, unbekannte und doppelte
   Spalten, nicht endliche Breiten, Mindest-/Maximalgrenzen und ausgeblendete Breiten schreiben.
2. Ein generisches Layout mit `columns` und `widths` sowie typisierten Breitendefinitionen
   implementieren.
3. Zuerst fehlschlagende Hook-Tests für Laden, geordnete partielle Profilupdates,
   Fehlerweitergabe und Legacy-Fallback schreiben.
4. Den gemeinsamen Profil-Hook mit geordneter Save-Queue implementieren. Lokale Vorschau und
   persistierender Commit müssen getrennt möglich sein.
5. Gezielte Tests ausführen und committen.

## Aufgabe 2: Zugänglicher gemeinsamer Resize-Handle

Dateien:

- Neu: `frontend/src/shared/ui/TableColumnResizeHandle.tsx`
- Neu: `frontend/src/shared/ui/TableColumnResizeHandle.test.tsx`
- Ändern: `frontend/src/styles/base.css`

Schritte:

1. Fehlschlagende Tests für Separator-Semantik, Pointer-Ziehen, Pfeiltasten, Umschalt-Schritte,
   Pos1/Ende und Doppelklick schreiben.
2. Pointer-Capture mit unmittelbarer Vorschau und genau einem Commit am Ziehende implementieren.
3. Tastaturänderungen unmittelbar committen und Events gegen Sortieraktionen abschirmen.
4. Den Handle im Tabellenkopf positionieren, Fokus und Hover sichtbar machen und ihn für kompakte
   Mobilansichten ausblenden.
5. Gezielte Tests ausführen und committen.

## Aufgabe 3: Fahrzeuglayout erweitern

Dateien:

- Ändern: `frontend/src/features/vehicles/vehicleTableColumns.ts`
- Ändern: `frontend/src/features/vehicles/vehicleTableColumns.test.ts`
- Ändern: `frontend/src/features/vehicles/useVehicleColumnPreferences.ts`
- Ändern: `frontend/src/features/vehicles/useVehicleColumnPreferences.test.tsx`
- Ändern: `frontend/src/features/vehicles/VehicleInventoryTable.tsx`
- Ändern: `frontend/src/features/vehicles/VehicleInventoryTable.test.tsx`
- Ändern: `frontend/src/features/vehicles/VehicleSetInventoryRow.tsx`
- Ändern: `frontend/src/features/vehicles/VehicleInventoryPanel.tsx`
- Ändern: `frontend/src/features/vehicles/VehiclesView.tsx`
- Ändern: `frontend/src/styles/vehicle-inventory.css`
- Ändern: `frontend/src/shared/i18n/de.ts`
- Ändern: `frontend/src/shared/i18n/en.ts`

Schritte:

1. Fehlschlagende Tests für Fahrzeugstandardbreiten, Legacy-Profilwerte, Begrenzung und Reset
   schreiben.
2. Breitendefinitionen und das versionierte Fahrzeuglayout ergänzen.
3. Den vorhandenen Hook auf die gemeinsame Profilgrundlage umstellen. Seine bestehenden
   Auswahl- und Reihenfolgeaktionen bleiben erhalten und Reset löscht auch Sonderbreiten.
4. Breiten und Resize-Callbacks durch View und Panel bis zur Tabelle führen.
5. In der Tabelle ein `colgroup`, eine berechnete Mindestbreite und Handles nur für Datenspalten
   rendern. Set-Zeilen verwenden durch dieselben Spalten automatisch dieselben Breiten.
6. CSS-Regeln von gleichmäßiger Zwangsverteilung auf konfigurierte Breiten und lokalen horizontalen
   Scroll umstellen.
7. Gezielte Fahrzeugtests und Produktions-Build ausführen und committen.

## Aufgabe 4: Zubehörlayout profilbasiert und geordnet machen

Dateien:

- Ändern: `frontend/src/features/accessories/articleTableColumns.ts`
- Ändern: `frontend/src/features/accessories/articleTableColumns.test.ts`
- Neu: `frontend/src/features/accessories/useArticleColumnPreferences.ts`
- Neu: `frontend/src/features/accessories/useArticleColumnPreferences.test.tsx`
- Ändern: `frontend/src/features/accessories/AccessoriesView.tsx`
- Ändern: `frontend/src/features/accessories/AccessoriesView.test.tsx`
- Ändern: `frontend/src/features/accessories/ArticleColumnPicker.tsx`
- Ändern: `frontend/src/features/accessories/ArticleColumnPicker.test.tsx`
- Ändern: `frontend/src/features/accessories/ArticleTable.tsx`
- Ändern: `frontend/src/features/accessories/ArticleTable.test.tsx`
- Ändern: `frontend/src/styles/accessories.css`
- Ändern: `frontend/src/shared/i18n/de.ts`
- Ändern: `frontend/src/shared/i18n/en.ts`

Schritte:

1. Fehlschlagende Tests für geordnete Zubehörspalten, Legacy-`localStorage`, Profilvorrang,
   Breitenvalidierung und Reset schreiben.
2. Das Zubehörmodell von einer ungeordneten Menge auf eine geordnete Liste mit Breitendefinitionen
   umstellen.
3. Den Profil-Hook implementieren und den lokalen Altwahrscheinlichkeitswert bei fehlendem
   Profilwert einmalig migrieren.
4. Den Picker um sichtbare Reihenfolge mit Hoch-/Runter-Aktionen ergänzen.
5. Die Tabelle anhand der geordneten Liste rendern, `colgroup`, Mindestbreite und Resize-Handles
   ergänzen. Auswahl und Aktionen bleiben fest.
6. Gezielte Zubehörtests und Produktions-Build ausführen und committen.

## Aufgabe 5: Gesamtprüfung, visuelle QA und Integration

Schritte:

1. `git diff --check` ausführen.
2. Backend vollständig mit `go test ./...` prüfen, da die Profil-API unverändert wiederverwendet wird.
3. Frontend vollständig mit `npm.cmd test -- --run` und `npm.cmd run build` prüfen.
4. Dokumentation mit `npm.cmd run check` prüfen.
5. Eine isolierte lokale Instanz starten und beide Tabellen in Hell und Dunkel bei 2580 px, 1440 px
   und 820 px prüfen. Breite per Pointer und Tastatur ändern, neu laden, denselben Benutzer in einer
   neuen Browser-Sitzung verwenden, Spalte ausblenden/einblenden und Reset prüfen.
6. Globalen Überlauf, lokalen Tabellenscroll, feste Auswahl-/Aktionsspalten, lange deutsche Texte,
   Fokusreihenfolge sowie Lade- und Fehlerzustände prüfen.
7. Branch pushen, zweisprachigen PR mit `Closes #141` erstellen, alle GitHub-Checks und konkrete
   Kommentare prüfen und bei vollständig grünem Zustand regulär mergen.
