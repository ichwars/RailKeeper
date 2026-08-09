# Anlagen-Dialog und erweitertes Anlagenprofil

## Ziel

Die Anlagenansicht erhält eine klare, werkzeugartige Bedienung. Anlagen werden nicht mehr in einem
dauerhaft sichtbaren Seitenformular angelegt oder bearbeitet. Ein gemeinsamer RailKeeper-Dialog
übernimmt beide Abläufe. Verzeichnis und Profil nutzen den frei werdenden Platz, um die bereits
vorhandenen Anlagendaten vollständig und ruhig darzustellen.

Die Änderung bleibt auf die Frontend-Darstellung und -Interaktion begrenzt. Datenbank, Backend,
OpenAPI-Vertrag und vorhandenes Anlagenmodell bleiben unverändert.

## Bedienkonzept

### Anlagenverzeichnis

Das Anlagenverzeichnis nutzt die verfügbare Inhaltsbreite. Im Panelkopf steht rechts die primäre
Aktion `Anlage anlegen`. Sie ist nur für Benutzer mit Anlagenplanungsrecht sichtbar.

Jeder Verzeichniseintrag zeigt:

- Bezeichnung,
- Anlagenart,
- Spurweite und Maßstab,
- aktiven oder archivierten Status,
- Version,
- Zeitpunkt der letzten Änderung.

Die ausgewählte Anlage bleibt eindeutig markiert. Das Anklicken eines Eintrags öffnet weiterhin
keine neue Route, sondern aktualisiert den darunterliegenden Anlagenarbeitsbereich.

### Anlagenprofil

Das Profil belegt die volle Panelbreite. Im Panelkopf steht rechts die Aktion `Bearbeiten`, sofern
der Benutzer Anlagen planen darf. Das dauerhaft sichtbare Bearbeitungsformular entfällt.

Das Profil zeigt ausschließlich bereits vorhandene oder aus vorhandenen Daten berechenbare Werte:

- Anlagenart,
- Status,
- Spurweite,
- Maßstab,
- Version,
- Anzahl der Anlageneinheiten beziehungsweise Module,
- Anzahl der Aufbaukonfigurationen,
- Erstellungsdatum,
- Zeitpunkt der letzten Änderung,
- Beschreibung oder einen ruhigen Leerzustand.

Es werden in dieser Stufe keine neuen Stammdaten wie Standort, Eigentümer oder Inbetriebnahmedatum
eingeführt.

## Gemeinsamer Anlagendialog

Ein fokussierter `LayoutFormDialog` bedient die Modi `create` und `edit`. Beide Modi verwenden
dieselbe Feldanordnung, Validierung und Aktionsleiste. Titel, Speichern-Beschriftung, Anfangswerte
und API-Aktion unterscheiden sich nach Modus.

Der Dialog enthält:

- Bezeichnung als RailKeeper-Texteingabe,
- Art als `AppSelect`,
- Spurweite und Maßstab als RailKeeper-Texteingaben in einem zweispaltigen Raster,
- Beschreibung als app-eigener Textbereich,
- Archivstatus als app-eigene Checkbox nur im Bearbeitungsmodus,
- `Abbrechen` und die jeweilige primäre Speicheraktion in der Fußzeile.

Für Textbereich und Checkbox werden kleine, wiederverwendbare Shared-UI-Komponenten ergänzt. Sie
verwenden zugängliche HTML-Grundelemente, kapseln aber Beschriftung, Zustände und RailKeeper-Styling,
damit keine plattformspezifische Browserdarstellung in der Oberfläche sichtbar bleibt.

## Dialogverhalten und Barrierefreiheit

Der Dialog folgt dem vorhandenen RailKeeper-Muster aus Artikel- und Fahrzeugdialogen:

- eigener abgedunkelter Layer und app-eigene Dialoghülle,
- `role="dialog"` und `aria-modal="true"`,
- initialer Fokus auf der Bezeichnung,
- Fokus bleibt während der Bedienung im Dialog,
- Schließen über X, `Abbrechen` und Escape,
- Rückgabe des Fokus an den auslösenden Button,
- Speichern über Formular-Submit,
- gesperrte Aktionen während einer laufenden Speicherung.

Wurden Werte verändert, fordert ein app-eigener Bestätigungsdialog vor dem Verwerfen zur
Bestätigung auf. Ein unveränderter Dialog schließt unmittelbar.

## Datenfluss und Fehler

`LayoutsView` verwaltet Öffnen und Schließen des Anlegedialogs. Nach erfolgreicher Anlage wird die
Liste neu geladen und die neue Anlage ausgewählt.

`LayoutWorkspace` öffnet denselben Dialog mit einer Kopie der ausgewählten Anlage. Nach erfolgreicher
Aktualisierung ersetzt es den lokalen Datensatz über den vorhandenen Callback. Der Dialog verändert
die dargestellte Anlage erst nach einer erfolgreichen API-Antwort.

Validierungs- und Serverfehler erscheinen direkt im Dialog und schließen ihn nicht. Ein
Versionskonflikt beim Bearbeiten behält die Eingaben bei und bietet die bestehende Aktion zum Laden
des Serverstands an. Nach dem Neuladen werden Profil und Dialogentwurf mit dem aktuellen Stand
synchronisiert.

## Komponenten und Abgrenzung

Die Umsetzung umfasst:

- einen gemeinsamen `LayoutFormDialog` im Anlagenfeature,
- app-eigene Shared-UI-Komponenten für Textbereich und Checkbox,
- die Umstellung von `LayoutsView` und `LayoutWorkspace`,
- ergänzte deutsche und englische Übersetzungen,
- fokussierte Styles für Verzeichnis, Profil und Dialog,
- Komponenten- und Interaktionstests.

Nicht Bestandteil dieser Stufe sind neue Backendfelder, neue Anlagenregister, Änderungen am Planer,
Änderungen an Modulen oder Aufbaukonfigurationen sowie eine eigene Detailroute pro Anlage.

## Responsive Verhalten

Auf Desktop nutzt das Verzeichnis die volle Breite. Profilwerte erscheinen in einem dichten,
mehrspaltigen Raster. Der Dialog bleibt kompakt und zentriert.

Bei schmalen Ansichten brechen Verzeichnismetadaten und Profilwerte kontrolliert um. Das Feldraster
im Dialog wird einspaltig, die Aktionen bleiben vollständig sichtbar und der Dialoginhalt kann
innerhalb des verfügbaren Viewports scrollen.

## Tests und Abnahme

Automatisierte Frontendtests decken mindestens ab:

- Öffnen und Schließen des Anlegedialogs,
- erfolgreiche Anlage und Auswahl des neuen Datensatzes,
- Öffnen mit korrekten Werten im Bearbeitungsmodus,
- erfolgreiche Aktualisierung des Profils,
- Anzeige eines API-Fehlers ohne Datenverlust,
- Verhalten bei Versionskonflikten,
- Warnung vor dem Verwerfen geänderter Werte,
- Escape, Fokusführung und Fokusrückgabe,
- rollenabhängige Sichtbarkeit von Anlage- und Bearbeitungsaktionen,
- deutsche und englische Beschriftungen.

Zusätzlich werden Frontend-Build und die Anlagenansicht im lokalen Server geprüft, einschließlich
Desktop, schmaler Ansicht, langem deutschen Text sowie heller und dunkler Darstellung.
