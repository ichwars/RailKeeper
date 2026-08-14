# Interaktiver Anlagenzwilling, Paket B: sichtbare Leseansicht

## Ziel

Paket B macht die in Paket A gespeicherten technischen Positionen in der Anlagenübersicht sichtbar.
Die Ansicht bleibt bewusst lesend. Bearbeitungsmodus, freie Konturen und Drag-and-drop folgen in
Paket C.

## Umfang

1. Reservierungen und Einbauten können optional auf eine technische Position verweisen.
   Die Zuordnung muss zum Zielmodul und Artikel passen. Ein Einbau aus einer Reservierung übernimmt
   die Positionszuordnung.
2. Ein neuer lesender Twin-Endpunkt liefert Konfiguration oder Einzelmodul, transformierte Konturen
   und Positionen, Artikelinformationen sowie aktive Reservierungen und Einbauten.
3. Der Status einer Position wird aus den aktiven Zuordnungen abgeleitet: geplant, reserviert,
   eingebaut, wartungsfällig und defekt. Mehrere Zustände dürfen gleichzeitig sichtbar sein.
4. Die Anlagenübersicht erhält eine SVG-Darstellung mit Konfigurations- oder Modulauswahl,
   Statusfiltern, Desktop-Hover und einem klick- und touchfähigen Inspektor.
5. OpenAPI, TypeScript-Client, deutsche und englische Texte sowie responsive Darstellung werden
   gemeinsam aktualisiert.

## Umsetzungsschritte

### 1. Positionsbezug in Zubehörzuordnungen

- `technicalPositionId` zu Reservierungs- und Einbau-DTOs ergänzen.
- Referenz beim Schreiben serverseitig gegen Modul, Artikel und Archivstatus prüfen.
- Verknüpfungstabellen aus Migration 0044 verwenden.
- Übernahme von der Reservierung in den Einbau absichern.
- Repository- und Service-Tests ergänzen.

### 2. Transformierte Twin-Leseansicht

- Anwendungsmodell für Grenzen, Module, Konturen, Positionen und Statusinformationen anlegen.
- Repository-Abfrage für eine Konfiguration oder ein einzelnes Modul implementieren.
- Lokale Koordinaten mit Verschiebung und Drehung in Anlagenkoordinaten transformieren.
- Fehlende Konturen durch Modulmaße ersetzen und Datenlücken als Warnungen liefern.
- Statusaggregation und Transformationslogik gezielt testen.

### 3. API-Vertrag

- `GET /api/v1/layouts/{id}/twin` mit `configurationId` oder `unitId` ergänzen.
- Berechtigung auf lesende Anlagenrollen begrenzen.
- OpenAPI-Schemas und TypeScript-Client synchron halten.
- Route, Rollen und Client-Pfad testen.

### 4. SVG-Ansicht und Inspektor

- Fokuskomponente für Anlagenzwilling statt Wachstum des Workspace-Hauptmoduls erstellen.
- Konfiguration oder Einzelmodul über vorhandene App-Komponenten auswählen.
- Konturen und Statusmarker barrierearm im SVG darstellen.
- Hover-Zusammenfassung für Maus und Fokus, Inspektor für Klick, Tastatur und Touch ergänzen.
- Statusfilter mit Text und Symbolik umsetzen, nicht ausschließlich über Farbe.
- Desktop, schmale Ansichten, Hell- und Dunkelmodus abdecken.

### 5. Verifikation

- Backend-Tests und Frontend-Tests vollständig ausführen.
- Frontend bauen.
- Den Ablauf im lokalen Browser mit echten lokalen Daten prüfen.
- Nur lokale Commits erstellen, kein Push, keine PR und kein Merge.

